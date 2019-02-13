# rdscache
--
    import "github.com/wencan/cachex/rdscache"


## Usage

```go
var (
	// Marshal 数据序列化函数
	Marshal = msgpack.Marshal

	// Unmarshal 数据反序列化函数
	Unmarshal = msgpack.Unmarshal
)
```

#### type NotFound

```go
type NotFound struct{}
```

NotFound 没找到错误

#### func (NotFound) Error

```go
func (NotFound) Error() string
```

#### func (NotFound) NotFound

```go
func (NotFound) NotFound()
```
NotFound 实现cachex.NotFound错误接口

#### type PoolConfig

```go
type PoolConfig struct {
	MaxIdle int

	MaxActive int

	IdleTimeout time.Duration

	Wait bool

	MaxConnLifetime time.Duration
}
```

PoolConfig redis连接池配置

#### type RdsCache

```go
type RdsCache struct {
}
```

RdsCache redis存储实现

#### func  NewRdsCache

```go
func NewRdsCache(network, address string, rdsCfg *RdsConfig) *RdsCache
```
NewRdsCache 创建redis缓存对象

#### func (*RdsCache) Del

```go
func (c *RdsCache) Del(key interface{}) error
```
Del 删除缓存数据

#### func (*RdsCache) Get

```go
func (c *RdsCache) Get(key, value interface{}) error
```
Get 获取缓存数据

#### func (*RdsCache) Set

```go
func (c *RdsCache) Set(key, value interface{}) error
```
Set 设置缓存数据

#### func (*RdsCache) SetWithTTL

```go
func (c *RdsCache) SetWithTTL(key, value interface{}, TTL time.Duration) error
```
SetWithTTL 设置缓存数据，并定制TTL

#### type RdsConfig

```go
type RdsConfig struct {
	Dial func(network, addr string) (net.Conn, error)

	DB int

	Password string

	PoolCfg *PoolConfig

	KeyPrefix string

	DefaultTTL time.Duration
}
```

RdsConfig rdscache配置
