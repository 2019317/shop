package main

import (
	"flag"
	"fmt"

	_ "github.com/lib/pq" // 注册 PostgreSQL 驱动
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"

	"github.com/yourname/stationery-shop/apps/api/internal/config"
	"github.com/yourname/stationery-shop/apps/api/internal/handler"
	"github.com/yourname/stationery-shop/apps/api/internal/svc"
)

var configFile = flag.String("f", "etc/api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 所有接口统一带 /api/v1 前缀，便于后续版本演进
	server := rest.MustNewServer(c.RestConf, rest.WithPrefix("/api/v1"))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting shop-api at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
