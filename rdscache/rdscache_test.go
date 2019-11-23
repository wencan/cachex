package rdscache

// wencan
// 2017-08-31

import (
	"context"
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

	ctx := context.Background()

	cache := NewRdsCache(ctx, "tcp", s.Addr(), PoolConfig{DB: 1})
	assert.Implements(t, (*cachex.Storage)(nil), cache)

	err = cache.Set(ctx, "exists", "exists")
	if assert.NoError(t, err) {
		var value string
		err = cache.Get(ctx, "exists", &value)
		assert.NoError(t, err)
		assert.Equal(t, "exists", value)
	}

	var value string
	err = cache.Get(ctx, "non-exists", &value)
	assert.Implements(t, (*cachex.NotFound)(nil), err)
}

func TestRdsCacheWithPool(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	ctx := context.Background()

	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", s.Addr())
		},
	}

	cache := NewRdsCacheWithPool(pool)
	assert.Implements(t, (*cachex.Storage)(nil), cache)

	err = cache.Set(ctx, "exists", "exists")
	if assert.NoError(t, err) {
		var value string
		err = cache.Get(ctx, "exists", &value)
		assert.NoError(t, err)
		assert.Equal(t, "exists", value)
	}

	var value string
	err = cache.Get(ctx, "non-exists", &value)
	assert.Implements(t, (*cachex.NotFound)(nil), err)
}

func TestRdsCacheExpire(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	ctx := context.Background()

	cache := NewRdsCache(ctx, "tcp", s.Addr(), PoolConfig{}, RdsDefaultTTLOption(time.Millisecond*100))
	assert.Implements(t, (*cachex.Storage)(nil), cache)

	err = cache.Set(ctx, "exists", "exists")
	if assert.NoError(t, err) {
		var value string
		err = cache.Get(ctx, "exists", &value)
		assert.NoError(t, err)
		assert.Equal(t, "exists", value)
	}

	s.FastForward(time.Millisecond * 100)

	var value string
	err = cache.Get(ctx, "exists", &value)
	assert.Implements(t, (*cachex.NotFound)(nil), err)
}

func TestRdsCacheDel(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	ctx := context.Background()

	cache := NewRdsCache(ctx, "tcp", s.Addr(), PoolConfig{DB: 1}, RdsDefaultTTLOption(time.Millisecond*100))
	assert.Implements(t, (*cachex.Storage)(nil), cache)

	err = cache.Set(ctx, "exists", "exists")
	if assert.NoError(t, err) {
		var value string
		err = cache.Get(ctx, "exists", &value)
		assert.NoError(t, err)
		assert.Equal(t, "exists", value)
	}

	err = cache.Del(ctx, "exists")
	if !assert.NoError(t, err) {
		var value string
		err = cache.Get(ctx, "exists", &value)
		assert.Implements(t, (*cachex.NotFound)(nil), err)
	}
}

func TestRdsCacheKeyPrefix(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	ctx := context.Background()

	// 测试无前缀
	cacheWithoutPrefix := NewRdsCache(ctx, "tcp", s.Addr(), PoolConfig{DB: 1})
	assert.Implements(t, (*cachex.Storage)(nil), cacheWithoutPrefix)

	err = cacheWithoutPrefix.Set(ctx, "exists", "exists-withoutPrefix")
	if assert.NoError(t, err) {
		_, err := s.DB(1).Get("exists")
		assert.NoError(t, err)
	}

	// 测试有前缀
	s.DB(1).FlushDB()
	keyPrefix := "prefix"

	cacheWithPrefix := NewRdsCache(ctx, "tcp", s.Addr(), PoolConfig{DB: 1}, RdsKeyPrefixOption(keyPrefix))
	assert.Implements(t, (*cachex.Storage)(nil), cacheWithPrefix)

	err = cacheWithPrefix.Set(ctx, "exists", "exists-withPrefix")
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

	ctx := context.Background()

	cache := NewRdsCache(ctx, "tcp", s.Addr(), PoolConfig{DB: 1})
	assert.Implements(t, (*cachex.Storage)(nil), cache)

	exists := testStringer{SKey: "exists"}
	err = cache.Set(ctx, exists, "exists")
	if assert.NoError(t, err) {
		_, err := s.DB(1).Get(exists.String())
		assert.NoError(t, err)
	}
}
