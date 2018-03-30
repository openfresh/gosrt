package gosrt

import (
	"github.com/openfresh/gosrt/internal/poll"
	"github.com/openfresh/gosrt/internal/srtapi"
)

// Wrapper around the socket system call that marks the returned file
// descriptor as nonblocking.
func srtSocket(family, sotype, proto int) (int, error) {
	s, err := socketFunc(family, sotype, proto)
	if err != nil {
		return -1, err
	}
	if err = srtapi.SetNonblock(s, true); err != nil {
		poll.CloseFunc(s)
		return -1, err
	}
	return s, nil
}
