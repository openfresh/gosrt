// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package srt

import (
	"context"
	"io"
	"net"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/openfresh/gosrt/internal/poll"
	"github.com/openfresh/gosrt/internal/testenv"
)

var prohibitionaryDialArgTests = []struct {
	network string
	address string
}{
	{"srt6", "127.0.0.1"},
	{"srt6", "::ffff:127.0.0.1"},
}

func TestProhibitionaryDialArg(t *testing.T) {
	testenv.MustHaveExternalNetwork(t)

	switch runtime.GOOS {
	case "plan9":
		t.Skipf("not supported on %s", runtime.GOOS)
	}
	if !supportsIPv4map() {
		t.Skip("mapping ipv4 address inside ipv6 address not supported")
	}

	ln, err := Listen("srt", "[::]:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	for i, tt := range prohibitionaryDialArgTests {
		c, err := Dial(tt.network, net.JoinHostPort(tt.address, port))
		if err == nil {
			c.Close()
			t.Errorf("#%d: %v", i, err)
		}
	}
}

func TestDialLocal(t *testing.T) {
	t.Skip("not supported")
	ln, err := newLocalListener("srt")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	c, err := Dial("srt", net.JoinHostPort("", port))
	if err != nil {
		t.Fatal(err)
	}
	c.Close()
}

func TestDialerDualStackFDLeak(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping test: not supported yet")
	}
	switch runtime.GOOS {
	case "plan9":
		t.Skipf("%s does not have full support of socktest", runtime.GOOS)
	case "windows":
		t.Skipf("not implemented a way to cancel dial racers in SRT SYN-SENT state on %s", runtime.GOOS)
	case "openbsd":
		testenv.SkipFlaky(t, 15157)
	}
	if !supportsIPv4() || !supportsIPv6() {
		t.Skip("both IPv4 and IPv6 are required")
	}

	before := sw.Sockets()
	origTestHookLookupIP := testHookLookupIP
	defer func() { testHookLookupIP = origTestHookLookupIP }()
	testHookLookupIP = lookupLocalhost
	handler := func(dss *dualStackServer, ln net.Listener) {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}
	dss, err := newDualStackServer()
	if err != nil {
		t.Fatal(err)
	}
	if err := dss.buildup(handler); err != nil {
		dss.teardown()
		t.Fatal(err)
	}

	const N = 10
	var wg sync.WaitGroup
	wg.Add(N)
	d := &Dialer{DualStack: true, Timeout: 5 * time.Second}
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			c, err := d.Dial("srt", net.JoinHostPort("localhost", dss.port))
			if err != nil {
				t.Error(err)
				return
			}
			c.Close()
		}()
	}
	wg.Wait()
	dss.teardown()
	after := sw.Sockets()
	if len(after) != len(before) {
		t.Errorf("got %d; want %d", len(after), len(before))
	}
}

// Define a pair of blackholed (IPv4, IPv6) addresses, for which dialSRT is
// expected to hang until the timeout elapses. These addresses are reserved
// for benchmarking by RFC 6890.
const (
	slowDst4 = "198.18.0.254"
	slowDst6 = "2001:2::254"
)

// In some environments, the slow IPs may be explicitly unreachable, and fail
// more quickly than expected. This test hook prevents dialSRT from returning
// before the deadline.
func slowDialSRT(ctx context.Context, network string, laddr, raddr *SRTAddr) (*SRTConn, error) {
	c, err := doDialSRT(ctx, network, laddr, raddr)
	if net.ParseIP(slowDst4).Equal(raddr.IP) || net.ParseIP(slowDst6).Equal(raddr.IP) {
		// Wait for the deadline, or indefinitely if none exists.
		<-ctx.Done()
	}
	return c, err
}

func dialClosedPort() (actual, expected time.Duration) {
	// Estimate the expected time for this platform.
	// On Windows, dialing a closed port takes roughly 1 second,
	// but other platforms should be instantaneous.
	if runtime.GOOS == "windows" {
		expected = 1500 * time.Millisecond
	} else if runtime.GOOS == "darwin" {
		expected = 150 * time.Millisecond
	} else {
		expected = 95 * time.Millisecond
	}

	l, err := Listen("srt", "127.0.0.1:0")
	if err != nil {
		return 999 * time.Hour, expected
	}
	addr := l.Addr().String()
	l.Close()
	// On OpenBSD, interference from TestSelfConnect is mysteriously
	// causing the first attempt to hang for a few seconds, so we throw
	// away the first result and keep the second.
	for i := 1; ; i++ {
		startTime := time.Now()
		c, err := Dial("srt", addr)
		if err == nil {
			c.Close()
		}
		elapsed := time.Now().Sub(startTime)
		if i == 2 {
			return elapsed, expected
		}
	}
}

func TestDialParallel(t *testing.T) {
	testenv.MustHaveExternalNetwork(t)

	if !supportsIPv4() || !supportsIPv6() {
		t.Skip("both IPv4 and IPv6 are required")
	}

	closedPortDelay, expectClosedPortDelay := dialClosedPort()
	if closedPortDelay > expectClosedPortDelay {
		t.Errorf("got %v; want <= %v", closedPortDelay, expectClosedPortDelay)
	}

	const instant time.Duration = 0
	const fallbackDelay = 200 * time.Millisecond

	// Some cases will run quickly when "connection refused" is fast,
	// or trigger the fallbackDelay on Windows. This value holds the
	// lesser of the two delays.
	var closedPortOrFallbackDelay time.Duration
	if closedPortDelay < fallbackDelay {
		closedPortOrFallbackDelay = closedPortDelay
	} else {
		closedPortOrFallbackDelay = fallbackDelay
	}

	origTestHookDialSRT := testHookDialSRT
	defer func() { testHookDialSRT = origTestHookDialSRT }()
	testHookDialSRT = slowDialSRT

	nCopies := func(s string, n int) []string {
		out := make([]string, n)
		for i := 0; i < n; i++ {
			out[i] = s
		}
		return out
	}

	var testCases = []struct {
		primaries       []string
		fallbacks       []string
		teardownNetwork string
		expectOk        bool
		expectElapsed   time.Duration
	}{
		// These should just work on the first try.
		{[]string{"127.0.0.1"}, []string{}, "", true, instant},
		{[]string{"::1"}, []string{}, "", true, instant},
		{[]string{"127.0.0.1", "::1"}, []string{slowDst6}, "srt6", true, instant},
		{[]string{"::1", "127.0.0.1"}, []string{slowDst4}, "srt4", true, instant},
		// Primary is slow; fallback should kick in.
		{[]string{slowDst4}, []string{"::1"}, "", true, fallbackDelay},
		// Skip a "connection refused" in the primary thread.
		{[]string{"127.0.0.1", "::1"}, []string{}, "srt4", true, closedPortDelay},
		{[]string{"::1", "127.0.0.1"}, []string{}, "srt6", true, closedPortDelay},
		// Skip a "connection refused" in the fallback thread.
		{[]string{slowDst4, slowDst6}, []string{"::1", "127.0.0.1"}, "srt6", true, fallbackDelay + closedPortDelay},
		// Primary refused, fallback without delay.
		{[]string{"127.0.0.1"}, []string{"::1"}, "srt4", true, closedPortOrFallbackDelay},
		{[]string{"::1"}, []string{"127.0.0.1"}, "srt6", true, closedPortOrFallbackDelay},
		// Everything is refused.
		{[]string{"127.0.0.1"}, []string{}, "srt4", false, closedPortDelay},
		// Nothing to do; fail instantly.
		{[]string{}, []string{}, "", false, instant},
		// Connecting to tons of addresses should not trip the deadline.
		{nCopies("::1", 1000), []string{}, "", true, instant},
	}

	handler := func(dss *dualStackServer, ln net.Listener) {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}

	// Convert a list of IP strings into SRTAddrs.
	makeAddrs := func(ips []string, port string) addrList {
		var out addrList
		for _, ip := range ips {
			addr, err := ResolveSRTAddr("srt", net.JoinHostPort(ip, port))
			if err != nil {
				t.Fatal(err)
			}
			out = append(out, addr)
		}
		return out
	}

	for i, tt := range testCases {
		dss, err := newDualStackServer()
		if err != nil {
			t.Fatal(err)
		}
		defer dss.teardown()
		if err := dss.buildup(handler); err != nil {
			t.Fatal(err)
		}
		if tt.teardownNetwork != "" {
			// Destroy one of the listening sockets, creating an unreachable port.
			dss.teardownNetwork(tt.teardownNetwork)
		}

		primaries := makeAddrs(tt.primaries, dss.port)
		fallbacks := makeAddrs(tt.fallbacks, dss.port)
		d := Dialer{
			FallbackDelay: fallbackDelay,
		}
		startTime := time.Now()
		dp := &dialParam{
			Dialer:  d,
			network: "srt",
			address: "?",
		}
		c, err := dialParallel(context.Background(), dp, primaries, fallbacks)
		elapsed := time.Since(startTime)

		if c != nil {
			c.Close()
		}

		if tt.expectOk && err != nil {
			t.Errorf("#%d: got %v; want nil", i, err)
		} else if !tt.expectOk && err == nil {
			t.Errorf("#%d: got nil; want non-nil", i)
		}

		expectElapsedMin := tt.expectElapsed - 95*time.Millisecond
		expectElapsedMax := tt.expectElapsed + 95*time.Millisecond
		if !(elapsed >= expectElapsedMin) {
			t.Errorf("#%d: got %v; want >= %v", i, elapsed, expectElapsedMin)
		} else if !(elapsed <= expectElapsedMax) {
			t.Errorf("#%d: got %v; want <= %v", i, elapsed, expectElapsedMax)
		}

		// Repeat each case, ensuring that it can be canceled quickly.
		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			time.Sleep(5 * time.Millisecond)
			cancel()
			wg.Done()
		}()
		startTime = time.Now()
		c, err = dialParallel(ctx, dp, primaries, fallbacks)
		if c != nil {
			c.Close()
		}
		elapsed = time.Now().Sub(startTime)
		if elapsed > 100*time.Millisecond {
			t.Errorf("#%d (cancel): got %v; want <= 100ms", i, elapsed)
		}
		wg.Wait()
	}
}

func lookupSlowFast(ctx context.Context, fn func(context.Context, string) ([]net.IPAddr, error), host string) ([]net.IPAddr, error) {
	switch host {
	case "slow6loopback4":
		// Returns a slow IPv6 address, and a local IPv4 address.
		return []net.IPAddr{
			{IP: net.ParseIP(slowDst6)},
			{IP: net.ParseIP("127.0.0.1")},
		}, nil
	default:
		return fn(ctx, host)
	}
}

func TestDialerFallbackDelay(t *testing.T) {
	testenv.MustHaveExternalNetwork(t)

	if !supportsIPv4() || !supportsIPv6() {
		t.Skip("both IPv4 and IPv6 are required")
	}

	origTestHookLookupIP := testHookLookupIP
	defer func() { testHookLookupIP = origTestHookLookupIP }()
	testHookLookupIP = lookupSlowFast

	origTestHookDialSRT := testHookDialSRT
	defer func() { testHookDialSRT = origTestHookDialSRT }()
	testHookDialSRT = slowDialSRT

	var testCases = []struct {
		dualstack     bool
		delay         time.Duration
		expectElapsed time.Duration
	}{
		// Use a very brief delay, which should fallback immediately.
		{true, 1 * time.Nanosecond, 0},
		// Use a 200ms explicit timeout.
		{true, 200 * time.Millisecond, 200 * time.Millisecond},
		// The default is 300ms.
		{true, 0, 300 * time.Millisecond},
	}

	handler := func(dss *dualStackServer, ln net.Listener) {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}
	dss, err := newDualStackServer()
	if err != nil {
		t.Fatal(err)
	}
	defer dss.teardown()
	if err := dss.buildup(handler); err != nil {
		t.Fatal(err)
	}

	for i, tt := range testCases {
		d := &Dialer{DualStack: tt.dualstack, FallbackDelay: tt.delay}

		startTime := time.Now()
		c, err := d.Dial("srt", net.JoinHostPort("slow6loopback4", dss.port))
		elapsed := time.Now().Sub(startTime)
		if err == nil {
			c.Close()
		} else if tt.dualstack {
			t.Error(err)
		}
		expectMin := tt.expectElapsed - 1*time.Millisecond
		expectMax := tt.expectElapsed + 95*time.Millisecond
		if !(elapsed >= expectMin) {
			t.Errorf("#%d: got %v; want >= %v", i, elapsed, expectMin)
		}
		if !(elapsed <= expectMax) {
			t.Errorf("#%d: got %v; want <= %v", i, elapsed, expectMax)
		}
	}
}

func TestDialParallelSpuriousConnection(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping test: not supported yet")
	}
	if !supportsIPv4() || !supportsIPv6() {
		t.Skip("both IPv4 and IPv6 are required")
	}

	var wg sync.WaitGroup
	wg.Add(2)
	handler := func(dss *dualStackServer, ln net.Listener) {
		// Accept one connection per address.
		c, err := ln.Accept()
		if err != nil {
			t.Fatal(err)
		}
		// The client should close itself, without sending data.
		c.SetReadDeadline(time.Now().Add(1 * time.Second))
		var b [1]byte
		if _, err := c.Read(b[:]); err != io.EOF {
			t.Errorf("got %v; want %v", err, io.EOF)
		}
		c.Close()
		wg.Done()
	}
	dss, err := newDualStackServer()
	if err != nil {
		t.Fatal(err)
	}
	defer dss.teardown()
	if err := dss.buildup(handler); err != nil {
		t.Fatal(err)
	}

	const fallbackDelay = 100 * time.Millisecond

	origTestHookDialSRT := testHookDialSRT
	defer func() { testHookDialSRT = origTestHookDialSRT }()
	testHookDialSRT = func(ctx context.Context, net string, laddr, raddr *SRTAddr) (*SRTConn, error) {
		// Sleep long enough for Happy Eyeballs to kick in, and inhibit cancelation.
		// This forces dialParallel to juggle two successful connections.
		time.Sleep(fallbackDelay * 2)

		// Now ignore the provided context (which will be canceled) and use a
		// different one to make sure this completes with a valid connection,
		// which we hope to be closed below:
		return doDialSRT(context.Background(), net, laddr, raddr)
	}

	d := Dialer{
		FallbackDelay: fallbackDelay,
	}
	dp := &dialParam{
		Dialer:  d,
		network: "srt",
		address: "?",
	}

	makeAddr := func(ip string) addrList {
		addr, err := ResolveSRTAddr("srt", net.JoinHostPort(ip, dss.port))
		if err != nil {
			t.Fatal(err)
		}
		return addrList{addr}
	}

	// dialParallel returns one connection (and closes the other.)
	c, err := dialParallel(context.Background(), dp, makeAddr("127.0.0.1"), makeAddr("::1"))
	if err != nil {
		t.Fatal(err)
	}
	c.Close()

	// The server should've seen both connections.
	wg.Wait()
}

func TestDialerPartialDeadline(t *testing.T) {
	now := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	var testCases = []struct {
		now            time.Time
		deadline       time.Time
		addrs          int
		expectDeadline time.Time
		expectErr      error
	}{
		// Regular division.
		{now, now.Add(12 * time.Second), 1, now.Add(12 * time.Second), nil},
		{now, now.Add(12 * time.Second), 2, now.Add(6 * time.Second), nil},
		{now, now.Add(12 * time.Second), 3, now.Add(4 * time.Second), nil},
		// Bump against the 2-second sane minimum.
		{now, now.Add(12 * time.Second), 999, now.Add(2 * time.Second), nil},
		// Total available is now below the sane minimum.
		{now, now.Add(1900 * time.Millisecond), 999, now.Add(1900 * time.Millisecond), nil},
		// Null deadline.
		{now, noDeadline, 1, noDeadline, nil},
		// Step the clock forward and cross the deadline.
		{now.Add(-1 * time.Millisecond), now, 1, now, nil},
		{now.Add(0 * time.Millisecond), now, 1, noDeadline, poll.ErrTimeout},
		{now.Add(1 * time.Millisecond), now, 1, noDeadline, poll.ErrTimeout},
	}
	for i, tt := range testCases {
		deadline, err := partialDeadline(tt.now, tt.deadline, tt.addrs)
		if err != tt.expectErr {
			t.Errorf("#%d: got %v; want %v", i, err, tt.expectErr)
		}
		if !deadline.Equal(tt.expectDeadline) {
			t.Errorf("#%d: got %v; want %v", i, deadline, tt.expectDeadline)
		}
	}
}

// Issue 18806: it should always be possible to net.Dial a
// net.Listener().Addr().String when the listen address was ":n", even
// if the machine has halfway configured IPv6 such that it can bind on
// "::" not connect back to that same address.
func TestDialListenerAddr(t *testing.T) {
	if testenv.Builder() == "" {
		testenv.MustHaveExternalNetwork(t)
	}
	ln, err := Listen("srt", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	addr := ln.Addr().String()
	c, err := Dial("srt", addr)
	if err != nil {
		t.Fatalf("for addr %q, dial error: %v", addr, err)
	}
	c.Close()
}
