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

// PoolConfig redis连接池配置
type PoolConfig struct {
	MaxIdle int

	MaxActive int

	IdleTimeout time.Duration

	Wait bool

	MaxConnLifetime time.Duration
}

// RdsConfig rdscache配置
type RdsConfig struct {
	Dial func(network, addr string) (net.Conn, error)

	DB int

	Password string

	PoolCfg *PoolConfig

	KeyPrefix string

	DefaultTTL time.Duration
}

// NewRdsCache 创建redis缓存对象
func NewRdsCache(network, address string, rdsCfg *RdsConfig) *RdsCache {
	var opts []redis.DialOption
	var poolCfg *PoolConfig
	var keyPrefix string
	var defaultTTL time.Duration

	if rdsCfg != nil {
		if rdsCfg.Dial != nil {
			opts = append(opts, redis.DialNetDial(rdsCfg.Dial))
		}
		if rdsCfg.DB != 0 {
			opts = append(opts, redis.DialDatabase(rdsCfg.DB))
		}
		if rdsCfg.Password != "" {
			opts = append(opts, redis.DialPassword(rdsCfg.Password))
		}
		if rdsCfg.PoolCfg != nil {
			poolCfg = rdsCfg.PoolCfg
		}
		if rdsCfg.KeyPrefix != "" {
			keyPrefix = rdsCfg.KeyPrefix
		}
		if rdsCfg.DefaultTTL != 0 {
			defaultTTL = rdsCfg.DefaultTTL
		}
	}

	rdsPool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial(network, address, opts...)
		},
	}
	if poolCfg != nil {
		rdsPool.MaxIdle = poolCfg.MaxIdle
		rdsPool.MaxActive = poolCfg.MaxActive
		rdsPool.IdleTimeout = poolCfg.IdleTimeout
		rdsPool.Wait = poolCfg.Wait
		rdsPool.MaxConnLifetime = poolCfg.MaxConnLifetime
	}

	return &RdsCache{
		rdsPool:    rdsPool,
		keyPrefix:  keyPrefix,
		defaultTTL: defaultTTL,
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
func (c *RdsCache) Del(key interface{}) error {
	skey, err := c.stringKey(key)
	if err != nil {
		return err
	}

	conn := c.rdsPool.Get()
	defer conn.Close()

	_, err = conn.Do("DEL", skey)
	if err != nil {
		return err
	}

	return nil
}
