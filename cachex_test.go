package cachex

import (
	"errors"
	"math/rand"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"

	"github.com/golang/mock/gomock"
	"github.com/wencan/cachex/mock_cachex"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func TestCachexGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	notFound := mock_cachex.NewMockNotFound(ctrl)
	cached := make(map[interface{}]interface{})
	mockStorage := mock_cachex.NewMockStorage(ctrl)
	mockStorage.EXPECT().Set(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		cached[key] = value
		return nil
	}).AnyTimes()
	mockStorage.EXPECT().Get(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		v, exist := cached[key]
		if !exist {
			return notFound
		}
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(v))
		return nil
	}).AnyTimes()

	queried := make(map[interface{}]interface{})
	mockQuery := mock_cachex.NewMockQuerier(ctrl)
	mockQuery.EXPECT().Query(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		num := key.(int)
		result, ok := queried[num]
		if ok { // 一个key，只执行查询一次，否则算错
			return errors.New("always return")
		}
		result = num * num
		queried[key] = result
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(result))
		return nil
	}).AnyTimes()

	c := NewCachex(mockStorage, mockQuery)

loop:
	for i, k := rand.Intn(1000), 0; k < 100; k++ {
		i += k
		for j := 0; j < rand.Intn(10); j++ {
			var value int
			err := c.Get(i, &value)
			skip := true
			if assert.Equal(t, nil, err) {
				if assert.Equal(t, i*i, value) {
					skip = false
					t.Log(value)
				}
			}
			if skip {
				break loop
			}
		}
	}
}

func TestCachexExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	notFound := mock_cachex.NewMockNotFound(ctrl)
	cached := make(map[interface{}]interface{})
	expires := make(map[interface{}]int64)
	mockStorage := mock_cachex.NewMockStorage(ctrl)
	mockStorage.EXPECT().Set(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		cached[key] = value
		expires[key] = time.Now().UnixNano() + 1000*1000*100 // 0.1秒
		return nil
	}).AnyTimes()
	mockStorage.EXPECT().Get(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		if expire, exist := expires[key]; exist && expire > time.Now().UnixNano() {
			reflect.ValueOf(value).Elem().Set(reflect.ValueOf(cached[key]))
			return nil
		}
		return notFound
	}).AnyTimes()

	base := 1
	mockQuery := mock_cachex.NewMockQuerier(ctrl)
	mockQuery.EXPECT().Query(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		num := key.(int)
		result := num * base
		base++
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(result))
		return nil
	}).AnyTimes()

	c := NewCachex(mockStorage, mockQuery)

	// 正常返回
	var value int
	err := c.Get(1, &value)
	if assert.NoError(t, err) {
		// 缓存未过期
		var newValue int
		err = c.Get(1, &newValue)
		assert.NoError(t, err)
		assert.Equal(t, value, newValue)

		// 缓存过期
		time.Sleep(time.Millisecond * 200)
		err = c.Get(1, &newValue)
		assert.NoError(t, err)
		assert.NotEqual(t, value, newValue)
	}
}

func TestCachexGetError(t *testing.T) {
	// 查询条件，即返回错误
	storageSetErr := errors.New("storage set error")
	storageGetErr := errors.New("storage get error")
	queryQueryErr := errors.New("query error")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	notFound := mock_cachex.NewMockNotFound(ctrl)
	mockStorage := mock_cachex.NewMockStorage(ctrl)
	mockStorage.EXPECT().Set(gomock.Eq(storageSetErr), gomock.Any()).Return(storageSetErr).AnyTimes()
	mockStorage.EXPECT().Set(gomock.Not(gomock.Eq(storageSetErr)), gomock.Any()).Return(nil).AnyTimes()
	mockStorage.EXPECT().Get(gomock.Eq(storageSetErr), gomock.Any()).Return(notFound).AnyTimes()
	mockStorage.EXPECT().Get(gomock.Eq(queryQueryErr), gomock.Any()).Return(notFound).AnyTimes()
	mockStorage.EXPECT().Get(gomock.Eq(storageGetErr), gomock.Any()).Return(storageGetErr).AnyTimes()
	mockStorage.EXPECT().Get(gomock.Not(gomock.Eq(storageGetErr)), gomock.Any()).Return(nil).AnyTimes()

	mockQuery := mock_cachex.NewMockQuerier(ctrl)
	mockQuery.EXPECT().Query(gomock.Eq(queryQueryErr), gomock.Any()).Return(queryQueryErr).AnyTimes()
	mockQuery.EXPECT().Query(gomock.Not(gomock.Eq(queryQueryErr)), gomock.Any()).Return(nil).AnyTimes()

	c := NewCachex(mockStorage, mockQuery)

	conditions := []error{storageSetErr, storageGetErr, queryQueryErr, nil}
	for _, cond := range conditions {
		var value error
		err := c.Get(cond, &value)
		assert.Equal(t, cond, err)
		assert.Equal(t, nil, value)
	}
}

func TestCachexGetConcurrency(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	routines := 500
	loopTimes := 1000
	total := int64(0)

	notFound := mock_cachex.NewMockNotFound(ctrl)
	cached := make(map[interface{}]interface{})
	mockStorage := mock_cachex.NewMockStorage(ctrl)
	mockStorage.EXPECT().Set(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		cached[key] = value
		return nil
	}).AnyTimes()
	mockStorage.EXPECT().Get(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		if result, exist := cached[key]; exist {
			reflect.ValueOf(value).Elem().Set(reflect.ValueOf(result))
			return nil
		}
		return notFound
	}).AnyTimes()

	mockQuery := mock_cachex.NewMockQuerier(ctrl)
	mockQuery.EXPECT().Query(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		atomic.AddInt64(&total, 1)
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(key))
		return nil
	}).AnyTimes()

	c := NewCachex(mockStorage, mockQuery)

	var g errgroup.Group
	for i := 0; i < routines; i++ {
		t.Run("", func(t *testing.T) {
			// t.Parallel()
			for j := 0; j < loopTimes; j++ {
				var value int
				err := c.Get(j, &value)
				// t.Log(value, err)
				if !assert.NoError(t, err) {
					return
				}
				if !assert.Equal(t, j, value) {
					return
				}
			}
		})
	}

	if g.Wait() == nil {
		assert.Equal(t, int64(loopTimes), total)
	}
}

func TestCachex_UseStaleWhenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	notFound := mock_cachex.NewMockNotFound(ctrl)
	expired := mock_cachex.NewMockExpired(ctrl)
	cached := make(map[interface{}]interface{})
	expires := make(map[interface{}]int64)
	mockStorage := mock_cachex.NewMockStorage(ctrl)
	mockStorage.EXPECT().Set(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		cached[key] = value
		expires[key] = time.Now().UnixNano() + int64(time.Nanosecond*2)
		return nil
	}).AnyTimes()
	mockStorage.EXPECT().Get(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		result, exist := cached[key]
		if !exist {
			return notFound
		}
		expire, _ := expires[key]

		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(result))
		if expire <= time.Now().UnixNano() {
			// 已过期
			return expired
		}
		return nil
	}).AnyTimes()

	var testError = errors.New("test")
	var returnErr error
	mockQuery := mock_cachex.NewMockQuerier(ctrl)
	mockQuery.EXPECT().Query(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		if returnErr != nil {
			return returnErr
		}
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(time.Now().Nanosecond()))
		return nil
	}).AnyTimes()

	c := NewCachex(mockStorage, mockQuery)
	c.UseStaleWhenError(true)

	// 正常返回
	var value int
	err := c.Get(1, &value)
	if assert.NoError(t, err) {
		// 缓存过期，查询出错，返回过期的缓存
		time.Sleep(time.Nanosecond * 2)
		returnErr = testError
		var newValue int
		err := c.Get(1, &newValue)
		if assert.Equal(t, testError, err) {
			if assert.Equal(t, value, newValue) {
				// 错误排除，返回新的正常数据
				returnErr = nil
				err = c.Get(1, &newValue)
				assert.NoError(t, err)
				assert.NotEqual(t, value, newValue)
			}
		}
	}
}

type testRequest struct {
	num int
}

func (request testRequest) Key() interface{} {
	return request.num
}

func TestCachexGetWithKeyable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	notFound := mock_cachex.NewMockNotFound(ctrl)
	cached := make(map[interface{}]interface{})
	mockStorage := mock_cachex.NewMockDeletableStorage(ctrl)
	mockStorage.EXPECT().Set(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		cached[key] = value
		return nil
	}).AnyTimes()
	mockStorage.EXPECT().Del(gomock.AssignableToTypeOf(1)).DoAndReturn(func(key interface{}) error {
		delete(cached, key)
		return nil
	}).AnyTimes()
	mockStorage.EXPECT().Get(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		v, exist := cached[key]
		if !exist {
			return notFound
		}
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(v))
		return nil
	}).AnyTimes()

	mockQuery := mock_cachex.NewMockQuerier(ctrl)
	mockQuery.EXPECT().Query(gomock.AssignableToTypeOf((*testRequest)(nil)), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		request := key.(*testRequest)
		result := request.num * request.num
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(result))
		return nil
	}).AnyTimes()

	c := NewCachex(mockStorage, mockQuery)

	request := &testRequest{num: 10}
	var value int
	err := c.Get(request, &value)
	assert.NoError(t, err)
	err = c.Set(request, 10)
	assert.NoError(t, err)
	err = c.Del(request)
	assert.NoError(t, err)
}

func TestCachexGetQueryOption(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	notFound := mock_cachex.NewMockNotFound(ctrl)
	cached := make(map[interface{}]interface{})
	mockStorage := mock_cachex.NewMockStorage(ctrl)
	mockStorage.EXPECT().Set(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		cached[key] = value
		return nil
	}).AnyTimes()
	mockStorage.EXPECT().Get(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		v, exist := cached[key]
		if !exist {
			return notFound
		}
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(v))
		return nil
	}).AnyTimes()

	mockQuery := mock_cachex.NewMockQuerier(ctrl)
	mockQuery.EXPECT().Query(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		num := key.(int)
		result := num * num
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(result))
		return nil
	}).AnyTimes()

	// 默认querier为nil
	c := NewCachex(mockStorage, nil)

	var value int
	// 定制querier
	err := c.Get(10, &value, GetQueryOption(mockQuery))
	assert.NoError(t, err)
	assert.Equal(t, 100, value)
}

func TestCachexGetTTLOption(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	notFound := mock_cachex.NewMockNotFound(ctrl)
	cached := make(map[interface{}]interface{})
	mockStorage := mock_cachex.NewMockSetWithTTLableStorage(ctrl)
	mockStorage.EXPECT().SetWithTTL(gomock.AssignableToTypeOf(1), gomock.Any(), gomock.AssignableToTypeOf(time.Minute)).DoAndReturn(func(key, value interface{}, ttl time.Duration) error {
		cached[key] = value
		return nil
	}).AnyTimes()
	mockStorage.EXPECT().Get(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		v, exist := cached[key]
		if !exist {
			return notFound
		}
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(v))
		return nil
	}).AnyTimes()

	mockQuery := mock_cachex.NewMockQuerier(ctrl)
	mockQuery.EXPECT().Query(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		num := key.(int)
		result := num * num
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(result))
		return nil
	}).AnyTimes()

	// 默认querier为nil
	c := NewCachex(mockStorage, mockQuery)

	var value int
	// 定制querier
	err := c.Get(10, &value, GetTTLOption(time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, 100, value)
}
