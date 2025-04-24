package model

const (
	ROOT_OPERATE_TYPE = "RootOperateType"

	ROOT_INSERT = "INSERT"
	ROOT_UPDATE = "UPDATE"
	ROOT_DELETE = "DELETE"
	ROOT_QUERY = "QUERY"
)

type Root struct {
	RootOperateType string `json:"RootOperateType"`
}

func (r *Root) getOperateType() string {
	return r.RootOperateType
}

func (r *Root) creat() {
	r.RootOperateType = ROOT_INSERT
}
func (r *Root) update() {
	r.RootOperateType = ROOT_UPDATE
}
func (r *Root) delete() {
	r.RootOperateType = ROOT_DELETE
}
func (r *Root) query()  {
	r.RootOperateType = ROOT_QUERY
}