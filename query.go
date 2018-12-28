/*
 * 查询接口
 *
 * wencan
 * 2018-12-26
 */

package cachex

// QueryFunc 查询过程签名
type QueryFunc func(key, value interface{}) error

// Query 查询过程实现Querier接口
func (fun QueryFunc) Query(key, value interface{}) error {
	return fun(key, value)
}

// Querier 查询接口
type Querier interface {
	// Query 查询。value必须是非nil指针。没找到返回Expired
	Query(key, value interface{}) error
}
