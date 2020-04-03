// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

package srtapi

// #cgo LDFLAGS: -lsrt
// #include <srt/srt.h>
import "C"
import (
	"io"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

// Startup call srt_startup
func Startup() (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	stat := C.srt_startup()
	if stat == APIError {
		err = getLastError()
	}
	return
}

// Cleanup call srt_cleanup
func Cleanup() (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	stat := C.srt_cleanup()
	if stat == APIError {
		err = getLastError()
	}
	return
}

// EpollCreate call srt_epoll_create
func EpollCreate() (epfd int, err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	epfd = int(C.srt_epoll_create())
	if epfd == APIError {
		err = getLastError()
	}
	return
}

// EpollAddUsock call srt_epoll_add_usock
func EpollAddUsock(epfd int, fd int, events int) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	stat := int(C.srt_epoll_add_usock(C.int(epfd), C.SRTSOCKET(fd), (*C.int)(unsafe.Pointer(&events))))
	if stat == APIError {
		err = getLastError()
	}
	return
}

// EpollRemoveUsock call srt_epoll_remove_usock
func EpollRemoveUsock(epfd int, fd int) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	stat := int(C.srt_epoll_remove_usock(C.int(epfd), C.SRTSOCKET(fd)))
	if stat == APIError {
		err = getLastError()
	}
	return
}

// EpollUpdateUsock call srt_epoll_update_usock
func EpollUpdateUsock(epfd int, fd int, events int) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	stat := int(C.srt_epoll_update_usock(C.int(epfd), C.SRTSOCKET(fd), (*C.int)(unsafe.Pointer(&events))))
	if stat == APIError {
		err = getLastError()
	}
	return
}

// EpollWait call srt_epoll_wait
func EpollWait(epfd int, rfds *SrtSocket, rfdslen *int, wfds *SrtSocket, wfdslen *int, timeout int64) (n int) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	rnum := C.int(*rfdslen)
	wnum := C.int(*wfdslen)
	n = int(C.srt_epoll_wait(C.int(epfd), (*C.SRTSOCKET)(unsafe.Pointer(rfds)), &rnum, (*C.SRTSOCKET)(unsafe.Pointer(wfds)), &wnum, C.int64_t(timeout), nil, nil, nil, nil))
	if n < 0 {
		err := getLastError()
		switch err {
		case ETIMEOUT:
		default:
			println("runtime: srt_epoll_wait on fd", epfd, "failed with", err.Error())
			panic("runtime: netpoll failed")
		}
		ClearLastError()
		n = 0
	}
	*rfdslen = int(rnum)
	*wfdslen = int(wnum)
	return
}

// EpollUwait call srt_epoll_uwait
func EpollUwait(epfd int, fdsSet *SrtEpollEvent, fdsSize int, msTimeOut int64) (n int) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	n = int(C.srt_epoll_uwait(C.int(epfd), (*C.SRT_EPOLL_EVENT)(fdsSet), C.int(fdsSize), C.int64_t(msTimeOut)))
	if n < 0 {
		err := getLastError()
		switch err {
		case ETIMEOUT:
		default:
			println("runtime: srt_epoll_uwait on fd", epfd, "failed with", err.Error())
			panic("runtime: netpoll failed")
		}
		ClearLastError()
		n = 0
	}
	return
}

// GetFdFromEpollEvent return fd from SrtEpollEvent
func GetFdFromEpollEvent(fds *SrtEpollEvent) SrtSocket {
	return SrtSocket(fds.fd)
}

// GetEventsFromEpollEvent return events from SrtEpollEvent
func GetEventsFromEpollEvent(fds *SrtEpollEvent) int {
	return int(fds.events)
}

// EpollSet call srt_epoll_set
func EpollSet(epfd int, flags int) (oflags int, err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	oflags = int(C.srt_epoll_set(C.int(epfd), C.int(flags)))
	if oflags == APIError {
		err = getLastError()
	}
	return
}

func accept(s int, rsa *syscall.RawSockaddrAny, addrlen *_Socklen) (fd int, err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	fd = int(C.srt_accept(C.SRTSOCKET(s), (*C.struct_sockaddr)(unsafe.Pointer(rsa)), (*C.int)(addrlen)))
	if fd == APIError {
		err = getLastError()
	}
	return
}

func getsockname(s int, rsa *syscall.RawSockaddrAny, addrlen *_Socklen) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	stat := C.srt_getsockname(C.SRTSOCKET(s), (*C.struct_sockaddr)(unsafe.Pointer(rsa)), (*C.int)(addrlen))
	if stat == APIError {
		err = getLastError()
	}
	return
}

func getpeername(s int, rsa *syscall.RawSockaddrAny, addrlen *_Socklen) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	stat := C.srt_getpeername(C.SRTSOCKET(s), (*C.struct_sockaddr)(unsafe.Pointer(rsa)), (*C.int)(addrlen))
	if stat == APIError {
		err = getLastError()
	}
	return
}

func bind(s int, addr unsafe.Pointer, addrlen _Socklen) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	stat := C.srt_bind(C.SRTSOCKET(s), (*C.struct_sockaddr)(unsafe.Pointer(addr)), C.int(addrlen))
	if stat == APIError {
		err = getLastError()
	}
	return
}

func connect(s int, addr unsafe.Pointer, addrlen _Socklen) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	stat := C.srt_connect(C.SRTSOCKET(s), (*C.struct_sockaddr)(unsafe.Pointer(addr)), C.int(addrlen))
	if stat == APIError {
		err = getLastError()
	}
	return
}

func socket(domain int, typ int, proto int) (fd int, err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	fd = int(C.srt_socket(C.int(domain), C.int(typ), C.int(proto)))
	if fd == APIError {
		err = getLastError()
	}
	return
}

func getsockflag(s int, name int, val unsafe.Pointer, vallen *_Socklen) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	stat := C.srt_getsockflag(C.SRTSOCKET(s), C.SRT_SOCKOPT(name), val, (*C.int)(vallen))
	if stat == APIError {
		err = getLastError()
	}
	return
}

func setsockflag(s int, name int, val unsafe.Pointer, vallen uintptr) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	stat := C.srt_setsockflag(C.SRTSOCKET(s), C.SRT_SOCKOPT(name), val, C.int(vallen))
	if stat == APIError {
		err = getLastError()
	}
	return
}

func getsockopt(s int, level int, name int, val unsafe.Pointer, vallen *_Socklen) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	stat := C.srt_getsockopt(C.SRTSOCKET(s), C.int(level), C.SRT_SOCKOPT(name), val, (*C.int)(vallen))
	if stat == APIError {
		err = getLastError()
	}
	return
}

func setsockopt(s int, level int, name int, val unsafe.Pointer, vallen uintptr) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	stat := C.srt_setsockopt(C.SRTSOCKET(s), C.int(level), C.SRT_SOCKOPT(name), val, C.int(vallen))
	if stat == APIError {
		err = getLastError()
	}
	return
}

// Listen call srt_listen
func Listen(s int, n int) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	stat := C.srt_listen(C.SRTSOCKET(s), C.int(n))
	if stat == APIError {
		err = getLastError()
	}
	return
}

// Close call srt_close
func Close(fd int) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	stat := C.srt_close(C.SRTSOCKET(fd))
	if stat == APIError {
		err = getLastError()
	}
	return
}

func read(fd int, p []byte) (n int, err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var _p0 unsafe.Pointer
	if len(p) > 0 {
		_p0 = unsafe.Pointer(&p[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	r0 := C.srt_recv(C.SRTSOCKET(fd), (*C.char)(_p0), C.int(len(p)))
	n = int(r0)
	if r0 == APIError {
		err = getLastError()
	}
	return
}

func sendfile(outfd int, r io.Reader, offset *int64, count int) (written int, err error) {
	f, ok := r.(*os.File)
	if !ok {
		return 0, nil
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	name := C.CString(f.Name())
	defer C.free(unsafe.Pointer(name))
	r0 := C.srt_sendfile(C.SRTSOCKET(outfd), name, (*C.int64_t)(offset), C.int64_t(count), DefaultSendfileBlock)
	if r0 == APIError {
		err = getLastError()
	}
	written = int(r0)
	return
}

func write(fd int, p []byte) (n int, err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var _p0 unsafe.Pointer
	if len(p) > 0 {
		_p0 = unsafe.Pointer(&p[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	r0 := C.srt_send(C.SRTSOCKET(fd), (*C.char)(_p0), C.int(len(p)))
	n = int(r0)
	if r0 == APIError {
		err = getLastError()
	}
	return
}

func getlasterror() int {
	return int(C.srt_getlasterror(nil))
}

func strerror(code int, errnoval int) string {
	return C.GoString(C.srt_strerror(C.int(code), C.int(errnoval)))
}

// ClearLastError call srt_clearlasterror
func ClearLastError() {
	C.srt_clearlasterror()
}

// SetLogLevel call srt_setloglevel
func SetLogLevel(level int) {
	C.srt_setloglevel(C.int(level))
}

// AddLogFA call srt_addlogfa
func AddLogFA(fa int) {
	C.srt_addlogfa(C.int(fa))
}

// SetLogFlags call srt_setlogflags
func SetLogFlags(flags int) {
	C.srt_setlogflags(C.int(flags))
}

func GetStats(fd int, clear bool) map[string]interface{} {
	var mon C.struct_CBytePerfMon
	clearStats := 0
	if clear {
		clearStats = 1
	}
	C.srt_bstats(C.SRTSOCKET(fd), &mon, C.int(clearStats))
	output := map[string]interface{}{
		"sid":  fd,
		"time": mon.msTimeStamp,
		"window": map[string]interface{}{
			"flow":       mon.pktFlowWindow,
			"congestion": mon.pktCongestionWindow,
			"flight":     mon.pktFlightSize,
		},
		"link": map[string]interface{}{
			"rtt":          mon.msRTT,
			"bandwidth":    mon.mbpsBandwidth,
			"maxBandwidth": mon.mbpsMaxBW,
		},
		"send": map[string]interface{}{
			"packets":              mon.pktSent,
			"packetsLost":          mon.pktSndLoss,
			"packetsDropped":       mon.pktSndDrop,
			"packetsRetransmitted": mon.pktRetrans,
			"packetsFilterExtra":   mon.pktSndFilterExtra,
			"bytes":                mon.byteSent,
			"bytesDropped":         mon.byteSndDrop,
			"mbitRate":             mon.mbpsSendRate,
		},
		"recv": map[string]interface{}{
			"packets":              mon.pktRecv,
			"packetsLost":          mon.pktRcvLoss,
			"packetsDropped":       mon.pktRcvDrop,
			"packetsRetransmitted": mon.pktRcvRetrans,
			"packetsBelated":       mon.pktRcvBelated,
			"packetsFilterExtra":   mon.pktRcvFilterExtra,
			"packetsFilterSupply":  mon.pktRcvFilterSupply,
			"packetsFilterLoss":    mon.pktRcvFilterLoss,
			"bytes":                mon.byteRecv,
			"bytesLost":            mon.byteRcvLoss,
			"bytesDropped":         mon.byteRcvDrop,
			"mbitRate":             mon.mbpsRecvRate,
		},
	}

	return output
}
