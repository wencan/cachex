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
	TTLSeconds int //seconds

	Mapping    *ListMap

	lock       AtomicMutex

	Deleter    func(key interface{}, value interface{})
}

func NewCache(maxEntries int, TTLSeconds int) *LRUCache {
	return &LRUCache{
		MaxEntries: maxEntries,
		TTLSeconds: TTLSeconds,
		Mapping: NewListMap(),
	}
}

func (c *LRUCache) Set(key interface{}, value interface{}) {
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
}

func (c *LRUCache) Get(key interface{}) (value interface{}, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	item, ok := c.Mapping.Get(key)
	if ok {
		entry := item.(*cacheEntry)
		if c.TTLSeconds != 0 {
			if int(time.Now().Unix() - entry.created) >= c.TTLSeconds {
				return nil, false
			}
		}

		c.Mapping.MoveToFront(key)
		return entry.value, true
	}

	return nil, false
}

func (c *LRUCache) Remove(key interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.Mapping.Pop(key)
}

func (c *LRUCache) RemoveOldest() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.Mapping.PopBack()
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