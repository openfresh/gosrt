// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package srt

import (
	"net"
	"reflect"
	"testing"
)

var testInetaddr = func(ip net.IPAddr) net.Addr { return &SRTAddr{IP: ip.IP, Port: 5682, Zone: ip.Zone} }

var addrListTests = []struct {
	filter    func(net.IPAddr) bool
	ips       []net.IPAddr
	inetaddr  func(net.IPAddr) net.Addr
	first     net.Addr
	primaries addrList
	fallbacks addrList
	err       error
}{
	{
		nil,
		[]net.IPAddr{
			{IP: net.IPv4(127, 0, 0, 1)},
			{IP: net.IPv6loopback},
		},
		testInetaddr,
		&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682},
		addrList{&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682}},
		addrList{&SRTAddr{IP: net.IPv6loopback, Port: 5682}},
		nil,
	},
	{
		nil,
		[]net.IPAddr{
			{IP: net.IPv6loopback},
			{IP: net.IPv4(127, 0, 0, 1)},
		},
		testInetaddr,
		&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682},
		addrList{&SRTAddr{IP: net.IPv6loopback, Port: 5682}},
		addrList{&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682}},
		nil,
	},
	{
		nil,
		[]net.IPAddr{
			{IP: net.IPv4(127, 0, 0, 1)},
			{IP: net.IPv4(192, 168, 0, 1)},
		},
		testInetaddr,
		&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682},
		addrList{
			&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682},
			&SRTAddr{IP: net.IPv4(192, 168, 0, 1), Port: 5682},
		},
		nil,
		nil,
	},
	{
		nil,
		[]net.IPAddr{
			{IP: net.IPv6loopback},
			{IP: net.ParseIP("fe80::1"), Zone: "eth0"},
		},
		testInetaddr,
		&SRTAddr{IP: net.IPv6loopback, Port: 5682},
		addrList{
			&SRTAddr{IP: net.IPv6loopback, Port: 5682},
			&SRTAddr{IP: net.ParseIP("fe80::1"), Port: 5682, Zone: "eth0"},
		},
		nil,
		nil,
	},
	{
		nil,
		[]net.IPAddr{
			{IP: net.IPv4(127, 0, 0, 1)},
			{IP: net.IPv4(192, 168, 0, 1)},
			{IP: net.IPv6loopback},
			{IP: net.ParseIP("fe80::1"), Zone: "eth0"},
		},
		testInetaddr,
		&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682},
		addrList{
			&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682},
			&SRTAddr{IP: net.IPv4(192, 168, 0, 1), Port: 5682},
		},
		addrList{
			&SRTAddr{IP: net.IPv6loopback, Port: 5682},
			&SRTAddr{IP: net.ParseIP("fe80::1"), Port: 5682, Zone: "eth0"},
		},
		nil,
	},
	{
		nil,
		[]net.IPAddr{
			{IP: net.IPv6loopback},
			{IP: net.ParseIP("fe80::1"), Zone: "eth0"},
			{IP: net.IPv4(127, 0, 0, 1)},
			{IP: net.IPv4(192, 168, 0, 1)},
		},
		testInetaddr,
		&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682},
		addrList{
			&SRTAddr{IP: net.IPv6loopback, Port: 5682},
			&SRTAddr{IP: net.ParseIP("fe80::1"), Port: 5682, Zone: "eth0"},
		},
		addrList{
			&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682},
			&SRTAddr{IP: net.IPv4(192, 168, 0, 1), Port: 5682},
		},
		nil,
	},
	{
		nil,
		[]net.IPAddr{
			{IP: net.IPv4(127, 0, 0, 1)},
			{IP: net.IPv6loopback},
			{IP: net.IPv4(192, 168, 0, 1)},
			{IP: net.ParseIP("fe80::1"), Zone: "eth0"},
		},
		testInetaddr,
		&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682},
		addrList{
			&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682},
			&SRTAddr{IP: net.IPv4(192, 168, 0, 1), Port: 5682},
		},
		addrList{
			&SRTAddr{IP: net.IPv6loopback, Port: 5682},
			&SRTAddr{IP: net.ParseIP("fe80::1"), Port: 5682, Zone: "eth0"},
		},
		nil,
	},
	{
		nil,
		[]net.IPAddr{
			{IP: net.IPv6loopback},
			{IP: net.IPv4(127, 0, 0, 1)},
			{IP: net.ParseIP("fe80::1"), Zone: "eth0"},
			{IP: net.IPv4(192, 168, 0, 1)},
		},
		testInetaddr,
		&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682},
		addrList{
			&SRTAddr{IP: net.IPv6loopback, Port: 5682},
			&SRTAddr{IP: net.ParseIP("fe80::1"), Port: 5682, Zone: "eth0"},
		},
		addrList{
			&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682},
			&SRTAddr{IP: net.IPv4(192, 168, 0, 1), Port: 5682},
		},
		nil,
	},

	{
		ipv4only,
		[]net.IPAddr{
			{IP: net.IPv4(127, 0, 0, 1)},
			{IP: net.IPv6loopback},
		},
		testInetaddr,
		&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682},
		addrList{&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682}},
		nil,
		nil,
	},
	{
		ipv4only,
		[]net.IPAddr{
			{IP: net.IPv6loopback},
			{IP: net.IPv4(127, 0, 0, 1)},
		},
		testInetaddr,
		&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682},
		addrList{&SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5682}},
		nil,
		nil,
	},

	{
		ipv6only,
		[]net.IPAddr{
			{IP: net.IPv4(127, 0, 0, 1)},
			{IP: net.IPv6loopback},
		},
		testInetaddr,
		&SRTAddr{IP: net.IPv6loopback, Port: 5682},
		addrList{&SRTAddr{IP: net.IPv6loopback, Port: 5682}},
		nil,
		nil,
	},
	{
		ipv6only,
		[]net.IPAddr{
			{IP: net.IPv6loopback},
			{IP: net.IPv4(127, 0, 0, 1)},
		},
		testInetaddr,
		&SRTAddr{IP: net.IPv6loopback, Port: 5682},
		addrList{&SRTAddr{IP: net.IPv6loopback, Port: 5682}},
		nil,
		nil,
	},

	{nil, nil, testInetaddr, nil, nil, nil, &net.AddrError{errNoSuitableAddress.Error(), "ADDR"}},

	{ipv4only, nil, testInetaddr, nil, nil, nil, &net.AddrError{errNoSuitableAddress.Error(), "ADDR"}},
	{ipv4only, []net.IPAddr{{IP: net.IPv6loopback}}, testInetaddr, nil, nil, nil, &net.AddrError{errNoSuitableAddress.Error(), "ADDR"}},

	{ipv6only, nil, testInetaddr, nil, nil, nil, &net.AddrError{errNoSuitableAddress.Error(), "ADDR"}},
	{ipv6only, []net.IPAddr{{IP: net.IPv4(127, 0, 0, 1)}}, testInetaddr, nil, nil, nil, &net.AddrError{errNoSuitableAddress.Error(), "ADDR"}},
}

func TestAddrList(t *testing.T) {
	if !supportsIPv4() || !supportsIPv6() {
		t.Skip("both IPv4 and IPv6 are required")
	}

	for i, tt := range addrListTests {
		addrs, err := filterAddrList(tt.filter, tt.ips, tt.inetaddr, "ADDR")
		if !reflect.DeepEqual(err, tt.err) {
			t.Errorf("#%v: got %v; want %v", i, err, tt.err)
		}
		if tt.err != nil {
			if len(addrs) != 0 {
				t.Errorf("#%v: got %v; want 0", i, len(addrs))
			}
			continue
		}
		first := addrs.first(isIPv4)
		if !reflect.DeepEqual(first, tt.first) {
			t.Errorf("#%v: got %v; want %v", i, first, tt.first)
		}
		primaries, fallbacks := addrs.partition(isIPv4)
		if !reflect.DeepEqual(primaries, tt.primaries) {
			t.Errorf("#%v: got %v; want %v", i, primaries, tt.primaries)
		}
		if !reflect.DeepEqual(fallbacks, tt.fallbacks) {
			t.Errorf("#%v: got %v; want %v", i, fallbacks, tt.fallbacks)
		}
		expectedLen := len(primaries) + len(fallbacks)
		if len(addrs) != expectedLen {
			t.Errorf("#%v: got %v; want %v", i, len(addrs), expectedLen)
		}
	}
}

func TestAddrListPartition(t *testing.T) {
	addrs := addrList{
		&net.IPAddr{IP: net.ParseIP("fe80::"), Zone: "eth0"},
		&net.IPAddr{IP: net.ParseIP("fe80::1"), Zone: "eth0"},
		&net.IPAddr{IP: net.ParseIP("fe80::2"), Zone: "eth0"},
	}
	cases := []struct {
		lastByte  byte
		primaries addrList
		fallbacks addrList
	}{
		{0, addrList{addrs[0]}, addrList{addrs[1], addrs[2]}},
		{1, addrList{addrs[0], addrs[2]}, addrList{addrs[1]}},
		{2, addrList{addrs[0], addrs[1]}, addrList{addrs[2]}},
		{3, addrList{addrs[0], addrs[1], addrs[2]}, nil},
	}
	for i, tt := range cases {
		// Inverting the function's output should not affect the outcome.
		for _, invert := range []bool{false, true} {
			primaries, fallbacks := addrs.partition(func(a net.Addr) bool {
				ip := a.(*net.IPAddr).IP
				return (ip[len(ip)-1] == tt.lastByte) != invert
			})
			if !reflect.DeepEqual(primaries, tt.primaries) {
				t.Errorf("#%v: got %v; want %v", i, primaries, tt.primaries)
			}
			if !reflect.DeepEqual(fallbacks, tt.fallbacks) {
				t.Errorf("#%v: got %v; want %v", i, fallbacks, tt.fallbacks)
			}
		}
	}
}
