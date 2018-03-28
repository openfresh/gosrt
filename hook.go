package gosrt

import "context"

var (
	// if non-nil, overrides dialTCP.
	testHookDialSRT func(ctx context.Context, net string, laddr, raddr *SRTAddr) (*SRTConn, error)
)
