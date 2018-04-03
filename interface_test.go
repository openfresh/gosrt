// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package gosrt

import "net"

// loopbackInterface returns an available logical network interface
// for loopback tests. It returns nil if no suitable interface is
// found.
func loopbackInterface() *net.Interface {
	ift, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, ifi := range ift {
		if ifi.Flags&net.FlagLoopback != 0 && ifi.Flags&net.FlagUp != 0 {
			return &ifi
		}
	}
	return nil
}

// ipv6LinkLocalUnicastAddr returns an IPv6 link-local unicast address
// on the given network interface for tests. It returns "" if no
// suitable address is found.
func ipv6LinkLocalUnicastAddr(ifi *net.Interface) string {
	if ifi == nil {
		return ""
	}
	ifat, err := ifi.Addrs()
	if err != nil {
		return ""
	}
	for _, ifa := range ifat {
		if ifa, ok := ifa.(*net.IPNet); ok {
			if ifa.IP.To4() == nil && ifa.IP.IsLinkLocalUnicast() {
				return ifa.IP.String()
			}
		}
	}
	return ""
}
