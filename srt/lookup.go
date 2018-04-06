// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package srt

import (
	"context"
	"net"
)

// DefaultResolver is the resolver used by the package-level Lookup
// functions and by Dialers without a specified Resolver.
var DefaultResolver = &Resolver{nr: net.DefaultResolver}

// A Resolver looks up names and numbers.
//
// A nil *Resolver is equivalent to a zero Resolver.
type Resolver struct {
	nr *net.Resolver
}

// LookupIPAddr looks up host using the local resolver.
// It returns a slice of that host's IPv4 and IPv6 addresses.
func (r *Resolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	return r.nr.LookupIPAddr(ctx, host)
}

// LookupPort looks up the port for the given network and service.
func (r *Resolver) LookupPort(ctx context.Context, network, service string) (port int, err error) {
	port, needsLookup := parsePort(service)
	if needsLookup {
		return 0, &net.AddrError{Err: "unknown port", Addr: network + "/" + service}
	}
	if 0 > port || port > 65535 {
		return 0, &net.AddrError{Err: "invalid port", Addr: service}
	}
	return port, nil
}
