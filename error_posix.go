// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package gosrt

import (
	"os"

	"github.com/openfresh/gosrt/srtapi"
)

// wrapSyscallError takes an error and a srtapi name. If the error is
// a srtapi.Errno, it wraps it in a os.SyscallError using the srtapi name.
func wrapSyscallError(name string, err error) error {
	if _, ok := err.(srtapi.Errno); ok {
		err = os.NewSyscallError(name, err)
	}
	return err
}
