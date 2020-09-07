package cachex

// wencan
// 2017-09-02 10:48

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSentinel_Wait(t *testing.T) {
	sentinel := NewSentinel()

	var sum int

	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var value int
			err := sentinel.Wait(context.TODO(), &value)
			assert.NoError(t, err)
			assert.Equal(t, 1, value)

			mu.Lock()
			defer mu.Unlock()
			sum += value
		}()
	}

	err := sentinel.Done(1, nil)
	if !assert.NoError(t, err) {
		return
	}

	wg.Wait()

	assert.Equal(t, 10, sum)
}
