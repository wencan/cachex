# lrucache
--
    import "github.com/wencan/cachex/lrucache"


## Usage

#### type Expired

```go
type Expired struct{}
```

Expired 数据已过期错误

#### func (Expired) Error

```go
func (Expired) Error() string
```

#### func (Expired) Expired

```go
func (Expired) Expired()
```
Expired 实现cachex.Expired错误接口

#### type LRUCache

```go
type LRUCache struct {
	MaxEntries int

	Mapping *ListMap
}
```

LRUCache 本地LRU缓存类，实现了cachex.DeletableStorage接口

#### func  NewLRUCache

```go
func NewLRUCache(maxEntries int, defaultTTL time.Duration) *LRUCache
```
NewLRUCache 新建本地LRU缓存类

#### func (*LRUCache) Clear

```go
func (c *LRUCache) Clear() error
```
Clear 清空缓存的数据

#### func (*LRUCache) Del

```go
func (c *LRUCache) Del(key interface{}) error
```
Del 删除缓存数据

#### func (*LRUCache) Get

```go
func (c *LRUCache) Get(key, value interface{}) error
```
Get 获取缓存数据

#### func (*LRUCache) Len

```go
func (c *LRUCache) Len() int
```
Len 缓存的数据的长度

#### func (*LRUCache) Remove

```go
func (c *LRUCache) Remove(key interface{})
```
Remove 删除缓存数据

#### func (*LRUCache) Set

```go
func (c *LRUCache) Set(key, value interface{}) error
```
Set 设置缓存数据

#### func (*LRUCache) SetWithTTL

```go
func (c *LRUCache) SetWithTTL(key, value interface{}, TTL time.Duration) error
```
SetWithTTL 设置缓存数据，并定制TTL


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
