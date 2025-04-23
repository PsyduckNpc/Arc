package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	MySQLConfig MySQLConfig
}

type MySQLConfig struct {
	Address        string `json:"address,optional"`
	Username       string `json:"username,optional"`
	Password       string `json:"password,optional"`
	MaxConnectTime int    `json:"maxConnectTime,default=0"`
}
