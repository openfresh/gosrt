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
