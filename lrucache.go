package cachex

// wencan
// 2017-08-31

import (
	"time"
)

type cacheEntry struct {
	value   interface{}
	created int64
}

type LRUCache struct {
	MaxEntries int
	TTLSeconds int

	Mapping    *ListMap

	lock       AtomicMutex
}

func NewLRUCache(maxEntries int, TTLSeconds int) *LRUCache {
	return &LRUCache{
		MaxEntries: maxEntries,
		TTLSeconds: TTLSeconds,
		Mapping: NewListMap(),
	}
}

func (c *LRUCache) Set(key, value interface{}) (err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	item, ok := c.Mapping.Get(key)
	if ok {
		entry := item.(cacheEntry)
		entry.value = value
		entry.created = time.Now().Unix()

		c.Mapping.MoveToFront(key)
	} else {
		entry := &cacheEntry{
			value: value,
			created: time.Now().Unix(),
		}

		len := c.Mapping.Len()

		c.Mapping.PushFront(key, entry)

		if c.MaxEntries != 0 && len >= c.MaxEntries {
			c.Mapping.PopBack()
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
			if int(time.Now().Unix() - entry.created) >= c.TTLSeconds {
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

	c.Mapping.Pop(key)
}

func (c *LRUCache) Len() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.Mapping.Len()
}

func (c *LRUCache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.Mapping.Clear()
}