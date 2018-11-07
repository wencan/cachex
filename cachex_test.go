package cachex

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
	squareQuery := func(key interface{}) (value interface{}, ok bool, err error) {
		num := key.(int)

		return num * num, true, nil
	}
	c := NewCachex(NewLRUCache(1000, time.Minute*5), squareQuery)

	value, err := c.Get(100)
	if err != nil {
		t.Fatal(err)
	} else if value.(int) != 10000 {
		t.FailNow()
	}
}

func TestCachexGetError(t *testing.T) {
	errorQuery := func(key interface{}) (value interface{}, ok bool, err error) {
		return nil, false, testError
	}
	c := NewCachex(NewLRUCache(1000, time.Minute*5), errorQuery)

	_, err := c.Get(1)
	if err != testError {
		t.Fatal(err)
	}

	var retError error
	returnErrorQuery := func(key interface{}) (value interface{}, ok bool, err error) {
		if retError != nil {
			return nil, false, retError
		}
		return nil, true, nil
	}
	c = NewCachex(NewLRUCache(1000, time.Minute*5), returnErrorQuery)

	retError = testError
	_, err = c.Get(1)
	if err != testError {
		t.Fatal(err)
	}
	retError = nil
	_, err = c.Get(1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Get(1)
	if err != nil {
		t.Fatal(err)
	}
	c.Del(1)
	_, err = c.Get(1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCachexGetConcurrency(t *testing.T) {
	routines := 1000
	loopTimes := 1000

	total := int64(0)
	returnKeyQuery := func(key interface{}) (value interface{}, ok bool, err error) {
		atomic.AddInt64(&total, 1)
		return key, true, nil
	}
	c := NewCachex(NewLRUCache(loopTimes, 0), returnKeyQuery)

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

	c = NewCachex(NewLRUCache(loopTimes, time.Second), returnKeyQuery)
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

func testStorageQuery(key interface{}) (value interface{}, ok bool, err error) {
	num, ok := key.(int)
	if !ok {
		return nil, false, errors.New("key type error")
	}

	rand.Seed(time.Now().Unix())
	return num + rand.Int(), true, nil
}

func TestCachex_Storage(t *testing.T) {
	s := &testStorage{}
	c := NewCachex(s, testStorageQuery)

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

func TestCachex_UseStaleWhenError(t *testing.T) {
	var retError error
	returnErrorQuery := func(key interface{}) (value interface{}, ok bool, err error) {
		if retError != nil {
			return nil, false, retError
		}
		return time.Now().Nanosecond(), true, nil
	}
	c := NewCachex(NewLRUCache(1000, time.Nanosecond), returnErrorQuery)
	c.UseStaleWhenError(true)

	value, err := c.Get(1)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Nanosecond * 2)
	retError = testError
	newValue, err := c.Get(1)
	if err != testError {
		t.FailNow()
	}
	if newValue != value {
		t.FailNow()
	}

	retError = nil
	newValue, err = c.Get(1)
	if err != nil {
		t.FailNow()
	}
	if newValue == value {
		t.FailNow()
	}
}
