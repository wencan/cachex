/*
 * Storage的错误接口
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

// NotFound 没找到错误接口。
// 需要存储后端找不到缓存时，返回一个实现了NotFound接口的错误。
type NotFound interface {
	error
	NotFound()
}
