package utils

import (
	"Arc/db/internal/comm/utils/xerr"
	"github.com/pkg/errors"
)

// NCopies 返回包含 n 个相同元素的切片 使用泛型需要go1.18+
func NCopies[T any](n int, elem T) ([]T, error) {
	if n < 0 {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "n 必须大于 0")
	}
	slice := make([]T, n)
	for i := range slice {
		slice[i] = elem
	}
	return slice, nil
}

// NCopiesString 只string
func NCopiesString(n int, elem string) ([]string, error) {
	if n < 0 {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "n 必须大于 0")
	}
	slice := make([]string, n)
	for i := range slice {
		slice[i] = elem
	}
	return slice, nil
}
