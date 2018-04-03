// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package poll

import (
	"errors"
	"sync"
	"syscall"
	"time"

	"github.com/openfresh/gosrt/internal/poll/runtime"
)

type pollDesc struct {
	runtimeCtx runtime.PollDesc
}

var serverInit sync.Once

func (pd *pollDesc) init(fd *FD) error {
	serverInit.Do(runtime.PollServerInit)
	ctx, errno := runtime.PollOpen(fd.Sysfd)
	if errno != 0 {
		if ctx != nil {
			ctx.Unblock()
			ctx.Close()
		}
		return syscall.Errno(errno)
	}
	pd.runtimeCtx = ctx
	return nil
}

func (pd *pollDesc) close() {
	if pd.runtimeCtx == nil {
		return
	}
	pd.runtimeCtx.Close()
	pd.runtimeCtx = nil
}

// Evict evicts fd from the pending list, unblocking any I/O running on fd.
func (pd *pollDesc) evict() {
	if pd.runtimeCtx == nil {
		return
	}
	pd.runtimeCtx.Unblock()
}

func (pd *pollDesc) prepare(mode int) error {
	if pd.runtimeCtx == nil {
		return nil
	}
	res := pd.runtimeCtx.Reset(mode)
	return convertErr(res)
}

func (pd *pollDesc) prepareRead() error {
	return pd.prepare('r')
}

func (pd *pollDesc) prepareWrite() error {
	return pd.prepare('w')
}

func (pd *pollDesc) wait(mode int) error {
	if pd.runtimeCtx == nil {
		return errors.New("waiting for unsupported file type")
	}
	res := pd.runtimeCtx.Wait(mode)
	return convertErr(res)
}

func (pd *pollDesc) waitRead() error {
	return pd.wait('r')
}

func (pd *pollDesc) waitWrite() error {
	return pd.wait('w')
}

func (pd *pollDesc) pollable() bool {
	return pd.runtimeCtx != nil
}

func convertErr(res int) error {
	switch res {
	case 0:
		return nil
	case 1:
		return errClosing()
	case 2:
		return ErrTimeout
	}
	println("unreachable: ", res)
	panic("unreachable")
}

// SetDeadline sets the read and write deadlines associated with fd.
func (fd *FD) SetDeadline(t time.Time) error {
	return setDeadlineImpl(fd, t, 'r'+'w')
}

// SetReadDeadline sets the read deadline associated with fd.
func (fd *FD) SetReadDeadline(t time.Time) error {
	return setDeadlineImpl(fd, t, 'r')
}

// SetWriteDeadline sets the write deadline associated with fd.
func (fd *FD) SetWriteDeadline(t time.Time) error {
	return setDeadlineImpl(fd, t, 'w')
}

func setDeadlineImpl(fd *FD, t time.Time, mode int) error {
	d := time.Until(t)
	if fd.pd.runtimeCtx == nil {
		return ErrNoDeadline
	}
	fd.pd.runtimeCtx.SetDeadline(d, mode)
	return nil
}

// Descriptor returns the descriptor being used by the poller,
// or ^uintptr(0) if there isn't one. This is only used for testing.
func Descriptor() int {
	return runtime.PollServerDescriptor()
}
