package rdscache

// wencan
// 2017-08-31

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis"

	"github.com/stretchr/testify/assert"
	"github.com/wencan/cachex"
)

func TestRdsCache(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	cache := NewRdsCache("tcp", s.Addr(), RdsDB(1), RdsTTL(time.Second/10))
	assert.Implements(t, (*cachex.Storage)(nil), cache)

	err = cache.Set("exits", "exits")
	if assert.NoError(t, err) {
		var value string
		err = cache.Get("exits", &value)
		assert.NoError(t, err)
		assert.Equal(t, "exits", value)
	}

	time.Sleep(time.Second / 10)

	var value string
	err = cache.Get("exists", &value)
	assert.Implements(t, (*cachex.NotFound)(nil), err)

	err = cache.Get("non-exists", &value)
	assert.Implements(t, (*cachex.NotFound)(nil), err)
}
