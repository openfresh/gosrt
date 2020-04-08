// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package srt

import (
	"context"
	"net"
	"os"
	"syscall"

	"github.com/openfresh/gosrt/internal/poll"
	"github.com/openfresh/gosrt/srtapi"
)

// A sockaddr represents a SRT network endpoint
// address that can be converted into a syscall.Sockaddr.
type sockaddr interface {
	net.Addr

	// family returns the platform-dependent address family
	// identifier.
	family() int

	// isWildcard reports whether the address is a wildcard
	// address.
	isWildcard() bool

	// sockaddr returns the address converted into a syscall
	// sockaddr type that implements syscall.Sockaddr
	// interface. It returns a nil interface when the address is
	// nil.
	sockaddr(family int) (syscall.Sockaddr, error)

	// toLocal maps the zero address to a local system address (127.0.0.1 or ::1)
	toLocal(net string) sockaddr
}

// socket returns a network file descriptor
func socket(ctx context.Context, net string, family, sotype, proto int, ipv6only bool, laddr, raddr sockaddr) (fd *netFD, err error) {
	s, err := srtSocket(family, sotype, proto)
	if err != nil {
		return nil, err
	}
	if err = setDefaultSockopts(s, family, sotype, ipv6only); err != nil {
		poll.CloseFunc(s)
		return nil, err
	}
	configure(ctx, s, bindPre)
	if fd, err = newFD(s, family, sotype, net); err != nil {
		poll.CloseFunc(s)
		return nil, err
	}

	if laddr != nil && raddr == nil {
		if err := fd.listen(laddr, listenerBacklog); err != nil {
			fd.Close()
			return nil, err
		}
		if callback := listenCallbackValue(ctx); callback != nil {
			if err := fd.listenCallback(callback); err != nil {
				fd.Close()
				return nil, err
			}
		}
		return fd, nil
	}
	if err := fd.dial(ctx, laddr, raddr); err != nil {
		fd.Close()
		return nil, err
	}
	return fd, nil
}

func (fd *netFD) addrFunc() func(syscall.Sockaddr) net.Addr {
	switch fd.family {
	case syscall.AF_INET, syscall.AF_INET6:
		return sockaddrToSRT
	}
	return func(syscall.Sockaddr) net.Addr { return nil }
}

func (fd *netFD) dial(ctx context.Context, laddr, raddr sockaddr) error {
	var err error
	var lsa syscall.Sockaddr
	if laddr != nil {
		if lsa, err = laddr.sockaddr(fd.family); err != nil {
			return err
		} else if lsa != nil {
			if err := srtapi.Bind(fd.pfd.Sysfd, lsa); err != nil {
				return os.NewSyscallError("bind", err)
			}
		}
	}
	var rsa syscall.Sockaddr  // remote address from the user
	var crsa syscall.Sockaddr // remote address we actually connected to
	if raddr != nil {
		if rsa, err = raddr.sockaddr(fd.family); err != nil {
			return err
		}
		if crsa, err = fd.connect(ctx, lsa, rsa); err != nil {
			return err
		}
		fd.isConnected = true
		configure(ctx, fd.pfd.Sysfd, bindPost)
	} else {
		if err := fd.init(); err != nil {
			return err
		}
	}
	// Record the local and remote addresses from the actual socket.
	// Get the local address by calling Getsockname.
	// For the remote address, use
	// 1) the one returned by the connect method, if any; or
	// 2) the one from Getpeername, if it succeeds; or
	// 3) the one passed to us as the raddr parameter.
	lsa, _ = srtapi.Getsockname(fd.pfd.Sysfd)

	// hack - if it could bit get zone ID, it forces to set 1
	switch sa := lsa.(type) {
	case *syscall.SockaddrInet6:
		var IP net.IP = sa.Addr[0:]
		if IP.IsLinkLocalUnicast() && sa.ZoneId == 0 {
			sa.ZoneId = 1
		}
	}

	if crsa != nil {
		fd.setAddr(fd.addrFunc()(lsa), fd.addrFunc()(crsa))
	} else if rsa, _ = srtapi.Getpeername(fd.pfd.Sysfd); rsa != nil {
		fd.setAddr(fd.addrFunc()(lsa), fd.addrFunc()(rsa))
	} else {
		fd.setAddr(fd.addrFunc()(lsa), raddr)
	}
	return nil
}

func (fd *netFD) listen(laddr sockaddr, backlog int) error {
	if err := setDefaultListenerSockopts(fd.pfd.Sysfd); err != nil {
		return err
	}
	if lsa, err := laddr.sockaddr(fd.family); err != nil {
		return err
	} else if lsa != nil {
		if err := srtapi.Bind(fd.pfd.Sysfd, lsa); err != nil {
			return os.NewSyscallError("bind", err)
		}
	}
	if err := listenFunc(fd.pfd.Sysfd, backlog); err != nil {
		return os.NewSyscallError("listen", err)
	}
	if err := fd.init(); err != nil {
		return err
	}
	lsa, _ := srtapi.Getsockname(fd.pfd.Sysfd)
	fd.setAddr(fd.addrFunc()(lsa), nil)
	return nil
}

func (fd *netFD) listenCallback(callback srtapi.SrtListenCallbackFunc) error {
	return srtapi.ListenCallback(fd.pfd.Sysfd, callback)
}
