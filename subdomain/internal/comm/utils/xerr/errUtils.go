package xerr

import (
	"fmt"
)

//错误输出方法:
//return slice, errors.Wrapf(xerr.REUQEST_PARAM_ERROR, "入参数有误,不符合json结构,检查ApiParam, 错误:%v", err)
//return slice, errors.WithMessagef(xerr.REUQEST_PARAM_ERROR, "入参数有误,不符合json结构,检查ApiParam, 错误:%v", err)
//Wrapf 参数1为返回给前端的错误信息, 参数2及以后为错误日志输出
//带有errors.Wrapf()会输出错误堆栈信息 WithMessagef不会输出堆栈信息

//

// 全局错误码
const (
	SUC_CODE                  = "00000"
	SERVER_COMMON_ERROR_CODE  = "10001"
	DATA_ERROR_CODE           = "10002"
	REUQEST_PARAM_ERROR_CODE  = "20002"
	TOKEN_EXPIRE_ERROR_CODE   = "30003"
	TOKEN_GENERATE_ERROR_CODE = "30004"
	DB_ERROR_CODE             = "40005"
)

var SERVER_COMMON_ERROR = NewErrCodeMsg(SERVER_COMMON_ERROR_CODE, "服务出错,请稍后再试")
var REUQEST_PARAM_ERROR = NewErrCodeMsg(REUQEST_PARAM_ERROR_CODE, "参数错误")
var TOKEN_EXPIRE_ERROR = NewErrCodeMsg(TOKEN_EXPIRE_ERROR_CODE, "token失效，请重新登陆")
var TOKEN_GENERATE_ERROR = NewErrCodeMsg(TOKEN_GENERATE_ERROR_CODE, "生成token失败")
var DB_ERROR = NewErrCodeMsg(DB_ERROR_CODE, "数据库繁忙,请稍后再试")
var PERSON_ERROR = func(format string, args ...interface{}) *CodeError { //自定义错误
	return NewErrCodeMsg(SERVER_COMMON_ERROR_CODE, fmt.Sprintf(format, args...))
}

type CodeError struct {
	errCode string
	errMsg  string
}

// 返回给前端的错误码
func (e *CodeError) GetErrCode() string {
	return e.errCode
}

// 返回给前端显示端错误信息
func (e *CodeError) GetErrMsg() string {
	return e.errMsg
}

func (e *CodeError) Error() string {
	return fmt.Sprintf("错误码:%d，错误信息:%s", e.errCode, e.errMsg)
}

func NewErrCodeMsg(errCode, errMsg string) *CodeError {
	return &CodeError{errCode: errCode, errMsg: errMsg}
}
