// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/openfresh/gosrt/srt"
)

func main() {
	sport := os.Getenv("SERVER_PORT")
	targetsEnv := os.Getenv("TARGETS")
	targets := strings.Split(targetsEnv, ",")
	fmt.Println("start")
	fmt.Printf("server port: %s\n", sport)
	for i := 0; i < len(targets); i++ {
		fmt.Printf("target %d: %s\n", i, targets[i])
	}
	chunksize := 1316

	ctx := srt.WithOptions(context.Background(), srt.Options("payloadsize", strconv.Itoa(chunksize)))
	fmt.Println("listen")
	l, err := srt.ListenContext(ctx, "srt", ":"+sport)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	i := 0
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("accepted: %s\n", conn.RemoteAddr())
		target := ""
		if i < len(targets) {
			target = targets[i]
		}
		go func(sc net.Conn, taddr string) {
			defer sc.Close()
			var d srt.Dialer
			fmt.Printf("connecting: %s\n", taddr)
			tc, err := d.DialContext(ctx, "srt", taddr)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("connected: %s\n", taddr)
			for {
				b := make([]byte, chunksize)
				n, err := sc.Read(b)
				if err != nil {
					log.Fatal(err)
				}
				tc.Write(b[:n])
			}
		}(conn, target)
		i++
	}
}
