package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	CenterDataRpc zrpc.RpcClientConf
	SubdomainRpc  zrpc.RpcClientConf
	Redis         redis.RedisKeyConf
}
