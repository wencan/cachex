package cachex

// wencan
// 2017-08-31

import (
	"testing"
	"time"
)

func TestLRUCacheMaxEntries(t *testing.T) {
	cache := NewLRUCache(10, 0)

	for i := 0; i < 11; i++ {
		err := cache.Set(i, i*i)
		if err != nil {
			t.Fatal(err)
		}
	}

	value, ok, err := cache.Get(5)
	if err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal("not found value by key:", 5)
	} else if value != 5*5 {
		t.Fatal("cached value missmatch")
	}

	value, ok, err = cache.Get(0)
	if err != nil {
		t.Fatal(err)
	} else if ok {
		t.Fatal("found value by key:", 0)
	}
}

func TestLRUCacheExpire(t *testing.T) {
	cache := NewLRUCache(0, time.Second)

	key := "test"
	value := "test"
	err := cache.Set(key, value)
	if err != nil {
		t.Fatal(err)
	}

	cached, ok, err := cache.Get(value)
	if err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal("not found value by key:", key)
	} else if cached != value {
		t.Fatal("cached value missmatch")
	}

	time.Sleep(time.Second)

	_, ok, err = cache.Get(value)
	if err != nil {
		t.Fatal(err)
	} else if ok {
		t.Fatal("found value by key:", key)
	}
}

func TestLRUCacheLength(t *testing.T) {
	cache := NewLRUCache(10, 0)

	for i := 0; i < 10; i++ {
		err := cache.Set(i, i*i)
		if err != nil {
			t.Fatal(err)
		}

		if cache.Len() != i+1 {
			t.Fatal("cache length error")
		}
	}

	err := cache.Set("test", "test")
	if err != nil {
		t.Fatal(err)
	} else if cache.Len() != 10 {
		t.Fatal("cache length error")
	}

	cache.Clear()
	if cache.Len() != 0 {
		t.Fatal("cache length error")
	}
}

func TestLRUCacheDel(t *testing.T) {
	cache := NewLRUCache(0, time.Second)

	key := "test"
	value := "test"
	err := cache.Set(key, value)
	if err != nil {
		t.Fatal(err)
	}

	cached, ok, err := cache.Get(value)
	if err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal("not found value by key:", key)
	} else if cached != value {
		t.Fatal("cached value missmatch")
	}

	err = cache.Del(key)
	if err != nil {
		t.Fatal(err)
	}

	_, ok, err = cache.Get(value)
	if err != nil {
		t.Fatal(err)
	} else if ok {
		t.Fatal("found value by key:", key)
	}
}
