/*
 * 错误定义
 *
 * wencan
 * 2018-12-26
 */

package driver

import "errors"

var (
	// ErrNotFound 没找到
	ErrNotFound = errors.New("not found")

	// ErrExpired 已过期
	ErrExpired = errors.New("expired")
)
