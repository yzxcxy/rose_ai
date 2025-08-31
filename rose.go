package main

import (
	"flag"
	"github.com/zeromicro/go-zero/core/logx"
	"rose/internal/config"
	"rose/internal/handler"
	"rose/internal/svc"
	"rose/pkg/snowflake"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/rose-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// Initialize some package

	// Initialize Snowflake ID generator
	err := snowflake.Init(c.SnowFlake.StartTime, c.SnowFlake.WorkerId)
	if err != nil {
		logx.Error(err)
		return
	}

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	logx.Infof("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
