package poll

// Implemented in runtime package.
func runtime_Semacquire(sema *uint32)
func runtime_Semrelease(sema *uint32)

func (fd *FD) readLock() error {
	return nil
}

func (fd *FD) readUnlock() {
}

func (fd *FD) writeLock() error {
	return nil
}

func (fd *FD) writeUnlock() {
}
