package gosrt

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"
import (
	"context"
	"net"
	"syscall"
	"unsafe"

	"github.com/openfresh/gosrt/def"
	"github.com/openfresh/gosrt/poll"
	"github.com/openfresh/gosrt/util"
)

// A sockaddr represents a TCP, UDP, IP or Unix network endpoint
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
	sockaddr(family int) (*syscall.RawSockaddrAny, error)

	// toLocal maps the zero address to a local system address (127.0.0.1 or ::1)
	toLocal(net string) sockaddr
}

// socket returns a network file descriptor
func socket(ctx context.Context, net string, family, sotype, proto int, ipv6only bool, laddr, raddr sockaddr) (fd *netFD, err error) {
	s := int(C.srt_socket(C.int(family), C.int(sotype), C.int(proto)))
	if s == def.SRT_ERROR {
		return nil, util.GetLastError("srt_socket")
	}
	if err = setDefaultSockopts(s, family, sotype, ipv6only); err != nil {
		poll.CloseFunc(s)
		return nil, err
	}
	if fd, err = newFD(s, family, sotype, net); err != nil {
		poll.CloseFunc(s)
		return nil, err
	}

	if laddr != nil && raddr == nil {
		if err := fd.listen(laddr, listenerBacklog); err != nil {
			fd.Close()
			return nil, err
		}
		return fd, nil
	}
	if err := fd.dial(ctx, laddr, raddr); err != nil {
		fd.Close()
		return nil, err
	}
	return fd, nil
}

func (fd *netFD) addrFunc() func(*syscall.RawSockaddrAny) net.Addr {
	switch fd.family {
	case syscall.AF_INET, syscall.AF_INET6:
		return sockaddrToSRT
	}
	return func(*syscall.RawSockaddrAny) net.Addr { return nil }
}

func (fd *netFD) dial(ctx context.Context, laddr, raddr sockaddr) error {
	var err error
	var lsa *syscall.RawSockaddrAny
	if laddr != nil {
		if lsa, err = laddr.sockaddr(fd.family); err != nil {
			return err
		} else if lsa != nil {
			stat := C.srt_bind(C.SRTSOCKET(fd.pfd.Sysfd), (*C.struct_sockaddr)(unsafe.Pointer(lsa)), C.int(unsafe.Sizeof(*lsa)))
			if stat == def.SRT_ERROR {
				return util.GetLastError("srt_bind")
			}
		}
	}
	var rsa *syscall.RawSockaddrAny  // remote address from the user
	var crsa *syscall.RawSockaddrAny // remote address we actually connected to
	if raddr != nil {
		if rsa, err = raddr.sockaddr(fd.family); err != nil {
			return err
		}
		if crsa, err = fd.connect(ctx, lsa, rsa); err != nil {
			return err
		}
		fd.isConnected = true
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
	var lsaa syscall.RawSockaddrAny
	var rsaa syscall.RawSockaddrAny
	var salen C.int
	C.srt_getsockname(C.SRTSOCKET(fd.pfd.Sysfd), (*C.struct_sockaddr)(unsafe.Pointer(&lsaa)), &salen)
	lsa = &lsaa
	if crsa != nil {
		fd.setAddr(fd.addrFunc()(lsa), fd.addrFunc()(crsa))
	} else if stat := C.srt_getpeername(C.SRTSOCKET(fd.pfd.Sysfd), (*C.struct_sockaddr)(unsafe.Pointer(&rsaa)), &salen); stat != def.SRT_ERROR {
		rsa = &rsaa
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
		if stat := C.srt_bind(C.SRTSOCKET(fd.pfd.Sysfd), (*C.struct_sockaddr)(unsafe.Pointer(lsa)), C.int(unsafe.Sizeof(*lsa))); stat == def.SRT_ERROR {
			return util.GetLastError("srt_bind")
		}
	}
	if err := listenFunc(fd.pfd.Sysfd, backlog); err != nil {
		return util.GetLastError("srt_listen")
	}
	if err := fd.init(); err != nil {
		return err
	}
	var lsa syscall.RawSockaddrAny
	var salen C.int
	C.srt_getsockname(C.SRTSOCKET(fd.pfd.Sysfd), (*C.struct_sockaddr)(unsafe.Pointer(&lsa)), &salen)
	fd.setAddr(fd.addrFunc()(&lsa), nil)
	return nil
}
