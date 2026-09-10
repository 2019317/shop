package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/yourname/stationery-shop/apps/api/internal/config"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/password"
	"github.com/yourname/stationery-shop/apps/api/internal/repo"
)

// 创建后台管理员：
//   go run ./apps/api/cmd/seed -f apps/api/etc/api.yaml \
//     -email admin@yourbrand.com -password 'xxx' -name 'Admin'
var (
	configFile = flag.String("f", "apps/api/etc/api.yaml", "the config file")
	email      = flag.String("email", "", "admin email")
	pwd        = flag.String("password", "", "admin password")
	name       = flag.String("name", "Admin", "admin display name")
	role       = flag.String("role", "admin", "role: admin | operator")
)

func main() {
	flag.Parse()

	if *email == "" || *pwd == "" {
		fmt.Println("usage: seed -email <email> -password <password> [-name <name>]")
		os.Exit(1)
	}
	if len(*pwd) < 8 {
		fmt.Println("password must be at least 8 characters")
		os.Exit(1)
	}

	var c config.Config
	conf.MustLoad(*configFile, &c)

	conn := sqlx.NewSqlConn("postgres", c.Database.Dsn)
	adminRepo := repo.NewAdminUserRepo(conn)
	ctx := context.Background()

	if exist, err := adminRepo.FindByEmail(ctx, *email); err == nil && exist != nil {
		fmt.Printf("admin %s already exists\n", *email)
		os.Exit(0)
	}

	hash, err := password.Hash(*pwd)
	if err != nil {
		panic(err)
	}

	id, err := adminRepo.Create(ctx, *email, hash, *name, *role)
	if err != nil {
		panic(err)
	}

	fmt.Printf("admin created: id=%s email=%s\n", id, *email)
}
