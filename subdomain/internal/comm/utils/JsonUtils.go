package utils

import (
	"github.com/bytedance/sonic"
)

func MustMarshal(v any) string {
	marshal, _ := sonic.Marshal(v)
	return string(marshal)
}

func MustUnmarshal[T any](s string) *T {
	var t T
	sonic.Unmarshal([]byte(s), &t)
	return &t
}
