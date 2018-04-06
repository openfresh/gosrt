// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

// This file implements API tests across platforms and will never have a build
// tag.

package srt

import (
	"context"
	"net"
	"testing"
	"time"
)

// someTimeout is used just to test that net.Conn implementations
// don't explode when their SetFooDeadline methods are called.
// It isn't actually used for testing timeouts.
const someTimeout = 10 * time.Second

func TestConnAndListener(t *testing.T) {
	ctx := WithOptions(context.Background(), Options("payloadsize", "128"))
	for i, network := range []string{"srt"} {
		if !testableNetwork(network) {
			t.Logf("skipping %s test", network)
			continue
		}

		ls, err := newLocalServerContext(ctx, network)
		if err != nil {
			t.Fatal(err)
		}
		defer ls.teardown()
		ch := make(chan error, 1)
		handler := func(ls *localServer, ln net.Listener) { transponder(ln, ch) }
		if err := ls.buildup(handler); err != nil {
			t.Fatal(err)
		}
		if ls.Listener.Addr().Network() != network {
			t.Fatalf("got %s; want %s", ls.Listener.Addr().Network(), network)
		}

		var d Dialer
		c, err := d.DialContext(ctx, ls.Listener.Addr().Network(), ls.Listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		if c.LocalAddr().Network() != network || c.RemoteAddr().Network() != network {
			t.Fatalf("got %s->%s; want %s->%s", c.LocalAddr().Network(), c.RemoteAddr().Network(), network, network)
		}
		c.SetDeadline(time.Now().Add(someTimeout))
		c.SetReadDeadline(time.Now().Add(someTimeout))
		c.SetWriteDeadline(time.Now().Add(someTimeout))

		if _, err := c.Write([]byte("CONN AND LISTENER TEST")); err != nil {
			t.Fatal(err)
		}
		rb := make([]byte, 128)
		if _, err := c.Read(rb); err != nil {
			t.Fatal(err)
		}

		for err := range ch {
			t.Errorf("#%d: %v", i, err)
		}
	}
}
