package cachex

// wencan
// 2017-09-01 16:33

import (
	"testing"
	"sync"
)

func TestAtomicMutex_Lock(t *testing.T) {
	var number uint64

	var lock AtomicMutex
	var wg sync.WaitGroup
	for i:=0; i<1000; i++ {
		wg.Add(1)
		go func() {
			for j:=0; j<1000; j++ {
				lock.Lock()
				number += 1
				lock.Unlock()
			}
			wg.Done()
		}()
	}

	wg.Wait()

	if number != 1000*1000 {
		t.FailNow()
	}
}