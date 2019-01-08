# cachex
Go业务层缓存中间件，自带内存LRU存储和Redis存储支持，并支持自定义存储后端

# Example
```go
query := func(key, value interface{}) error {
	dt := value.(*DateTime)
	err = db.Get(dt, "SELECT date('now') as date, time('now') as time, random() as rand;")
	if err != nil {
		// log.Println(err)
		return err
	}
	return nil
}

s := lrucache.NewLRUCache(1000, time.Second)
// or s := rdscache.NewRdsCache("tcp", rds.Addr(), rdscache.RdsDB(1), rdscache.RdsKeyPrefix("cache"))
cache := cachex.NewCachex(s, cachex.QueryFunc(query))

var dt DateTime
err = cache.Get(time.Now().Second(), &dt)
if err != nil {
	log.Println(err)
	return
}
```

# API

## cachex API

#### variables

```go
var (
	// ErrNotFound 没找到
	ErrNotFound = errors.New("not found")

	// ErrNotSupported 操作不支持
	ErrNotSupported = errors.New("not supported operation")
)
```

#### type Cachex

```go
type Cachex struct {
}
```

Cachex 缓存处理类

#### func  NewCachex

```go
func NewCachex(storage Storage, querier Querier) (c *Cachex)
```
NewCachex 新建缓存处理对象

#### func (*Cachex) Get

```go
func (c *Cachex) Get(key, value interface{}) error
```
Get 获取

#### func (*Cachex) Set

```go
func (c *Cachex) Set(key, value interface{}) (err error)
```
Set 更新

#### func (*Cachex) Del

```go
func (c *Cachex) Del(key interface{}) (err error)
```
Del 删除

#### func (*Cachex) UseStaleWhenError

```go
func (c *Cachex) UseStaleWhenError(use bool)
```
UseStaleWhenError 设置当查询发生错误时，使用过期的缓存数据。该特性需要Storage支持（Get返回并继续暂存过期的缓存数据）。默认关闭。

#### type Querier

```go
type Querier interface {
	// Query 查询。value必须是非nil指针。没找到返回Expired
	Query(key, value interface{}) error
}
```

Querier 查询接口

#### type QueryFunc

```go
type QueryFunc func(key, value interface{}) error
```

QueryFunc 查询过程签名

#### func (QueryFunc) Query

```go
func (fun QueryFunc) Query(key, value interface{}) error
```
Query 查询过程实现Querier接口

## cachex/lrucache API

#### func  NewLRUCache

```go
func NewLRUCache(maxEntries int, TTL time.Duration) *LRUCache
```
NewLRUCache 新建本地LRU缓存类

## cachex/rdscache API

#### variables

```go
var (
	// Marshal 数据序列化函数
	Marshal = msgpack.Marshal

	// Unmarshal 数据反序列化函数
	Unmarshal = msgpack.Unmarshal
)
```

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
func NewRdsCache(network, address string, rdsCfgs ...RdsConfig) *RdsCache
```
NewRdsCache 创建redis缓存对象

#### type RdsConfig

```go
type RdsConfig struct {
}
```

RdsConfig rdscache配置

#### func  RdsDB

```go
func RdsDB(db int) RdsConfig
```
RdsDB redis db配置

#### func  RdsDial

```go
func RdsDial(dial func(network, addr string) (net.Conn, error)) RdsConfig
```
RdsDial redis连接函数

#### func  RdsKeyPrefix

```go
func RdsKeyPrefix(keyPrefix string) RdsConfig
```
RdsKeyPrefix redis key前缀

#### func  RdsPassword

```go
func RdsPassword(password string) RdsConfig
```
RdsPassword redis密码

#### func  RdsPoolConfig

```go
func RdsPoolConfig(poolCfg PoolConfig) RdsConfig
```
RdsPoolConfig redis连接池配置对象

#### func  RdsDefaultTTL

```go
func RdsDefaultTTL(ttl time.Duration) RdsConfig
```
RdsDefaultTTL redis key生存时间
