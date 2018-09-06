package cachex

// wencan
// 2017-09-02 14:06

import (
	"errors"
	"math/rand"
	"sync"
	"testing"
	"time"
)

var testError error = errors.New("test")

func makeSquareMaker(key interface{}) (value interface{}, ok bool, err error) {
	time.Sleep(time.Second)

	num := key.(int)

	return num * num, true, nil
}

func makeRandomMaker(key interface{}) (value interface{}, ok bool, err error) {
	time.Sleep(time.Second)

	num := key.(int)

	rand.Seed(time.Now().Unix())
	return num + rand.Int(), true, nil
}

func makeErrorMaker(key interface{}) (value interface{}, ok bool, err error) {
	return nil, false, testError
}

func TestCachexGet(t *testing.T) {
	c := NewCachex(NewLRUCache(1000, 60*5), makeSquareMaker)

	value, err := c.Get(100)
	if err != nil {
		t.Fatal(err)
	} else if value.(int) != 10000 {
		t.FailNow()
	}
}

func TestCachexGetError(t *testing.T) {
	c := NewCachex(NewLRUCache(1000, 60*5), makeErrorMaker)

	_, err := c.Get(100)
	if err != testError {
		t.Fatal(err)
	}
}

func TestCachexGetConcurrency(t *testing.T) {
	c := NewCachex(NewLRUCache(1000, 60*5), makeRandomMaker)

	ch := make(chan int, 100)
	for i := 0; i < 100; i++ {
		go func() {
			value, err := c.Get(100)
			if err != nil {
				t.Fatal(err)
			}

			ch <- value.(int)
		}()
	}

	var number int
	for i := 0; i < 100; i++ {
		if i == 0 {
			number = <-ch
		} else {
			value := <-ch
			if number != value {
				t.FailNow()
			}
		}
	}
}

type testStorage struct {
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

func (s *testStorage) Del(key interface{}) (err error) {
	s.mapping.Delete(key)
	return nil
}

func testStorageMaker(key interface{}) (value interface{}, ok bool, err error) {
	num, ok := key.(int)
	if !ok {
		return nil, false, errors.New("key type error")
	}

	rand.Seed(time.Now().Unix())
	return num + rand.Int(), true, nil
}

func TestCachex_Storage(t *testing.T) {
	s := &testStorage{}
	c := NewCachex(s, testStorageMaker)

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
