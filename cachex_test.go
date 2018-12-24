package cachex

import (
	"errors"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
	"github.com/wencan/cachex/mock_cachex"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func TestCachexGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cached := make(map[interface{}]interface{})
	mockStorage := mock_cachex.NewMockStorage(ctrl)
	mockStorage.EXPECT().Set(gomock.Any(), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		cached[key] = value
		return nil
	}).AnyTimes()
	mockStorage.EXPECT().Get(gomock.AssignableToTypeOf(1)).DoAndReturn(func(key interface{}) (interface{}, bool, error) {
		value, exist := cached[key]
		return value, exist, nil
	}).AnyTimes()

	queried := make(map[interface{}]interface{})
	mockQuery := mock_cachex.NewMockQuerier(ctrl)
	mockQuery.EXPECT().Query(gomock.AssignableToTypeOf(1)).DoAndReturn(func(key interface{}) (interface{}, bool, error) {
		num := key.(int)
		result, ok := queried[num]
		if ok { // 一个key，只执行查询一次，否则算错
			return nil, false, errors.New("always return")
		}
		result = num * num
		queried[key] = result
		return result, true, nil
	}).AnyTimes()

	c := NewCachex(mockStorage, mockQuery)

	for i, k := rand.Intn(1000), 0; k < 100; k++ {
		i += k
		for j := 0; j < rand.Intn(10); j++ {
			value, err := c.Get(i)
			if assert.Equal(t, nil, err) {
				assert.Equal(t, i*i, value.(int))
			}
		}
	}
}

func TestCachexExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cached := make(map[interface{}]interface{})
	expires := make(map[interface{}]int64)
	mockStorage := mock_cachex.NewMockStorage(ctrl)
	mockStorage.EXPECT().Set(gomock.Any(), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		cached[key] = value
		expires[key] = time.Now().UnixNano() + 1000*1000*100 // 0.1秒
		return nil
	}).AnyTimes()
	mockStorage.EXPECT().Get(gomock.AssignableToTypeOf(1)).DoAndReturn(func(key interface{}) (interface{}, bool, error) {
		if expire, exist := expires[key]; exist && expire > time.Now().UnixNano() {
			return cached[key], true, nil
		}
		return nil, false, nil
	}).AnyTimes()

	base := 1
	mockQuery := mock_cachex.NewMockQuerier(ctrl)
	mockQuery.EXPECT().Query(gomock.AssignableToTypeOf(1)).DoAndReturn(func(key interface{}) (interface{}, bool, error) {
		num := key.(int)
		result := num * base
		base++
		return result, true, nil
	}).AnyTimes()

	c := NewCachex(mockStorage, mockQuery)

	// 正常返回
	value, err := c.Get(1)
	if assert.NoError(t, err) {
		// 缓存未过期
		newValue, err := c.Get(1)
		assert.NoError(t, err)
		assert.Equal(t, value, newValue)

		// 缓存过期
		time.Sleep(time.Millisecond * 200)
		newValue, err = c.Get(1)
		assert.NoError(t, err)
		assert.NotEqual(t, value, newValue)
	}
}

func TestCachexGetError(t *testing.T) {
	// 查询条件
	storageSetErr := errors.New("storage set error")
	storageGetErr := errors.New("storage get error")
	queryQueryErr := errors.New("query error")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock_cachex.NewMockStorage(ctrl)
	mockStorage.EXPECT().Set(gomock.Not(gomock.Eq(storageSetErr)), gomock.Any()).Return(nil).AnyTimes()
	mockStorage.EXPECT().Set(gomock.Eq(storageSetErr), gomock.Any()).Return(storageSetErr).AnyTimes()
	mockStorage.EXPECT().Get(gomock.Not(gomock.Eq(storageGetErr))).Return(nil, false, nil).AnyTimes()
	mockStorage.EXPECT().Get(gomock.Eq(storageGetErr)).Return(nil, false, storageGetErr).AnyTimes()

	mockQuery := mock_cachex.NewMockQuerier(ctrl)
	mockQuery.EXPECT().Query(gomock.Not(gomock.Eq(queryQueryErr))).Return(nil, true, nil).AnyTimes()
	mockQuery.EXPECT().Query(gomock.Eq(queryQueryErr)).Return(nil, false, queryQueryErr).AnyTimes()

	c := NewCachex(mockStorage, mockQuery)

	conditions := []error{storageSetErr, storageGetErr, queryQueryErr}
	for _, cond := range conditions {
		value, err := c.Get(cond)
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

	cached := make(map[interface{}]interface{})
	mockStorage := mock_cachex.NewMockStorage(ctrl)
	mockStorage.EXPECT().Set(gomock.Any(), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		cached[key] = value
		return nil
	}).AnyTimes()
	mockStorage.EXPECT().Get(gomock.AssignableToTypeOf(1)).DoAndReturn(func(key interface{}) (interface{}, bool, error) {
		if value, exist := cached[key]; exist {
			return value, true, nil
		}
		return nil, false, nil
	}).AnyTimes()

	mockQuery := mock_cachex.NewMockQuerier(ctrl)
	mockQuery.EXPECT().Query(gomock.AssignableToTypeOf(1)).DoAndReturn(func(key interface{}) (interface{}, bool, error) {
		atomic.AddInt64(&total, 1)
		return key, true, nil
	}).AnyTimes()

	c := NewCachex(mockStorage, mockQuery)

	var g errgroup.Group
	for i := 0; i < routines; i++ {
		t.Run("", func(t *testing.T) {
			// t.Parallel()
			for j := 0; j < loopTimes; j++ {
				value, err := c.Get(j)
				// t.Log(value, err)
				if !assert.NoError(t, err) {
					return
				}
				if !assert.Equal(t, j, value.(int)) {
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

	cached := make(map[interface{}]interface{})
	expires := make(map[interface{}]int64)
	mockStorage := mock_cachex.NewMockStorage(ctrl)
	mockStorage.EXPECT().Set(gomock.Any(), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		cached[key] = value
		expires[key] = time.Now().UnixNano() + int64(time.Nanosecond*2)
		return nil
	}).AnyTimes()
	mockStorage.EXPECT().Get(gomock.AssignableToTypeOf(1)).DoAndReturn(func(key interface{}) (interface{}, bool, error) {
		value, exist := cached[key]
		if !exist {
			return nil, false, nil
		}
		expire, exist := expires[key]
		if exist && expire <= time.Now().UnixNano() {
			return value, false, nil
		}
		return value, true, nil
	}).AnyTimes()

	var testError = errors.New("test")
	var returnErr error
	mockQuery := mock_cachex.NewMockQuerier(ctrl)
	mockQuery.EXPECT().Query(gomock.AssignableToTypeOf(1)).DoAndReturn(func(key interface{}) (interface{}, bool, error) {
		if returnErr != nil {
			return nil, false, returnErr
		}
		return time.Now().Nanosecond(), true, nil
	}).AnyTimes()

	c := NewCachex(mockStorage, mockQuery)
	c.UseStaleWhenError(true)

	// 正常返回
	value, err := c.Get(1)
	if assert.NoError(t, err) {
		// 缓存过期，查询出错，返回过期的缓存
		time.Sleep(time.Nanosecond * 2)
		returnErr = testError
		newValue, err := c.Get(1)
		if assert.Equal(t, testError, err) {
			if assert.Equal(t, value, newValue) {
				// 错误排除，返回新的正常数据
				returnErr = nil
				newValue, err = c.Get(1)
				assert.NoError(t, err)
				assert.NotEqual(t, value, newValue)
			}
		}
	}
}
