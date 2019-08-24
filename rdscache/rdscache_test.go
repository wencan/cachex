package rdscache

// wencan
// 2017-08-31

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/wencan/cachex"
)

func TestRdsCache(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	cache := NewRdsCache("tcp", s.Addr(), PoolConfig{DB: 1})
	assert.Implements(t, (*cachex.Storage)(nil), cache)

	err = cache.Set("exists", "exists")
	if assert.NoError(t, err) {
		var value string
		err = cache.Get("exists", &value)
		assert.NoError(t, err)
		assert.Equal(t, "exists", value)
	}

	var value string
	err = cache.Get("non-exists", &value)
	assert.Implements(t, (*cachex.NotFound)(nil), err)
}

func TestRdsCacheWithPool(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", s.Addr())
		},
	}

	cache := NewRdsCacheWithPool(pool)
	assert.Implements(t, (*cachex.Storage)(nil), cache)

	err = cache.Set("exists", "exists")
	if assert.NoError(t, err) {
		var value string
		err = cache.Get("exists", &value)
		assert.NoError(t, err)
		assert.Equal(t, "exists", value)
	}

	var value string
	err = cache.Get("non-exists", &value)
	assert.Implements(t, (*cachex.NotFound)(nil), err)
}

func TestRdsCacheExpire(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	cache := NewRdsCache("tcp", s.Addr(), PoolConfig{}, RdsDefaultTTLOption(time.Millisecond*100))
	assert.Implements(t, (*cachex.Storage)(nil), cache)

	err = cache.Set("exists", "exists")
	if assert.NoError(t, err) {
		var value string
		err = cache.Get("exists", &value)
		assert.NoError(t, err)
		assert.Equal(t, "exists", value)
	}

	s.FastForward(time.Millisecond * 100)

	var value string
	err = cache.Get("exists", &value)
	assert.Implements(t, (*cachex.NotFound)(nil), err)
}

func TestRdsCacheDel(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	cache := NewRdsCache("tcp", s.Addr(), PoolConfig{DB: 1}, RdsDefaultTTLOption(time.Millisecond*100))
	assert.Implements(t, (*cachex.Storage)(nil), cache)

	err = cache.Set("exists", "exists")
	if assert.NoError(t, err) {
		var value string
		err = cache.Get("exists", &value)
		assert.NoError(t, err)
		assert.Equal(t, "exists", value)
	}

	err = cache.Del("exists")
	if !assert.NoError(t, err) {
		var value string
		err = cache.Get("exists", &value)
		assert.Implements(t, (*cachex.NotFound)(nil), err)
	}
}

func TestRdsCacheKeyPrefix(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	// 测试无前缀

	cacheWithoutPrefix := NewRdsCache("tcp", s.Addr(), PoolConfig{DB: 1})
	assert.Implements(t, (*cachex.Storage)(nil), cacheWithoutPrefix)

	err = cacheWithoutPrefix.Set("exists", "exists-withoutPrefix")
	if assert.NoError(t, err) {
		_, err := s.DB(1).Get("exists")
		assert.NoError(t, err)
	}

	// 测试有前缀
	s.DB(1).FlushDB()
	keyPrefix := "prefix"

	cacheWithPrefix := NewRdsCache("tcp", s.Addr(), PoolConfig{DB: 1}, RdsKeyPrefixOption(keyPrefix))
	assert.Implements(t, (*cachex.Storage)(nil), cacheWithPrefix)

	err = cacheWithPrefix.Set("exists", "exists-withPrefix")
	if assert.NoError(t, err) {
		_, err := s.DB(1).Get("prefix:exists")
		assert.NoError(t, err)
	}
}

type testStringer struct {
	SKey string
}

func (stringer testStringer) String() string {
	return stringer.SKey
}

func TestRdsCacheStringerKey(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	cache := NewRdsCache("tcp", s.Addr(), PoolConfig{DB: 1})
	assert.Implements(t, (*cachex.Storage)(nil), cache)

	exists := testStringer{SKey: "exists"}
	err = cache.Set(exists, "exists")
	if assert.NoError(t, err) {
		_, err := s.DB(1).Get(exists.String())
		assert.NoError(t, err)
	}
}
