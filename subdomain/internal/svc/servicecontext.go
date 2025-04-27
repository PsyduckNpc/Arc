package svc

import (
	"Arc/db/work/dbs"
	"Arc/subdomain/internal/config"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	CenterDataRpc dbs.DbsClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		CenterDataRpc: dbs.NewDbsClient(zrpc.MustNewClient(c.CenterDataRpc).Conn()),
	}
}


