// Code generated by MockGen. DO NOT EDIT.
// Source: query.go

// Package mock_cachex is a generated GoMock package.
package mock_cachex

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockQuerier is a mock of Querier interface
type MockQuerier struct {
	ctrl     *gomock.Controller
	recorder *MockQuerierMockRecorder
}

// MockQuerierMockRecorder is the mock recorder for MockQuerier
type MockQuerierMockRecorder struct {
	mock *MockQuerier
}

// NewMockQuerier creates a new mock instance
func NewMockQuerier(ctrl *gomock.Controller) *MockQuerier {
	mock := &MockQuerier{ctrl: ctrl}
	mock.recorder = &MockQuerierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockQuerier) EXPECT() *MockQuerierMockRecorder {
	return m.recorder
}

// Query mocks base method
func (m *MockQuerier) Query(ctx context.Context, request, value interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Query", ctx, request, value)
	ret0, _ := ret[0].(error)
	return ret0
}

// Query indicates an expected call of Query
func (mr *MockQuerierMockRecorder) Query(ctx, request, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockQuerier)(nil).Query), ctx, request, value)
}
