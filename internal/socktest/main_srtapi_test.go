// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package socktest_test

import "github.com/openfresh/gosrt/srtapi"

var (
	socketFunc func(int, int, int) (int, error)
	closeFunc  func(int) error
)

func installTestHooks() {
	socketFunc = sw.Socket
	closeFunc = sw.Close
}

func uninstallTestHooks() {
	socketFunc = srtapi.Socket
	closeFunc = srtapi.Close
}
