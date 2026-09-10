package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

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

	// 安全自检：后台 JWT 密钥缺失/过短时直接拒绝启动，避免签发可被伪造的令牌
	if len(c.AdminAuth.AccessSecret) < 32 {
		fmt.Fprintln(os.Stderr, "FATAL: AdminAuth.AccessSecret must be at least 32 chars; generate with: openssl rand -base64 48")
		os.Exit(1)
	}
	if c.Mode == "pro" {
		if c.Payment.WebhookSecret == "" {
			fmt.Fprintln(os.Stderr, "FATAL: Payment.WebhookSecret is required in production mode (used to verify payment webhooks)")
			os.Exit(1)
		}
		if c.CorsOrigins == "" || strings.Contains(c.CorsOrigins, "*") {
			fmt.Fprintf(os.Stderr, "WARNING: CorsOrigins allows all origins (%q); tighten it before going live\n", c.CorsOrigins)
		}
	}

	// 所有接口统一带 /api/v1 前缀，便于后续版本演进
	// 统一 /api/v1 前缀由 handler 注册路由时通过 rest.WithPrefix 承载
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting shop-api at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
