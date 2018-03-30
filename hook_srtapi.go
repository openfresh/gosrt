package gosrt

import (
	"syscall"

	"github.com/openfresh/gosrt/internal/srtapi"
)

var (
	testHookDialChannel  = func() {}
	testHookCanceledDial = func() {}

	// Placeholders for socket srt calls.
	socketFunc        func(int, int, int) (int, error)  = srtapi.Socket
	connectFunc       func(int, syscall.Sockaddr) error = srtapi.Connect
	listenFunc        func(int, int) error              = srtapi.Listen
	getsockoptIntFunc func(int, int, int) (int, error)  = srtapi.GetsockoptInt
)
