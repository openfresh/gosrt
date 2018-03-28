package gosrt

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"
import (
	"context"
	"net"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/openfresh/gosrt/def"
	"github.com/openfresh/gosrt/poll"
	"github.com/openfresh/gosrt/util"
	"github.com/pkg/errors"
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

func (fd *netFD) connect(ctx context.Context, la, ra *syscall.RawSockaddrAny) (rsa *syscall.RawSockaddrAny, ret error) {
	switch err := connectFunc(fd.pfd.Sysfd, ra); err {
	case syscall.EINPROGRESS, syscall.EALREADY, syscall.EINTR:
	case nil, syscall.EISCONN:
		select {
		case <-ctx.Done():
			return nil, mapErr(ctx.Err())
		default:
		}
		if err := fd.pfd.Init(fd.net, true); err != nil {
			return nil, err
		}
		runtime.KeepAlive(fd)
		return nil, nil
	default:
		return nil, err
	}
	if err := fd.pfd.Init(fd.net, true); err != nil {
		return nil, err
	}
	if deadline, _ := ctx.Deadline(); !deadline.IsZero() {
		fd.pfd.SetWriteDeadline(deadline)
		defer fd.pfd.SetWriteDeadline(noDeadline)
	}
	var rsaa syscall.RawSockaddrAny
	var namelen C.int
	stat := C.srt_getpeername(C.SRTSOCKET(fd.pfd.Sysfd), (*C.struct_sockaddr)(unsafe.Pointer(&rsaa)), &namelen)
	if stat == def.SRT_ERROR {
		return nil, util.GetLastError("srt_getpeername")
	}
	return &rsaa, nil
}

func (fd *netFD) Close() error {
	runtime.SetFinalizer(fd, nil)
	return fd.pfd.Close()
}

func (fd *netFD) Read(p []byte) (n int, err error) {
	return fd.pfd.Read(p)
}

func (fd *netFD) Write(p []byte) (nn int, err error) {
	return fd.pfd.Write(p)
}

func (fd *netFD) accept() (netfd *netFD, err error) {
	d, rsa, errcall, err := fd.pfd.Accept()
	if err != nil {
		if errcall != "" {
			err = errors.Wrap(err, errcall)
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
	var lsa syscall.RawSockaddrAny
	var namelen C.int
	C.srt_getsockname(C.SRTSOCKET(netfd.pfd.Sysfd), (*C.struct_sockaddr)(unsafe.Pointer(&lsa)), &namelen)
	netfd.setAddr(netfd.addrFunc()(&lsa), netfd.addrFunc()(rsa))
	return netfd, nil
}
