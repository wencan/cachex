package cachex

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/wencan/cachex/mock_cachex"
)

func TestNopStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuery := mock_cachex.NewMockQuerier(ctrl)
	mockQuery.EXPECT().Query(gomock.AssignableToTypeOf(1), gomock.Any()).DoAndReturn(func(key, value interface{}) error {
		num := key.(int)
		result := num * num
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(result))
		return nil
	}).AnyTimes()

	c := NewCachex(NopStorage{}, mockQuery)

	var value int
	err := c.Get(10, &value)
	assert.NoError(t, err)
	assert.Equal(t, 100, value)
}
