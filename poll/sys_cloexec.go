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

// Wrapper around the accept system call that marks the returned file
// descriptor as nonblocking and close-on-exec.
func accept(s int) (int, *syscall.RawSockaddrAny, string, error) {
	// See ../syscall/exec_unix.go for description of ForkLock.
	// It is probably okay to hold the lock across syscall.Accept
	// because we have put fd.sysfd into non-blocking mode.
	// However, a call to the File method will put it back into
	// blocking mode. We can't take that risk, so no use of ForkLock here.
	ns, sa, err := AcceptFunc(s)
	if err != nil {
		return -1, nil, "accept", err
	}
	yes := 0
	result := C.srt_setsockopt(C.SRTSOCKET(ns), 0, C.SRTO_SNDSYN, unsafe.Pointer(&yes), C.int(unsafe.Sizeof(yes)))
	if result == def.SRT_ERROR {
		CloseFunc(ns)
		return -1, nil, "setnonblock", util.GetLastError("srt_setsockopt")
	}
	result = C.srt_setsockopt(C.SRTSOCKET(ns), 0, C.SRTO_RCVSYN, unsafe.Pointer(&yes), C.int(unsafe.Sizeof(yes)))
	if result == def.SRT_ERROR {
		CloseFunc(ns)
		return -1, nil, "setnonblock", util.GetLastError("srt_setsockopt")
	}
	return ns, sa, "", nil
}
