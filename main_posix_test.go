package gosrt

import (
	"github.com/openfresh/gosrt/internal/poll"
)

var (
	// Placeholders for saving original socket system calls.
	origSocket        = socketFunc
	origClose         = poll.CloseFunc
	origConnect       = connectFunc
	origListen        = listenFunc
	origAccept        = poll.AcceptFunc
	origGetsockoptInt = getsockoptIntFunc

	extraTestHookInstallers   []func()
	extraTestHookUninstallers []func()
)

func installTestHooks() {
	socketFunc = sw.Socket
	poll.CloseFunc = sw.Close
	connectFunc = sw.Connect
	listenFunc = sw.Listen
	poll.AcceptFunc = sw.Accept
	getsockoptIntFunc = sw.GetsockoptInt

	for _, fn := range extraTestHookInstallers {
		fn()
	}
}

func uninstallTestHooks() {
	socketFunc = origSocket
	poll.CloseFunc = origClose
	connectFunc = origConnect
	listenFunc = origListen
	poll.AcceptFunc = origAccept
	getsockoptIntFunc = origGetsockoptInt

	for _, fn := range extraTestHookUninstallers {
		fn()
	}
}

// forceCloseSockets must be called only from TestMain.
func forceCloseSockets() {
	for s := range sw.Sockets() {
		poll.CloseFunc(s)
	}
}
