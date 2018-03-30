package srtapi

func SetNonblock(fd int, nonblocking bool) (err error) {
	val := 1
	if nonblocking {
		val = 0
	}
	if err = SetsockoptInt(fd, 0, SRTO_SNDSYN, val); err != nil {
		return err
	}
	if err = SetsockoptInt(fd, 0, SRTO_RCVSYN, val); err != nil {
		return err
	}
	return err
}
