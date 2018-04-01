package config

import (
	"github.com/openfresh/gosrt/srtapi"
)

// config
var (
	Verbose     = false
	LogLevel    = srtapi.LogError
	LogFas      = []int{srtapi.LogFAGeneral}
	LogFile     = ""
	LogInternal = false
	FullStats   = false
)
