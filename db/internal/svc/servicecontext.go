package svc

import (
	"Arc/db/internal/config"
	"Arc/db/internal/config/dbs"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config
	MySQL  sqlx.SqlConn
}

func NewServiceContext(c config.Config) *ServiceContext {
	connect := dbs.NewMySQLConnect(c.MySQLConfig)
	return &ServiceContext{
		Config: c,
		MySQL:  connect,
	}
}
