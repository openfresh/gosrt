package gosrt

import (
	"context"
	"net"
	"sync"
)

type ipStackCapabilities struct {
	sync.Once             // guards following
	ipv4Enabled           bool
	ipv6Enabled           bool
	ipv4MappedIPv6Enabled bool
}

var ipStackCaps ipStackCapabilities

// supportsIPv4 reports whether the platform supports IPv4 networking
// functionality.
func supportsIPv4() bool {
	ipStackCaps.Once.Do(ipStackCaps.probe)
	return ipStackCaps.ipv4Enabled
}

// supportsIPv6 reports whether the platform supports IPv6 networking
// functionality.
func supportsIPv6() bool {
	ipStackCaps.Once.Do(ipStackCaps.probe)
	return ipStackCaps.ipv6Enabled
}

// supportsIPv4map reports whether the platform supports mapping an
// IPv4 address inside an IPv6 address at transport layer
// protocols. See RFC 4291, RFC 4038 and RFC 3493.
func supportsIPv4map() bool {
	ipStackCaps.Once.Do(ipStackCaps.probe)
	return ipStackCaps.ipv4MappedIPv6Enabled
}

// An addrList represents a list of network endpoint addresses.
type addrList []net.Addr

// isIPv4 reports whether addr contains an IPv4 address.
func isIPv4(addr net.Addr) bool {
	switch addr := addr.(type) {
	case *SRTAddr:
		return addr.IP.To4() != nil
	}
	return false
}

// isNotIPv4 reports whether addr does not contain an IPv4 address.
func isNotIPv4(addr net.Addr) bool { return !isIPv4(addr) }

// forResolve returns the most appropriate address in address for
// a call to ResolveTCPAddr, ResolveUDPAddr, or ResolveIPAddr.
// IPv4 is preferred, unless addr contains an IPv6 literal.
func (addrs addrList) forResolve(network, addr string) net.Addr {
	var want6 bool
	switch network {
	case "srt":
		// IPv6 literal. (addr contains a port, so look for '[')
		want6 = count(addr, '[') > 0
	}
	if want6 {
		return addrs.first(isNotIPv4)
	}
	return addrs.first(isIPv4)
}

// first returns the first address which satisfies strategy, or if
// none do, then the first address of any kind.
func (addrs addrList) first(strategy func(net.Addr) bool) net.Addr {
	for _, addr := range addrs {
		if strategy(addr) {
			return addr
		}
	}
	return addrs[0]
}

// partition divides an address list into two categories, using a
// strategy function to assign a boolean label to each address.
// The first address, and any with a matching label, are returned as
// primaries, while addresses with the opposite label are returned
// as fallbacks. For non-empty inputs, primaries is guaranteed to be
// non-empty.
func (addrs addrList) partition(strategy func(net.Addr) bool) (primaries, fallbacks addrList) {
	var primaryLabel bool
	for i, addr := range addrs {
		label := strategy(addr)
		if i == 0 || label == primaryLabel {
			primaryLabel = label
			primaries = append(primaries, addr)
		} else {
			fallbacks = append(fallbacks, addr)
		}
	}
	return
}

// filterAddrList applies a filter to a list of IP addresses,
// yielding a list of Addr objects. Known filters are nil, ipv4only,
// and ipv6only. It returns every address when the filter is nil.
// The result contains at least one address when error is nil.
func filterAddrList(filter func(net.IPAddr) bool, ips []net.IPAddr, inetaddr func(net.IPAddr) net.Addr, originalAddr string) (addrList, error) {
	var addrs addrList
	for _, ip := range ips {
		if filter == nil || filter(ip) {
			addrs = append(addrs, inetaddr(ip))
		}
	}
	if len(addrs) == 0 {
		return nil, &net.AddrError{Err: errNoSuitableAddress.Error(), Addr: originalAddr}
	}
	return addrs, nil
}

// ipv4only reports whether addr is an IPv4 address.
func ipv4only(addr net.IPAddr) bool {
	return addr.IP.To4() != nil
}

// ipv6only reports whether addr is an IPv6 address except IPv4-mapped IPv6 address.
func ipv6only(addr net.IPAddr) bool {
	return len(addr.IP) == net.IPv6len && addr.IP.To4() == nil
}

func splitHostZone(s string) (host, zone string) {
	// The IPv6 scoped addressing zone identifier starts after the
	// last percent sign.
	if i := last(s, '%'); i > 0 {
		host, zone = s[:i], s[i+1:]
	} else {
		host = s
	}
	return
}

// internetAddrList resolves addr, which may be a literal IP
// address or a DNS name, and returns a list of internet protocol
// family addresses. The result contains at least one address when
// error is nil.
func (r *Resolver) internetAddrList(ctx context.Context, network, addr string) (addrList, error) {
	var (
		err        error
		host, port string
		portnum    int
	)
	switch network {
	case "srt", "srt4", "srt6":
		if addr != "" {
			if host, port, err = net.SplitHostPort(addr); err != nil {
				return nil, err
			}
			if portnum, err = r.LookupPort(ctx, network, port); err != nil {
				return nil, err
			}
		}
	default:
		return nil, net.UnknownNetworkError(network)
	}
	inetaddr := func(ip net.IPAddr) net.Addr {
		switch network {
		case "srt", "srt4", "srt6":
			return &SRTAddr{IP: ip.IP, Port: portnum, Zone: ip.Zone}
		default:
			panic("unexpected network: " + network)
		}
	}
	if host == "" {
		return addrList{inetaddr(net.IPAddr{})}, nil
	}

	// Try as a literal IP address, then as a DNS name.
	var ips []net.IPAddr
	if ip := parseIPv4(host); ip != nil {
		ips = []net.IPAddr{{IP: ip}}
	} else if ip, zone := parseIPv6(host, true); ip != nil {
		ips = []net.IPAddr{{IP: ip, Zone: zone}}
		// Issue 18806: if the machine has halfway configured
		// IPv6 such that it can bind on "::" (IPv6unspecified)
		// but not connect back to that same address, fall
		// back to dialing 0.0.0.0.
		if ip.Equal(net.IPv6unspecified) {
			ips = append(ips, net.IPAddr{IP: net.IPv4zero})
		}
	} else {
		// Try as a DNS name.
		ips, err = r.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
	}

	var filter func(net.IPAddr) bool
	if network != "" && network[len(network)-1] == '4' {
		filter = ipv4only
	}
	if network != "" && network[len(network)-1] == '6' {
		filter = ipv6only
	}
	return filterAddrList(filter, ips, inetaddr, host)
}

func loopbackIP(network string) net.IP {
	if network != "" && network[len(network)-1] == '6' {
		return net.IPv6loopback
	}
	return net.IP{127, 0, 0, 1}
}
