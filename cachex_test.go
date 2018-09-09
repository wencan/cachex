package cachex

// wencan
// 2017-09-02 14:06

import (
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var testError error = errors.New("test")

func TestCachexGet(t *testing.T) {
	makeSquareMaker := func(key interface{}) (value interface{}, ok bool, err error) {
		num := key.(int)

		return num * num, true, nil
	}
	c := NewCachex(NewLRUCache(1000, 60*5), makeSquareMaker)

	value, err := c.Get(100)
	if err != nil {
		t.Fatal(err)
	} else if value.(int) != 10000 {
		t.FailNow()
	}
}

func TestCachexGetError(t *testing.T) {
	makeErrorMaker := func(key interface{}) (value interface{}, ok bool, err error) {
		return nil, false, testError
	}
	c := NewCachex(NewLRUCache(1000, 60*5), makeErrorMaker)

	_, err := c.Get(100)
	if err != testError {
		t.Fatal(err)
	}

	var retError error
	returnErrorMaker := func(key interface{}) (value interface{}, ok bool, err error) {
		if retError != nil {
			return nil, false, retError
		}
		return nil, true, nil
	}
	c = NewCachex(NewLRUCache(1000, 60*5), returnErrorMaker)

	retError = testError
	_, err = c.Get(nil)
	if err != testError {
		t.Fatal(err)
	}
	retError = nil
	_, err = c.Get(nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCachexGetConcurrency(t *testing.T) {
	routines := 1000
	loopTimes := 1000

	total := int64(0)
	returnKeyMaker := func(key interface{}) (value interface{}, ok bool, err error) {
		atomic.AddInt64(&total, 1)
		return key, true, nil
	}
	c := NewCachex(NewLRUCache(loopTimes, 0), returnKeyMaker)

	var wg sync.WaitGroup
	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < loopTimes; j++ {
				value, err := c.Get(j)
				if err != nil {
					t.Fatal(err)
				}
				n := value.(int)
				if n != j {
					t.Fatalf("value missmatch, got: %d, want: %d", n, j)
				}
			}
		}()
	}
	wg.Wait()

	if total != int64(loopTimes) {
		t.Fatalf("total missmatch, got: %d, want: %d", total, loopTimes)
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
