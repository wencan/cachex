package lrucache

// wencan
// 2017-08-31

import (
	"reflect"
	"sync"
	"time"
)

// NotFound 没找到错误
type NotFound struct{}

// NotFound 实现cachex.NotFound错误接口
func (NotFound) NotFound() {}
func (NotFound) Error() string {
	return "not found"
}

// Expired 数据已过期错误
type Expired struct{}

// Expired 实现cachex.Expired错误接口
func (Expired) Expired() {}
func (Expired) Error() string {
	return "expired"
}

var notFound = NotFound{}
var expired = Expired{}

type cacheEntry struct {
	value   interface{}
	created time.Time
}

// LRUCache 本地LRU缓存类，实现了cachex.DeletableStorage接口
type LRUCache struct {
	MaxEntries int
	TTL        time.Duration

	Mapping *ListMap

	lock sync.Mutex

	entryPool sync.Pool
}

// NewLRUCache 新建本地LRU缓存类
func NewLRUCache(maxEntries int, TTL time.Duration) *LRUCache {
	return &LRUCache{
		MaxEntries: maxEntries,
		TTL:        TTL,
		Mapping:    NewListMap(),
		entryPool: sync.Pool{
			New: func() interface{} {
				return &cacheEntry{}
			},
		},
	}
}

// Set 设置缓存数据
func (c *LRUCache) Set(key, value interface{}) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	item, ok := c.Mapping.Get(key)
	if ok {
		entry := item.(*cacheEntry)
		entry.value = value
		entry.created = time.Now()

		c.Mapping.MoveToFront(key)
	} else {
		entry := c.entryPool.Get().(*cacheEntry)
		entry.value = value
		entry.created = time.Now()

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

// Get 获取缓存数据
func (c *LRUCache) Get(key, value interface{}) error {
	if v := reflect.ValueOf(value); v.Kind() != reflect.Ptr || v.IsNil() {
		panic("value not is non-nil pointer")
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	item, ok := c.Mapping.Get(key)
	if ok {
		entry := item.(*cacheEntry)
		if c.TTL != 0 {
			if time.Now().Sub(entry.created) >= c.TTL {
				// 将过期数据移到队列后方，而不是删除
				// 如果查询出错，还可能使用保留的过期数据
				c.Mapping.MoveToBack(key)
				// c.Mapping.Pop(key)
				// c.entryPool.Put(entry)
				reflect.ValueOf(value).Elem().Set(reflect.ValueOf(entry.value))
				return expired
			}
		}

		c.Mapping.MoveToFront(key)
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(entry.value))
		return nil
	}

	return notFound
}

// Remove 删除缓存数据
func (c *LRUCache) Remove(key interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	entry, _ := c.Mapping.Pop(key)
	if entry != nil {
		c.entryPool.Put(entry)
	}
}

// Del 删除缓存数据
func (c *LRUCache) Del(key interface{}) error {
	c.Remove(key)
	return nil
}

// Len 缓存的数据的长度
func (c *LRUCache) Len() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.Mapping.Len()
}

// Clear 清空缓存的数据
func (c *LRUCache) Clear() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	for c.Mapping.Len() != 0 {
		entry, _ := c.Mapping.PopBack()
		if entry != nil {
			c.entryPool.Put(entry)
		}
	}
	return nil
}
