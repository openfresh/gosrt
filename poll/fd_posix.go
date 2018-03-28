package poll

import (
	"io"
)

func (fd *FD) eofError(n int, err error) error {
	if n == 0 && err == nil {
		return io.EOF
	}
	return err
}
