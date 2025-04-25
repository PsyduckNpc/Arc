package utils

import (
	"Arc/db/internal/comm/utils/xerr"
	"database/sql"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/pkg/errors"
	"reflect"
)

// StructToProto go结构体转Proto文件中的结构体
func StructToProto[T any](src any) (*T, error) {
	dstType := reflect.TypeOf((*T)(nil)).Elem() // 获取目标类型 T 的信息
	dstVal := reflect.New(dstType).Elem()       // 创建目标类型的实例

	srcVal := reflect.ValueOf(src)
	// 如果 src 是指针，取其指向的值
	if srcVal.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
	}

	if srcVal.Kind() != reflect.Struct {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "来源必须是结构体或结构体指针")
	}

	// 遍历目标结构体的所有字段
	for i := 0; i < dstType.NumField(); i++ {
		dstField := dstVal.Field(i)
		dstFieldType := dstType.Field(i)                  // 目标字段的元信息
		srcField := srcVal.FieldByName(dstFieldType.Name) // 按字段名匹配源字段

		if !srcField.IsValid() {
			continue // 源结构体无此字段，跳过
		}

		// 处理 sql.NullString → string 的转换
		if srcField.Type() == reflect.TypeOf(sql.NullString{}) && dstField.Kind() == reflect.String {
			nullString := srcField.Interface().(sql.NullString)
			if nullString.Valid {
				dstField.SetString(nullString.String)
			} else {
				dstField.SetString("") // 默认空字符串，或根据业务需求调整
			}
			continue
		}

		// 其他类型兼容性检查（原逻辑）
		if srcField.Type().AssignableTo(dstField.Type()) {
			dstField.Set(srcField)
		} else {
			return nil, fmt.Errorf("field %s: type mismatch (src %s vs dst %s)",
				dstFieldType.Name, srcField.Type(), dstField.Type())
		}
	}

	return dstVal.Addr().Interface().(*T), nil
}

func SliceToProto[S any, T any](srcSlice []S) ([]*T, error) {
	var result []*T
	for i, v := range srcSlice {
		conv, err := StructToProto[T](v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert element at index %d: %w", i, err)
		}
		result = append(result, conv)
	}
	return result, nil
}

// IsDefaultValue 判断是否是默认值
func IsDefaultValue(v any) bool {
	if v == nil {
		return true
	}

	switch val := v.(type) {
	case int, int8, int16, int32, int64:
		return val == 0
	case uint, uint8, uint16, uint32, uint64:
		return val == 0
	case float32, float64:
		return val == 0.0
	case string:
		return val == ""
	case bool:
		return val == false
	default:
		// 其他复杂类型（如结构体、切片等）
		return reflect.ValueOf(v).IsZero() // 使用反射
	}
}

// 检查JSON的类型 是切片还是结构体
func checkJSONType(jsonStr string) (string, error) {
	var data interface{}
	err := sonic.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return "", errors.Wrapf(xerr.SERVER_COMMON_ERROR, "json格式不规范:%s", jsonStr)
	}

	switch data.(type) {
	case []interface{}:
		return "ARRAY", nil
	case map[string]interface{}:
		return "OBJECT", nil
	default:
		return "", errors.Wrapf(xerr.SERVER_COMMON_ERROR, "json为不匹配的类型,json:%s 类型:%s", jsonStr, reflect.TypeOf(data))
	}
}
