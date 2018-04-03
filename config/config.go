// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

package config

import (
	"github.com/openfresh/gosrt/srtapi"
)

// config
var (
	Verbose  = false
	LogLevel = srtapi.LogError
	//LogLevel    = srtapi.LogDebug
	LogFas = []int{}
	//LogFas      = []int{srtapi.LogFABstats, srtapi.LogFAControl, srtapi.LogFAData, srtapi.LogFATsbpd, srtapi.LogFARexmit}
	LogFile     = ""
	LogInternal = false
	FullStats   = false
)
