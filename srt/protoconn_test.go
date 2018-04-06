// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements API tests across platforms and will never have a build
// tag.

package srt

import (
	"context"
	"net"
	"runtime"
	"testing"
	"time"
)

// The full stack test cases for IPConn have been moved to the
// following:
//	golang.org/x/net/ipv4
//	golang.org/x/net/ipv6
//	golang.org/x/net/icmp

func TestSRTListenerSpecificMethods(t *testing.T) {
	switch runtime.GOOS {
	case "plan9":
		t.Skipf("not supported on %s", runtime.GOOS)
	}

	la, err := ResolveSRTAddr("srt4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ln, err := ListenSRT("srt4", la)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	ln.Addr()
	ln.SetDeadline(time.Now().Add(30 * time.Nanosecond))

	if c, err := ln.Accept(); err != nil {
		if !err.(net.Error).Timeout() {
			t.Fatal(err)
		}
	} else {
		c.Close()
	}
	if c, err := ln.AcceptSRT(); err != nil {
		if !err.(net.Error).Timeout() {
			t.Fatal(err)
		}
	} else {
		c.Close()
	}
}

func TestSRTConnSpecificMethods(t *testing.T) {
	ctx := WithOptions(context.Background(), Options("payloadsize", "128"))
	ln, err := ListenContext(ctx, "srt4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan error, 1)
	handler := func(ls *localServer, ln net.Listener) { transponder(ls.Listener, ch) }
	ls, err := (&streamListener{Listener: ln}).newLocalServer()
	if err != nil {
		t.Fatal(err)
	}
	defer ls.teardown()
	if err := ls.buildup(handler); err != nil {
		t.Fatal(err)
	}

	var d Dialer
	c, err := d.DialContext(ctx, "srt4", ls.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.LocalAddr()
	c.RemoteAddr()
	c.SetDeadline(time.Now().Add(someTimeout))
	c.SetReadDeadline(time.Now().Add(someTimeout))
	c.SetWriteDeadline(time.Now().Add(someTimeout))

	if _, err := c.Write([]byte("SRTCONN TEST")); err != nil {
		t.Fatal(err)
	}
	rb := make([]byte, 128)
	if _, err := c.Read(rb); err != nil {
		t.Fatal(err)
	}

	for err := range ch {
		t.Error(err)
	}
}
