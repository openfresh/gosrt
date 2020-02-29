// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go

package socktest

import (
	"syscall"

	"github.com/openfresh/gosrt/srtapi"
)

// Socket wraps srtapi.Socket.
func (sw *Switch) Socket(family, sotype, proto int) (s int, err error) {
	sw.once.Do(sw.init)

	so := &Status{Cookie: cookie(family, sotype, proto)}
	sw.fmu.RLock()
	f := sw.fltab[FilterSocket]
	sw.fmu.RUnlock()

	af, err := f.apply(so)
	if err != nil {
		return -1, err
	}
	s, so.Err = srtapi.Socket(family, sotype, proto)
	if err = af.apply(so); err != nil {
		if so.Err == nil {
			srtapi.Close(s)
		}
		return -1, err
	}

	sw.smu.Lock()
	defer sw.smu.Unlock()
	if so.Err != nil {
		sw.stats.getLocked(so.Cookie).OpenFailed++
		return -1, so.Err
	}
	nso := sw.addLocked(s, family, sotype, proto)
	sw.stats.getLocked(nso.Cookie).Opened++
	return s, nil
}

// Close wraps srtapi.Close.
func (sw *Switch) Close(s int) (err error) {
	so := sw.sockso(s)
	if so == nil {
		return srtapi.Close(s)
	}
	sw.fmu.RLock()
	f := sw.fltab[FilterClose]
	sw.fmu.RUnlock()

	af, err := f.apply(so)
	if err != nil {
		return err
	}
	so.Err = srtapi.Close(s)
	if err = af.apply(so); err != nil {
		return err
	}

	sw.smu.Lock()
	defer sw.smu.Unlock()
	if so.Err != nil {
		sw.stats.getLocked(so.Cookie).CloseFailed++
		return so.Err
	}
	delete(sw.sotab, s)
	sw.stats.getLocked(so.Cookie).Closed++
	return nil
}

// Connect wraps srtapi.Connect.
func (sw *Switch) Connect(s int, sa syscall.Sockaddr) (err error) {
	so := sw.sockso(s)
	if so == nil {
		return srtapi.Connect(s, sa)
	}
	sw.fmu.RLock()
	f := sw.fltab[FilterConnect]
	sw.fmu.RUnlock()

	af, err := f.apply(so)
	if err != nil {
		return err
	}
	so.Err = srtapi.Connect(s, sa)
	if err = af.apply(so); err != nil {
		return err
	}

	sw.smu.Lock()
	defer sw.smu.Unlock()
	if so.Err != nil {
		sw.stats.getLocked(so.Cookie).ConnectFailed++
		return so.Err
	}
	sw.stats.getLocked(so.Cookie).Connected++
	return nil
}

// Listen wraps srtapi.Listen.
func (sw *Switch) Listen(s, backlog int) (err error) {
	so := sw.sockso(s)
	if so == nil {
		return srtapi.Listen(s, backlog)
	}
	sw.fmu.RLock()
	f := sw.fltab[FilterListen]
	sw.fmu.RUnlock()

	af, err := f.apply(so)
	if err != nil {
		return err
	}
	so.Err = srtapi.Listen(s, backlog)
	if err = af.apply(so); err != nil {
		return err
	}

	sw.smu.Lock()
	defer sw.smu.Unlock()
	if so.Err != nil {
		sw.stats.getLocked(so.Cookie).ListenFailed++
		return so.Err
	}
	sw.stats.getLocked(so.Cookie).Listened++
	return nil
}

// Accept wraps srtapi.Accept.
func (sw *Switch) Accept(s int) (ns int, sa syscall.Sockaddr, err error) {
	so := sw.sockso(s)
	if so == nil {
		return srtapi.Accept(s)
	}
	sw.fmu.RLock()
	f := sw.fltab[FilterAccept]
	sw.fmu.RUnlock()

	af, err := f.apply(so)
	if err != nil {
		return -1, nil, err
	}
	ns, sa, so.Err = srtapi.Accept(s)
	if err = af.apply(so); err != nil {
		if so.Err == nil {
			srtapi.Close(ns)
		}
		return -1, nil, err
	}

	sw.smu.Lock()
	defer sw.smu.Unlock()
	if so.Err != nil {
		sw.stats.getLocked(so.Cookie).AcceptFailed++
		return -1, nil, so.Err
	}
	nso := sw.addLocked(ns, so.Cookie.Family(), so.Cookie.Type(), so.Cookie.Protocol())
	sw.stats.getLocked(nso.Cookie).Accepted++
	return ns, sa, nil
}

// GetsockoptInt wraps srtapi.GetsockoptInt.
func (sw *Switch) GetsockoptInt(s, level, opt int) (status int, err error) {
	so := sw.sockso(s)
	if so == nil {
		return srtapi.GetsockoptInt(s, level, opt)
	}
	sw.fmu.RLock()
	f := sw.fltab[FilterGetsockoptInt]
	sw.fmu.RUnlock()

	af, err := f.apply(so)
	if err != nil {
		return -1, err
	}
	status, so.Err = srtapi.GetsockoptInt(s, level, opt)
	so.SockStatus = status
	if err = af.apply(so); err != nil {
		return -1, err
	}

	if so.Err != nil {
		return -1, so.Err
	}
	if opt == srtapi.OptionState && so.SockStatus == srtapi.StatusConnected {
		sw.smu.Lock()
		sw.stats.getLocked(so.Cookie).Connected++
		sw.smu.Unlock()
	}
	return status, nil
}

// GetsockflagInt wraps srtapi.GetsockflagInt.
func (sw *Switch) GetsockflagInt(s, opt int) (status int, err error) {
	so := sw.sockso(s)
	if so == nil {
		return srtapi.GetsockflagInt(s, opt)
	}
	sw.fmu.RLock()
	f := sw.fltab[FilterGetsockoptInt]
	sw.fmu.RUnlock()

	af, err := f.apply(so)
	if err != nil {
		return -1, err
	}
	status, so.Err = srtapi.GetsockflagInt(s, opt)
	so.SockStatus = status
	if err = af.apply(so); err != nil {
		return -1, err
	}

	if so.Err != nil {
		return -1, so.Err
	}
	if opt == srtapi.OptionState && so.SockStatus == srtapi.StatusConnected {
		sw.smu.Lock()
		sw.stats.getLocked(so.Cookie).Connected++
		sw.smu.Unlock()
	}
	return status, nil
}
