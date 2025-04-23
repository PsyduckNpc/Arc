package model

import "database/sql"

type CenterDataApi struct {
	AfterSql      sql.NullString `db:"AfterSql"`
	SqlParam      sql.NullString `db:"SqlParam"`
	ApiId         sql.NullInt32  `db:"ApiId"`
	CenterName    sql.NullString `db:"CenterName"`
	ApiName       sql.NullString `db:"ApiName"`
	ApiPath       sql.NullString `db:"ApiPath"`
	OpType        sql.NullString `db:"OpType"`
	CallSource    sql.NullString `db:"CallSource"`
	ApiParam      sql.NullString `db:"ApiParam"`
	BeforeSql     sql.NullString `db:"BeforeSql"`
	DecryptFlag   sql.NullString `db:"DecryptFlag"`
	DecryptFld    sql.NullString `db:"DecryptFld"`
	BeforeExtend  sql.NullString `db:"BeforeExtend"`
	BeforeExtend2 sql.NullString `db:"BeforeExtend2"`
}















