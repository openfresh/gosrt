// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrt

import (
	"fmt"
	"net"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/openfresh/gosrt/internal/poll"
	"github.com/openfresh/gosrt/internal/socktest"
	"github.com/openfresh/gosrt/internal/testenv"
)

var dialTimeoutTests = []struct {
	timeout time.Duration
	delta   time.Duration // for deadline

	guard time.Duration
	max   time.Duration
}{
	// Tests that dial timeouts, deadlines in the past work.
	{-5 * time.Second, 0, -5 * time.Second, 100 * time.Millisecond},
	{0, -5 * time.Second, -5 * time.Second, 100 * time.Millisecond},
	{-5 * time.Second, 5 * time.Second, -5 * time.Second, 100 * time.Millisecond}, // timeout over deadline
	{-1 << 63, 0, time.Second, 100 * time.Millisecond},
	{0, -1 << 63, time.Second, 100 * time.Millisecond},

	{50 * time.Millisecond, 0, 100 * time.Millisecond, time.Second},
	{0, 50 * time.Millisecond, 100 * time.Millisecond, time.Second},
	{50 * time.Millisecond, 5 * time.Second, 100 * time.Millisecond, time.Second}, // timeout over deadline
}

func TestDialTimeout(t *testing.T) {
	// Cannot use t.Parallel - modifies global hooks.
	origTestHookDialChannel := testHookDialChannel
	defer func() { testHookDialChannel = origTestHookDialChannel }()
	defer sw.Set(socktest.FilterConnect, nil)

	for i, tt := range dialTimeoutTests {
		switch runtime.GOOS {
		case "plan9", "windows":
			testHookDialChannel = func() { time.Sleep(tt.guard) }
			if runtime.GOOS == "plan9" {
				break
			}
			fallthrough
		default:
			sw.Set(socktest.FilterConnect, func(so *socktest.Status) (socktest.AfterFilter, error) {
				time.Sleep(tt.guard)
				return nil, errTimedout
			})
		}

		ch := make(chan error)
		d := Dialer{Timeout: tt.timeout}
		if tt.delta != 0 {
			d.Deadline = time.Now().Add(tt.delta)
		}
		max := time.NewTimer(tt.max)
		defer max.Stop()
		go func() {
			// This dial never starts to send any SRT SYN
			// segment because of above socket filter and
			// test hook.
			c, err := d.Dial("srt", "127.0.0.1:0")
			if err == nil {
				err = fmt.Errorf("unexpectedly established: srt:%s->%s", c.LocalAddr(), c.RemoteAddr())
				c.Close()
			}
			ch <- err
		}()

		select {
		case <-max.C:
			t.Fatalf("#%d: Dial didn't return in an expected time", i)
		case err := <-ch:
			if perr := parseDialError(err); perr != nil {
				t.Errorf("#%d: %v", i, perr)
			}
			if nerr, ok := err.(net.Error); !ok || !nerr.Timeout() {
				t.Fatalf("#%d: %v", i, err)
			}
		}
	}
}

var dialTimeoutMaxDurationTests = []struct {
	timeout time.Duration
	delta   time.Duration // for deadline
}{
	// Large timeouts that will overflow an int64 unix nanos.
	{1<<63 - 1, 0},
	{0, 1<<63 - 1},
}

func TestDialTimeoutMaxDuration(t *testing.T) {
	if runtime.GOOS == "openbsd" {
		testenv.SkipFlaky(t, 15157)
	}

	ln, err := newLocalListener("srt")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	for i, tt := range dialTimeoutMaxDurationTests {
		ch := make(chan error)
		max := time.NewTimer(250 * time.Millisecond)
		defer max.Stop()
		go func() {
			d := Dialer{Timeout: tt.timeout}
			if tt.delta != 0 {
				d.Deadline = time.Now().Add(tt.delta)
			}
			c, err := d.Dial(ln.Addr().Network(), ln.Addr().String())
			if err == nil {
				c.Close()
			}
			ch <- err
		}()

		select {
		case <-max.C:
			t.Fatalf("#%d: Dial didn't return in an expected time", i)
		case err := <-ch:
			if perr := parseDialError(err); perr != nil {
				t.Error(perr)
			}
			if err != nil {
				t.Errorf("#%d: %v", i, err)
			}
		}
	}
}

var acceptTimeoutTests = []struct {
	timeout time.Duration
	xerrs   [2]error // expected errors in transition
}{
	// Tests that accept deadlines in the past work, even if
	// there's incoming connections available.
	{-5 * time.Second, [2]error{poll.ErrTimeout, poll.ErrTimeout}},

	{50 * time.Millisecond, [2]error{nil, poll.ErrTimeout}},
}

func TestAcceptTimeout(t *testing.T) {
	testenv.SkipFlaky(t, 17948)
	t.Parallel()

	switch runtime.GOOS {
	case "plan9":
		t.Skipf("not supported on %s", runtime.GOOS)
	}

	ln, err := newLocalListener("srt")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	var wg sync.WaitGroup
	for i, tt := range acceptTimeoutTests {
		if tt.timeout < 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				d := Dialer{Timeout: 100 * time.Millisecond}
				c, err := d.Dial(ln.Addr().Network(), ln.Addr().String())
				if err != nil {
					t.Error(err)
					return
				}
				c.Close()
			}()
		}

		if err := ln.(*SRTListener).SetDeadline(time.Now().Add(tt.timeout)); err != nil {
			t.Fatalf("$%d: %v", i, err)
		}
		for j, xerr := range tt.xerrs {
			for {
				c, err := ln.Accept()
				if xerr != nil {
					if perr := parseAcceptError(err); perr != nil {
						t.Errorf("#%d/%d: %v", i, j, perr)
					}
					if nerr, ok := err.(net.Error); !ok || !nerr.Timeout() {
						t.Fatalf("#%d/%d: %v", i, j, err)
					}
				}
				if err == nil {
					c.Close()
					time.Sleep(10 * time.Millisecond)
					continue
				}
				break
			}
		}
	}
	wg.Wait()
}
