package cachex

// wencan
// 2017-08-31

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"
)

var (
	// ErrNotFound 没找到
	ErrNotFound = errors.New("not found")

	// ErrNotSupported 操作不支持
	ErrNotSupported = errors.New("not supported operation")
)

// Keyable 如果需要包装多个参数为key，包装结构体实现该接口，以支持获取包装结构体的缓存Key。
type Keyable interface {
	// Key 给Storage的缓存Key
	CacheKey() interface{}
}

// Cachex 缓存处理类
type Cachex struct {
	storage Storage

	querier Querier

	sentinels sync.Map

	// useStale UseStaleWhenError
	useStale bool

	deletableStorage   DeletableStorage
	withTTLableStorage SetWithTTLableStorage
}

// NewCachex 新建缓存处理对象
func NewCachex(storage Storage, querier Querier) (c *Cachex) {
	c = &Cachex{
		storage: storage,
		querier: querier,
	}
	c.deletableStorage, _ = storage.(DeletableStorage)
	c.withTTLableStorage, _ = storage.(SetWithTTLableStorage)
	return c
}

// getOptions Get方法的可选参数项
type getOptions struct {
	querier Querier
	ttl     time.Duration
}

// GetOption Get方法的可选参数项结构，不需要直接调用。
type GetOption struct {
	apply func(options *getOptions)
}

// GetQueryOption 为Get操作定制查询过程。
func GetQueryOption(querier Querier) GetOption {
	return GetOption{
		apply: func(options *getOptions) {
			options.querier = querier
		},
	}
}

// GetTTLOption 为Get操作定制TTL。
// 需要存储后端支持，否则报错。
func GetTTLOption(ttl time.Duration) GetOption {
	return GetOption{
		apply: func(options *getOptions) {
			options.ttl = ttl
		},
	}
}

// Get 获取
func (c *Cachex) Get(ctx context.Context, key, value interface{}, opts ...GetOption) error {
	if v := reflect.ValueOf(value); v.Kind() != reflect.Ptr || v.IsNil() {
		panic("value not is non-nil pointer")
	}

	// 可选参数
	var options getOptions
	for _, opt := range opts {
		opt.apply(&options)
	}
	// 查询过程
	querier := c.querier
	if options.querier != nil {
		querier = options.querier
	}
	// ttl
	var ttl time.Duration
	if options.ttl != 0 {
		if c.withTTLableStorage == nil {
			return ErrNotSupported
		}
		ttl = options.ttl
	}

	// 支持包装结构体的key
	request := key
	if keyable, ok := key.(Keyable); ok {
		key = keyable.CacheKey()
	}

	err := c.storage.Get(ctx, key, value)
	if err == nil {
		return nil
	} else if _, ok := err.(NotFound); ok {
		// 下面查询
	} else if _, ok := err.(Expired); ok {
		// 数据已过期，下面查询
	} else if err != nil {
		return err
	}

	if querier == nil {
		return ErrNotFound
	}

	// 在一份实例中
	// 不同时发起重复的查询请求——解决缓存失效风暴
	newSentinel := NewSentinel()
	actual, loaded := c.sentinels.LoadOrStore(key, newSentinel)
	sentinel := actual.(*Sentinel)
	if loaded {
		newSentinel.Close()
	} else {
		// 确保生产者总是能发出通知，并解锁
		defer c.sentinels.Delete(key)
		defer sentinel.CloseIfUnclose()
	}

	// 双重检查
	var staled interface{}
	err = c.storage.Get(ctx, key, value)
	if err == nil {
		if !loaded {
			// 将结果通知等待的过程
			sentinel.Done(reflect.ValueOf(value).Elem().Interface(), nil)
		}
		return nil
	} else if err == nil {
		return nil
	} else if _, ok := err.(NotFound); ok {
		// 下面查询
	} else if _, ok := err.(Expired); ok {
		// 保存过期数据，如果下面查询失败，且useStale，返回过期数据
		staled = reflect.ValueOf(value).Elem().Interface()
	} else if err != nil {
		if !loaded {
			// 将错误通知等待的过程
			sentinel.Done(nil, err)
		}
		return err
	}

	if !loaded {
		err := querier.Query(ctx, request, value)
		if err != nil && c.useStale && staled != nil {
			// 当查询发生错误时，使用过期的缓存数据。该特性需要Storage支持
			reflect.ValueOf(value).Elem().Set(reflect.ValueOf(staled))
			sentinel.Done(staled, err)
			return err
		}

		if _, ok := err.(NotFound); ok {
			err = ErrNotFound
		}
		if err != nil {
			sentinel.Done(nil, err)
			return err
		}

		// 更新到存储后端
		elem := reflect.ValueOf(value).Elem().Interface()
		if ttl != 0 {
			err = c.withTTLableStorage.SetWithTTL(ctx, key, elem, ttl)
		} else {
			err = c.storage.Set(ctx, key, elem)
		}

		sentinel.Done(elem, nil)

		return err
	}

	return sentinel.Wait(ctx, value)
}

// Set 更新
func (c *Cachex) Set(ctx context.Context, key, value interface{}) error {
	if keyable, ok := key.(Keyable); ok {
		key = keyable.CacheKey()
	}
	return c.storage.Set(ctx, key, value)
}

// SetWithTTL 更新，并定制TTL
func (c *Cachex) SetWithTTL(ctx context.Context, key, value interface{}, TTL time.Duration) error {
	if c.withTTLableStorage != nil {
		if keyable, ok := key.(Keyable); ok {
			key = keyable.CacheKey()
		}
		c.withTTLableStorage.SetWithTTL(ctx, key, value, TTL)
	}
	return ErrNotSupported
}

// Del 删除
func (c *Cachex) Del(ctx context.Context, keys ...interface{}) error {
	if c.deletableStorage == nil {
		return ErrNotSupported
	}

	for idx, key := range keys {
		if keyable, ok := key.(Keyable); ok {
			keys[idx] = keyable.CacheKey()
		}
	}
	return c.deletableStorage.Del(ctx, keys...)
}

// UseStaleWhenError 设置当查询发生错误时，使用过期的缓存数据。该特性需要Storage支持（Get返回过期的缓存数据和Expired错误实现）。默认关闭。
func (c *Cachex) UseStaleWhenError(use bool) {
	c.useStale = use
}
