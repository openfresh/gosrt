// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package srt_test

import (
	"io"
	"log"
	"net"

	"github.com/openfresh/gosrt/srt"
)

//lint:ignore U1000 Dummy interface ffor Testable Example
var Listener interface{}

func ExampleListener() {
	// Listen on UDP port 2000 on all available unicast and
	// anycast IP addresses of the local system.
	l, err := srt.Listen("srt", ":2000")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		srtConn := conn.(*srt.SRTConn)
		streamID, err := srtConn.StreamID()
		if err != nil {
			log.Fatal(err)
		}
		log.Print(streamID)
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn) {
			// Echo all incoming data.
			io.Copy(c, c)
			// Shut down the connection.
			c.Close()
		}(conn)
	}
}
