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
	Del(key interface{}) error
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
