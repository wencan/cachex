package cachex

// wencan
// 2017-09-01 16:31

import (
	"runtime"
	"sync/atomic"
)

type AtomicMutex struct {
	bay uint32
}

func (mu *AtomicMutex) Lock() {
	for {
		if atomic.CompareAndSwapUint32(&mu.bay, 0, 1) {
			break
		} else {
			runtime.Gosched()
		}
	}
}

func (mu *AtomicMutex) Unlock() {
	atomic.StoreUint32(&mu.bay, 0)
}