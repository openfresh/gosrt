// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9

package srt

import (
	"fmt"
	"net"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/openfresh/gosrt/internal/testenv"
)

func (ln *SRTListener) port() string {
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		return ""
	}
	return port
}

var srtListenerTests = []struct {
	network string
	address string
}{
	{"srt", ""},
	{"srt", "0.0.0.0"},
	{"srt", "::ffff:0.0.0.0"},
	{"srt", "::"},

	{"srt", "127.0.0.1"},
	{"srt", "::ffff:127.0.0.1"},
	{"srt", "::1"},

	{"srt4", ""},
	{"srt4", "0.0.0.0"},
	{"srt4", "::ffff:0.0.0.0"},

	{"srt4", "127.0.0.1"},
	{"srt4", "::ffff:127.0.0.1"},

	{"srt6", ""},
	{"srt6", "::"},

	{"srt6", "::1"},
}

// TestSRTListener tests both single and double listen to a test
// listener with same address family, same listening address and
// same port.
func TestSRTListener(t *testing.T) {
	switch runtime.GOOS {
	case "plan9":
		t.Skipf("not supported on %s", runtime.GOOS)
	}

	for _, tt := range srtListenerTests {
		if !testableListenArgs(tt.network, net.JoinHostPort(tt.address, "0"), "") {
			t.Logf("skipping %s test", tt.network+" "+tt.address)
			continue
		}

		ln1, err := Listen(tt.network, net.JoinHostPort(tt.address, "0"))
		if err != nil {
			t.Fatal(err)
		}
		if err := checkFirstListener(tt.network, ln1); err != nil {
			ln1.Close()
			t.Fatal(err)
		}
		ln2, err := Listen(tt.network, net.JoinHostPort(tt.address, ln1.(*SRTListener).port()))
		if err == nil {
			ln2.Close()
		}
		if err := checkSecondListener(tt.network, tt.address, err); err != nil {
			ln1.Close()
			t.Fatal(err)
		}
		ln1.Close()
	}
}

var dualStackSRTListenerTests = []struct {
	network1, address1 string // first listener
	network2, address2 string // second listener
	xerr               error  // expected error value, nil or other
}{
	// Test cases and expected results for the attempting 2nd listen on the same port
	// 1st listen                2nd listen                 darwin  freebsd  linux  openbsd
	// ------------------------------------------------------------------------------------
	// "srt"  ""                 "srt"  ""                    -        -       -       -
	// "srt"  ""                 "srt"  "0.0.0.0"             -        -       -       -
	// "srt"  "0.0.0.0"          "srt"  ""                    -        -       -       -
	// ------------------------------------------------------------------------------------
	// "srt"  ""                 "srt"  "[::]"                -        -       -       ok
	// "srt"  "[::]"             "srt"  ""                    -        -       -       ok
	// "srt"  "0.0.0.0"          "srt"  "[::]"                -        -       -       ok
	// "srt"  "[::]"             "srt"  "0.0.0.0"             -        -       -       ok
	// "srt"  "[::ffff:0.0.0.0]" "srt"  "[::]"                -        -       -       ok
	// "srt"  "[::]"             "srt"  "[::ffff:0.0.0.0]"    -        -       -       ok
	// ------------------------------------------------------------------------------------
	// "srt4" ""                 "srt6" ""                    ok       ok      ok      ok
	// "srt6" ""                 "srt4" ""                    ok       ok      ok      ok
	// "srt4" "0.0.0.0"          "srt6" "[::]"                ok       ok      ok      ok
	// "srt6" "[::]"             "srt4" "0.0.0.0"             ok       ok      ok      ok
	// ------------------------------------------------------------------------------------
	// "srt"  "127.0.0.1"        "srt"  "[::1]"               ok       ok      ok      ok
	// "srt"  "[::1]"            "srt"  "127.0.0.1"           ok       ok      ok      ok
	// "srt4" "127.0.0.1"        "srt6" "[::1]"               ok       ok      ok      ok
	// "srt6" "[::1]"            "srt4" "127.0.0.1"           ok       ok      ok      ok
	//
	// Platform default configurations:
	// darwin, kernel version 11.3.0
	//	net.inet6.ip6.v6only=0 (overridable by sysctl or IPV6_V6ONLY option)
	// freebsd, kernel version 8.2
	//	net.inet6.ip6.v6only=1 (overridable by sysctl or IPV6_V6ONLY option)
	// linux, kernel version 3.0.0
	//	net.ipv6.bindv6only=0 (overridable by sysctl or IPV6_V6ONLY option)
	// openbsd, kernel version 5.0
	//	net.inet6.ip6.v6only=1 (overriding is prohibited)

	{"srt", "", "srt", "", syscall.EADDRINUSE},
	{"srt", "", "srt", "0.0.0.0", syscall.EADDRINUSE},
	{"srt", "0.0.0.0", "srt", "", syscall.EADDRINUSE},

	{"srt", "", "srt", "::", syscall.EADDRINUSE},
	{"srt", "::", "srt", "", syscall.EADDRINUSE},
	{"srt", "0.0.0.0", "srt", "::", syscall.EADDRINUSE},
	{"srt", "::", "srt", "0.0.0.0", syscall.EADDRINUSE},
	{"srt", "::ffff:0.0.0.0", "srt", "::", syscall.EADDRINUSE},
	{"srt", "::", "srt", "::ffff:0.0.0.0", syscall.EADDRINUSE},

	{"srt4", "", "srt6", "", nil},
	{"srt6", "", "srt4", "", nil},
	{"srt4", "0.0.0.0", "srt6", "::", nil},
	{"srt6", "::", "srt4", "0.0.0.0", nil},

	{"srt", "127.0.0.1", "srt", "::1", nil},
	{"srt", "::1", "srt", "127.0.0.1", nil},
	{"srt4", "127.0.0.1", "srt6", "::1", nil},
	{"srt6", "::1", "srt4", "127.0.0.1", nil},
}

// TestDualStackSRTListener tests both single and double listen
// to a test listener with various address families, different
// listening address and same port.
//
// On DragonFly BSD, we expect the kernel version of node under test
// to be greater than or equal to 4.4.
func TestDualStackSRTListener(t *testing.T) {
	switch runtime.GOOS {
	case "nacl", "plan9":
		t.Skipf("not supported on %s", runtime.GOOS)
	}
	if !supportsIPv4() || !supportsIPv6() {
		t.Skip("both IPv4 and IPv6 are required")
	}

	for _, tt := range dualStackSRTListenerTests {
		if !testableListenArgs(tt.network1, net.JoinHostPort(tt.address1, "0"), "") {
			t.Logf("skipping %s test", tt.network1+" "+tt.address1)
			continue
		}

		if !supportsIPv4map() && differentWildcardAddr(tt.address1, tt.address2) {
			tt.xerr = nil
		}
		var firstErr, secondErr error
		for i := 0; i < 5; i++ {
			lns, err := newDualStackListener()
			if err != nil {
				t.Fatal(err)
			}
			port := lns[0].port()
			for _, ln := range lns {
				ln.Close()
			}
			var ln1 net.Listener
			ln1, firstErr = Listen(tt.network1, net.JoinHostPort(tt.address1, port))
			if firstErr != nil {
				continue
			}
			if err := checkFirstListener(tt.network1, ln1); err != nil {
				ln1.Close()
				t.Fatal(err)
			}
			ln2, err := Listen(tt.network2, net.JoinHostPort(tt.address2, ln1.(*SRTListener).port()))
			if err == nil {
				ln2.Close()
			}
			if secondErr = checkDualStackSecondListener(tt.network2, tt.address2, err, tt.xerr); secondErr != nil {
				ln1.Close()
				continue
			}
			ln1.Close()
			break
		}
		if firstErr != nil {
			t.Error(firstErr)
		}
		if secondErr != nil {
			t.Error(secondErr)
		}
	}
}

func differentWildcardAddr(i, j string) bool {
	if (i == "" || i == "0.0.0.0" || i == "::ffff:0.0.0.0") && (j == "" || j == "0.0.0.0" || j == "::ffff:0.0.0.0") {
		return false
	}
	if i == "[::]" && j == "[::]" {
		return false
	}
	return true
}

func checkFirstListener(network string, ln interface{}) error {
	switch network {
	case "srt":
		fd := ln.(*SRTListener).fd
		if err := checkDualStackAddrFamily(fd); err != nil {
			return err
		}
	case "srt4":
		fd := ln.(*SRTListener).fd
		if fd.family != syscall.AF_INET {
			return fmt.Errorf("%v got %v; want %v", fd.laddr, fd.family, syscall.AF_INET)
		}
	case "srt6":
		fd := ln.(*SRTListener).fd
		if fd.family != syscall.AF_INET6 {
			return fmt.Errorf("%v got %v; want %v", fd.laddr, fd.family, syscall.AF_INET6)
		}
	default:
		return net.UnknownNetworkError(network)
	}
	return nil
}

func checkSecondListener(network, address string, err error) error {
	switch network {
	case "srt", "srt4", "srt6":
		if err == nil {
			return fmt.Errorf("%s should fail", network+" "+address)
		}
	default:
		return net.UnknownNetworkError(network)
	}
	return nil
}

func checkDualStackSecondListener(network, address string, err, xerr error) error {
	switch network {
	case "srt", "srt4", "srt6":
		if xerr == nil && err != nil || xerr != nil && err == nil {
			return fmt.Errorf("%s got %v; want %v", network+" "+address, err, xerr)
		}
	default:
		return net.UnknownNetworkError(network)
	}
	return nil
}

func checkDualStackAddrFamily(fd *netFD) error {
	switch a := fd.laddr.(type) {
	case *SRTAddr:
		// If a node under test supports both IPv6 capability
		// and IPv6 IPv4-mapping capability, we can assume
		// that the node listens on a wildcard address with an
		// AF_INET6 socket.
		if supportsIPv4map() && fd.laddr.(*SRTAddr).isWildcard() {
			if fd.family != syscall.AF_INET6 {
				return fmt.Errorf("Listen(%s, %v) returns %v; want %v", fd.net, fd.laddr, fd.family, syscall.AF_INET6)
			}
		} else {
			if fd.family != a.family() {
				return fmt.Errorf("Listen(%s, %v) returns %v; want %v", fd.net, fd.laddr, fd.family, a.family())
			}
		}
	default:
		return fmt.Errorf("unexpected protocol address type: %T", a)
	}
	return nil
}

func TestWildWildcardListener(t *testing.T) {
	testenv.MustHaveExternalNetwork(t)

	switch runtime.GOOS {
	case "plan9":
		t.Skipf("not supported on %s", runtime.GOOS)
	}

	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("panicked: %v", p)
		}
	}()

	if ln, err := Listen("srt", ""); err == nil {
		ln.Close()
	}
	if ln, err := ListenSRT("srt", nil); err == nil {
		ln.Close()
	}
}

// Issue 21856.
func TestClosingListener(t *testing.T) {
	ln, err := newLocalListener("srt")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr()

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()

	// Let the goroutine start. We don't sleep long: if the
	// goroutine doesn't start, the test will pass without really
	// testing anything, which is OK.
	time.Sleep(time.Millisecond)

	ln.Close()

	ln2, err := Listen("srt", addr.String())
	if err != nil {
		t.Fatal(err)
	}
	ln2.Close()
}
