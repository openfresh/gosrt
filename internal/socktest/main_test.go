package socktest_test

import (
	"os"
	"sync"
	"syscall"
	"testing"

	socktest "github.com/openfresh/gosrt/internal/socktest"
)

var sw socktest.Switch

func TestMain(m *testing.M) {
	installTestHooks()

	st := m.Run()

	for s := range sw.Sockets() {
		closeFunc(s)
	}
	uninstallTestHooks()
	os.Exit(st)
}

func TestSwitch(t *testing.T) {
	const N = 10
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			for _, family := range []int{syscall.AF_INET, syscall.AF_INET6} {
				socketFunc(family, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
			}
		}()
	}
	wg.Wait()
}

func TestSocket(t *testing.T) {
	for _, f := range []socktest.Filter{
		func(st *socktest.Status) (socktest.AfterFilter, error) { return nil, nil },
		nil,
	} {
		sw.Set(socktest.FilterSocket, f)
		for _, family := range []int{syscall.AF_INET, syscall.AF_INET6} {
			socketFunc(family, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
		}
	}
}
