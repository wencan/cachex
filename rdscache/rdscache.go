/*
 * redis存储支持
 *
 * wencan
 * 2018-12-30
 */

package rdscache

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/vmihailenco/msgpack"
)

var (
	// Marshal 数据序列化函数
	Marshal = msgpack.Marshal

	// Unmarshal 数据反序列化函数
	Unmarshal = msgpack.Unmarshal
)

// NotFound 没找到错误
type NotFound struct{}

// NotFound 实现cachex.NotFound错误接口
func (NotFound) NotFound() {}
func (NotFound) Error() string {
	return "not found"
}

var notFound = NotFound{}

// RdsCache redis存储实现
type RdsCache struct {
	rdsPool *redis.Pool

	keyPrefix string

	defaultTTL time.Duration
}

// PoolConfig redis池连接参数
type PoolConfig struct {
	Dial func(network, addr string) (net.Conn, error)

	DB int

	Password string

	MaxIdle int

	MaxActive int

	IdleTimeout time.Duration

	Wait bool

	MaxConnLifetime time.Duration
}

type rdsOptions struct {
	keyPrefix string

	defaultTTL time.Duration
}

// RdsOption rdscache配置
type RdsOption struct {
	f func(*rdsOptions)
}

// RdsKeyPrefixOption 配置key前缀
func RdsKeyPrefixOption(keyPrefix string) RdsOption {
	return RdsOption{func(options *rdsOptions) {
		options.keyPrefix = keyPrefix
	}}
}

// RdsDefaultTTLOption 配置key默认生存时间
func RdsDefaultTTLOption(defaultTTL time.Duration) RdsOption {
	return RdsOption{func(options *rdsOptions) {
		options.defaultTTL = defaultTTL
	}}
}

// NewRdsCache 创建redis缓存对象
// 内部创建redis连接池
func NewRdsCache(network, address string, poolCfg PoolConfig, options ...RdsOption) *RdsCache {
	var opts []redis.DialOption
	if poolCfg.Dial != nil {
		opts = append(opts, redis.DialNetDial(poolCfg.Dial))
	}
	if poolCfg.DB != 0 {
		opts = append(opts, redis.DialDatabase(poolCfg.DB))
	}
	if poolCfg.Password != "" {
		opts = append(opts, redis.DialPassword(poolCfg.Password))
	}

	rdsPool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial(network, address, opts...)
		},
	}

	rdsPool.MaxIdle = poolCfg.MaxIdle
	rdsPool.MaxActive = poolCfg.MaxActive
	rdsPool.IdleTimeout = poolCfg.IdleTimeout
	rdsPool.Wait = poolCfg.Wait
	rdsPool.MaxConnLifetime = poolCfg.MaxConnLifetime

	return NewRdsCacheWithPool(rdsPool, options...)
}

// NewRdsCacheWithPool 创建redis缓存对象
// 使用现有redis连接池
func NewRdsCacheWithPool(rdsPool *redis.Pool, options ...RdsOption) *RdsCache {
	var opts rdsOptions
	for _, option := range options {
		option.f(&opts)
	}

	return &RdsCache{
		rdsPool:    rdsPool,
		keyPrefix:  opts.keyPrefix,
		defaultTTL: opts.defaultTTL,
	}
}

// stringKey 将interface{} key转为字符串并加上前缀，不支持类型返回错误
func (c *RdsCache) stringKey(key interface{}) (string, error) {
	var skey string
	switch t := key.(type) {
	case fmt.Stringer:
		skey = t.String()
	case string, []byte, int, int32, int64, uint, uint32, uint64, float32, float64, bool:
		skey = fmt.Sprint(key)
	default:
		return "", errors.New("key type is unacceptable")
	}

	if c.keyPrefix != "" {
		skey = strings.Join([]string{c.keyPrefix, skey}, ":")
	}
	return skey, nil
}

// Set 设置缓存数据
func (c *RdsCache) Set(key, value interface{}) error {
	return c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL 设置缓存数据，并定制TTL
func (c *RdsCache) SetWithTTL(key, value interface{}, TTL time.Duration) error {
	skey, err := c.stringKey(key)
	if err != nil {
		return err
	}

	data, err := Marshal(value)
	if err != nil {
		return err
	}

	conn := c.rdsPool.Get()
	defer conn.Close()

	if TTL != 0 {
		_, err = conn.Do("SET", skey, data, "NX", "PX", int(TTL/time.Millisecond))
	} else {
		_, err = conn.Do("SET", skey, data)
	}
	if err != nil {
		return err
	}

	return nil
}

// Get 获取缓存数据
func (c *RdsCache) Get(key, value interface{}) error {
	skey, err := c.stringKey(key)
	if err != nil {
		return err
	}

	conn := c.rdsPool.Get()
	data, err := redis.Bytes(conn.Do("GET", skey))
	conn.Close()
	if err == redis.ErrNil {
		return notFound
	} else if err != nil {
		return err
	}

	err = Unmarshal(data, value)
	if err != nil {
		return err
	}

	return nil
}

// Del 删除缓存数据
func (c *RdsCache) Del(keys ...interface{}) error {
	var err error
	for idx, key := range keys {
		keys[idx], err = c.stringKey(key)
		if err != nil {
			return err
		}
	}

	conn := c.rdsPool.Get()
	defer conn.Close()

	_, err = conn.Do("DEL", keys...)
	if err != nil {
		return err
	}

	return nil
}
