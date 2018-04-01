package gosrt

import (
	"syscall"

	"golang.org/x/net/route"
)

func interfaceMessages(ifindex int) ([]route.Message, error) {
	typ := route.RIBType(syscall.NET_RT_IFLISTL)
	rib, err := route.FetchRIB(syscall.AF_UNSPEC, typ, ifindex)
	if err != nil {
		typ = route.RIBType(syscall.NET_RT_IFLIST)
		rib, err = route.FetchRIB(syscall.AF_UNSPEC, typ, ifindex)
	}
	if err != nil {
		return nil, err
	}
	return route.ParseRIB(typ, rib)
}
