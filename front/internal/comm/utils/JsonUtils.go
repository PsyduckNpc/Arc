package utils

import "encoding/json"

func MustMarshal(v any) string {
	marshal, _ := json.Marshal(v)
	return string(marshal)
}

func MustUnmarshal[T any](s string) *T {
	var t T
	json.Unmarshal([]byte(s), &t)
	return &t
}
