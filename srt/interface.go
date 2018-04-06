// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package srt

import (
	"net"
	"sync"
	"time"
)

// An ipv6ZoneCache represents a cache holding partial network
// interface information. It is used for reducing the cost of IPv6
// addressing scope zone resolution.
//
// Multiple names sharing the index are managed by first-come
// first-served basis for consistency.
type ipv6ZoneCache struct {
	sync.RWMutex                // guard the following
	lastFetched  time.Time      // last time routing information was fetched
	toIndex      map[string]int // interface name to its index
	toName       map[int]string // interface index to its name
}

var zoneCache = ipv6ZoneCache{
	toIndex: make(map[string]int),
	toName:  make(map[int]string),
}

func (zc *ipv6ZoneCache) update(ift []net.Interface) {
	zc.Lock()
	defer zc.Unlock()
	now := time.Now()
	if zc.lastFetched.After(now.Add(-60 * time.Second)) {
		return
	}
	zc.lastFetched = now
	if len(ift) == 0 {
		var err error
		if ift, err = interfaceTable(0); err != nil {
			return
		}
	}
	zc.toIndex = make(map[string]int, len(ift))
	zc.toName = make(map[int]string, len(ift))
	for _, ifi := range ift {
		zc.toIndex[ifi.Name] = ifi.Index
		if _, ok := zc.toName[ifi.Index]; !ok {
			zc.toName[ifi.Index] = ifi.Name
		}
	}
}

func (zc *ipv6ZoneCache) name(index int) string {
	if index == 0 {
		return ""
	}
	zoneCache.update(nil)
	zoneCache.RLock()
	defer zoneCache.RUnlock()
	name, ok := zoneCache.toName[index]
	if !ok {
		name = uitoa(uint(index))
	}
	return name
}

func (zc *ipv6ZoneCache) index(name string) int {
	if name == "" {
		return 0
	}
	zoneCache.update(nil)
	zoneCache.RLock()
	defer zoneCache.RUnlock()
	index, ok := zoneCache.toIndex[name]
	if !ok {
		index, _, _ = dtoi(name)
	}
	return index
}
