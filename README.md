# cachex
 业务层本地缓存

# Example
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
	cfg := cachex.Config{
		MaxEntries: 1000,
		TTLSeconds: 60 * 5,
		Maker:      mysql_query,
	}
	cache := cachex.NewCachex(&cfg)

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
