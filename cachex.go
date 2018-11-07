package cachex

// wencan
// 2017-08-31

import (
	"errors"
	"sync"
)

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

type QueryFunc func(key interface{}) (value interface{}, ok bool, err error)

func (fun QueryFunc) Query(key interface{}) (value interface{}, ok bool, err error) {
	return fun(key)
}

type Querier interface {
	Query(key interface{}) (value interface{}, ok bool, err error)
}

type Cachex struct {
	storage Storage

	querier Querier

	sentinels sync.Map

	NotFound error
}

func NewCachexWithQuerier(storage Storage, querier Querier) (c *Cachex) {
	c = &Cachex{
		storage: storage,
		querier: querier,
	}
	return c
}

func NewCachex(storage Storage, query QueryFunc) (c *Cachex) {
	return NewCachexWithQuerier(storage, QueryFunc(query))
}

func (c *Cachex) Get(key interface{}) (value interface{}, err error) {
	value, ok, err := c.storage.Get(key)
	if err != nil {
		return nil, err
	} else if ok {
		return value, nil
	}

	if c.querier == nil {
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
		value, ok, err := c.querier.Query(key)
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
