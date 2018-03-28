package gosrt

import (
	"net"
)

// ipEmptyString is like ip.String except that it returns
// an empty string when ip is unset.
func ipEmptyString(ip net.IP) string {
	if len(ip) == 0 {
		return ""
	}
	return ip.String()
}
