// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrt

import (
	"context"
	"net"
	"runtime"
	"testing"
)

var srtServerTests = []struct {
	snet, saddr string // server endpoint
	tnet, taddr string // target endpoint for client
}{
	{snet: "srt", saddr: ":0", tnet: "srt", taddr: "127.0.0.1"},
	{snet: "srt", saddr: "0.0.0.0:0", tnet: "srt", taddr: "127.0.0.1"},
	{snet: "srt", saddr: "[::ffff:0.0.0.0]:0", tnet: "srt", taddr: "127.0.0.1"},
	{snet: "srt", saddr: "[::]:0", tnet: "srt", taddr: "::1"},

	{snet: "srt", saddr: ":0", tnet: "srt", taddr: "::1"},
	{snet: "srt", saddr: "0.0.0.0:0", tnet: "srt", taddr: "::1"},
	{snet: "srt", saddr: "[::ffff:0.0.0.0]:0", tnet: "srt", taddr: "::1"},
	{snet: "srt", saddr: "[::]:0", tnet: "srt", taddr: "127.0.0.1"},

	{snet: "srt", saddr: ":0", tnet: "srt4", taddr: "127.0.0.1"},
	{snet: "srt", saddr: "0.0.0.0:0", tnet: "srt4", taddr: "127.0.0.1"},
	{snet: "srt", saddr: "[::ffff:0.0.0.0]:0", tnet: "srt4", taddr: "127.0.0.1"},
	{snet: "srt", saddr: "[::]:0", tnet: "srt6", taddr: "::1"},

	{snet: "srt", saddr: ":0", tnet: "srt6", taddr: "::1"},
	{snet: "srt", saddr: "0.0.0.0:0", tnet: "srt6", taddr: "::1"},
	{snet: "srt", saddr: "[::ffff:0.0.0.0]:0", tnet: "srt6", taddr: "::1"},
	{snet: "srt", saddr: "[::]:0", tnet: "srt4", taddr: "127.0.0.1"},

	{snet: "srt", saddr: "127.0.0.1:0", tnet: "srt", taddr: "127.0.0.1"},
	{snet: "srt", saddr: "[::ffff:127.0.0.1]:0", tnet: "srt", taddr: "127.0.0.1"},
	{snet: "srt", saddr: "[::1]:0", tnet: "srt", taddr: "::1"},

	{snet: "srt4", saddr: ":0", tnet: "srt4", taddr: "127.0.0.1"},
	{snet: "srt4", saddr: "0.0.0.0:0", tnet: "srt4", taddr: "127.0.0.1"},
	{snet: "srt4", saddr: "[::ffff:0.0.0.0]:0", tnet: "srt4", taddr: "127.0.0.1"},

	{snet: "srt4", saddr: "127.0.0.1:0", tnet: "srt4", taddr: "127.0.0.1"},

	{snet: "srt6", saddr: ":0", tnet: "srt6", taddr: "::1"},
	{snet: "srt6", saddr: "[::]:0", tnet: "srt6", taddr: "::1"},

	{snet: "srt6", saddr: "[::1]:0", tnet: "srt6", taddr: "::1"},
}

// TestSRTServer tests concurrent accept-read-write servers.
func TestSRTServer(t *testing.T) {
	switch runtime.GOOS {
	case "linux":
		t.Skipf("not supported on %s", runtime.GOOS)
	}

	const N = 3
	ctx := WithOptions(context.Background(), Options("payloadsize", "15"))

	for i, tt := range srtServerTests {
		if !testableListenArgs(tt.snet, tt.saddr, tt.taddr) {
			t.Logf("skipping %s test", tt.snet+" "+tt.saddr+"<-"+tt.taddr)
			continue
		}

		ln, err := ListenContext(ctx, tt.snet, tt.saddr)
		if err != nil {
			if perr := parseDialError(err); perr != nil {
				t.Error(perr)
			}
			t.Fatal(err)
		}

		var lss []*localServer
		var tpchs []chan error
		defer func() {
			for _, ls := range lss {
				ls.teardown()
			}
		}()
		for i := 0; i < N; i++ {
			ls, err := (&streamListener{Listener: ln}).newLocalServer()
			if err != nil {
				t.Fatal(err)
			}
			lss = append(lss, ls)
			tpchs = append(tpchs, make(chan error, 1))
		}
		for i := 0; i < N; i++ {
			ch := tpchs[i]
			handler := func(ls *localServer, ln net.Listener) { transponder(ln, ch) }
			if err := lss[i].buildup(handler); err != nil {
				t.Fatal(err)
			}
		}

		var trchs []chan error
		for i := 0; i < N; i++ {
			_, port, err := net.SplitHostPort(lss[i].Listener.Addr().String())
			if err != nil {
				t.Fatal(err)
			}
			d := Dialer{Timeout: someTimeout}
			c, err := d.DialContext(ctx, tt.tnet, net.JoinHostPort(tt.taddr, port))
			if err != nil {
				if perr := parseDialError(err); perr != nil {
					t.Error(perr)
				}
				t.Fatal(err)
			}
			defer c.Close()
			trchs = append(trchs, make(chan error, 1))
			go transceiver(c, []byte("SRT SERVER TEST"), trchs[i])
		}

		for _, ch := range trchs {
			for err := range ch {
				t.Errorf("#%d: %v", i, err)
			}
		}
		for _, ch := range tpchs {
			for err := range ch {
				t.Errorf("#%d: %v", i, err)
			}
		}
	}
}
