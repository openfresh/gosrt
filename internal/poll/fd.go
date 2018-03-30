package poll

import "errors"

// ErrNetClosing is returned when a network descriptor is used after
// it has been closed. Keep this string consistent because of issue
// #4373: since historically programs have not been able to detect
// this error, they look for the string.
var ErrNetClosing = errors.New("use of closed network connection")

// ErrFileClosing is returned when a file descriptor is used after it
// has been closed.
var ErrFileClosing = errors.New("use of closed file")

// ErrNoDeadline is returned when a request is made to set a deadline
// on a file type that does not use the poller.
var ErrNoDeadline = errors.New("file type does not support deadline")

// Return the appropriate closing error based on isFile.
func errClosing() error {
	return ErrNetClosing
}

// ErrTimeout is returned for an expired deadline.
var ErrTimeout error = &TimeoutError{}

// TimeoutError is returned for an expired deadline.
type TimeoutError struct{}

// Implement the net.Error interface.
func (e *TimeoutError) Error() string   { return "i/o timeout" }
func (e *TimeoutError) Timeout() bool   { return true }
func (e *TimeoutError) Temporary() bool { return true }
