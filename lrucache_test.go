package cachex

// wencan
// 2017-08-31

import (
	"testing"
	"time"
)

func TestCacheMaxEntries(t *testing.T) {
	cache := NewCache(10, 0)

	for i :=0; i<11; i++ {
		cache.Set(i, i*i)
	}

	value, ok := cache.Get(5)
	if !ok {
		t.Fatal("not found value by key:", 5)
	} else if value != 5*5 {
		t.Fatal("cached value missmatch")
	}

	value, ok = cache.Get(0)
	if ok {
		t.Fatal("found value by key:", 0)
	}
}

func TestCacheExpire(t *testing.T) {
	cache := NewCache(0, 1)

	key := "test"
	value := "test"
	cache.Set(key, value)

	cached, ok := cache.Get(value)
	if !ok {
		t.Fatal("not found value by key:", key)
	} else if cached != value {
		t.Fatal("cached value missmatch")
	}

	time.Sleep(time.Second * 2)

	_, ok = cache.Get(value)
	if ok {
		t.Fatal("found value by key:", key)
	}
}

func TestCacheLength(t *testing.T) {
	cache := NewCache(10, 0)

	for i :=0; i<10; i++ {
		cache.Set(i, i*i)

		if cache.Len() != i+1 {
			t.Fatal("cache length error")
		}
	}

	cache.Set("test", "test")
	if cache.Len() != 10 {
		t.Fatal("cache length error")
	}

	cache.Clear()
	if cache.Len() != 0 {
		t.Fatal("cache length error")
	}
}
