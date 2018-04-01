package gosrt

import (
	"fmt"
	"net"
	"sync"
	"time"
)

func newLocalListener(network string) (net.Listener, error) {
	switch network {
	case "srt":
		if supportsIPv4() {
			if ln, err := Listen("srt4", "127.0.0.1:0"); err == nil {
				return ln, nil
			}
		}
		if supportsIPv6() {
			return Listen("srt6", "[::1]:0")
		}
	case "srt4":
		if supportsIPv4() {
			return Listen("srt4", "127.0.0.1:0")
		}
	case "srt6":
		if supportsIPv6() {
			return Listen("srt6", "[::1]:0")
		}
	}
	return nil, fmt.Errorf("%s is not supported", network)
}

type localServer struct {
	lnmu sync.RWMutex
	net.Listener
	done chan bool // signal that indicates server stopped
}

func (ls *localServer) buildup(handler func(*localServer, net.Listener)) error {
	go func() {
		handler(ls, ls.Listener)
		close(ls.done)
	}()
	return nil
}

func (ls *localServer) teardown() error {
	ls.lnmu.Lock()
	if ls.Listener != nil {
		ls.Listener.Close()
		<-ls.done
		ls.Listener = nil
	}
	ls.lnmu.Unlock()
	return nil
}

func transponder(ln net.Listener, ch chan<- error) {
	defer close(ch)

	switch ln := ln.(type) {
	case *SRTListener:
		ln.SetDeadline(time.Now().Add(someTimeout))
	}
	c, err := ln.Accept()
	if err != nil {
		if perr := parseAcceptError(err); perr != nil {
			ch <- perr
		}
		ch <- err
		return
	}
	defer c.Close()

	network := ln.Addr().Network()
	if c.LocalAddr().Network() != network || c.RemoteAddr().Network() != network {
		ch <- fmt.Errorf("got %v->%v; expected %v->%v", c.LocalAddr().Network(), c.RemoteAddr().Network(), network, network)
		return
	}
	c.SetDeadline(time.Now().Add(someTimeout))
	c.SetReadDeadline(time.Now().Add(someTimeout))
	c.SetWriteDeadline(time.Now().Add(someTimeout))

	b := make([]byte, 256)
	n, err := c.Read(b)
	if err != nil {
		if perr := parseReadError(err); perr != nil {
			ch <- perr
		}
		ch <- err
		return
	}
	if _, err := c.Write(b[:n]); err != nil {
		if perr := parseWriteError(err); perr != nil {
			ch <- perr
		}
		ch <- err
		return
	}
}

func newLocalServer(network string) (*localServer, error) {
	ln, err := newLocalListener(network)
	if err != nil {
		return nil, err
	}
	return &localServer{Listener: ln, done: make(chan bool)}, nil
}

type streamListener struct {
	network, address string
	net.Listener
	done chan bool // signal that indicates server stopped
}

func (sl *streamListener) newLocalServer() (*localServer, error) {
	return &localServer{Listener: sl.Listener, done: make(chan bool)}, nil
}
