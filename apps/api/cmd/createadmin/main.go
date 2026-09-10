// Command createadmin 用于初始化/重置后台管理员账号，并自校验密码哈希。
//
// 用法（在仓库根目录执行，需能连上数据库）：
//
//	go run ./apps/api/cmd/createadmin -email admin@example.com -password 'your-password'
//
// 连接串从环境变量 DATABASE_URL 读取；也可用 -dsn 显式指定。
// 行为：邮箱已存在则重置密码（并重新激活账号），否则新建（role=admin）。
// 最后会把库中的哈希读回并用同一套 Verify 校验，便于排查登录失败。
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/yourname/stationery-shop/apps/api/internal/pkg/password"
)

func main() {
	email := flag.String("email", "", "管理员邮箱（必填）")
	pwd := flag.String("password", "", "管理员密码（必填）")
	name := flag.String("name", "Admin", "显示名称")
	dsn := flag.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL 连接串，默认取环境变量 DATABASE_URL")
	flag.Parse()

	if *email == "" || *pwd == "" {
		log.Fatal("必须提供 -email 与 -password")
	}
	if *dsn == "" {
		log.Fatal("缺少数据库连接串：请设置 DATABASE_URL 或使用 -dsn")
	}

	conn := sqlx.NewSqlConn("postgres", *dsn)
	ctx := context.Background()

	hash, err := password.Hash(*pwd)
	if err != nil {
		log.Fatalf("生成密码哈希失败: %v", err)
	}

	var id string
	err = conn.QueryRowCtx(ctx, &id,
		`SELECT id FROM admin.admin_users WHERE lower(email) = lower($1) LIMIT 1`, *email)

	switch {
	case err == nil:
		if _, err := conn.ExecCtx(ctx,
			`UPDATE admin.admin_users SET password_hash=$1, name=$2, status='active', updated_at=now() WHERE id=$3`,
			hash, *name, id); err != nil {
			log.Fatalf("重置密码失败: %v", err)
		}
		fmt.Printf("已重置管理员密码：%s (id=%s)\n", *email, id)

	case err == sql.ErrNoRows:
		if err := conn.QueryRowCtx(ctx, &id,
			`INSERT INTO admin.admin_users (email, password_hash, name, role, status)
			 VALUES ($1,$2,$3,'admin','active') RETURNING id`,
			*email, hash, *name); err != nil {
			log.Fatalf("创建管理员失败: %v", err)
		}
		fmt.Printf("已创建管理员：%s (id=%s)\n", *email, id)

	default:
		log.Fatalf("查询管理员失败: %v", err)
	}

	// 自校验：读回库中哈希，用同一套 Verify 验证
	var saved string
	if err := conn.QueryRowCtx(ctx, &saved,
		`SELECT password_hash FROM admin.admin_users WHERE lower(email) = lower($1) LIMIT 1`,
		*email); err != nil {
		log.Fatalf("读取已保存哈希失败: %v", err)
	}

	fmt.Println("---- 自校验 ----")
	fmt.Printf("[1] 新生成哈希长度 = %d\n", len(hash))
	fmt.Printf("[2] 库中哈希长度   = %d\n", len(saved))
	fmt.Printf("[3] 库中哈希 == 新生成哈希 ? %v\n", saved == hash)
	fmt.Printf("[4] Verify(明文, 库中哈希)   = %v\n", password.Verify(*pwd, saved))
	fmt.Printf("[5] Verify(明文, 新生成哈希) = %v\n", password.Verify(*pwd, hash))
}
