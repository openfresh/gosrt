package gosrt

import (
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"

	socktest "github.com/openfresh/gosrt/internal/socktest"
)

var sw socktest.Switch

var (
	// uninstallTestHooks runs just before a run of benchmarks.
	testHookUninstaller sync.Once
)

var (
	testSRTBig = flag.Bool("srtbig", false, "whether to test massive size of data per read or write call on SRT connection")

	testDNSFlood = flag.Bool("dnsflood", false, "whether to test DNS query flooding")

	// If external IPv4 connectivity exists, we can try dialing
	// non-node/interface local scope IPv4 addresses.
	// On Windows, Lookup APIs may not return IPv4-related
	// resource records when a node has no external IPv4
	// connectivity.
	testIPv4 = flag.Bool("ipv4", true, "assume external IPv4 connectivity exists")

	// If external IPv6 connectivity exists, we can try dialing
	// non-node/interface local scope IPv6 addresses.
	// On Windows, Lookup APIs may not return IPv6-related
	// resource records when a node has no external IPv6
	// connectivity.
	testIPv6 = flag.Bool("ipv6", false, "assume external IPv6 connectivity exists")
)

func TestMain(m *testing.M) {
	setupTestData()
	installTestHooks()

	st := m.Run()

	testHookUninstaller.Do(uninstallTestHooks)
	if testing.Verbose() {
		printRunningGoroutines()
		printInflightSockets()
		printSocketStats()
	}
	forceCloseSockets()
	Shutdown()
	os.Exit(st)
}

type ipv6LinkLocalUnicastTest struct {
	network, address string
	nameLookup       bool
}

var (
	ipv6LinkLocalUnicastSRTTests []ipv6LinkLocalUnicastTest
)

func setupTestData() {
	if supportsIPv4() {
		resolveSRTAddrTests = append(resolveSRTAddrTests, []resolveSRTAddrTest{
			{"srt", "localhost:1", &SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1}, nil},
			{"srt4", "localhost:2", &SRTAddr{IP: net.IPv4(127, 0, 0, 1), Port: 2}, nil},
		}...)
	}

	if supportsIPv6() {
		resolveSRTAddrTests = append(resolveSRTAddrTests, resolveSRTAddrTest{"srt6", "localhost:3", &SRTAddr{IP: net.IPv6loopback, Port: 3}, nil})

		// Issue 20911: don't return IPv4 addresses for
		// Resolve*Addr calls of the IPv6 unspecified address.
		resolveSRTAddrTests = append(resolveSRTAddrTests, resolveSRTAddrTest{"srt", "[::]:4", &SRTAddr{IP: net.IPv6unspecified, Port: 4}, nil})
	}

	ifi := loopbackInterface()
	if ifi != nil {
		index := fmt.Sprintf("%v", ifi.Index)
		resolveSRTAddrTests = append(resolveSRTAddrTests, []resolveSRTAddrTest{
			{"srt6", "[fe80::1%" + ifi.Name + "]:1", &SRTAddr{IP: net.ParseIP("fe80::1"), Port: 1, Zone: zoneCache.name(ifi.Index)}, nil},
			{"srt6", "[fe80::1%" + index + "]:2", &SRTAddr{IP: net.ParseIP("fe80::1"), Port: 2, Zone: index}, nil},
		}...)
	}

	addr := ipv6LinkLocalUnicastAddr(ifi)
	if addr != "" {
		if runtime.GOOS != "dragonfly" {
			ipv6LinkLocalUnicastSRTTests = append(ipv6LinkLocalUnicastSRTTests, []ipv6LinkLocalUnicastTest{
				{"srt", "[" + addr + "%" + ifi.Name + "]:0", false},
			}...)
		}
		ipv6LinkLocalUnicastSRTTests = append(ipv6LinkLocalUnicastSRTTests, []ipv6LinkLocalUnicastTest{
			{"srt6", "[" + addr + "%" + ifi.Name + "]:0", false},
		}...)
		switch runtime.GOOS {
		case "darwin", "dragonfly", "freebsd", "openbsd", "netbsd":
			ipv6LinkLocalUnicastSRTTests = append(ipv6LinkLocalUnicastSRTTests, []ipv6LinkLocalUnicastTest{
				{"srt", "[localhost%" + ifi.Name + "]:0", true},
				{"srt6", "[localhost%" + ifi.Name + "]:0", true},
			}...)
		case "linux":
			ipv6LinkLocalUnicastSRTTests = append(ipv6LinkLocalUnicastSRTTests, []ipv6LinkLocalUnicastTest{
				{"srt", "[ip6-localhost%" + ifi.Name + "]:0", true},
				{"srt6", "[ip6-localhost%" + ifi.Name + "]:0", true},
			}...)
		}
	}
}

func printRunningGoroutines() {
	gss := runningGoroutines()
	if len(gss) == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "Running goroutines:\n")
	for _, gs := range gss {
		fmt.Fprintf(os.Stderr, "%v\n", gs)
	}
	fmt.Fprintf(os.Stderr, "\n")
}

// runningGoroutines returns a list of remaining goroutines.
func runningGoroutines() []string {
	var gss []string
	b := make([]byte, 2<<20)
	b = b[:runtime.Stack(b, true)]
	for _, s := range strings.Split(string(b), "\n\n") {
		ss := strings.SplitN(s, "\n", 2)
		if len(ss) != 2 {
			continue
		}
		stack := strings.TrimSpace(ss[1])
		if !strings.Contains(stack, "created by net") {
			continue
		}
		gss = append(gss, stack)
	}
	sort.Strings(gss)
	return gss
}

func printInflightSockets() {
	sos := sw.Sockets()
	if len(sos) == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "Inflight sockets:\n")
	for s, so := range sos {
		fmt.Fprintf(os.Stderr, "%v: %v\n", s, so)
	}
	fmt.Fprintf(os.Stderr, "\n")
}

func printSocketStats() {
	sts := sw.Stats()
	if len(sts) == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "Socket statistical information:\n")
	for _, st := range sts {
		fmt.Fprintf(os.Stderr, "%v\n", st)
	}
	fmt.Fprintf(os.Stderr, "\n")
}
