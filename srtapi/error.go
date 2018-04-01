package srtapi

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"
import "fmt"

// SRTError records an error from a specific srt api call.
type SRTError struct {
	SrtAPI  string
	Message string
	Code    int
	Errno   int
}

func (e *SRTError) Error() string {
	return fmt.Sprintf("FAILURE %s:[%d] [%d] %s", e.SrtAPI, e.Errno, e.Code, e.Message)
}

// GetLastError return last SRT API error and clear error
func GetLastError(src string) error {
	var errno C.int
	message := C.srt_getlasterror_str()
	errCode := C.srt_getlasterror(&errno)
	C.srt_clearlasterror()
	return &SRTError{SrtAPI: src, Message: C.GoString(message), Code: int(errCode), Errno: int(errno)}
}
