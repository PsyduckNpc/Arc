package model

import "database/sql"

type CDataSrvRela struct {
	RelaId            sql.NullInt32  	`db:"RelaId"`
	ApiId             sql.NullInt32  	`db:"ApiId"`
	SqlLogic          sql.NullString 	`db:"SqlLogic"`
	SqlCondition      sql.NullString 	`db:"SqlCondition"`
	SqlSort           sql.NullString 	`db:"SqlSort"`
	FldTypeObhjName   sql.NullString 	`db:"FldTypeObhjName"`
	DataModelObhjName sql.NullString 	`db:"DataModelObhjName"`
	DataObjId         sql.NullString 	`db:"DataObjId"`
	RelaDataObjId     sql.NullString 	`db:"RelaDataObjId"`
	AttrMapping       sql.NullString 	`db:"AttrMapping"`
	RelaMapping       sql.NullString 	`db:"RelaMapping"`
}











