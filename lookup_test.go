// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package gosrt

import (
	"context"
	"net"
)

func lookupLocalhost(ctx context.Context, fn func(context.Context, string) ([]net.IPAddr, error), host string) ([]net.IPAddr, error) {
	switch host {
	case "localhost":
		return []net.IPAddr{
			{IP: net.IPv4(127, 0, 0, 1)},
			{IP: net.IPv6loopback},
		}, nil
	default:
		return fn(ctx, host)
	}
}
