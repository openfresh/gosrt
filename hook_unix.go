package gosrt

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"
import (
	"syscall"
	"unsafe"

	"github.com/openfresh/gosrt/def"
	"github.com/openfresh/gosrt/util"
)

var (
	testHookDialChannel  = func() {}
	testHookCanceledDial = func() {}

	// Placeholders for socket srt calls.
	socketFunc  = callSrtSocket
	connectFunc = callSrtConnect
	listenFunc  = callSrtListen
)

func callSrtSocket(family int, sotype int, proto int) (int, error) {
	stat := C.srt_socket(C.int(family), C.int(sotype), C.int(proto))
	if stat == def.SRT_ERROR {
		return int(stat), util.GetLastError("srt_socket")
	}
	return int(stat), nil
}

func callSrtConnect(sock int, name *syscall.RawSockaddrAny) error {
	stat := C.srt_connect(C.SRTSOCKET(sock), (*C.struct_sockaddr)(unsafe.Pointer(name)), C.int(unsafe.Sizeof(name)))
	if stat == def.SRT_ERROR {
		if C.srt_getlasterror(nil) == C.SRT_ECONNSOCK {
			return syscall.EISCONN
		}
		return util.GetLastError("srt_connect")
	}
	return nil
}

func callSrtListen(sock int, backlog int) error {
	stat := C.srt_listen(C.SRTSOCKET(sock), C.int(backlog))
	if stat == def.SRT_ERROR {
		return util.GetLastError("srt_listen")
	}
	return nil
}
