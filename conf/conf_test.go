// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package conf

import (
	"os"
	"testing"

	"github.com/openfresh/gosrt/srtapi"
)

func TestConf(t *testing.T) {
	tests := []struct {
		verbose     string
		logLevel    string
		logFAs      string
		logFile     string
		logInternal string
		fullStats   string
		want        Conf
	}{
		{
			verbose:     "true",
			logLevel:    "7",
			logFAs:      "all",
			logFile:     "",
			logInternal: "false",
			fullStats:   "false",
			want: Conf{
				verbose:     true,
				logLevel:    srtapi.LogDebug,
				logFAs:      []int{srtapi.LogFABstats, srtapi.LogFAControl, srtapi.LogFAData, srtapi.LogFATsbpd, srtapi.LogFARexmit},
				logFile:     "",
				logInternal: false,
				fullStats:   false,
			},
		},
		{
			verbose:     "false",
			logLevel:    "fatal",
			logFAs:      "control,tsbpd,rexmit",
			logFile:     "/path/gosrt.log",
			logInternal: "true",
			fullStats:   "true",
			want: Conf{
				verbose:     false,
				logLevel:    srtapi.LogFatal,
				logFAs:      []int{srtapi.LogFAControl, srtapi.LogFATsbpd, srtapi.LogFARexmit},
				logFile:     "/path/gosrt.log",
				logInternal: true,
				fullStats:   true,
			},
		},
	}

	for _, tt := range tests {
		os.Setenv("SRT_VERBOSE", tt.verbose)
		os.Setenv("SRT_LOGLEVEL", tt.logLevel)
		os.Setenv("SRT_LOGFA", tt.logFAs)
		os.Setenv("SRT_LOGFILE", tt.logFile)
		os.Setenv("SRT_LOGINTERNAL", tt.logInternal)
		os.Setenv("SRT_FULLSTATS", tt.fullStats)
		initConfVal()
		if confVal.Verbose() != tt.want.verbose {
			t.Errorf("verbose = %v; want %v", confVal.Verbose(), tt.want.verbose)
		}
		if confVal.LogLevel() != tt.want.logLevel {
			t.Errorf("logLevel = %v; want %v", confVal.LogLevel(), tt.want.logLevel)
		}
		same := true
		if len(confVal.LogFAs()) != len(tt.want.logFAs) {
			same = false
		}
		for i := 0; i < len(confVal.LogFAs()); i++ {
			if confVal.LogFAs()[i] != tt.want.logFAs[i] {
				same = false
			}
		}
		if !same {
			t.Errorf("logFAs = %+v; want %+v", confVal.LogFAs(), tt.want.logFAs)
		}
		if confVal.LogFile() != tt.want.logFile {
			t.Errorf("logFile = %v; want %v", confVal.LogFile(), tt.want.logFile)
		}
		if confVal.LogInternal() != tt.want.logInternal {
			t.Errorf("logInternal = %v; want %v", confVal.LogInternal(), tt.want.logInternal)
		}
		if confVal.FullStats() != tt.want.fullStats {
			t.Errorf("fullStats = %v; want %v", confVal.FullStats(), tt.want.fullStats)
		}
	}
}

func TestSystemConf(t *testing.T) {
	SystemConf()
}
