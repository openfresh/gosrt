package poll

import (
	"sync"
)

type fdMutex struct {
	rlock sync.Mutex
	wlock sync.Mutex
}

func (fdmu *fdMutex) init() {
	fdmu.rlock = sync.Mutex{}
	fdmu.wlock = sync.Mutex{}
}

func (fd *FD) readLock() error {
	fd.fdmu.rlock.Lock()
	return nil
}

func (fd *FD) readUnlock() {
	fd.fdmu.rlock.Unlock()
}

func (fd *FD) writeLock() error {
	fd.fdmu.wlock.Lock()
	return nil
}

func (fd *FD) writeUnlock() {
	fd.fdmu.wlock.Unlock()
}
