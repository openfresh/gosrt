// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

package srtapi

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"
import (
	"syscall"
	"unsafe"
)

func accept(s int, rsa *syscall.RawSockaddrAny, addrlen *_Socklen) (fd int, err error) {
	fd = int(C.srt_accept(C.SRTSOCKET(s), (*C.struct_sockaddr)(unsafe.Pointer(rsa)), (*C.int)(addrlen)))
	if fd == APIError {
		err = Errno(C.srt_getlasterror(nil))
	}
	return
}

func getsockname(s int, rsa *syscall.RawSockaddrAny, addrlen *_Socklen) (err error) {
	stat := C.srt_getsockname(C.SRTSOCKET(s), (*C.struct_sockaddr)(unsafe.Pointer(rsa)), (*C.int)(addrlen))
	if stat == APIError {
		err = Errno(C.srt_getlasterror(nil))
	}
	return
}

func getpeername(s int, rsa *syscall.RawSockaddrAny, addrlen *_Socklen) (err error) {
	stat := C.srt_getpeername(C.SRTSOCKET(s), (*C.struct_sockaddr)(unsafe.Pointer(rsa)), (*C.int)(addrlen))
	if stat == APIError {
		err = Errno(C.srt_getlasterror(nil))
	}
	return
}

func bind(s int, addr unsafe.Pointer, addrlen _Socklen) (err error) {
	stat := C.srt_bind(C.SRTSOCKET(s), (*C.struct_sockaddr)(unsafe.Pointer(addr)), C.int(addrlen))
	if stat == APIError {
		err = Errno(C.srt_getlasterror(nil))
	}
	return
}

func connect(s int, addr unsafe.Pointer, addrlen _Socklen) (err error) {
	stat := C.srt_connect(C.SRTSOCKET(s), (*C.struct_sockaddr)(unsafe.Pointer(addr)), C.int(addrlen))
	if stat == APIError {
		err = Errno(C.srt_getlasterror(nil))
	}
	return
}

func socket(domain int, typ int, proto int) (fd int, err error) {
	fd = int(C.srt_socket(C.int(domain), C.int(typ), C.int(proto)))
	if fd == APIError {
		err = Errno(C.srt_getlasterror(nil))
	}
	return
}

func getsockopt(s int, level int, name int, val unsafe.Pointer, vallen *_Socklen) (err error) {
	stat := C.srt_getsockopt(C.SRTSOCKET(s), C.int(level), C.SRT_SOCKOPT(name), val, (*C.int)(vallen))
	if stat == APIError {
		err = Errno(C.srt_getlasterror(nil))
	}
	return
}

func setsockopt(s int, level int, name int, val unsafe.Pointer, vallen uintptr) (err error) {
	stat := C.srt_setsockopt(C.SRTSOCKET(s), C.int(level), C.SRT_SOCKOPT(name), val, C.int(vallen))
	if stat == APIError {
		err = Errno(C.srt_getlasterror(nil))
	}
	return
}

// Listen call srt_listen
func Listen(s int, n int) (err error) {
	stat := C.srt_listen(C.SRTSOCKET(s), C.int(n))
	if stat == APIError {
		err = Errno(C.srt_getlasterror(nil))
	}
	return
}

// Close call srt_close
func Close(fd int) (err error) {
	stat := C.srt_close(C.SRTSOCKET(fd))
	if stat == APIError {
		err = Errno(C.srt_getlasterror(nil))
	}
	return
}

func read(fd int, p []byte) (n int, err error) {
	var _p0 unsafe.Pointer
	if len(p) > 0 {
		_p0 = unsafe.Pointer(&p[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	r0 := C.srt_recv(C.SRTSOCKET(fd), (*C.char)(_p0), C.int(len(p)))
	n = int(r0)
	if r0 == APIError {
		err = Errno(C.srt_getlasterror(nil))
	}
	return
}

func write(fd int, p []byte) (n int, err error) {
	var _p0 unsafe.Pointer
	if len(p) > 0 {
		_p0 = unsafe.Pointer(&p[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	r0 := C.srt_send(C.SRTSOCKET(fd), (*C.char)(_p0), C.int(len(p)))
	n = int(r0)
	if r0 == APIError {
		err = Errno(C.srt_getlasterror(nil))
	}
	return
}

func getlasterror() int {
	return int(C.srt_getlasterror(nil))
}

func strerror(code int, errnoval int) string {
	return C.GoString(C.srt_strerror(C.int(code), C.int(errnoval)))
}
