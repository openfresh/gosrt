package srtapi

import (
	"syscall"
	"unsafe"
)

func Read(fd int, p []byte) (n int, err error) {
	n, err = read(fd, p)
	return
}

func Write(fd int, p []byte) (n int, err error) {
	n, err = write(fd, p)
	return
}

func Bind(fd int, sa syscall.Sockaddr) (err error) {
	ptr, n, err := sockaddr(sa)
	if err != nil {
		return err
	}
	return bind(fd, ptr, n)
}

func Connect(fd int, sa syscall.Sockaddr) (err error) {
	ptr, n, err := sockaddr(sa)
	if err != nil {
		return err
	}
	return connect(fd, ptr, n)
}

func Getpeername(fd int) (sa syscall.Sockaddr, err error) {
	var rsa syscall.RawSockaddrAny
	var len _Socklen = SizeofSockaddrAny
	if err = getpeername(fd, &rsa, &len); err != nil {
		return
	}
	return anyToSockaddr(&rsa)
}

func GetsockoptInt(fd, level, opt int) (value int, err error) {
	var n int32
	vallen := _Socklen(4)
	err = getsockopt(fd, level, opt, unsafe.Pointer(&n), &vallen)
	return int(n), err
}

func SetsockoptByte(fd, level, opt int, value byte) (err error) {
	return setsockopt(fd, level, opt, unsafe.Pointer(&value), 1)
}

func SetsockoptInt(fd, level, opt int, value int) (err error) {
	var n = int32(value)
	return setsockopt(fd, level, opt, unsafe.Pointer(&n), 4)
}

func SetsockoptString(fd, level, opt int, s string) (err error) {
	return setsockopt(fd, level, opt, unsafe.Pointer(&[]byte(s)[0]), uintptr(len(s)))
}

func Socket(domain, typ, proto int) (fd int, err error) {
	if domain == syscall.AF_INET6 && syscall.SocketDisableIPv6 {
		return -1, syscall.EAFNOSUPPORT
	}
	fd, err = socket(domain, typ, proto)
	return
}
