package model

type Page struct {
	CurPage  int64 `json:"curPage"`  //当前页码
	PageSize int64 `json:"pageSize"` //页大小
	AllNum   int64 `json:"allNum"`   //总数
	//PageNum  int64 `json:"pageNum"`  //页数
	//ExpFlag  bool  `json:"expFlag"`  //导出标志
}
