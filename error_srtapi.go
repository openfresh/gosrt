package gosrt

import "syscall"

func isPlatformError(err error) bool {
	_, ok := err.(syscall.Errno)
	return ok
}
