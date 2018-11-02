// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/openfresh/gosrt/srt"
)

func main() {
	sport := os.Getenv("SERVER_PORT")
	targetsEnv := os.Getenv("TARGETS")
	targets := strings.Split(targetsEnv, ",")
	statsReportStr := os.Getenv("STATS_REPORT")
	statsReport := 0
	v, err := strconv.Atoi(statsReportStr)
	if err == nil {
		statsReport = v
	}
	fmt.Println("start")
	fmt.Printf("server port: %s\n", sport)
	for i := 0; i < len(targets); i++ {
		fmt.Printf("target %d: %s\n", i, targets[i])
	}
	chunksize := 1316

	srt.SetLoggingHandler(func(level int, file string, line int, area string, message string) {
		now := time.Now()
		buf := fmt.Sprintf("[%v, %s:%d(%s)]{%d} %s", now, file, line, area, level, message)
		println(buf)
	})

	defer srt.Shutdown()
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
			counter := 0
			for {
				b := make([]byte, chunksize)
				n, err := sc.Read(b)
				if err != nil {
					log.Fatal(err)
				}
				tc.Write(b[:n])

				if statsReport > 0 && (counter%statsReport) == statsReport-1 {
					printSrtStats(sc)
					printSrtStats(tc)
				}
				counter++
			}
		}(conn, target)
		i++
	}
}

func printSrtStats(conn net.Conn) {
	mon := conn.(*srt.SRTConn).Stats()
	s, _ := json.MarshalIndent(mon, "", "\t")
	fmt.Println(string(s))
}
