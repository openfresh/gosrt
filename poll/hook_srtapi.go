package poll

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"

import (
	"syscall"
	"unsafe"

	"github.com/openfresh/gosrt/def"
	"github.com/openfresh/gosrt/util"
)

// CloseFunc is used to hook the close call.
var CloseFunc func(int) error = callSrtSocket

// AcceptFunc is used to hook the accept call.
var AcceptFunc func(int) (int, *syscall.RawSockaddrAny, error) = callSrtAccept

func callSrtSocket(sock int) error {
	stat := C.srt_close(C.SRTSOCKET(sock))
	if stat == def.SRT_ERROR {
		return util.GetLastError("srt_close")
	}
	return nil
}

func callSrtAccept(bindsock int) (int, *syscall.RawSockaddrAny, error) {
	var scl syscall.RawSockaddrAny
	var sclen C.int = C.int(unsafe.Sizeof(scl))
	sock := C.srt_accept(C.SRTSOCKET(bindsock), (*C.struct_sockaddr)(unsafe.Pointer(&scl)), &sclen)
	if sock == def.SRT_ERROR {
		return int(sock), nil, util.GetLastError("srt_accept")
	}
	return int(sock), &scl, nil
}
