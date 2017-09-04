package cachex

// wencan
// 2017-09-02 14:06

import (
	"testing"
	"time"
	"errors"
	"math/rand"
	"sync"
)

var testError error = errors.New("test")

func make_square_maker(key interface{}) (value interface{}, err error) {
	time.Sleep(time.Second)

	num := key.(int)

	return num*num, nil
}

func make_random_maker(key interface{}) (value interface{}, err error) {
	time.Sleep(time.Second)

	num := key.(int)

	rand.Seed(time.Now().Unix())
	return num+rand.Int(), nil
}

func make_error_maker(key interface{}) (value interface{}, err error) {
	return nil, testError
}

func TestCachex_Get(t *testing.T) {
	c := NewCachex(NewLRUCache(1000, 60*5), make_square_maker)

	value, err := c.Get(100)
	if err != nil {
		t.Fatal(err)
	} else if value.(int) != 10000 {
		t.FailNow()
	}
}

func TestCachex_Get_Error(t *testing.T) {
	c := NewCachex(NewLRUCache(1000, 60*5), make_error_maker)

	_, err := c.Get(100)
	if err != testError {
		t.Fatal(err)
	}
}

func TestCachex_Get_Concurrency(t *testing.T) {
	c := NewCachex(NewLRUCache(1000, 60*5), make_random_maker)

	ch := make(chan int, 100)
	for i:=0; i<100; i++ {
		go func() {
			value, err := c.Get(100)
			if err != nil {
				t.Fatal(err)
			}

			ch <- value.(int)
		}()
	}

	var number int
	for i:=0; i<100; i++ {
		if i==0 {
			number = <- ch
		} else {
			value := <- ch
			if number != value {
				t.FailNow()
			}
		}
	}
}

type testStorage struct{
	mapping sync.Map
}

func (s *testStorage) Get(key interface{}) (value interface{}, ok bool, err error) {
	value, ok = s.mapping.Load(key)
	return value, ok, nil
}

func (s *testStorage) Set(key, value interface{}) (err error) {
	s.mapping.Store(key, value)
	return nil
}

func testStorage_maker(key interface{}) (value interface{}, err error) {
	num, ok := key.(int)
	if !ok {
		return nil, errors.New("key type error")
	}

	rand.Seed(time.Now().Unix())
	return num+rand.Int(), nil
}

func TestCachex_Storage(t *testing.T) {
	s := &testStorage{}
	c := NewCachex(s, testStorage_maker)

	value1, err := c.Get(100)
	if err != nil {
		t.Fatal(err)
	}

	value2, err := c.Get(100)
	if err != nil {
		t.Fatal(err)
	}

	if value1 != value2 {
		t.FailNow()
	}
}