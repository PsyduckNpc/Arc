package xerr

import (
	"fmt"
)

// 全局错误码
const (
	SUC_CODE                  = "00000"
	SERVER_COMMON_ERROR_CODE  = "10001"
	REUQEST_PARAM_ERROR_CODE  = "20002"
	TOKEN_EXPIRE_ERROR_CODE   = "30003"
	TOKEN_GENERATE_ERROR_CODE = "30004"
	DB_ERROR_CODE             = "40005"
)

var message map[string]string

var SERVER_COMMON_ERROR = NewErrCodeMsg(SERVER_COMMON_ERROR_CODE, message[SERVER_COMMON_ERROR_CODE])
var REUQEST_PARAM_ERROR = NewErrCodeMsg(REUQEST_PARAM_ERROR_CODE, message[REUQEST_PARAM_ERROR_CODE])
var TOKEN_EXPIRE_ERROR = NewErrCodeMsg(TOKEN_EXPIRE_ERROR_CODE, message[TOKEN_EXPIRE_ERROR_CODE])
var TOKEN_GENERATE_ERROR = NewErrCodeMsg(TOKEN_GENERATE_ERROR_CODE, message[TOKEN_GENERATE_ERROR_CODE])
var DB_ERROR = NewErrCodeMsg(DB_ERROR_CODE, message[DB_ERROR_CODE])

func init() {
	message = make(map[string]string)
	message[SUC_CODE] = "success"
	message[SERVER_COMMON_ERROR_CODE] = "服务出错,请稍后再试"
	message[REUQEST_PARAM_ERROR_CODE] = "参数错误"
	message[TOKEN_EXPIRE_ERROR_CODE] = "token失效，请重新登陆"
	message[TOKEN_GENERATE_ERROR_CODE] = "生成token失败"
	message[DB_ERROR_CODE] = "数据库繁忙,请稍后再试"
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
func NewErrCode(errCode string) *CodeError {
	return &CodeError{errCode: errCode, errMsg: MapErrMsg(errCode)}
}

func NewErrMsg(errMsg string) *CodeError {
	return &CodeError{errCode: SERVER_COMMON_ERROR_CODE, errMsg: errMsg}
}

func MapErrMsg(errcode string) string {
	if msg, ok := message[errcode]; ok {
		return msg
	} else {
		return "服务器出错, 请稍后重试"
	}
}

func IsCodeErr(errcode string) bool {
	if _, ok := message[errcode]; ok {
		return true
	} else {
		return false
	}
}
