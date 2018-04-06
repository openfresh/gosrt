// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package srt

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"syscall"

	"github.com/openfresh/gosrt/internal/poll"
	"github.com/openfresh/gosrt/srtapi"
)

// Network file descriptor.
type netFD struct {
	pfd poll.FD

	// immutable until Close
	family      int
	sotype      int
	isConnected bool
	net         string
	laddr       net.Addr
	raddr       net.Addr
}

func newFD(sysfd, family, sotype int, net string) (*netFD, error) {
	ret := &netFD{
		pfd: poll.FD{
			Sysfd: sysfd,
		},
		family: family,
		sotype: sotype,
		net:    net,
	}
	return ret, nil
}

func (fd *netFD) init() error {
	return fd.pfd.Init(fd.net, true)
}

func (fd *netFD) setAddr(laddr, raddr net.Addr) {
	fd.laddr = laddr
	fd.raddr = raddr
	runtime.SetFinalizer(fd, (*netFD).Close)
}

func (fd *netFD) name() string {
	var ls, rs string
	if fd.laddr != nil {
		ls = fd.laddr.String()
	}
	if fd.raddr != nil {
		rs = fd.raddr.String()
	}
	return fd.net + ":" + ls + "->" + rs
}

func (fd *netFD) connect(ctx context.Context, la, ra syscall.Sockaddr) (rsa syscall.Sockaddr, ret error) {
	if err := connectFunc(fd.pfd.Sysfd, ra); err != nil {
		return nil, os.NewSyscallError("connect", err)
	}
	state, err := getsockoptIntFunc(fd.pfd.Sysfd, 0, srtapi.OptionState)
	if err != nil {
		return nil, os.NewSyscallError("getsockopt", err)
	}
	switch state {
	case srtapi.StatusConnecting:
	case srtapi.StatusConnected:
		return nil, nil
	default:
		return nil, fmt.Errorf("unexpected socket state %d", state)
	}
	if err := fd.pfd.Init(fd.net, true); err != nil {
		return nil, err
	}
	if deadline, _ := ctx.Deadline(); !deadline.IsZero() {
		fd.pfd.SetWriteDeadline(deadline)
		defer fd.pfd.SetWriteDeadline(noDeadline)
	}

	// Start the "interrupter" goroutine, if this context might be canceled.
	// (The background context cannot)
	//
	// The interrupter goroutine waits for the context to be done and
	// interrupts the dial (by altering the fd's write deadline, which
	// wakes up waitWrite).
	if ctx != context.Background() {
		// Wait for the interrupter goroutine to exit before returning
		// from connect.
		done := make(chan struct{})
		interruptRes := make(chan error)
		defer func() {
			close(done)
			if ctxErr := <-interruptRes; ctxErr != nil && ret == nil {
				// The interrupter goroutine called SetWriteDeadline,
				// but the connect code below had returned from
				// waitWrite already and did a successful connect (ret
				// == nil). Because we've now poisoned the connection
				// by making it unwritable, don't return a successful
				// dial. This was issue 16523.
				ret = ctxErr
				fd.Close() // prevent a leak
			}
		}()
		go func() {
			select {
			case <-ctx.Done():
				// Force the runtime's poller to immediately give up
				// waiting for writability, unblocking waitWrite
				// below.
				fd.pfd.SetWriteDeadline(aLongTimeAgo)
				testHookCanceledDial()
				interruptRes <- ctx.Err()
			case <-done:
				interruptRes <- nil
			}
		}()
	}

	for {
		if err := fd.pfd.WaitWrite(); err != nil {
			select {
			case <-ctx.Done():
				return nil, mapErr(ctx.Err())
			default:
			}
			return nil, err
		}
		state, err := getsockoptIntFunc(fd.pfd.Sysfd, 0, srtapi.OptionState)
		if err != nil {
			return nil, os.NewSyscallError("getsockopt", err)
		}
		switch state {
		case srtapi.StatusConnecting:
		case srtapi.StatusConnected:
			return nil, nil
		default:
			return nil, fmt.Errorf("unexpected socket state %d", state)
		}
	}
}

func (fd *netFD) Close() error {
	runtime.SetFinalizer(fd, nil)
	return fd.pfd.Close()
}

func (fd *netFD) Read(p []byte) (n int, err error) {
	n, err = fd.pfd.Read(p)
	return n, wrapSyscallError("read", err)
}

func (fd *netFD) Write(p []byte) (nn int, err error) {
	nn, err = fd.pfd.Write(p)
	return nn, wrapSyscallError("write", err)
}

func (fd *netFD) accept() (netfd *netFD, err error) {
	d, rsa, errcall, err := fd.pfd.Accept()
	if err != nil {
		if errcall != "" {
			err = wrapSyscallError(errcall, err)
		}
		return nil, err
	}

	if netfd, err = newFD(d, fd.family, fd.sotype, fd.net); err != nil {
		poll.CloseFunc(d)
		return nil, err
	}
	if err = netfd.init(); err != nil {
		fd.Close()
		return nil, err
	}
	lsa, _ := srtapi.Getsockname(netfd.pfd.Sysfd)
	netfd.setAddr(netfd.addrFunc()(lsa), netfd.addrFunc()(rsa))
	return netfd, nil
}
