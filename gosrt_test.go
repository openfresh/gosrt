// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrt

import (
	"errors"
	"fmt"
	"io"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/openfresh/gosrt/internal/socktest"
)

func TestConnClose(t *testing.T) {
	for _, network := range []string{"srt"} {
		if !testableNetwork(network) {
			t.Logf("skipping %s test", network)
			continue
		}

		ln, err := newLocalListener(network)
		if err != nil {
			t.Fatal(err)
		}
		defer ln.Close()

		c, err := Dial(ln.Addr().Network(), ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()

		if err := c.Close(); err != nil {
			if perr := parseCloseError(err, false); perr != nil {
				t.Error(perr)
			}
			t.Fatal(err)
		}
		var b [1]byte
		n, err := c.Read(b[:])
		if n != 0 || err == nil {
			t.Fatalf("got (%d, %v); want (0, error)", n, err)
		}
	}
}

func TestListenerClose(t *testing.T) {
	for _, network := range []string{"srt"} {
		if !testableNetwork(network) {
			t.Logf("skipping %s test", network)
			continue
		}

		ln, err := newLocalListener(network)
		if err != nil {
			t.Fatal(err)
		}

		dst := ln.Addr().String()
		if err := ln.Close(); err != nil {
			if perr := parseCloseError(err, false); perr != nil {
				t.Error(perr)
			}
			t.Fatal(err)
		}
		c, err := ln.Accept()
		if err == nil {
			c.Close()
			t.Fatal("should fail")
		}

		if network == "srt" {
			time.Sleep(time.Millisecond)

			cc, err := Dial("srt", dst)
			if err == nil {
				t.Error("Dial to closed SRT listener succeeded.")
				cc.Close()
			}
		}
	}
}

// nacl was previous failing to reuse an address.
func TestListenCloseListen(t *testing.T) {
	const maxTries = 10
	for tries := 0; tries < maxTries; tries++ {
		ln, err := newLocalListener("srt")
		if err != nil {
			t.Fatal(err)
		}
		addr := ln.Addr().String()
		if err := ln.Close(); err != nil {
			if perr := parseCloseError(err, false); perr != nil {
				t.Error(perr)
			}
			t.Fatal(err)
		}
		ln, err = Listen("srt", addr)
		if err == nil {
			// Success. nacl couldn't do this before.
			ln.Close()
			return
		}
		t.Errorf("failed on try %d/%d: %v", tries+1, maxTries, err)
	}
	t.Fatalf("failed to listen/close/listen on same address after %d tries", maxTries)
}

// See golang.org/issue/6163, golang.org/issue/6987.
func TestAcceptIgnoreAbortedConnRequest(t *testing.T) {
	switch runtime.GOOS {
	case "plan9":
		t.Skipf("%s does not have full support of socktest", runtime.GOOS)
	}

	syserr := make(chan error)
	go func() {
		defer close(syserr)
		for _, err := range abortedConnRequestErrors {
			syserr <- err
		}
	}()
	sw.Set(socktest.FilterAccept, func(so *socktest.Status) (socktest.AfterFilter, error) {
		if err, ok := <-syserr; ok {
			return nil, err
		}
		return nil, nil
	})
	defer sw.Set(socktest.FilterAccept, nil)

	operr := make(chan error, 1)
	handler := func(ls *localServer, ln net.Listener) {
		defer close(operr)
		c, err := ln.Accept()
		if err != nil {
			if perr := parseAcceptError(err); perr != nil {
				operr <- perr
			}
			operr <- err
			return
		}
		c.Close()
	}
	ls, err := newLocalServer("srt")
	if err != nil {
		t.Fatal(err)
	}
	defer ls.teardown()
	if err := ls.buildup(handler); err != nil {
		t.Fatal(err)
	}

	c, err := Dial(ls.Listener.Addr().Network(), ls.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	c.Close()

	for err := range operr {
		t.Error(err)
	}
}

func TestZeroByteRead(t *testing.T) {
	for _, network := range []string{"srt"} {
		if !testableNetwork(network) {
			t.Logf("skipping %s test", network)
			continue
		}

		ln, err := newLocalListener(network)
		if err != nil {
			t.Fatal(err)
		}
		connc := make(chan net.Conn, 1)
		go func() {
			defer ln.Close()
			c, err := ln.Accept()
			if err != nil {
				t.Error(err)
			}
			connc <- c // might be nil
		}()
		c, err := Dial(network, ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		sc := <-connc
		if sc == nil {
			continue
		}
		defer sc.Close()

		if runtime.GOOS == "windows" {
			// A zero byte read on Windows caused a wait for readability first.
			// Rather than change that behavior, satisfy it in this test.
			// See Issue 15735.
			go io.WriteString(sc, "a")
		}

		n, err := c.Read(nil)
		if n != 0 || err != nil {
			t.Errorf("%s: zero byte client read = %v, %v; want 0, nil", network, n, err)
		}

		if runtime.GOOS == "windows" {
			// Same as comment above.
			go io.WriteString(c, "a")
		}
		n, err = sc.Read(nil)
		if n != 0 || err != nil {
			t.Errorf("%s: zero byte server read = %v, %v; want 0, nil", network, n, err)
		}
	}
}

// withSRTConnPair sets up a SRT connection between two peers, then
// runs peer1 and peer2 concurrently. withSRTConnPair returns when
// both have completed.
func withSRTConnPair(t *testing.T, peer1, peer2 func(c *SRTConn) error) {
	ln, err := newLocalListener("srt")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	errc := make(chan error, 2)
	go func() {
		c1, err := ln.Accept()
		if err != nil {
			errc <- err
			return
		}
		defer c1.Close()
		errc <- peer1(c1.(*SRTConn))
	}()
	go func() {
		c2, err := Dial("srt", ln.Addr().String())
		if err != nil {
			errc <- err
			return
		}
		defer c2.Close()
		errc <- peer2(c2.(*SRTConn))
	}()
	for i := 0; i < 2; i++ {
		if err := <-errc; err != nil {
			t.Fatal(err)
		}
	}
}

// Tests that a blocked Read is interrupted by a concurrent SetReadDeadline
// modifying that Conn's read deadline to the past.
// See golang.org/cl/30164 which documented this. The net/http package
// depends on this.
func TestReadTimeoutUnblocksRead(t *testing.T) {
	serverDone := make(chan struct{})
	server := func(cs *SRTConn) error {
		defer close(serverDone)
		errc := make(chan error, 1)
		go func() {
			defer close(errc)
			go func() {
				// TODO: find a better way to wait
				// until we're blocked in the cs.Read
				// call below. Sleep is lame.
				time.Sleep(100 * time.Millisecond)

				// Interrupt the upcoming Read, unblocking it:
				cs.SetReadDeadline(time.Unix(123, 0)) // time in the past
			}()
			var buf [1]byte
			n, err := cs.Read(buf[:1])
			if n != 0 || err == nil {
				errc <- fmt.Errorf("Read = %v, %v; want 0, non-nil", n, err)
			}
		}()
		select {
		case err := <-errc:
			return err
		case <-time.After(5 * time.Second):
			buf := make([]byte, 2<<20)
			buf = buf[:runtime.Stack(buf, true)]
			println("Stacks at timeout:\n", string(buf))
			return errors.New("timeout waiting for Read to finish")
		}

	}
	// Do nothing in the client. Never write. Just wait for the
	// server's half to be done.
	client := func(*SRTConn) error {
		<-serverDone
		return nil
	}
	withSRTConnPair(t, client, server)
}

// Issue 17695: verify that a blocked Read is woken up by a Close.
func TestCloseUnblocksRead(t *testing.T) {
	t.Parallel()
	server := func(cs *SRTConn) error {
		// Give the client time to get stuck in a Read:
		time.Sleep(20 * time.Millisecond)
		cs.Close()
		return nil
	}
	client := func(ss *SRTConn) error {
		n, err := ss.Read([]byte{0})
		if n != 0 || err != io.EOF {
			return fmt.Errorf("Read = %v, %v; want 0, EOF", n, err)
		}
		return nil
	}
	withSRTConnPair(t, client, server)
}
