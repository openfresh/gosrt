// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package gosrt

import (
	"syscall"

	"golang.org/x/net/route"
)

func interfaceMessages(ifindex int) ([]route.Message, error) {
	typ := route.RIBType(syscall.NET_RT_IFLISTL)
	rib, err := route.FetchRIB(syscall.AF_UNSPEC, typ, ifindex)
	if err != nil {
		typ = route.RIBType(syscall.NET_RT_IFLIST)
		rib, err = route.FetchRIB(syscall.AF_UNSPEC, typ, ifindex)
	}
	if err != nil {
		return nil, err
	}
	return route.ParseRIB(typ, rib)
}
