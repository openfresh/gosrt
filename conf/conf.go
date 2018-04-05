// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package conf

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/openfresh/gosrt/srtapi"
)

// Conf represents a system's network configuration.
type Conf struct {
	goos string // the runtime.GOOS, to ease testing

	verbose     bool
	logLevel    int
	logFAs      []int
	logFile     string
	logInternal bool
	fullStats   bool
}

var (
	confOnce sync.Once // guards init of confVal via initConfVal
	confVal  = &Conf{goos: runtime.GOOS}
)

var logLevels = map[string]int{
	"alert":   srtapi.LogAlert,
	"crit":    srtapi.LogFatal,
	"debug":   srtapi.LogDebug,
	"emerg":   srtapi.LogEmerg,
	"err":     srtapi.LogError,
	"fatal":   srtapi.LogFatal,
	"info":    srtapi.LogInfo,
	"notice":  srtapi.LogNote,
	"note":    srtapi.LogNote,
	"warning": srtapi.LogWarning,
}

var logNames = []string{"general", "bstats", "control", "data", "tsbpd", "rexmit"}

// SystemConf returns the machine's network configuration.
func SystemConf() *Conf {
	confOnce.Do(initConfVal)
	return confVal
}

func initConfVal() {
	confVal.verbose = false
	if env := os.Getenv("SRT_VERBOSE"); env != "" {
		if val, err := strconv.ParseBool(env); err == nil {
			confVal.verbose = val
		}
	}

	confVal.logLevel = srtapi.LogError
	if env := os.Getenv("SRT_LOGLEVEL"); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			confVal.logLevel = val
		} else if val, ok := logLevels[strings.ToLower(env)]; ok {
			confVal.logLevel = val
		}
	}

	confVal.logFAs = []int{}
	if fa := os.Getenv("SRT_LOGFA"); fa != "" {
		if fa == "all" {
			confVal.logFAs = append(confVal.logFAs, srtapi.LogFABstats)
			confVal.logFAs = append(confVal.logFAs, srtapi.LogFAControl)
			confVal.logFAs = append(confVal.logFAs, srtapi.LogFAData)
			confVal.logFAs = append(confVal.logFAs, srtapi.LogFATsbpd)
			confVal.logFAs = append(confVal.logFAs, srtapi.LogFARexmit)
		} else {
			fa = strings.ToLower(fa)
			xfas := strings.Split(fa, ",")
			for _, fa := range xfas {
				nfa := 0
				for ; nfa < len(logNames); nfa++ {
					if fa == logNames[nfa] {
						if nfa > 0 {
							confVal.logFAs = append(confVal.logFAs, nfa)
						}
					}
				}
			}
		}
	}

	confVal.logFile = os.Getenv("SRT_LOGFILE")

	confVal.logInternal = false
	if env := os.Getenv("SRT_LOGINTERNAL"); env != "" {
		if val, err := strconv.ParseBool(env); err == nil {
			confVal.logInternal = val
		}
	}

	confVal.fullStats = false
	if env := os.Getenv("SRT_FULLSTATS"); env != "" {
		if val, err := strconv.ParseBool(env); err == nil {
			confVal.fullStats = val
		}
	}
}

// Verbose reports whether verbose log is enabled
func (c *Conf) Verbose() bool {
	return c.verbose
}

func (c *Conf) LogLevel() int {
	return c.logLevel
}

func (c *Conf) LogFAs() []int {
	return c.logFAs
}

func (c *Conf) LogFile() string {
	return c.logFile
}

func (c *Conf) LogInternal() bool {
	return c.logInternal
}

func (c *Conf) FullStats() bool {
	return c.fullStats
}
