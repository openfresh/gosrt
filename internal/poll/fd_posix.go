// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package poll

import (
	"io"

	"github.com/openfresh/gosrt/srtapi"
)

func (fd *FD) eofError(n int, err error) error {
	if n == 0 && (err == nil || err == srtapi.EINVALMSGAPI) {
		return io.EOF
	}
	return err
}
