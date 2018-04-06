// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package srt

import (
	"context"
	"net"
)

var (
	// if non-nil, overrides dialSRT.
	testHookDialSRT func(ctx context.Context, net string, laddr, raddr *SRTAddr) (*SRTConn, error)

	testHookLookupIP = func(
		ctx context.Context,
		fn func(context.Context, string) ([]net.IPAddr, error),
		host string,
	) ([]net.IPAddr, error) {
		return fn(ctx, host)
	}
)
