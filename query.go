/*
 * 查询接口
 *
 * wencan
 * 2018-12-26
 */

package cachex

import "context"

// QueryFunc 查询过程签名
type QueryFunc func(ctx context.Context, request, value interface{}) error

// Query 查询过程实现Querier接口
func (fun QueryFunc) Query(ctx context.Context, request, value interface{}) error {
	return fun(ctx, request, value)
}

// Querier 查询接口
type Querier interface {
	// Query 查询。value必须是非nil指针。没找到返回NotFound错误实现
	Query(ctx context.Context, request, value interface{}) error
}
