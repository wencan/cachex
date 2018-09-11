package cachex

// wencan
// 2017-08-31

import (
	"errors"
	"sync"
)

type MakerFunc func(key interface{}) (value interface{}, ok bool, err error)

var (
	ErrNotFound error = errors.New("not found")

	ErrNotSupported error = errors.New("not supported operation")
)

type Storage interface {
	Get(key interface{}) (value interface{}, ok bool, err error)
	Set(key, value interface{}) (err error)
}

type DeletableStorage interface {
	Storage
	Del(key interface{}) (err error)
}

type Cachex struct {
	storage Storage

	maker MakerFunc

	sentinels sync.Map

	NotFound error
}

func NewCachex(storage Storage, maker MakerFunc) (c *Cachex) {
	c = &Cachex{
		storage: storage,
		maker:   maker,
	}
	return c
}

func (c *Cachex) Get(key interface{}) (value interface{}, err error) {
	value, ok, err := c.storage.Get(key)
	if err != nil {
		return nil, err
	} else if ok {
		return value, nil
	}

	if c.maker == nil {
		if c.NotFound != nil {
			return nil, c.NotFound
		} else {
			return nil, ErrNotFound
		}
	}

	newSentinel := NewSentinel()
	actual, loaded := c.sentinels.LoadOrStore(key, newSentinel)
	sentinel := actual.(*Sentinel)
	if loaded {
		newSentinel.Destroy()
	}

	value, ok, err = c.storage.Get(key)
	if err != nil {
		return nil, err
	} else if ok {
		if !loaded {
			sentinel.Done(value, err)
			c.sentinels.Delete(key)
		}
		return value, nil
	}

	if !loaded {
		value, ok, err := c.maker(key)
		if err == nil && !ok {
			if c.NotFound != nil {
				err = c.NotFound
			} else {
				err = ErrNotFound
			}
		}
		if err != nil {
			sentinel.Done(value, err)
			c.sentinels.Delete(key)
			return nil, err
		}

		err = c.storage.Set(key, value)

		sentinel.Done(value, nil)
		c.sentinels.Delete(key)

		return value, err
	} else {
		return sentinel.Wait()
	}
}

func (c *Cachex) Set(key, value interface{}) (err error) {
	return c.storage.Set(key, value)
}

func (c *Cachex) Del(key interface{}) (err error) {
	s, ok := c.storage.(DeletableStorage)
	if ok {
		s.Del(key)
	}
	return ErrNotSupported
}
