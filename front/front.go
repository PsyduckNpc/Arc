package main

import (
	"Arc/front/internal/comm/utils"
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/rest/httpx"

	"Arc/front/internal/config"
	"Arc/front/internal/handler"
	"Arc/front/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/front.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	//server.Use()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	httpx.SetOkHandler(utils.SucHandler)
	httpx.SetErrorHandlerCtx(utils.ErrHandler(c.Name))

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
