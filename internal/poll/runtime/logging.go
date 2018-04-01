package runtime

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
// #include "udt_wrapper.h"
/*
   void logHandler_cgo(void* opaque, int level, const char* file, int line, const char* area, const char* message);
*/
import "C"
import (
	"fmt"
	"time"
	"unsafe"

	"github.com/openfresh/gosrt/config"
	"github.com/openfresh/gosrt/srtapi"
)

//export logHandler
func logHandler(opaque unsafe.Pointer, level C.int, file *C.char, line C.int, area *C.char, message *C.char) {
	now := time.Now()
	buf := fmt.Sprintf("[%v, %s:%d(%s)]{%d} %s", now, C.GoString(file), line, C.GoString(area), level, C.GoString(message))
	println(buf)
}

func setlog() {
	C.srt_setloglevel(C.int(config.LogLevel))
	for fa := range config.LogFas {
		C.srt_addlogfa(C.int(fa))
	}
	NAME := C.CString("SRTLIB")
	defer C.free(unsafe.Pointer(NAME))
	if config.LogInternal {
		C.srt_setlogflags(0 | srtapi.LogFlagDisableTime | srtapi.LogFlagDisableSeverity | srtapi.LogFlagDisableThreadname | srtapi.LogFlagDisableEOF)
		C.srt_setloghandler(unsafe.Pointer(NAME), (*C.SRT_LOG_HANDLER_FN)(C.logHandler_cgo))
	} else if config.LogFile != "" {
		p := C.CString(config.LogFile)
		defer C.free(unsafe.Pointer(p))
		C.udtSetLogStream(p)
	}
}
