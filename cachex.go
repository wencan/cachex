package cachex

// wencan
// 2017-08-31

import (
	"errors"
	"sync"
)

type MakerFunc func(key interface{}) (value interface{}, err error)
type DeleterFunc func(key interface{}, value interface{})

var ErrorNotFound error = errors.New("not found")

type Config struct {
	MaxEntries int
	TTLSeconds int

	Maker MakerFunc
	Deleter DeleterFunc

	NotFound error
}

type Cachex struct {
	storage   *LRUCache

	maker     MakerFunc

	sentinels sync.Map

	NotFound error
}

func NewCachex(cfg *Config) (c *Cachex) {
	c = &Cachex{
		storage: &LRUCache{
			MaxEntries: cfg.MaxEntries,
			TTLSeconds: cfg.TTLSeconds,
			Mapping: NewListMap(),
			Deleter: cfg.Deleter,
		},
		maker: cfg.Maker,
	}

	return c
}

func (c *Cachex) Get(key interface{}) (value interface{}, err error) {
	actual, ok := c.sentinels.Load(key)
	if ok {
		sentinel := actual.(*Sentinel)
		return sentinel.Wait()
	}

	value, ok = c.storage.Get(key)
	if ok {
		return value, nil
	}

	if c.maker == nil {
		if c.NotFound != nil {
			return nil, c.NotFound
		} else {
			return nil, ErrorNotFound
		}
	}

	newSentinel := NewSentinel()
	actual, loaded := c.sentinels.LoadOrStore(key, newSentinel)
	sentinel := actual.(*Sentinel)
	if loaded {
		newSentinel.Destroy()
	}

	if !loaded {
		value, err = c.maker(key)
		if err != nil {
			sentinel.Done(value, err)
			return nil, err
		}

		sentinel.Done(value, nil)

		c.Set(key, value)

		c.sentinels.Delete(key)

		return value, nil
	} else {
		return sentinel.Wait()
	}
}

func (c *Cachex) Set(key interface{}, value interface{}) {
	c.storage.Set(key, value)
}

func (c *Cachex) Remove(key interface{}) {
	c.storage.Remove(key)
}

func (c *Cachex) Clear() {
	c.storage.Clear()
}