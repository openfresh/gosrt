// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package srt

import (
	"github.com/openfresh/gosrt/srtapi"
)

var (
	errTimedout       = srtapi.ETIMEOUT
	errOpNotSupported = srtapi.EINVOP
)

func isPlatformError(err error) bool {
	_, ok := err.(srtapi.Errno)
	return ok
}
