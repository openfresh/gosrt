package gosrt

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"
import (
	"context"
	"net"
	"syscall"
	"unsafe"
)

func sockaddrToSRT(sa *syscall.RawSockaddrAny) net.Addr {
	switch sa.Addr.Family {
	case syscall.AF_INET:
		pp := (*syscall.RawSockaddrInet4)(unsafe.Pointer(sa))
		p := (*[2]byte)(unsafe.Pointer(&pp.Port))
		port := int(p[0])<<8 + int(p[1])
		return &SRTAddr{IP: pp.Addr[0:], Port: port}
	case syscall.AF_INET6:
		pp := (*syscall.RawSockaddrInet6)(unsafe.Pointer(sa))
		p := (*[2]byte)(unsafe.Pointer(&pp.Port))
		port := int(p[0])<<8 + int(p[1])
		ifi, err := net.InterfaceByIndex(int(pp.Scope_id))
		if err != nil {
			return nil
		}
		return &SRTAddr{IP: pp.Addr[0:], Port: port, Zone: ifi.Name}
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

func (a *SRTAddr) sockaddr(family int) (*syscall.RawSockaddrAny, error) {
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
	fd, err := srtSocket(ctx, network, laddr, raddr, syscall.SOCK_STREAM, 0, "dial")
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
	fd, err := srtSocket(ctx, network, laddr, nil, syscall.SOCK_STREAM, 0, "listen")
	if err != nil {
		return nil, err
	}
	return &SRTListener{fd}, nil
}
