# cachex
Go业务层缓存，自带内存LRU存储,支持自定义Redis存储实现

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
