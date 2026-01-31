// Package utils 提供通用工具函数
package utils

import "io"

// CloseAll 安全地关闭多个资源，收集所有错误
// 如果有多个错误发生，会返回第一个错误
func CloseAll(closers ...io.Closer) error {
	var firstErr error
	for _, c := range closers {
		if c == nil {
			continue
		}
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
