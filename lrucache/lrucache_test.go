package lrucache

// wencan
// 2017-08-31

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLRUCacheMaxEntries(t *testing.T) {
	cache := NewLRUCache(10, 0)

	for i := 0; i < 11; i++ {
		err := cache.Set(i, i*i)
		if !assert.NoError(t, err) {
			return
		}
	}

	value, ok, err := cache.Get(5)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 5*5, value)

	value, ok, err = cache.Get(0)
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestLRUCacheExpire(t *testing.T) {
	cache := NewLRUCache(0, time.Millisecond*10)

	key := "test"
	value := "test"
	err := cache.Set(key, value)
	if !assert.NoError(t, err) {
		return
	}

	cached, ok, err := cache.Get(value)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, value, cached)

	time.Sleep(time.Millisecond * 20)

	cached, ok, err = cache.Get(value)
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Equal(t, value, cached) // 支持StaleWhenError
}

func TestLRUCacheLength(t *testing.T) {
	cache := NewLRUCache(10, 0)

	for i := 0; i < 10; i++ {
		err := cache.Set(i, i*i)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, i+1, cache.Len())
	}

	err := cache.Set("test", "test")
	assert.NoError(t, err)
	assert.Equal(t, 10, cache.Len())

	cache.Clear()
	assert.Equal(t, 0, cache.Len())
}

func TestLRUCacheDel(t *testing.T) {
	cache := NewLRUCache(0, time.Second)

	key := "test"
	value := "test"
	err := cache.Set(key, value)
	if !assert.NoError(t, err) {
		return
	}

	cached, ok, err := cache.Get(value)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, value, cached)

	err = cache.Del(key)
	assert.NoError(t, err)

	cached, ok, err = cache.Get(value)
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Nil(t, cached)
}
