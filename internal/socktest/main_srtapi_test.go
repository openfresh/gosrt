package socktest_test

import "github.com/openfresh/gosrt/srtapi"

var (
	socketFunc func(int, int, int) (int, error)
	closeFunc  func(int) error
)

func installTestHooks() {
	socketFunc = sw.Socket
	closeFunc = sw.Close
}

func uninstallTestHooks() {
	socketFunc = srtapi.Socket
	closeFunc = srtapi.Close
}
