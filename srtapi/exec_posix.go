// Copyright (c) 2018 CyberAgent, Inc. All rights reserved.
// https://github.com/openfresh/gosrt

package srtapi

// SetNonblock set non-blocking mode
func SetNonblock(fd int, nonblocking bool) (err error) {
	val := 1
	if nonblocking {
		val = 0
	}
	if err = SetsockoptInt(fd, 0, OptionSndsyn, val); err != nil {
		return err
	}
	if err = SetsockoptInt(fd, 0, OptionRcvsyn, val); err != nil {
		return err
	}
	return err
}
