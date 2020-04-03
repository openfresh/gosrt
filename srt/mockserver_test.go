// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package srt

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

func newLocalListener(network string) (net.Listener, error) {
	return newLocalListenerContext(context.Background(), network)
}

func newLocalListenerContext(ctx context.Context, network string) (net.Listener, error) {
	switch network {
	case "srt":
		if supportsIPv4() {
			if ln, err := ListenContext(ctx, "srt4", "127.0.0.1:0"); err == nil {
				return ln, nil
			}
		}
		if supportsIPv6() {
			return ListenContext(ctx, "srt6", "[::1]:0")
		}
	case "srt4":
		if supportsIPv4() {
			return ListenContext(ctx, "srt4", "127.0.0.1:0")
		}
	case "srt6":
		if supportsIPv6() {
			return ListenContext(ctx, "srt6", "[::1]:0")
		}
	}
	return nil, fmt.Errorf("%s is not supported", network)
}

func newDualStackListener() (lns []*SRTListener, err error) {
	var args = []struct {
		network string
		SRTAddr
	}{
		{"srt4", SRTAddr{IP: net.IPv4(127, 0, 0, 1)}},
		{"srt6", SRTAddr{IP: net.IPv6loopback}},
	}
	for i := 0; i < 64; i++ {
		var port int
		var lns []*SRTListener
		for _, arg := range args {
			arg.SRTAddr.Port = port
			ln, err := ListenSRT(arg.network, &arg.SRTAddr)
			if err != nil {
				continue
			}
			port = ln.Addr().(*SRTAddr).Port
			lns = append(lns, ln)
		}
		if len(lns) != len(args) {
			for _, ln := range lns {
				ln.Close()
			}
			continue
		}
		return lns, nil
	}
	return nil, errors.New("no dualstack port available")
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

type dualStackServer struct {
	lnmu sync.RWMutex
	lns  []streamListener
	port string

	cmu sync.RWMutex
	cs  []net.Conn // established connections at the passive open side
}

func (dss *dualStackServer) buildup(handler func(*dualStackServer, net.Listener)) error {
	for i := range dss.lns {
		go func(i int) {
			handler(dss, dss.lns[i].Listener)
			close(dss.lns[i].done)
		}(i)
	}
	return nil
}

func (dss *dualStackServer) teardownNetwork(network string) error {
	dss.lnmu.Lock()
	for i := range dss.lns {
		if network == dss.lns[i].network && dss.lns[i].Listener != nil {
			dss.lns[i].Listener.Close()
			<-dss.lns[i].done
			dss.lns[i].Listener = nil
		}
	}
	dss.lnmu.Unlock()
	return nil
}

func (dss *dualStackServer) teardown() error {
	dss.lnmu.Lock()
	for i := range dss.lns {
		if dss.lns[i].Listener != nil {
			dss.lns[i].Listener.Close()
			<-dss.lns[i].done
		}
	}
	dss.lns = dss.lns[:0]
	dss.lnmu.Unlock()
	dss.cmu.Lock()
	for _, c := range dss.cs {
		c.Close()
	}
	dss.cs = dss.cs[:0]
	dss.cmu.Unlock()
	return nil
}

func newDualStackServer() (*dualStackServer, error) {
	lns, err := newDualStackListener()
	if err != nil {
		return nil, err
	}
	_, port, err := net.SplitHostPort(lns[0].Addr().String())
	if err != nil {
		lns[0].Close()
		lns[1].Close()
		return nil, err
	}
	return &dualStackServer{
		lns: []streamListener{
			{network: "srt4", address: lns[0].Addr().String(), Listener: lns[0], done: make(chan bool)},
			{network: "srt6", address: lns[1].Addr().String(), Listener: lns[1], done: make(chan bool)},
		},
		port: port,
	}, nil
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
	time.Sleep(time.Second * 1)
}

func transceiver(c net.Conn, wb []byte, ch chan<- error) {
	defer close(ch)

	c.SetDeadline(time.Now().Add(someTimeout))
	c.SetReadDeadline(time.Now().Add(someTimeout))
	c.SetWriteDeadline(time.Now().Add(someTimeout))

	n, err := c.Write(wb)
	if err != nil {
		if perr := parseWriteError(err); perr != nil {
			ch <- perr
		}
		ch <- err
		return
	}
	if n != len(wb) {
		ch <- fmt.Errorf("wrote %d; want %d", n, len(wb))
	}
	rb := make([]byte, len(wb))
	n, err = c.Read(rb)
	if err != nil {
		if perr := parseReadError(err); perr != nil {
			ch <- perr
		}
		ch <- err
		return
	}
	if n != len(wb) {
		ch <- fmt.Errorf("read %d; want %d", n, len(wb))
	}
}

func newLocalServer(network string) (*localServer, error) {
	return newLocalServerContext(context.Background(), network)
}

func newLocalServerContext(ctx context.Context, network string) (*localServer, error) {
	ln, err := newLocalListenerContext(ctx, network)
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
