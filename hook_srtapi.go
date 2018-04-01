package gosrt

import (
	"github.com/openfresh/gosrt/srtapi"
)

var (
	testHookDialChannel  = func() {}
	testHookCanceledDial = func() {}

	// Placeholders for socket srt calls.
	socketFunc        = srtapi.Socket
	connectFunc       = srtapi.Connect
	listenFunc        = srtapi.Listen
	getsockoptIntFunc = srtapi.GetsockoptInt
)
