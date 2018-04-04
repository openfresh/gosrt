// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrt

import (
	"net"
	"strings"

	"github.com/openfresh/gosrt/internal/testenv"
)

// testableNetwork reports whether network is testable on the current
// platform configuration.
func testableNetwork(network string) bool {
	ss := strings.Split(network, ":")
	switch ss[0] {
	case "srt4":
		if !supportsIPv4() {
			return false
		}
	case "srt6":
		if !supportsIPv6() {
			return false
		}
	}
	return true
}

// testableAddress reports whether address of network is testable on
// the current platform configuration.
func testableAddress(network, address string) bool {
	return true
}

// testableListenArgs reports whether arguments are testable on the
// current platform configuration.
func testableListenArgs(network, address, client string) bool {
	if !testableNetwork(network) || !testableAddress(network, address) {
		return false
	}

	var err error
	var addr net.Addr
	switch ss := strings.Split(network, ":"); ss[0] {
	case "srt", "srt4", "srt6":
		addr, err = ResolveSRTAddr("srt", address)
	default:
		return true
	}
	if err != nil {
		return false
	}
	var ip net.IP
	var wildcard bool
	switch addr := addr.(type) {
	case *SRTAddr:
		ip = addr.IP
		wildcard = addr.isWildcard()
	}

	// Test wildcard IP addresses.
	if wildcard && !testenv.HasExternalNetwork() {
		return false
	}

	// Test functionality of IPv4 communication using AF_INET and
	// IPv6 communication using AF_INET6 sockets.
	if !supportsIPv4() && ip.To4() != nil {
		return false
	}
	if !supportsIPv6() && ip.To16() != nil && ip.To4() == nil {
		return false
	}
	cip := net.ParseIP(client)
	if cip != nil {
		if !supportsIPv4() && cip.To4() != nil {
			return false
		}
		if !supportsIPv6() && cip.To16() != nil && cip.To4() == nil {
			return false
		}
	}

	// Test functionality of IPv4 communication using AF_INET6
	// sockets.
	if !supportsIPv4map() && supportsIPv4() && (network == "srt") && wildcard {
		// At this point, we prefer IPv4 when ip is nil.
		// See favoriteAddrFamily for further information.
		if ip.To16() != nil && ip.To4() == nil && cip.To4() != nil { // a pair of IPv6 server and IPv4 client
			return false
		}
		if (ip.To4() != nil || ip == nil) && cip.To16() != nil && cip.To4() == nil { // a pair of IPv4 server and IPv6 client
			return false
		}
	}

	return true
}
