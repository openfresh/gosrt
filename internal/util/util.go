package util

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"
import (
	"fmt"
)

func GetLastError(src string) error {
	var errno C.int
	message := C.srt_getlasterror_str()
	errCode := C.srt_getlasterror(&errno)
	C.srt_clearlasterror()
	return fmt.Errorf("FAILURE %s:[%d] [%d] %s", src, errno, errCode, C.GoString(message))
}
