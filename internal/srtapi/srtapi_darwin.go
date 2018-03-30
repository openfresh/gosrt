package srtapi

import (
	"syscall"
	"unsafe"
)

func Accept(fd int) (nfd int, sa syscall.Sockaddr, err error) {
	return 0, nil, nil
}

func Getsockname(fd int) (sa syscall.Sockaddr, err error) {
	return nil, nil
}

func sockaddr(sa syscall.Sockaddr) (unsafe.Pointer, _Socklen, error) {
	return nil, 0, nil
}

func anyToSockaddr(rsa *syscall.RawSockaddrAny) (syscall.Sockaddr, error) {
	return nil, nil
}
