# cachex
Go业务层缓存，自带内存LRU存储,支持自定义Redis存储实现

# Example
### memory lrucache
```go
func mysql_query(query interface{}) (result interface{}, err error) {
	t := reflect.TypeOf(query)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	v := reflect.New(t)
	result = v.Interface()

	err = db.Where(query).First(result).Error

	return result, err
}

func main() {
	storage := cachex.NewLRUCache(1000, 60*5)
	cache := cachex.NewCachex(storage, mysql_query)

	value, err := cache.Get(&Student{Id: "123"})
	if err == cachex.ErrorNotFound {
		log.Fatalln("not found")
	} else if err != nil {
		log.Fatalln(err)
	}

	student := value.(*Student)
	log.Println(student)
}
```
### redis cache
```go
func mysql_query(query interface{}) (result interface{}, err error) {
	t := reflect.TypeOf(query)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	v := reflect.New(t)
	result = v.Interface()

	err = mydb.Where(query).First(result).Error

	return result, err
}

type RedisCache struct {
	RdsClient *redis.Client
	Expires   time.Duration
}

func (s *RedisCache) Get(key interface{}) (value interface{}, ok bool, err error) {
	keys := fmt.Sprint(key)

	buff, err := s.RdsClient.Get(keys).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}else if err != nil {
		return nil, false, err
	}

	t := reflect.TypeOf(key)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	v := reflect.New(t)
	value = v.Interface()

	buffer := bytes.NewBuffer(buff)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(value)
	if err != nil {
		return nil, false, err
	}

	return value, true, err
}

func (s *RedisCache) Set(key, value interface{}) (err error) {
	keys := fmt.Sprint(key)

	buffer := bytes.Buffer{}
	enc := gob.NewEncoder(&buffer)
	err = enc.Encode(value)
	if err != nil {
		return err
	}

	return s.RdsClient.Set(keys, buffer.Bytes(), s.Expires).Err()
}

func main() {
	s := &RedisCache{
		RdsClient: rds,
		Expires: time.Minute*5,
	}
	cache := cachex.NewCachex(s, mysql_query)

	value, err := cache.Get(&Student{Id: "123"})
	if err == cachex.ErrorNotFound {
		log.Fatalln("not found")
	} else if err != nil {
		log.Fatalln(err)
	}

	student := value.(*Student)
	log.Println(student)
}
```
