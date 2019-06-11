# cachex
Go业务层缓存中间件，自带内存LRU存储和Redis存储支持，并支持自定义存储后端

# 工作机制

cachex由两部分组成：查询引擎和存储后端。

用户逻辑通过查询引擎向存储后端查询缓存，如果查到返回缓存的结果；否则查询引擎调用查询接口获取新的结果，将新结果存储到存储后端，并返回给用户逻辑。

# 特性

- 支持内存LRU存储、Redis存储，支持自定义存储实现

- 通过哨兵机制解决了单实例内的缓存失效风暴问题

- 支持缓存TTL、查询接口失败返回过期的结果（均需要存储后端支持）

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
// or s := rdscache.NewRdsCache("tcp", rds.Addr(), &rdscache.RdsConfig{DB: 1, KeyPrefix: "cache"})
cache := cachex.NewCachex(s, cachex.QueryFunc(query))

var dt DateTime
err = cache.Get(time.Now().Second(), &dt)
if err != nil {
	log.Println(err)
	return
}
```

# API
## cachex
--
    import "github.com/wencan/cachex"


### Usage

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

#### func (*Cachex) Del

```go
func (c *Cachex) Del(key interface{}) error
```
Del 删除

#### func (*Cachex) Get

```go
func (c *Cachex) Get(key, value interface{}) error
```
Get 获取

#### func (*Cachex) Set

```go
func (c *Cachex) Set(key, value interface{}) error
```
Set 更新

#### func (*Cachex) SetWithTTL

```go
func (c *Cachex) SetWithTTL(key, value interface{}, TTL time.Duration) error
```
SetWithTTL 更新，并定制TTL

#### func (*Cachex) UseStaleWhenError

```go
func (c *Cachex) UseStaleWhenError(use bool)
```
UseStaleWhenError
设置当查询发生错误时，使用过期的缓存数据。该特性需要Storage支持（Get返回过期的缓存数据和Expired错误实现）。默认关闭。

#### type ClearableStorage

```go
type ClearableStorage interface {
        Storage
        Clear() error
}
```

ClearableStorage 支持清理操作的存储后端接口

#### type DeletableStorage

```go
type DeletableStorage interface {
        Storage
        Del(key interface{}) error
}
```

DeletableStorage 支持删除操作的存储后端接口

#### type Expired

```go
type Expired interface {
        error
        Expired()
}
```

Expired 已过期错误接口

#### type NotFound

```go
type NotFound interface {
        error
        NotFound()
}
```

NotFound 没找到错误接口

#### type Querier

```go
type Querier interface {
        // Query 查询。value必须是非nil指针。没找到返回NotFound错误实现
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

#### type SetWithTTLableStorage

```go
type SetWithTTLableStorage interface {
        Storage
        SetWithTTL(key, value interface{}, TTL time.Duration) error
}
```

SetWithTTLableStorage 支持定制TTL的存储后端接口

#### type Storage

```go
type Storage interface {
        // Get 获取缓存的数据。value必须是非nil指针。没找到返回NotFound；数据已经过期返回过期数据加NotFound
        Get(key, value interface{}) error

        // Set 缓存数据
        Set(key, value interface{}) error
}
```

Storage 存储后端接口


