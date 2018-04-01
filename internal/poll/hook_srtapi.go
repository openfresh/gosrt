package poll

import (
	"github.com/openfresh/gosrt/srtapi"
)

// CloseFunc is used to hook the close call.
var CloseFunc = srtapi.Close

// AcceptFunc is used to hook the accept call.
var AcceptFunc = srtapi.Accept
