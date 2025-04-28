package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	MySQLConfig MySQLConfig
	DBSCache    cache.CacheConf
	LocalCache  []LocalCacheItem
}

type MySQLConfig struct {
	Address        string `json:"address,optional"`
	Username       string `json:"username,optional"`
	Password       string `json:"password,optional"`
	MaxConnectTime int    `json:"maxConnectTime,default=0"`
}

// LocalCacheItem 本地缓存项
type LocalCacheItem struct {
	Name   string `json:",optional"`
	Time   int64  `json:",optional"`
	MaxNum int    `json:",optional"`
}
