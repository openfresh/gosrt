package srtapi

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"
import (
	"syscall"
	"unsafe"
)

func accept(s int, rsa *syscall.RawSockaddrAny, addrlen *_Socklen) (fd int, err error) {
	stat := C.srt_accept(C.SRTSOCKET(s), (*C.struct_sockaddr)(unsafe.Pointer(rsa)), (*C.int)(addrlen))
	if stat == APIError {
		err = GetLastError("srt_accept")
	}
	return
}

func getsockname(s int, rsa *syscall.RawSockaddrAny, addrlen *_Socklen) (err error) {
	stat := C.srt_getsockname(C.SRTSOCKET(s), (*C.struct_sockaddr)(unsafe.Pointer(rsa)), (*C.int)(addrlen))
	if stat == APIError {
		err = GetLastError("srt_getsockname")
	}
	return
}

func getpeername(s int, rsa *syscall.RawSockaddrAny, addrlen *_Socklen) (err error) {
	stat := C.srt_getpeername(C.SRTSOCKET(s), (*C.struct_sockaddr)(unsafe.Pointer(rsa)), (*C.int)(addrlen))
	if stat == APIError {
		err = GetLastError("srt_getpeername")
	}
	return
}

func bind(s int, addr unsafe.Pointer, addrlen _Socklen) (err error) {
	stat := C.srt_bind(C.SRTSOCKET(s), (*C.struct_sockaddr)(unsafe.Pointer(addr)), C.int(addrlen))
	if stat == APIError {
		err = GetLastError("srt_bind")
	}
	return
}

func connect(s int, addr unsafe.Pointer, addrlen _Socklen) (err error) {
	stat := C.srt_connect(C.SRTSOCKET(s), (*C.struct_sockaddr)(unsafe.Pointer(addr)), C.int(addrlen))
	if stat == APIError {
		err = GetLastError("srt_connect")
	}
	return
}

func socket(domain int, typ int, proto int) (fd int, err error) {
	fd = int(C.srt_socket(C.int(domain), C.int(typ), C.int(proto)))
	if fd == APIError {
		err = GetLastError("srt_socket")
	}
	return
}

func getsockopt(s int, level int, name int, val unsafe.Pointer, vallen *_Socklen) (err error) {
	stat := C.srt_getsockopt(C.SRTSOCKET(s), C.int(level), C.SRT_SOCKOPT(name), val, (*C.int)(vallen))
	if stat == APIError {
		err = GetLastError("srt_getsockopt")
	}
	return
}

func setsockopt(s int, level int, name int, val unsafe.Pointer, vallen uintptr) (err error) {
	stat := C.srt_setsockopt(C.SRTSOCKET(s), C.int(level), C.SRT_SOCKOPT(name), val, C.int(vallen))
	if stat == APIError {
		err = GetLastError("srt_setsockopt")
	}
	return
}

// Listen call srt_listen
func Listen(s int, n int) (err error) {
	stat := C.srt_listen(C.SRTSOCKET(s), C.int(n))
	if stat == APIError {
		err = GetLastError("srt_listen")
	}
	return
}

// Close call srt_close
func Close(fd int) (err error) {
	stat := C.srt_close(C.SRTSOCKET(fd))
	if stat == APIError {
		err = GetLastError("srt_close")
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
		err = GetLastError("srt_recv")
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
		err = GetLastError("srt_send")
	}
	return
}
