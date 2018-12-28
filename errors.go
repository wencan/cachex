/*
 * 错误接口
 *
 * wencan
 * 2018-12-26
 */

package cachex

// Expired 已过期错误接口
type Expired interface {
	error
	Expired()
}

// NotFound 没找到错误接口
type NotFound interface {
	error
	NotFound()
}
