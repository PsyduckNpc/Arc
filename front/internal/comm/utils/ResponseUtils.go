package utils

import (
	"Arc/front/internal/comm/utils/xerr"
	"context"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
)

type Http struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

type Err struct {
	code string
	err  error
}

func Success(data any) *Http {
	return &Http{
		Code: xerr.SUC_CODE,
		Msg:  "success",
		Data: data,
	}
}

func Fail(code, msg string) *Http {
	return &Http{
		Code: code,
		Msg:  msg,
	}
}

func FailData(code, msg string, data any) *Http {
	return &Http{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func SucHandler(ctx context.Context, a any) any {
	return &Http{
		Code: xerr.SUC_CODE,
		Msg:  "success",
		Data: a,
	}
}

func ErrHandler(name string) func(ctx context.Context, err error) (int, any) {
	return func(ctx context.Context, err error) (int, any) {

		logx.WithContext(ctx).Errorf("[%s] 错误日志: %+v", name, err)
		if ce, ok := errors.Cause(err).(*xerr.CodeError); ok {
			return http.StatusOK, Fail(ce.GetErrCode(), ce.GetErrMsg())
		}
		logx.WithContext(ctx).Errorf("[%s] 异常类型错误", name)
		//return http.StatusBadRequest, xerr.SERVER_COMMON_ERROR
		return http.StatusOK, xerr.SERVER_COMMON_ERROR

	}
}
