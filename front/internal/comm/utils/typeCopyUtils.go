package utils

import (
	"Arc/db/work/dbs"
	"Arc/front/internal/comm/utils/xerr"
	"fmt"
	gogo "github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"reflect"
)

// StructCopy 结构体拷贝
func StructCopy[T any](source any) (T, error) {
	var t T

	// 检查 source 是否为指针
	srcVal := reflect.ValueOf(source)
	if srcVal.Kind() != reflect.Ptr {
		return t, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "结构体拷贝来源参数必须是指针")
	}

	// 获取 source 节点代表的值
	srcVal = srcVal.Elem()
	if srcVal.Kind() != reflect.Struct {
		return t, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "结构体拷贝来源参数必须指向结构体")
	}

	// 获取目标值（t）的反射值，t 本身为零值，需要取地址后 Elem 才能赋值
	tgtVal := reflect.ValueOf(&t).Elem()
	if tgtVal.Kind() != reflect.Struct {
		return t, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "结构体拷贝目标必须是结构体")
	}

	// 遍历源结构体的每个字段，查找目标结构体中同名且类型相同的字段，然后赋值
	srcType := srcVal.Type()
	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		srcFieldName := srcType.Field(i).Name

		tgtField := tgtVal.FieldByName(srcFieldName)
		if tgtField.IsValid() && tgtField.CanSet() && srcField.Type() == tgtField.Type() {
			tgtField.Set(srcField)
		}
	}

	return t, nil
}

// SliceCopy 切片深拷贝
func SliceCopy[S any, T any](src []S) ([]T, error) {
	var result []T
	for i, v := range src {
		conv, err := StructCopy[T](v)
		if err != nil {
			return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "无法在索引处转换元素, 索引下标%d, 错误%w", i, err)
		}
		result = append(result, conv)
	}
	return result, nil
}

// ProtoToSlice 将 DataMapVO 转换为目标结构体切片
// 要求目标结构体字段通过 protobuf 标签声明字段映射关系
// 入参 dm 中心数据微服务返回参数
// 出参 []T 结果切片
// 出参 int64 总数量
// 出参 error 错误异常
func ProtoToSlice[T any](dm *dbs.DataMapVO) ([]T, int64, error) {
	var result []T

	if dm == nil {
		return result, 0, nil
	}

	// 获取目标类型的反射信息
	targetType := reflect.TypeOf((*T)(nil)).Elem()
	if targetType.Kind() != reflect.Struct {
		return nil, 0, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "目标类型必须是结构体")
	}

	for _, anyMap := range dm.Maps {
		// 创建新实例
		instance := reflect.New(targetType).Elem()

		// 遍历结构体字段
		for i := 0; i < targetType.NumField(); i++ {
			field := targetType.Field(i)
			tag := field.Tag.Get("db")
			if tag == "" {
				//continue // 跳过无标签字段
				tag = field.Name //没有标签默认使用字段名
			}

			// 解析 protobuf 标签获取字段名
			//parts := strings.Split(tag, ",")
			//
			//var fieldName string
			//for _, part := range parts {
			//	if strings.HasPrefix(part, "name=") {
			//		fieldName = strings.TrimPrefix(part, "name=")
			//		break
			//	}
			//}
			fieldName := tag

			// 获取对应的 Any 值
			anyVal, exists := anyMap.Data[fieldName]
			if !exists || anyVal == nil {
				continue // 字段不存在或值为空
			}

			// 类型转换
			value, err := anyToGoType(anyVal, field.Type)
			if err != nil {
				return nil, 0, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "字段 %s 转换 错误: %w", fieldName, err)
			}

			// 设置字段值
			if value != nil {
				instance.Field(i).Set(reflect.ValueOf(value))
			}
		}

		result = append(result, instance.Interface().(T))
	}

	return result, dm.Total, nil
}

// anyToGoType 将 Any 转换为具体的 Go 类型
func anyToGoType(anyVal *anypb.Any, targetType reflect.Type) (interface{}, error) {
	// 处理常见包装类型
	switch {
	case anyVal.MessageIs(&wrapperspb.StringValue{}):
		s := &wrapperspb.StringValue{}
		if err := anyVal.UnmarshalTo(s); err != nil {
			return nil, err
		}
		return s.Value, nil

	case anyVal.MessageIs(&wrapperspb.Int32Value{}):
		i := &wrapperspb.Int32Value{}
		if err := anyVal.UnmarshalTo(i); err != nil {
			return nil, err
		}
		return i.Value, nil

	case anyVal.MessageIs(&wrapperspb.Int64Value{}):
		i := &wrapperspb.Int64Value{}
		if err := anyVal.UnmarshalTo(i); err != nil {
			return nil, err
		}
		return i.Value, nil

	case anyVal.MessageIs(&wrapperspb.DoubleValue{}):
		f := &wrapperspb.DoubleValue{}
		if err := anyVal.UnmarshalTo(f); err != nil {
			return nil, err
		}
		return f.Value, nil

	case anyVal.MessageIs(&wrapperspb.BoolValue{}):
		b := &wrapperspb.BoolValue{}
		if err := anyVal.UnmarshalTo(b); err != nil {
			return nil, err
		}
		return b.Value, nil

	default:
		// 处理自定义消息类型
		msgType := gogo.MessageType(anyVal.TypeUrl)
		if msgType == nil {
			return nil, fmt.Errorf("未知的Message类型: %s", anyVal.TypeUrl)
		}

		msg := reflect.New(msgType.Elem()).Interface().(proto.Message)
		if err := anyVal.UnmarshalTo(msg); err != nil {
			return nil, err
		}

		// 如果目标类型是 proto.Message 则直接返回
		if targetType.Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) {
			return msg, nil
		}

		// 否则尝试提取基础类型值
		return extractProtoValue(msg)
	}
}

// extractProtoValue 从 proto.Message 提取基础值
func extractProtoValue(msg proto.Message) (interface{}, error) {
	switch v := msg.(type) {
	case *wrapperspb.StringValue:
		return v.Value, nil
	case *wrapperspb.Int32Value:
		return v.Value, nil
	case *wrapperspb.Int64Value:
		return v.Value, nil
	case *wrapperspb.DoubleValue:
		return v.Value, nil
	case *wrapperspb.BoolValue:
		return v.Value, nil
	default:
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "不支持的Proto message类型:%T", msg)
	}
}
