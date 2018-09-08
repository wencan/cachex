package cachex

// wencan
// 2017-08-31

import (
	"sync"
	"time"
)

type cacheEntry struct {
	value   interface{}
	created int64
}

type LRUCache struct {
	MaxEntries int
	TTLSeconds int

	Mapping *ListMap

	lock AtomicMutex

	entryPool sync.Pool
}

func NewLRUCache(maxEntries int, TTLSeconds int) *LRUCache {
	return &LRUCache{
		MaxEntries: maxEntries,
		TTLSeconds: TTLSeconds,
		Mapping:    NewListMap(),
		entryPool: sync.Pool{
			New: func() interface{} {
				return &cacheEntry{}
			},
		},
	}
}

func (c *LRUCache) Set(key, value interface{}) (err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	item, ok := c.Mapping.Get(key)
	if ok {
		entry := item.(*cacheEntry)
		entry.value = value
		entry.created = time.Now().Unix()

		c.Mapping.MoveToFront(key)
	} else {
		entry := c.entryPool.Get().(*cacheEntry)
		entry.value = value
		entry.created = time.Now().Unix()

		c.Mapping.PushFront(key, entry)

		if c.MaxEntries > 0 {
			for c.Mapping.Len() > c.MaxEntries {
				entry, _ := c.Mapping.PopBack()
				if entry != nil {
					c.entryPool.Put(entry)
				}
			}
		}
	}

	return nil
}

func (c *LRUCache) Get(key interface{}) (value interface{}, ok bool, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	item, ok := c.Mapping.Get(key)
	if ok {
		entry := item.(*cacheEntry)
		if c.TTLSeconds != 0 {
			if int(time.Now().Unix()-entry.created) >= c.TTLSeconds {
				c.Mapping.Pop(key)
				c.entryPool.Put(entry)
				return nil, false, nil
			}
		}

		c.Mapping.MoveToFront(key)
		return entry.value, true, nil
	}

	return nil, false, nil
}

func (c *LRUCache) Remove(key interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	entry, _ := c.Mapping.Pop(key)
	if entry != nil {
		c.entryPool.Put(entry)
	}
}

func (c *LRUCache) Del(key interface{}) (err error) {
	c.Remove(key)
	return nil
}

func (c *LRUCache) Len() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.Mapping.Len()
}

func (c *LRUCache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	for c.Mapping.Len() != 0 {
		entry, _ := c.Mapping.PopBack()
		if entry != nil {
			c.entryPool.Put(entry)
		}
	}
}
