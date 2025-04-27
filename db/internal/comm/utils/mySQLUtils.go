package utils

import (
	"Arc/db/internal/comm/utils/xerr"
	"Arc/db/internal/svc"
	"Arc/db/work/dbs"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func QueryRowSlice[T any](ctx context.Context, svcCtx *svc.ServiceContext, sql string, args ...any) ([]T, error) {
	var t []T
	if err := svcCtx.MySQL.QueryRowsPartialCtx(ctx, &t, sql, args...); err != nil {
		logx.Error("query err:", err.Error())
		return nil, status.Error(511, err.Error())
	}
	logx.Info("query rows: ", len(t))
	return t, nil
}

// QueryRowDataMapVO 执行select sql，将数据库返回信息整合到dbs.DataMapVO中
func QueryRowDataMapVO(ctx context.Context, svcCtx *svc.ServiceContext, sql string, args ...any) (*dbs.DataMapVO, error) {
	logx.Info("执行SQL:[%s] 参数:[%+v]", sql, args)
	db, _ := svcCtx.MySQL.RawDB()
	rows, err := db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, errors.Wrapf(xerr.DB_ERROR, "执行sql查询异常, 异常:%v", err)
	}
	defer rows.Close()

	dataMapVO := &dbs.DataMapVO{}
	columns, err := rows.Columns()
	if err != nil {
		return nil, errors.Wrapf(xerr.DB_ERROR, "查询内容获取行内容异常, 异常:%v", err)
	}

	colTypes, _ := rows.ColumnTypes()

	var rowsAffected int64
	for rows.Next() {
		//values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			// 根据列类型创建对应接收器
			valuePtrs[i] = createScanner(colTypes[i])
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, errors.Wrapf(xerr.DB_ERROR, "查询内容扫描到切片异常, 异常:%v", err)
		}

		anyMap := &dbs.AnyMap{Data: make(map[string]*anypb.Any)}

		for i, col := range columns {
			// 类型转换核心逻辑
			msg, err := convertValueToAny(valuePtrs[i], colTypes[i])
			if err != nil {
				return nil, fmt.Errorf("column %s conversion failed: %w", col, err)
			}

			anyVal, err := anypb.New(msg)
			if err != nil {
				return nil, err
			}
			anyMap.Data[col] = anyVal
		}

		dataMapVO.Maps = append(dataMapVO.Maps, anyMap)
		rowsAffected ++
	}
	dataMapVO.Total = rowsAffected

	return dataMapVO, rows.Err()
}

func ConvertRowsToProto(rows []map[string]any) (*dbs.DataMapVO, error) {
	result := &dbs.DataMapVO{}

	for _, row := range rows {
		anyMap := &dbs.AnyMap{
			Data: make(map[string]*anypb.Any),
		}
		for key, val := range row {
			pbVal, err := structpb.NewValue(val)
			if err != nil {
				return nil, fmt.Errorf("转换 key %s 的值失败: %w", key, err)
			}
			anyVal, err := anypb.New(pbVal)
			if err != nil {
				return nil, fmt.Errorf("封装 key %s 的值失败: %w", key, err)
			}
			anyMap.Data[key] = anyVal
		}
		result.Maps = append(result.Maps, anyMap)
	}
	return result, nil
}

// 创建适合数据库类型的扫描器
func createScanner(ct *sql.ColumnType) interface{} {
	switch ct.DatabaseTypeName() {
	case "BOOL":
		return &sql.NullBool{}
	case "INT", "BIGINT":
		return &sql.NullInt64{}
	case "FLOAT", "DOUBLE":
		return &sql.NullFloat64{}
	case "DATETIME", "TIMESTAMP":
		return &sql.NullTime{}
	default:
		return &sql.NullString{}
	}
}

// 类型转换核心方法
func convertValueToAny(rawValue interface{}, ct *sql.ColumnType) (proto.Message, error) {
	// 处理NULL值
	if scanner, ok := rawValue.(sql.Scanner); ok {
		//if err := scanner.Scan(rawValue); err != nil {
		//	return nil, err
		//}
		if isNull(scanner) {
			return wrapperspb.String("NULL"), nil // 或返回空消息
		}
	}

	// 实际类型转换
	switch v := rawValue.(type) {
	case *sql.NullBool:
		return wrapperspb.Bool(v.Bool), nil
	case *sql.NullInt64:
		if ct.DatabaseTypeName() == "BIGINT" {
			return wrapperspb.Int64(v.Int64), nil
		}
		return wrapperspb.Int32(int32(v.Int64)), nil
	case *sql.NullFloat64:
		return wrapperspb.Double(v.Float64), nil
	case *sql.NullString:
		return wrapperspb.String(v.String), nil
	case *sql.NullTime:
		return timestamppb.New(v.Time), nil
	default:
		// 扩展点：添加自定义类型处理
		return convertComplexType(v, ct)
	}
}

// 复杂类型处理扩展
func convertComplexType(v interface{}, ct *sql.ColumnType) (proto.Message, error) {
	switch ct.DatabaseTypeName() {
	case "JSON":
		return structpb.NewValue(json.RawMessage(v.([]byte)))
	case "DECIMAL":
		return wrapperspb.String(fmt.Sprintf("%v", v)), nil
	default:
		// 最终保底处理
		return wrapperspb.String(fmt.Sprintf("%v", v)), nil
	}
}

// NULL检测工具方法
func isNull(scanner sql.Scanner) bool {
	switch v := scanner.(type) {
	case *sql.NullBool:
		return !v.Valid
	case *sql.NullInt64:
		return !v.Valid
	case *sql.NullFloat64:
		return !v.Valid
	case *sql.NullString:
		return !v.Valid
	case *sql.NullTime:
		return !v.Valid
	default:
		return true
	}
}
