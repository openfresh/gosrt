package runtime

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"
import (
	"sync"
	"sync/atomic"

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
	setlog()
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
	<-done
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
	delete(pds, fd)
	return int(C.srt_epoll_remove_usock(C.int(epfd), C.SRTSOCKET(fd)))
}

func run() {
	var rfdslen, wfdslen C.int
	var rfds, wfds [128]C.SRTSOCKET

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
			defer pdsLock.RUnlock()
			for i := 0; i < int(rfdslen); i++ {
				fd := int(rfds[i])
				pd := pds[fd]

				netpollready(pd, 'r')
			}
			for i := 0; i < int(wfdslen); i++ {
				fd := int(wfds[i])
				pd := pds[fd]

				netpollready(pd, 'w')
			}
		}
	}

	for s, pd := range pds {
		if !pd.closing {
			srtapi.Close(s)
		}
	}
	C.srt_cleanup()
	done <- true
}
