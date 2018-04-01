package testenv

import (
	"testing"
)

// MustHaveExternalNetwork checks that the current system can use
// external (non-localhost) networks.
// If not, MustHaveExternalNetwork calls t.Skip with an explanation.
func MustHaveExternalNetwork(t testing.TB) {
	if testing.Short() {
		t.Skipf("skipping test: no external network in -short mode")
	}
}
