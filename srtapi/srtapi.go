package srtapi

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"

// Single-word zero for use when we need a valid pointer to 0 bytes.
// See mksyscall.pl.
var _zero uintptr
