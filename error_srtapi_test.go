// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package gosrt

import (
	"os"

	"github.com/openfresh/gosrt/srtapi"
)

var (
	errTimedout       = srtapi.ETIMEOUT
	errOpNotSupported = srtapi.EINVOP

	abortedConnRequestErrors = []error{srtapi.ECONNLOST} // see accept in fd_unix.go
)

func isPlatformError(err error) bool {
	_, ok := err.(srtapi.Errno)
	return ok
}

func samePlatformError(err, want error) bool {
	if op, ok := err.(*OpError); ok {
		err = op.Err
	}
	if sys, ok := err.(*os.SyscallError); ok {
		err = sys.Err
	}
	return err == want
}
