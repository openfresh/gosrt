package gosrt

import (
	"sync"

	socktest "github.com/openfresh/gosrt/internal/socktest"
)

var sw socktest.Switch

var (
	// uninstallTestHooks runs just before a run of benchmarks.
	testHookUninstaller sync.Once
)
