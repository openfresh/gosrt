package gosrt

import (
	"context"
	"net"
	"syscall"
)

func sockaddrToSRT(sa syscall.Sockaddr) net.Addr {
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		return &SRTAddr{IP: sa.Addr[0:], Port: sa.Port}
	case *syscall.SockaddrInet6:
		return &SRTAddr{IP: sa.Addr[0:], Port: sa.Port, Zone: zoneCache.name(int(sa.ZoneId))}
	}
	return nil
}

func (a *SRTAddr) family() int {
	if a == nil || len(a.IP) <= net.IPv4len {
		return syscall.AF_INET
	}
	if a.IP.To4() != nil {
		return syscall.AF_INET
	}
	return syscall.AF_INET6
}

func (a *SRTAddr) sockaddr(family int) (syscall.Sockaddr, error) {
	if a == nil {
		return nil, nil
	}
	return ipToSockaddr(family, a.IP, a.Port, a.Zone)
}

func (a *SRTAddr) toLocal(net string) sockaddr {
	return &SRTAddr{loopbackIP(net), a.Port, a.Zone}
}

func dialSRT(ctx context.Context, network string, laddr, raddr *SRTAddr) (*SRTConn, error) {
	if testHookDialSRT != nil {
		return testHookDialSRT(ctx, network, laddr, raddr)
	}
	return doDialSRT(ctx, network, laddr, raddr)
}

func doDialSRT(ctx context.Context, network string, laddr, raddr *SRTAddr) (*SRTConn, error) {
	fd, err := internetSocket(ctx, network, laddr, raddr, syscall.SOCK_DGRAM, 0, "dial")
	if err != nil {
		return nil, err
	}
	return newSRTConn(fd), nil
}

func (ln *SRTListener) ok() bool { return ln != nil && ln.fd != nil }

func (ln *SRTListener) accept() (*SRTConn, error) {
	fd, err := ln.fd.accept()
	if err != nil {
		return nil, err
	}
	return newSRTConn(fd), nil
}

func (ln *SRTListener) close() error {
	return ln.fd.Close()
}

func listenSRT(ctx context.Context, network string, laddr *SRTAddr) (*SRTListener, error) {
	fd, err := internetSocket(ctx, network, laddr, nil, syscall.SOCK_DGRAM, 0, "listen")
	if err != nil {
		return nil, err
	}
	return &SRTListener{fd}, nil
}
