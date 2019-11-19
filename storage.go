/*
 * 存储后端接口
 *
 * wencan
 * 2018-12-26
 */

package cachex

import "time"

// Storage 存储后端接口
type Storage interface {
	// Get 获取缓存的数据。value必须是非nil指针。没找到返回NotFound；数据已经过期返回过期数据加NotFound
	Get(key, value interface{}) error

	// Set 缓存数据
	Set(key, value interface{}) error
}

// DeletableStorage 支持删除操作的存储后端接口
type DeletableStorage interface {
	Storage
	Del(keys ...interface{}) error
}

// ClearableStorage 支持清理操作的存储后端接口
type ClearableStorage interface {
	Storage
	Clear() error
}

// SetWithTTLableStorage 支持定制TTL的存储后端接口
type SetWithTTLableStorage interface {
	Storage
	SetWithTTL(key, value interface{}, TTL time.Duration) error
}

// NopStorage 一个什么都不干的存储后端。
// 可以用NopStorage加CacheX组合出一个单实例内不重复查询的机制。
type NopStorage struct {
}

type nopNotFound struct{}

// Error 实现error接口
func (nopNotFound) Error() string {
	return "not found"
}

// NotFound 实现NotFound错误接口
func (nopNotFound) NotFound() {}

// Get 实现Storage接口，只返回NotFound错误。
func (NopStorage) Get(key, value interface{}) error {
	return nopNotFound{}
}

// Set 实现Storage接口，只返回nil。
func (NopStorage) Set(key, value interface{}) error {
	return nil
}

// Del 实现DeletableStorage接口，只返回nil。
func (NopStorage) Del(keys ...interface{}) error {
	return nil
}

// Clear 实现ClearableStorage接口，只返回nil。
func (NopStorage) Clear() error {
	return nil
}

// SetWithTTL 实现SetWithTTLableStorage接口，只返回nil。
func (NopStorage) SetWithTTL(key, value interface{}, TTL time.Duration) error {
	return nil
}
