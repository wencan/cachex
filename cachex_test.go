package cachex

// wencan
// 2017-09-02 14:06

import (
	"testing"
	"time"
	"errors"
	"math/rand"
)

var testError error = errors.New("test")

func make_square_handler(key interface{}) (value interface{}, err error) {
	time.Sleep(time.Second)

	num := key.(int)

	return num*num, nil
}

func make_random_handler(key interface{}) (value interface{}, err error) {
	time.Sleep(time.Second)

	num := key.(int)

	rand.Seed(time.Now().Unix())
	return num*rand.Int(), nil
}

func make_error_handler(key interface{}) (value interface{}, err error) {
	return nil, testError
}

func TestCachex_Get(t *testing.T) {
	c := NewCachex(Config{
		Maker: make_square_handler,
	})

	value, err := c.Get(100)
	if err != nil {
		t.FailNow()
	} else if value.(int) != 10000 {
		t.FailNow()
	}
}

func TestCachex_Get_Error(t *testing.T) {
	c := NewCachex(Config{
		Maker: make_error_handler,
	})

	_, err := c.Get(100)
	if err != testError {
		t.FailNow()
	}
}

func TestCachex_Get_Concurrency(t *testing.T) {
	c := NewCachex(Config{
		Maker: make_random_handler,
	})

	ch := make(chan int, 100)
	for i:=0; i<100; i++ {
		go func() {
			value, err := c.Get(100)
			if err != nil {
				t.FailNow()
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