// +build darwin dragonfly freebsd netbsd openbsd

package gosrt

import (
	"syscall"

	"golang.org/x/net/route"
)

func interfaceMessages(ifindex int) ([]route.Message, error) {
	rib, err := route.FetchRIB(syscall.AF_UNSPEC, syscall.NET_RT_IFLIST, ifindex)
	if err != nil {
		return nil, err
	}
	return route.ParseRIB(syscall.NET_RT_IFLIST, rib)
}
