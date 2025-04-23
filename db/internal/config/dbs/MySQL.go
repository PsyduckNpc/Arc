package dbs

import (
	"Arc/db/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"time"
)

func NewMySQLConnect(config config.MySQLConfig) sqlx.SqlConn {
	mysql := sqlx.NewMysql(config.Username + ":" + config.Password + "@tcp(" + config.Address + ")/Arc")
	db, err := mysql.RawDB()
	if err != nil {
		logx.Error("Attention: MySQL connect error! ", err)
		return mysql
	}
	if config.MaxConnectTime > 0 {
		db.SetConnMaxLifetime(time.Duration(config.MaxConnectTime) * time.Second)
	}
	return mysql
}
