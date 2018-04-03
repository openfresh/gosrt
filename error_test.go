// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package gosrt

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/openfresh/gosrt/internal/poll"
)

// parseAcceptError parses nestedErr and reports whether it is a valid
// error value from Accept functions.
// It returns nil when nestedErr is valid.
func parseAcceptError(nestedErr error) error {
	if nestedErr == nil {
		return nil
	}

	switch err := nestedErr.(type) {
	case *net.OpError:
		nestedErr = err.Err
		goto second
	}
	return fmt.Errorf("unexpected type on 1st nested level: %T", nestedErr)

second:
	if isPlatformError(nestedErr) {
		return nil
	}
	switch err := nestedErr.(type) {
	case *os.SyscallError:
		nestedErr = err.Err
		goto third
	case *os.PathError: // for Plan 9
		nestedErr = err.Err
		goto third
	}
	switch nestedErr {
	case poll.ErrNetClosing, poll.ErrTimeout:
		return nil
	}
	return fmt.Errorf("unexpected type on 2nd nested level: %T", nestedErr)

third:
	if isPlatformError(nestedErr) {
		return nil
	}
	return fmt.Errorf("unexpected type on 3rd nested level: %T", nestedErr)
}

// parseReadError parses nestedErr and reports whether it is a valid
// error value from Read functions.
// It returns nil when nestedErr is valid.
func parseReadError(nestedErr error) error {
	if nestedErr == nil {
		return nil
	}

	switch err := nestedErr.(type) {
	case *net.OpError:
		nestedErr = err.Err
		goto second
	}
	if nestedErr == io.EOF {
		return nil
	}
	return fmt.Errorf("unexpected type on 1st nested level: %T", nestedErr)

second:
	if isPlatformError(nestedErr) {
		return nil
	}
	switch err := nestedErr.(type) {
	case *os.SyscallError:
		nestedErr = err.Err
		goto third
	}
	switch nestedErr {
	case poll.ErrNetClosing, poll.ErrTimeout:
		return nil
	}
	return fmt.Errorf("unexpected type on 2nd nested level: %T", nestedErr)

third:
	if isPlatformError(nestedErr) {
		return nil
	}
	return fmt.Errorf("unexpected type on 3rd nested level: %T", nestedErr)
}

// parseWriteError parses nestedErr and reports whether it is a valid
// error value from Write functions.
// It returns nil when nestedErr is valid.
func parseWriteError(nestedErr error) error {
	if nestedErr == nil {
		return nil
	}

	switch err := nestedErr.(type) {
	case *net.OpError:
		nestedErr = err.Err
		goto second
	}
	return fmt.Errorf("unexpected type on 1st nested level: %T", nestedErr)

second:
	if isPlatformError(nestedErr) {
		return nil
	}
	switch err := nestedErr.(type) {
	case *net.AddrError, *net.DNSError, net.InvalidAddrError, *net.ParseError, *poll.TimeoutError, net.UnknownNetworkError:
		return nil
	case *os.SyscallError:
		nestedErr = err.Err
		goto third
	}
	switch nestedErr {
	case errCanceled, poll.ErrNetClosing, errMissingAddress, poll.ErrTimeout, net.ErrWriteToConnected, io.ErrUnexpectedEOF:
		return nil
	}
	return fmt.Errorf("unexpected type on 2nd nested level: %T", nestedErr)

third:
	if isPlatformError(nestedErr) {
		return nil
	}
	return fmt.Errorf("unexpected type on 3rd nested level: %T", nestedErr)
}
