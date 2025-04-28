package svc

import (
	"Arc/db/work/dbs"
	"Arc/front/internal/config"
	"Arc/subdomain/work/subdomains"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config        config.Config
	CenterDataRpc dbs.DbsClient
	SubdomainRpc  subdomains.SubdomainsClient
	RedisClient   *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {

	//redis服务必须提前启动
	redisClient := redis.MustNewRedis(c.Redis)

	//中心数据服务必须启动
	//centerDataClient, err := zrpc.NewClient(c.CenterDataRpc)
	//if err != nil {
	//	logx.Errorf("新建rpc客户端异常,未在ETCD中找到CenterData.rpc服务名,错误:%+v", err)
	//}

	//中台服务可选启动
	subdomainClient, err := zrpc.NewClient(c.SubdomainRpc)
	if err != nil {
		logx.Errorf("新建rpc客户端异常,未在ETCD中找到Subdomain.rpc服务名,错误:%+v", err)
	}

	return &ServiceContext{
		Config:        c,
		CenterDataRpc: dbs.NewDbsClient(zrpc.MustNewClient(c.CenterDataRpc).Conn()),
		SubdomainRpc:  subdomains.NewSubdomainsClient(subdomainClient.Conn()),
		RedisClient:   redisClient,
	}
}
