package main

import (
	"Arc/db/work/dbs"
	"flag"
	"fmt"

	"Arc/db/internal/config"
	"Arc/db/internal/server"
	"Arc/db/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/db.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	srv := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		dbs.RegisterDbsServer(grpcServer, server.NewDbsServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer srv.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	srv.Start()
}
