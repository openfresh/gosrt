// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

// This file implements sysSocket and accept for platforms that
// provide a fast path for setting SetNonblock and CloseOnExec.

package gosrt

import (
	"github.com/openfresh/gosrt/internal/poll"
	"github.com/openfresh/gosrt/srtapi"
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
