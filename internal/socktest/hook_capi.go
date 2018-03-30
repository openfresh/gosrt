package socktest

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"
import (
	"github.com/openfresh/gosrt/internal/srtapi"
	"github.com/openfresh/gosrt/internal/util"
)

func callSrtSocket(domain, typ, proto int) (int, error) {
	stat := C.srt_socket(C.int(domain), C.int(typ), C.int(proto))
	if stat == srtapi.SRT_ERROR {
		return int(stat), util.GetLastError("srt_socket")
	}
	return int(stat), nil
}

func callSrtClose(fd int) (err error) {
	stat := C.srt_close(C.SRTSOCKET(fd))
	if stat == srtapi.SRT_ERROR {
		err = util.GetLastError("srt_close")
	}
	return
}
