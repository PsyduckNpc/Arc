package main

import (
	"Arc/front/internal/comm/utils"
	"flag"
	"github.com/zeromicro/go-zero/core/logx"
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

	server := rest.MustNewServer(c.RestConf) //初始化服务配置
	defer server.Stop()

	//server.Use()

	ctx := svc.NewServiceContext(c)       //初始化服务上下文
	handler.RegisterHandlers(server, ctx) //注册路由

	httpx.SetOkHandler(utils.SucHandler)               //全局正常处理
	httpx.SetErrorHandlerCtx(utils.ErrHandler(c.Name)) //全局异常处理

	logx.Info("Starting Front server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
