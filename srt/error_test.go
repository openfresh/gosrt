// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package srt

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/openfresh/gosrt/internal/poll"
	"github.com/openfresh/gosrt/internal/socktest"
)

func (e *OpError) isValid() error {
	if e.Op == "" {
		return fmt.Errorf("OpError.Op is empty: %v", e)
	}
	if e.Net == "" {
		return fmt.Errorf("OpError.Net is empty: %v", e)
	}
	for _, addr := range []net.Addr{e.Source, e.Addr} {
		switch addr := addr.(type) {
		case nil:
		case *SRTAddr:
			if addr == nil {
				return fmt.Errorf("OpError.Source or Addr is non-nil interface: %#v, %v", addr, e)
			}
		default:
			return fmt.Errorf("OpError.Source or Addr is unknown type: %T, %v", addr, e)
		}
	}
	if e.Err == nil {
		return fmt.Errorf("OpError.Err is empty: %v", e)
	}
	return nil
}

// parseDialError parses nestedErr and reports whether it is a valid
// error value from Dial, Listen functions.
// It returns nil when nestedErr is valid.
func parseDialError(nestedErr error) error {
	if nestedErr == nil {
		return nil
	}

	switch err := nestedErr.(type) {
	case *OpError:
		if err := err.isValid(); err != nil {
			return err
		}
		nestedErr = err.Err
		goto second
	}
	return fmt.Errorf("unexpected type on 1st nested level: %T", nestedErr)

second:
	if isPlatformError(nestedErr) {
		return nil
	}
	switch err := nestedErr.(type) {
	case *net.AddrError, *net.DNSError, net.InvalidAddrError, *net.ParseError, *poll.TimeoutError, net.UnknownNetworkError:
		return nil
	case *os.SyscallError:
		nestedErr = err.Err
		goto third
	case *os.PathError: // for Plan 9
		nestedErr = err.Err
		goto third
	}
	switch nestedErr {
	case errCanceled, poll.ErrNetClosing, errMissingAddress, errNoSuitableAddress,
		context.DeadlineExceeded, context.Canceled:
		return nil
	}
	return fmt.Errorf("unexpected type on 2nd nested level: %T", nestedErr)

third:
	if isPlatformError(nestedErr) {
		return nil
	}
	return fmt.Errorf("unexpected type on 3rd nested level: %T", nestedErr)
}

var dialErrorTests = []struct {
	network, address string
}{
	{"foo", ""},
	{"bar", "baz"},
	{"datakit", "mh/astro/r70"},
	{"srt", ""},
	{"srt", "127.0.0.1:☺"},
	{"srt", "no-such-name:80"},
	{"srt", "mh/astro/r70:http"},

	{"srt", net.JoinHostPort("127.0.0.1", "-1")},
	{"srt", net.JoinHostPort("127.0.0.1", "123456789")},
}

func TestDialError(t *testing.T) {
	switch runtime.GOOS {
	case "plan9":
		t.Skipf("%s does not have full support of socktest", runtime.GOOS)
	}

	origTestHookLookupIP := testHookLookupIP
	defer func() { testHookLookupIP = origTestHookLookupIP }()
	testHookLookupIP = func(ctx context.Context, fn func(context.Context, string) ([]net.IPAddr, error), host string) ([]net.IPAddr, error) {
		return nil, &net.DNSError{Err: "dial error test", Name: "name", Server: "server", IsTimeout: true}
	}
	sw.Set(socktest.FilterConnect, func(so *socktest.Status) (socktest.AfterFilter, error) {
		return nil, errOpNotSupported
	})
	defer sw.Set(socktest.FilterConnect, nil)

	d := Dialer{Timeout: someTimeout}
	for i, tt := range dialErrorTests {
		c, err := d.Dial(tt.network, tt.address)
		if err == nil {
			t.Errorf("#%d: should fail; %s:%s->%s", i, c.LocalAddr().Network(), c.LocalAddr(), c.RemoteAddr())
			c.Close()
			continue
		}
		if tt.network == "srt" {
			nerr := err
			if op, ok := nerr.(*OpError); ok {
				nerr = op.Err
			}
			if sys, ok := nerr.(*os.SyscallError); ok {
				nerr = sys.Err
			}
			if nerr == errOpNotSupported {
				t.Errorf("#%d: should fail without %v; %s:%s->", i, nerr, tt.network, tt.address)
				continue
			}
		}
		if c != nil {
			t.Errorf("Dial returned non-nil interface %T(%v) with err != nil", c, c)
		}
		if err = parseDialError(err); err != nil {
			t.Errorf("#%d: %v", i, err)
			continue
		}
	}
}

func TestProtocolDialError(t *testing.T) {
	switch runtime.GOOS {
	case "nacl", "solaris":
		t.Skipf("not supported on %s", runtime.GOOS)
	}

	for _, network := range []string{"srt"} {
		var err error
		switch network {
		case "srt":
			_, err = DialSRT(network, nil, &SRTAddr{Port: 1 << 16})
		}
		if err == nil {
			t.Errorf("%s: should fail", network)
			continue
		}
		if err = parseDialError(err); err != nil {
			t.Errorf("%s: %v", network, err)
			continue
		}
	}
}

func TestDialAddrError(t *testing.T) {
	switch runtime.GOOS {
	case "nacl", "plan9":
		t.Skipf("not supported on %s", runtime.GOOS)
	}
	if !supportsIPv4() || !supportsIPv6() {
		t.Skip("both IPv4 and IPv6 are required")
	}

	for _, tt := range []struct {
		network string
		lit     string
		addr    *SRTAddr
	}{
		{"srt4", "::1", nil},
		{"srt4", "", &SRTAddr{IP: net.IPv6loopback}},
		// We don't test the {"srt6", "byte sequence", nil}
		// case for now because there is no easy way to
		// control name resolution.
		{"srt6", "", &SRTAddr{IP: net.IP{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}}},
	} {
		var err error
		var c net.Conn
		var op string
		if tt.lit != "" {
			c, err = Dial(tt.network, net.JoinHostPort(tt.lit, "0"))
			op = fmt.Sprintf("Dial(%q, %q)", tt.network, net.JoinHostPort(tt.lit, "0"))
		} else {
			c, err = DialSRT(tt.network, nil, tt.addr)
			op = fmt.Sprintf("DialSRT(%q, %q)", tt.network, tt.addr)
		}
		if err == nil {
			c.Close()
			t.Errorf("%s succeeded, want error", op)
			continue
		}
		if perr := parseDialError(err); perr != nil {
			t.Errorf("%s: %v", op, perr)
			continue
		}
		operr := err.(*OpError).Err
		aerr, ok := operr.(*net.AddrError)
		if !ok {
			t.Errorf("%s: %v is %T, want *AddrError", op, err, operr)
			continue
		}
		want := tt.lit
		if tt.lit == "" {
			want = tt.addr.IP.String()
		}
		if aerr.Addr != want {
			t.Errorf("%s: %v, error Addr=%q, want %q", op, err, aerr.Addr, want)
		}
	}
}

var listenErrorTests = []struct {
	network, address string
}{
	{"foo", ""},
	{"bar", "baz"},
	{"datakit", "mh/astro/r70"},
	{"srt", "127.0.0.1:☺"},
	{"srt", "no-such-name:80"},
	{"srt", "mh/astro/r70:http"},

	{"srt", net.JoinHostPort("127.0.0.1", "-1")},
	{"srt", net.JoinHostPort("127.0.0.1", "123456789")},
}

func TestListenError(t *testing.T) {
	switch runtime.GOOS {
	case "plan9":
		t.Skipf("%s does not have full support of socktest", runtime.GOOS)
	}

	origTestHookLookupIP := testHookLookupIP
	defer func() { testHookLookupIP = origTestHookLookupIP }()
	testHookLookupIP = func(_ context.Context, fn func(context.Context, string) ([]net.IPAddr, error), host string) ([]net.IPAddr, error) {
		return nil, &net.DNSError{Err: "listen error test", Name: "name", Server: "server", IsTimeout: true}
	}
	sw.Set(socktest.FilterListen, func(so *socktest.Status) (socktest.AfterFilter, error) {
		return nil, errOpNotSupported
	})
	defer sw.Set(socktest.FilterListen, nil)

	for i, tt := range listenErrorTests {
		ln, err := Listen(tt.network, tt.address)
		if err == nil {
			t.Errorf("#%d: should fail; %s:%s->", i, ln.Addr().Network(), ln.Addr())
			ln.Close()
			continue
		}
		if tt.network == "srt" {
			nerr := err
			if op, ok := nerr.(*OpError); ok {
				nerr = op.Err
			}
			if sys, ok := nerr.(*os.SyscallError); ok {
				nerr = sys.Err
			}
			if nerr == errOpNotSupported {
				t.Errorf("#%d: should fail without %v; %s:%s->", i, nerr, tt.network, tt.address)
				continue
			}
		}
		if ln != nil {
			t.Errorf("Listen returned non-nil interface %T(%v) with err != nil", ln, ln)
		}
		if err = parseDialError(err); err != nil {
			t.Errorf("#%d: %v", i, err)
			continue
		}
	}
}

func TestProtocolListenError(t *testing.T) {
	switch runtime.GOOS {
	case "nacl", "plan9":
		t.Skipf("not supported on %s", runtime.GOOS)
	}

	for _, network := range []string{"srt"} {
		var err error
		switch network {
		case "srt":
			_, err = ListenSRT(network, &SRTAddr{Port: 1 << 16})
		}
		if err == nil {
			t.Errorf("%s: should fail", network)
			continue
		}
		if err = parseDialError(err); err != nil {
			t.Errorf("%s: %v", network, err)
			continue
		}
	}
}

// parseReadError parses nestedErr and reports whether it is a valid
// error value from Read functions.
// It returns nil when nestedErr is valid.
func parseReadError(nestedErr error) error {
	if nestedErr == nil {
		return nil
	}

	switch err := nestedErr.(type) {
	case *OpError:
		if err := err.isValid(); err != nil {
			return err
		}
		nestedErr = err.Err
		goto second
	}
	if nestedErr == io.EOF {
		return nil
	}
	return fmt.Errorf("unexpected type on 1st nested level: %T", nestedErr)

second:
	if isPlatformError(nestedErr) {
		return nil
	}
	switch err := nestedErr.(type) {
	case *os.SyscallError:
		nestedErr = err.Err
		goto third
	}
	switch nestedErr {
	case poll.ErrNetClosing, poll.ErrTimeout:
		return nil
	}
	return fmt.Errorf("unexpected type on 2nd nested level: %T", nestedErr)

third:
	if isPlatformError(nestedErr) {
		return nil
	}
	return fmt.Errorf("unexpected type on 3rd nested level: %T", nestedErr)
}

// parseWriteError parses nestedErr and reports whether it is a valid
// error value from Write functions.
// It returns nil when nestedErr is valid.
func parseWriteError(nestedErr error) error {
	if nestedErr == nil {
		return nil
	}

	switch err := nestedErr.(type) {
	case *OpError:
		if err := err.isValid(); err != nil {
			return err
		}
		nestedErr = err.Err
		goto second
	}
	return fmt.Errorf("unexpected type on 1st nested level: %T", nestedErr)

second:
	if isPlatformError(nestedErr) {
		return nil
	}
	switch err := nestedErr.(type) {
	case *net.AddrError, *net.DNSError, net.InvalidAddrError, *net.ParseError, *poll.TimeoutError, net.UnknownNetworkError:
		return nil
	case *os.SyscallError:
		nestedErr = err.Err
		goto third
	}
	switch nestedErr {
	case errCanceled, poll.ErrNetClosing, errMissingAddress, poll.ErrTimeout, net.ErrWriteToConnected, io.ErrUnexpectedEOF:
		return nil
	}
	return fmt.Errorf("unexpected type on 2nd nested level: %T", nestedErr)

third:
	if isPlatformError(nestedErr) {
		return nil
	}
	return fmt.Errorf("unexpected type on 3rd nested level: %T", nestedErr)
}

// parseCloseError parses nestedErr and reports whether it is a valid
// error value from Close functions.
// It returns nil when nestedErr is valid.
func parseCloseError(nestedErr error, isShutdown bool) error {
	if nestedErr == nil {
		return nil
	}

	// Because historically we have not exported the error that we
	// return for an operation on a closed network connection,
	// there are programs that test for the exact error string.
	// Verify that string here so that we don't break those
	// programs unexpectedly. See issues #4373 and #19252.
	want := "use of closed network connection"
	if !isShutdown && !strings.Contains(nestedErr.Error(), want) {
		return fmt.Errorf("error string %q does not contain expected string %q", nestedErr, want)
	}

	switch err := nestedErr.(type) {
	case *OpError:
		if err := err.isValid(); err != nil {
			return err
		}
		nestedErr = err.Err
		goto second
	}
	return fmt.Errorf("unexpected type on 1st nested level: %T", nestedErr)

second:
	if isPlatformError(nestedErr) {
		return nil
	}
	switch err := nestedErr.(type) {
	case *os.SyscallError:
		nestedErr = err.Err
		goto third
	case *os.PathError: // for Plan 9
		nestedErr = err.Err
		goto third
	}
	switch nestedErr {
	case poll.ErrNetClosing:
		return nil
	}
	return fmt.Errorf("unexpected type on 2nd nested level: %T", nestedErr)

third:
	if isPlatformError(nestedErr) {
		return nil
	}
	switch nestedErr {
	case os.ErrClosed: // for Plan 9
		return nil
	}
	return fmt.Errorf("unexpected type on 3rd nested level: %T", nestedErr)
}

func TestCloseError(t *testing.T) {
	ln, err := newLocalListener("srt")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	c, err := Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	for i := 0; i < 3; i++ {
		err = c.Close()
		if perr := parseCloseError(err, false); perr != nil {
			t.Errorf("#%d: %v", i, perr)
		}
		err = ln.Close()
		if perr := parseCloseError(err, false); perr != nil {
			t.Errorf("#%d: %v", i, perr)
		}
	}
}

// parseAcceptError parses nestedErr and reports whether it is a valid
// error value from Accept functions.
// It returns nil when nestedErr is valid.
func parseAcceptError(nestedErr error) error {
	if nestedErr == nil {
		return nil
	}

	switch err := nestedErr.(type) {
	case *OpError:
		if err := err.isValid(); err != nil {
			return err
		}
		nestedErr = err.Err
		goto second
	}
	return fmt.Errorf("unexpected type on 1st nested level: %T", nestedErr)

second:
	if isPlatformError(nestedErr) {
		return nil
	}
	switch err := nestedErr.(type) {
	case *os.SyscallError:
		nestedErr = err.Err
		goto third
	case *os.PathError: // for Plan 9
		nestedErr = err.Err
		goto third
	}
	switch nestedErr {
	case poll.ErrNetClosing, poll.ErrTimeout:
		return nil
	}
	return fmt.Errorf("unexpected type on 2nd nested level: %T", nestedErr)

third:
	if isPlatformError(nestedErr) {
		return nil
	}
	return fmt.Errorf("unexpected type on 3rd nested level: %T", nestedErr)
}

func TestAcceptError(t *testing.T) {
	handler := func(ls *localServer, ln net.Listener) {
		for {
			ln.(*SRTListener).SetDeadline(time.Now().Add(5 * time.Millisecond))
			c, err := ln.Accept()
			if perr := parseAcceptError(err); perr != nil {
				t.Error(perr)
			}
			if err != nil {
				if c != nil {
					t.Errorf("Accept returned non-nil interface %T(%v) with err != nil", c, c)
				}
				if nerr, ok := err.(net.Error); !ok || (!nerr.Timeout() && !nerr.Temporary()) {
					return
				}
				continue
			}
			c.Close()
		}
	}
	ls, err := newLocalServer("srt")
	if err != nil {
		t.Fatal(err)
	}
	if err := ls.buildup(handler); err != nil {
		ls.teardown()
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)
	ls.teardown()
}
