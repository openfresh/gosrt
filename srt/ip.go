// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

// IP address manipulations
//
// IPv4 addresses are 4 bytes; IPv6 addresses are 16 bytes.
// An IPv4 address can be converted to an IPv6 address by
// adding a canonical prefix (10 zeros, 2 0xFFs).
// This library accepts either size of byte slice but always
// returns 16-byte addresses.

package srt

import (
	"net"
)

// ipEmptyString is like ip.String except that it returns
// an empty string when ip is unset.
func ipEmptyString(ip net.IP) string {
	if len(ip) == 0 {
		return ""
	}
	return ip.String()
}

// Parse IPv4 address (d.d.d.d).
func parseIPv4(s string) net.IP {
	var p [net.IPv4len]byte
	for i := 0; i < net.IPv4len; i++ {
		if len(s) == 0 {
			// Missing octets.
			return nil
		}
		if i > 0 {
			if s[0] != '.' {
				return nil
			}
			s = s[1:]
		}
		n, c, ok := dtoi(s)
		if !ok || n > 0xFF {
			return nil
		}
		s = s[c:]
		p[i] = byte(n)
	}
	if len(s) != 0 {
		return nil
	}
	return net.IPv4(p[0], p[1], p[2], p[3])
}

// parseIPv6 parses s as a literal IPv6 address described in RFC 4291
// and RFC 5952.  It can also parse a literal scoped IPv6 address with
// zone identifier which is described in RFC 4007 when zoneAllowed is
// true.
func parseIPv6(s string, zoneAllowed bool) (ip net.IP, zone string) {
	ip = make(net.IP, net.IPv6len)
	ellipsis := -1 // position of ellipsis in ip

	if zoneAllowed {
		s, zone = splitHostZone(s)
	}

	// Might have leading ellipsis
	if len(s) >= 2 && s[0] == ':' && s[1] == ':' {
		ellipsis = 0
		s = s[2:]
		// Might be only ellipsis
		if len(s) == 0 {
			return ip, zone
		}
	}

	// Loop, parsing hex numbers followed by colon.
	i := 0
	for i < net.IPv6len {
		// Hex number.
		n, c, ok := xtoi(s)
		if !ok || n > 0xFFFF {
			return nil, zone
		}

		// If followed by dot, might be in trailing IPv4.
		if c < len(s) && s[c] == '.' {
			if ellipsis < 0 && i != net.IPv6len-net.IPv4len {
				// Not the right place.
				return nil, zone
			}
			if i+net.IPv4len > net.IPv6len {
				// Not enough room.
				return nil, zone
			}
			ip4 := parseIPv4(s)
			if ip4 == nil {
				return nil, zone
			}
			ip[i] = ip4[12]
			ip[i+1] = ip4[13]
			ip[i+2] = ip4[14]
			ip[i+3] = ip4[15]
			s = ""
			i += net.IPv4len
			break
		}

		// Save this 16-bit chunk.
		ip[i] = byte(n >> 8)
		ip[i+1] = byte(n)
		i += 2

		// Stop at end of string.
		s = s[c:]
		if len(s) == 0 {
			break
		}

		// Otherwise must be followed by colon and more.
		if s[0] != ':' || len(s) == 1 {
			return nil, zone
		}
		s = s[1:]

		// Look for ellipsis.
		if s[0] == ':' {
			if ellipsis >= 0 { // already have one
				return nil, zone
			}
			ellipsis = i
			s = s[1:]
			if len(s) == 0 { // can be at end
				break
			}
		}
	}

	// Must have used entire string.
	if len(s) != 0 {
		return nil, zone
	}

	// If didn't parse enough, expand ellipsis.
	if i < net.IPv6len {
		if ellipsis < 0 {
			return nil, zone
		}
		n := net.IPv6len - i
		for j := i - 1; j >= ellipsis; j-- {
			ip[j+n] = ip[j]
		}
		for j := ellipsis + n - 1; j >= ellipsis; j-- {
			ip[j] = 0
		}
	} else if ellipsis >= 0 {
		// Ellipsis must represent at least one 0 group.
		return nil, zone
	}
	return ip, zone
}
