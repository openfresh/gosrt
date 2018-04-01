// +build darwin dragonfly freebsd netbsd openbsd

package gosrt

import (
	"net"
	"syscall"

	"golang.org/x/net/route"
)

// If the ifindex is zero, interfaceTable returns mappings of all
// network interfaces. Otherwise it returns a mapping of a specific
// interface.
func interfaceTable(ifindex int) ([]net.Interface, error) {
	msgs, err := interfaceMessages(ifindex)
	if err != nil {
		return nil, err
	}
	n := len(msgs)
	if ifindex != 0 {
		n = 1
	}
	ift := make([]net.Interface, n)
	n = 0
	for _, m := range msgs {
		switch m := m.(type) {
		case *route.InterfaceMessage:
			if ifindex != 0 && ifindex != m.Index {
				continue
			}
			ift[n].Index = m.Index
			ift[n].Name = m.Name
			ift[n].Flags = linkFlags(m.Flags)
			if sa, ok := m.Addrs[syscall.RTAX_IFP].(*route.LinkAddr); ok && len(sa.Addr) > 0 {
				ift[n].HardwareAddr = make([]byte, len(sa.Addr))
				copy(ift[n].HardwareAddr, sa.Addr)
			}
			for _, sys := range m.Sys() {
				if imx, ok := sys.(*route.InterfaceMetrics); ok {
					ift[n].MTU = imx.MTU
					break
				}
			}
			n++
			if ifindex == m.Index {
				return ift[:n], nil
			}
		}
	}
	return ift[:n], nil
}

func linkFlags(rawFlags int) net.Flags {
	var f net.Flags
	if rawFlags&syscall.IFF_UP != 0 {
		f |= net.FlagUp
	}
	if rawFlags&syscall.IFF_BROADCAST != 0 {
		f |= net.FlagBroadcast
	}
	if rawFlags&syscall.IFF_LOOPBACK != 0 {
		f |= net.FlagLoopback
	}
	if rawFlags&syscall.IFF_POINTOPOINT != 0 {
		f |= net.FlagPointToPoint
	}
	if rawFlags&syscall.IFF_MULTICAST != 0 {
		f |= net.FlagMulticast
	}
	return f
}
