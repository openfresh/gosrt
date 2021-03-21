// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package srtapi

import (
	"io"
	"syscall"
	"unsafe"
)

// An Errno is an number describing an error condition.
type Errno int

func (e Errno) Error() string {
	return strerror(int(e), 0)
}

// Temporary return if it is temprary error
func (e Errno) Temporary() bool {
	return e.Timeout()
}

// Timeout return if it is timeout error
func (e Errno) Timeout() bool {
	return e == EASYNCFAIL || e == EASYNCSND || e == EASYNCRCV || e == ETIMEOUT || e == ECONGEST
}

// Read call srt_recv
func Read(fd int, p []byte) (n int, err error) {
	n, err = read(fd, p)
	return
}

// Write call srt_send
func Write(fd int, p []byte) (n int, err error) {
	n, err = write(fd, p)
	return
}

// Bind call srt_bind
func Bind(fd int, sa syscall.Sockaddr) (err error) {
	ptr, n, err := sockaddr(sa)
	if err != nil {
		return err
	}
	return bind(fd, ptr, n)
}

// Connect call srt_connect
func Connect(fd int, sa syscall.Sockaddr) (err error) {
	ptr, n, err := sockaddr(sa)
	if err != nil {
		return err
	}
	return connect(fd, ptr, n)
}

// Getpeername call srt_getpeername
func Getpeername(fd int) (sa syscall.Sockaddr, err error) {
	var rsa syscall.RawSockaddrAny
	var len _Socklen = SizeofSockaddrAny
	if err = getpeername(fd, &rsa, &len); err != nil {
		return
	}
	return anyToSockaddr(&rsa)
}

// GetsockoptInt call srt_getsockopt
func GetsockoptInt(fd, level, opt int) (value int, err error) {
	var n int32
	vallen := _Socklen(4)
	err = getsockopt(fd, level, opt, unsafe.Pointer(&n), &vallen)
	return int(n), err
}

// GetsockoptString returns the string value of the socket option opt for the
// socket associated with fd at the given socket level.
func GetsockoptString(fd, level, opt int) (string, error) {
	buf := make([]byte, 256)
	vallen := _Socklen(len(buf))
	err := getsockopt(fd, level, opt, unsafe.Pointer(&buf[0]), &vallen)
	if err != nil {
		return "", err
	}
	return string(buf[:vallen]), nil
}

// GetsockflagInt call srt_getsockflag
func GetsockflagInt(fd, opt int) (value int, err error) {
	var n int32
	vallen := _Socklen(4)
	err = getsockflag(fd, opt, unsafe.Pointer(&n), &vallen)
	return int(n), err
}

// GetsockflagString returns the string value of the socket flag for the
// socket associated with a fd
func GetsockflagString(fd, opt int) (string, error) {
	buf := make([]byte, 256)
	vallen := _Socklen(len(buf))
	err := getsockflag(fd, opt, unsafe.Pointer(&buf[0]), &vallen)
	if err != nil {
		return "", err
	}
	return string(buf[:vallen]), nil
}

// SetsockoptByte call srt_setsockopt
func SetsockoptByte(fd, level, opt int, value byte) (err error) {
	return setsockopt(fd, level, opt, unsafe.Pointer(&value), 1)
}

// SetsockoptInt call srt_setsockopt
func SetsockoptInt(fd, level, opt int, value int) (err error) {
	var n = int32(value)
	return setsockopt(fd, level, opt, unsafe.Pointer(&n), 4)
}

// SetsockoptInt64 call srt_setsockopt
func SetsockoptInt64(fd, level, opt int, value int64) (err error) {
	var n = value
	return setsockopt(fd, level, opt, unsafe.Pointer(&n), 8)
}

// SetsockoptString call srt_setsockopt
func SetsockoptString(fd, level, opt int, s string) (err error) {
	return setsockopt(fd, level, opt, unsafe.Pointer(&[]byte(s)[0]), uintptr(len(s)))
}

// SetsockoptBool call srt_setsockopt
func SetsockoptBool(fd, level, opt int, value bool) (err error) {
	var n = int32(0)
	if value {
		n = 1
	}
	return setsockopt(fd, level, opt, unsafe.Pointer(&n), 4)
}

// SetsockflagByte call srt_setsockopt
func SetsockflagByte(fd, opt int, value byte) (err error) {
	return setsockflag(fd, opt, unsafe.Pointer(&value), 1)
}

// SetsockflagInt call srt_setsockopt
func SetsockflagInt(fd, opt int, value int) (err error) {
	var n = int32(value)
	return setsockflag(fd, opt, unsafe.Pointer(&n), 4)
}

// SetsockflagInt64 call srt_setsockopt
func SetsockflagInt64(fd, opt int, value int64) (err error) {
	var n = value
	return setsockflag(fd, opt, unsafe.Pointer(&n), 8)
}

// SetsockflagString call srt_setsockopt
func SetsockflagString(fd, opt int, s string) (err error) {
	return setsockflag(fd, opt, unsafe.Pointer(&[]byte(s)[0]), uintptr(len(s)))
}

// SetsockflagBool call srt_setsockopt
func SetsockflagBool(fd, opt int, value bool) (err error) {
	var n = int32(0)
	if value {
		n = 1
	}
	return setsockflag(fd, opt, unsafe.Pointer(&n), 4)
}

// Socket call srt_socket
func Socket() (fd int, err error) {
	fd, err = socket()
	return
}

func Sendfile(outfd int, r io.Reader, offset *int64, count int) (written int, err error) {
	return sendfile(outfd, r, offset, count)
}

// Accept call srt_accept
func Accept(fd int) (nfd int, sa syscall.Sockaddr, err error) {
	var rsa syscall.RawSockaddrAny
	var len _Socklen = SizeofSockaddrAny
	nfd, err = accept(fd, &rsa, &len)
	if err != nil {
		return
	}
	sa, err = anyToSockaddr(&rsa)
	if err != nil {
		Close(nfd)
		nfd = 0
	}
	return
}

// Getsockname call srt_getsockname
func Getsockname(fd int) (sa syscall.Sockaddr, err error) {
	var rsa syscall.RawSockaddrAny
	var len _Socklen = SizeofSockaddrAny
	if err = getsockname(fd, &rsa, &len); err != nil {
		return
	}
	return anyToSockaddr(&rsa)
}

func sockaddr(sa syscall.Sockaddr) (unsafe.Pointer, _Socklen, error) {
	if sa == nil {
		return nil, 0, EINVPARAM
	}
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		return sockaddrInet4((*syscall.SockaddrInet4)(sa))
	case *syscall.SockaddrInet6:
		return sockaddrInet6((*syscall.SockaddrInet6)(sa))
	}
	return nil, 0, syscall.EAFNOSUPPORT
}

func sockaddrInet4(sa *syscall.SockaddrInet4) (unsafe.Pointer, _Socklen, error) {
	if sa.Port < 0 || sa.Port > 0xFFFF {
		return nil, 0, EINVPARAM
	}
	var raw syscall.RawSockaddrInet4
	raw.Family = syscall.AF_INET
	p := (*[2]byte)(unsafe.Pointer(&raw.Port))
	p[0] = byte(sa.Port >> 8)
	p[1] = byte(sa.Port)
	for i := 0; i < len(sa.Addr); i++ {
		raw.Addr[i] = sa.Addr[i]
	}
	return unsafe.Pointer(&raw), SizeofSockaddrInet4, nil
}

func sockaddrInet6(sa *syscall.SockaddrInet6) (unsafe.Pointer, _Socklen, error) {
	if sa.Port < 0 || sa.Port > 0xFFFF {
		return nil, 0, EINVPARAM
	}
	var raw syscall.RawSockaddrInet6
	raw.Family = syscall.AF_INET6
	p := (*[2]byte)(unsafe.Pointer(&raw.Port))
	p[0] = byte(sa.Port >> 8)
	p[1] = byte(sa.Port)
	raw.Scope_id = sa.ZoneId
	for i := 0; i < len(sa.Addr); i++ {
		raw.Addr[i] = sa.Addr[i]
	}
	return unsafe.Pointer(&raw), SizeofSockaddrInet6, nil
}

func anyToSockaddr(rsa *syscall.RawSockaddrAny) (syscall.Sockaddr, error) {
	switch rsa.Addr.Family {
	case syscall.AF_INET:
		pp := (*syscall.RawSockaddrInet4)(unsafe.Pointer(rsa))
		sa := new(syscall.SockaddrInet4)
		p := (*[2]byte)(unsafe.Pointer(&pp.Port))
		sa.Port = int(p[0])<<8 + int(p[1])
		for i := 0; i < len(sa.Addr); i++ {
			sa.Addr[i] = pp.Addr[i]
		}
		return sa, nil

	case syscall.AF_INET6:
		pp := (*syscall.RawSockaddrInet6)(unsafe.Pointer(rsa))
		sa := new(syscall.SockaddrInet6)
		p := (*[2]byte)(unsafe.Pointer(&pp.Port))
		sa.Port = int(p[0])<<8 + int(p[1])
		sa.ZoneId = pp.Scope_id
		for i := 0; i < len(sa.Addr); i++ {
			sa.Addr[i] = pp.Addr[i]
		}
		return sa, nil
	}
	return nil, syscall.EAFNOSUPPORT
}

func getLastError() error {
	return Errno(getlasterror())
}
