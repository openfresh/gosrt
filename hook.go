package gosrt

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
