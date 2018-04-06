// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package runtime

import (
	"sync"
	"sync/atomic"

	"github.com/openfresh/gosrt/logging"
	"github.com/openfresh/gosrt/srtapi"
)

var (
	epfd     = -1 // epoll descriptor
	pds      = make(map[int]*pollDesc)
	pdsLock  = &sync.RWMutex{}
	intState int32
	done     = make(chan bool, 1)
)

func netpollinit() {
	srtapi.Startup()
	logging.Init()
	var err error
	epfd, err = srtapi.EpollCreate()
	if err == nil {
		go run()
		return
	}
	println("runtime: srt_epoll_create failed with", err.Error())
	panic("runtime: netpollinit failed")
}

func netpollshutdown() {
	atomic.CompareAndSwapInt32(&intState, 0, 1)
	if epfd >= 0 {
		<-done
	}
}

func netpolldescriptor() int {
	return epfd
}

func netpollopen(fd int, pd *pollDesc) error {
	events := srtapi.EpollIn | srtapi.EpollOut | srtapi.EpollErr
	pdsLock.Lock()
	pds[fd] = pd
	pdsLock.Unlock()
	return srtapi.EpollAddUsock(epfd, fd, events)
}

func netpollclose(fd int) error {
	pdsLock.Lock()
	delete(pds, fd)
	pdsLock.Unlock()
	return srtapi.EpollRemoveUsock(epfd, fd)
}

func run() {
	var rfdslen, wfdslen int
	var rfds, wfds [128]srtapi.SrtSocket

	defer func() {
		for s, pd := range pds {
			if !pd.closing {
				srtapi.Close(s)
			}
		}
		srtapi.Cleanup()
		done <- true
	}()

	for atomic.LoadInt32(&intState) == 0 {
		rfdslen = len(rfds)
		wfdslen = len(wfds)
		n := srtapi.EpollWait(epfd, &rfds[0], &rfdslen, &wfds[0], &wfdslen, 100)
		if n > 0 {
			pdsLock.RLock()
			for i := 0; i < rfdslen; i++ {
				fd := int(rfds[i])
				if pd := pds[fd]; pd != nil {
					netpollready(pd, 'r')
				}
			}
			for i := 0; i < wfdslen; i++ {
				fd := int(wfds[i])
				if pd := pds[fd]; pd != nil {
					netpollready(pd, 'w')
				}
			}
			pdsLock.RUnlock()
		}
	}
}
