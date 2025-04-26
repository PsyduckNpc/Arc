package model

const (
	ROOT_OPERATE_TYPE = "RootOperateType"

	ROOT_INSERT = "INSERT"
	ROOT_UPDATE = "UPDATE"
	ROOT_DELETE = "DELETE"
	ROOT_QUERY  = "QUERY"
)

type Root struct {
	RootOperateType string `json:"RootOperateType"`
}
