package gosrt

import "syscall"

func maxListenerBacklog() int {
	fd, err := open("/proc/sys/net/core/somaxconn")
	if err != nil {
		return syscall.SOMAXCONN
	}
	defer fd.close()
	l, ok := fd.readLine()
	if !ok {
		return syscall.SOMAXCONN
	}
	f := getFields(l)
	n, _, ok := dtoi(f[0])
	if n == 0 || !ok {
		return syscall.SOMAXCONN
	}
	// Linux stores the backlog in a uint16.
	// Truncate number to avoid wrapping.
	// See issue 5030.
	if n > 1<<16-1 {
		n = 1<<16 - 1
	}
	return n
}
