// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package poll

import (
	"github.com/openfresh/gosrt/srtapi"
)

// CloseFunc is used to hook the close call.
var CloseFunc = srtapi.Close

// AcceptFunc is used to hook the accept call.
var AcceptFunc = srtapi.Accept
