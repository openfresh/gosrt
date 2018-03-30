package poll

import (
	"syscall"

	"github.com/openfresh/gosrt/internal/srtapi"
)

// CloseFunc is used to hook the close call.
var CloseFunc func(int) error = srtapi.Close

// AcceptFunc is used to hook the accept call.
var AcceptFunc func(int) (int, syscall.Sockaddr, error) = srtapi.Accept
