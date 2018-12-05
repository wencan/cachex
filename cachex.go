package cachex

// wencan
// 2017-08-31

import (
	"errors"
	"sync"
)

var (
	// ErrNotFound 没找到
	ErrNotFound = errors.New("not found")

	// ErrNotSupported 操作不支持
	ErrNotSupported = errors.New("not supported operation")
)

// Storage 存储后端接口
type Storage interface {
	// Get 获取缓存的数据。 如果返回value != nil && !ok && err == nil，表示返回已经过期的数据
	Get(key interface{}) (value interface{}, ok bool, err error)

	// Set 缓存数据
	Set(key, value interface{}) (err error)
}

// DeletableStorage 支持删除操作的存储后端接口
type DeletableStorage interface {
	Storage
	Del(key interface{}) (err error)
}

// QueryFunc 查询过程签名
type QueryFunc func(key interface{}) (value interface{}, ok bool, err error)

// Query 查询过程实现Querier接口
func (fun QueryFunc) Query(key interface{}) (value interface{}, ok bool, err error) {
	return fun(key)
}

// Querier 查询接口
type Querier interface {
	Query(key interface{}) (value interface{}, ok bool, err error)
}

// Cachex 缓存处理类
type Cachex struct {
	storage Storage

	querier Querier

	sentinels sync.Map

	NotFound error

	// UseStale UseStaleWhenError
	UseStale bool
}

// NewCachexWithQuerier 新建缓存处理对象
func NewCachexWithQuerier(storage Storage, querier Querier) (c *Cachex) {
	c = &Cachex{
		storage: storage,
		querier: querier,
	}
	return c
}

// NewCachex 新建缓存处理对象
func NewCachex(storage Storage, query QueryFunc) (c *Cachex) {
	return NewCachexWithQuerier(storage, QueryFunc(query))
}

// Get 获取
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

	// 在一份示例中
	// 不同时发起重复的查询请求——解决缓存失效风暴
	newSentinel := NewSentinel()
	actual, loaded := c.sentinels.LoadOrStore(key, newSentinel)
	sentinel := actual.(*Sentinel)
	if loaded {
		newSentinel.Destroy()
	}

	var staled interface{}
	value, ok, err = c.storage.Get(key) // 双重检查
	if err != nil {
		return nil, err
	} else if ok {
		if !loaded {
			sentinel.Done(value, err)
			c.sentinels.Delete(key)
		}
		return value, nil
	} else if value != nil {
		// 如果返回value != nil && !ok && err == nil，表示返回已经过期的数据
		staled = value
	}

	if !loaded {
		value, ok, err := c.querier.Query(key)
		if err != nil && c.UseStale && staled != nil {
			// 当查询发生错误时，使用过期的缓存数据。该特性需要Storage支持
			sentinel.Done(staled, err)
			c.sentinels.Delete(key)
			return staled, err
		}

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

// Set 更新
func (c *Cachex) Set(key, value interface{}) (err error) {
	return c.storage.Set(key, value)
}

// Del 删除
func (c *Cachex) Del(key interface{}) (err error) {
	s, ok := c.storage.(DeletableStorage)
	if ok {
		s.Del(key)
	}
	return ErrNotSupported
}

// UseStaleWhenError 设置当查询发生错误时，使用过期的缓存数据。该特性需要Storage支持（Get返回并继续暂存过期的缓存数据）。默认关闭。
func (c *Cachex) UseStaleWhenError(use bool) {
	c.UseStale = use
}
