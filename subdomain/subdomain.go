package main

import (
	"flag"
	"github.com/zeromicro/go-zero/core/logx"

	"Arc/subdomain/internal/config"
	"Arc/subdomain/internal/server"
	"Arc/subdomain/internal/svc"
	"Arc/subdomain/work/subdomains"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/subdomain.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		subdomains.RegisterSubdomainsServer(grpcServer, server.NewSubdomainsServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	logx.Info("Starting Subdomain rpc server at %s...\n", c.ListenOn)
	s.Start()
}
