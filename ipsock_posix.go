// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package gosrt

import (
	"context"
	"net"
	"syscall"

	"github.com/openfresh/gosrt/internal/poll"
	"github.com/openfresh/gosrt/srtapi"
)

// Probe probes IPv4, IPv6 and IPv4-mapped IPv6 communication
// capabilities which are controlled by the IPV6_V6ONLY socket option
// and kernel configuration.
//
// Should we try to use the IPv4 socket interface if we're only
// dealing with IPv4 sockets? As long as the host system understands
// IPv4-mapped IPv6, it's okay to pass IPv4-mapeed IPv6 addresses to
// the IPv6 interface. That simplifies our code and is most
// general. Unfortunately, we need to run on kernels built without
// IPv6 support too. So probe the kernel to figure it out.
func (p *ipStackCapabilities) probe() {
	s, err := srtSocket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		poll.CloseFunc(s)
		p.ipv4Enabled = true
	}
	var probes = []struct {
		laddr SRTAddr
		value int
	}{
		// IPv6 communication capability
		{laddr: SRTAddr{IP: net.ParseIP("::1")}, value: 1},
		// IPv4-mapped IPv6 address communication capability
		{laddr: SRTAddr{IP: net.IPv4(127, 0, 0, 1)}, value: 0},
	}
	for i := range probes {
		s, err := srtSocket(syscall.AF_INET6, syscall.SOCK_DGRAM, 0)
		if err != nil {
			continue
		}
		defer poll.CloseFunc(s)
		sa, err := probes[i].laddr.sockaddr(syscall.AF_INET6)
		if err != nil {
			continue
		}
		if err := srtapi.Bind(s, sa); err != nil {
			continue
		}
		if i == 0 {
			p.ipv6Enabled = true
		} else {
			p.ipv4MappedIPv6Enabled = true
		}
	}
}

// favoriteAddrFamily returns the appropriate address family for the
// given network, laddr, raddr and mode.
//
// If mode indicates "listen" and laddr is a wildcard, we assume that
// the user wants to make a passive-open connection with a wildcard
// address family, both AF_INET and AF_INET6, and a wildcard address
// like the following:
//
//	- A listen for a wildcard communication domain, "srt",
//	  with a wildcard address: If the platform supports
//	  both IPv6 and IPv4-mapped IPv6 communication capabilities,
//	  or does not support IPv4, we use a dual stack, AF_INET6 and
//	  IPV6_V6ONLY=0, wildcard address listen. The dual stack
//	  wildcard address listen may fall back to an IPv6-only,
//	  AF_INET6 and IPV6_V6ONLY=1, wildcard address listen.
//	  Otherwise we prefer an IPv4-only, AF_INET, wildcard address
//	  listen.
//
//	- A listen for a wildcard communication domain, "srt", with an IPv4 wildcard address: same as above.
//
//	- A listen for a wildcard communication domain, "srt", with an IPv6 wildcard address: same as above.
//
//	- A listen for an IPv4 communication domain, "srt4",
//	  with an IPv4 wildcard address: We use an IPv4-only, AF_INET,
//	  wildcard address listen.
//
//	- A listen for an IPv6 communication domain, "srt6",
//	  with an IPv6 wildcard address: We use an IPv6-only, AF_INET6
//	  and IPV6_V6ONLY=1, wildcard address listen.
//
// Otherwise guess: If the addresses are IPv4 then returns AF_INET,
// or else returns AF_INET6. It also returns a boolean value what
// designates IPV6_V6ONLY option.
//
// Note that the latest DragonFly BSD and OpenBSD kernels allow
// neither "net.inet6.ip6.v6only=1" change nor IPPROTO_IPV6 level
// IPV6_V6ONLY socket option setting.
func favoriteAddrFamily(network string, laddr, raddr sockaddr, mode string) (family int, ipv6only bool) {
	switch network[len(network)-1] {
	case '4':
		return syscall.AF_INET, false
	case '6':
		return syscall.AF_INET6, true
	}

	if mode == "listen" && (laddr == nil || laddr.isWildcard()) {
		if supportsIPv4map() || !supportsIPv4() {
			return syscall.AF_INET6, false
		}
		if laddr == nil {
			return syscall.AF_INET, false
		}
		return laddr.family(), false
	}

	if (laddr == nil || laddr.family() == syscall.AF_INET) &&
		(raddr == nil || raddr.family() == syscall.AF_INET) {
		return syscall.AF_INET, false
	}
	return syscall.AF_INET6, false
}

func internetSocket(ctx context.Context, net string, laddr, raddr sockaddr, sotype, proto int, mode string) (fd *netFD, err error) {
	family, ipv6only := favoriteAddrFamily(net, laddr, raddr, mode)
	return socket(ctx, net, family, sotype, proto, ipv6only, laddr, raddr)
}

func ipToSockaddr(family int, ip net.IP, port int, zone string) (syscall.Sockaddr, error) {
	switch family {
	case syscall.AF_INET:
		if len(ip) == 0 {
			ip = net.IPv4zero
		}
		ip4 := ip.To4()
		if ip4 == nil {
			return nil, &net.AddrError{Err: "non-IPv4 address", Addr: ip.String()}
		}
		sa := &syscall.SockaddrInet4{Port: port}
		copy(sa.Addr[:], ip4)
		return sa, nil
	case syscall.AF_INET6:
		// In general, an IP wildcard address, which is either
		// "0.0.0.0" or "::", means the entire IP addressing
		// space. For some historical reason, it is used to
		// specify "any available address" on some operations
		// of IP node.
		//
		// When the IP node supports IPv4-mapped IPv6 address,
		// we allow an listener to listen to the wildcard
		// address of both IP addressing spaces by specifying
		// IPv6 wildcard address.
		if len(ip) == 0 || ip.Equal(net.IPv4zero) {
			ip = net.IPv6zero
		}
		// We accept any IPv6 address including IPv4-mapped
		// IPv6 address.
		ip6 := ip.To16()
		if ip6 == nil {
			return nil, &net.AddrError{Err: "non-IPv6 address", Addr: ip.String()}
		}
		sa := &syscall.SockaddrInet6{Port: port, ZoneId: uint32(zoneCache.index(zone))}
		copy(sa.Addr[:], ip6)
		return sa, nil
	}
	return nil, &net.AddrError{Err: "invalid address family", Addr: ip.String()}
}
