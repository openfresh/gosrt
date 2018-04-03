// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

// +build darwin dragonfly freebsd netbsd openbsd linux plan9 windows nacl solaris

package gosrt

func setDefaultSockopts(s, family, sotype int, ipv6only bool) error {
	return nil
}

func setDefaultListenerSockopts(s int) error {
	return nil
}
