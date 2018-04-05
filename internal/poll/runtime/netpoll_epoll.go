// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package runtime

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"
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
	C.srt_startup()
	logging.Init()
	epfd = int(C.srt_epoll_create())
	if epfd >= 0 {
		go run()
		return
	}
	println("runtime: srt_epoll_create failed with", -epfd)
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

func netpollopen(fd int, pd *pollDesc) int {
	var events C.int = C.SRT_EPOLL_IN | C.SRT_EPOLL_OUT | C.SRT_EPOLL_ERR
	pdsLock.Lock()
	pds[fd] = pd
	pdsLock.Unlock()
	return int(C.srt_epoll_add_usock(C.int(epfd), C.SRTSOCKET(fd), &events))
}

func netpollclose(fd int) int {
	pdsLock.Lock()
	delete(pds, fd)
	pdsLock.Unlock()
	return int(C.srt_epoll_remove_usock(C.int(epfd), C.SRTSOCKET(fd)))
}

func run() {
	var rfdslen, wfdslen C.int
	var rfds, wfds [128]C.SRTSOCKET

	defer func() {
		for s, pd := range pds {
			if !pd.closing {
				srtapi.Close(s)
			}
		}
		C.srt_cleanup()
		done <- true
	}()

	for atomic.LoadInt32(&intState) == 0 {
		rfdslen = C.int(len(rfds))
		wfdslen = C.int(len(wfds))
		n := C.srt_epoll_wait(C.int(epfd), &rfds[0], &rfdslen, &wfds[0], &wfdslen, 100, nil, nil, nil, nil)
		if n < 0 {
			err := srtapi.GetLastError()
			if err != srtapi.ETIMEOUT {
				println("runtime: srt_epoll_wait on fd", epfd, "failed with", err)
				panic("runtime: netpoll failed")
			}
			C.srt_clearlasterror()
			continue
		}
		if n > 0 {
			pdsLock.RLock()
			for i := 0; i < int(rfdslen); i++ {
				fd := int(rfds[i])
				if pd := pds[fd]; pd != nil {
					netpollready(pd, 'r')
				}
			}
			for i := 0; i < int(wfdslen); i++ {
				fd := int(wfds[i])
				if pd := pds[fd]; pd != nil {
					netpollready(pd, 'w')
				}
			}
			pdsLock.RUnlock()
		}
	}
}
