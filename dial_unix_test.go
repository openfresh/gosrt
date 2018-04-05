// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package gosrt

import (
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/openfresh/gosrt/srtapi"
)

// Issue 16523
func TestDialContextCancelRace(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping test")
	}
	oldConnectFunc := connectFunc
	oldGetsockoptIntFunc := getsockoptIntFunc
	oldTestHookCanceledDial := testHookCanceledDial
	defer func() {
		connectFunc = oldConnectFunc
		getsockoptIntFunc = oldGetsockoptIntFunc
		testHookCanceledDial = oldTestHookCanceledDial
	}()

	ln, err := newLocalListener("srt")
	if err != nil {
		t.Fatal(err)
	}
	listenerDone := make(chan struct{})
	go func() {
		defer close(listenerDone)
		c, err := ln.Accept()
		if err == nil {
			c.Close()
		}
	}()
	defer func() { <-listenerDone }()
	defer ln.Close()

	sawCancel := make(chan bool, 1)
	testHookCanceledDial = func() {
		sawCancel <- true
	}

	ctx, cancelCtx := context.WithCancel(context.Background())

	connectFunc = func(fd int, addr syscall.Sockaddr) error {
		err := oldConnectFunc(fd, addr)
		t.Logf("connect(%d, addr) = %v", fd, err)
		return err
	}

	getsockoptIntFunc = func(fd, level, opt int) (val int, err error) {
		val, err = oldGetsockoptIntFunc(fd, level, opt)
		t.Logf("getsockoptIntFunc(%d, %d, %d) = (%v, %v)", fd, level, opt, val, err)
		if level == 0 && opt == srtapi.OptionState && err == nil && val == srtapi.StatusConnected {
			t.Logf("canceling context")

			// Cancel the context at just the moment which
			// caused the race in issue 16523.
			cancelCtx()

			// And wait for the "interrupter" goroutine to
			// cancel the dial by messing with its write
			// timeout before returning.
			select {
			case <-sawCancel:
				t.Logf("saw cancel")
			case <-time.After(5 * time.Second):
				t.Errorf("didn't see cancel after 5 seconds")
			}
		}
		return
	}

	var d Dialer
	c, err := d.DialContext(ctx, "srt", ln.Addr().String())
	if err == nil {
		c.Close()
		t.Fatal("unexpected successful dial; want context canceled error")
	}

	select {
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("expected context to be canceled")
	}

	oe, ok := err.(*OpError)
	if !ok || oe.Op != "dial" {
		t.Fatalf("Dial error = %#v; want dial *OpError", err)
	}
	if oe.Err != ctx.Err() {
		t.Errorf("DialContext = (%v, %v); want OpError with error %v", c, err, ctx.Err())
	}
}
