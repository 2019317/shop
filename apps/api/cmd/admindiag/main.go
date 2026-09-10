// Command admindiag 用于排查后台登录失败问题。
//
// 它会依次检查：
//  1. 进程内 password.Hash -> password.Verify 是否自洽
//  2. 从数据库读出的哈希，能否通过 password.Verify
//  3. 该哈希的拆解（分段数、迭代次数、盐长度、密钥长度）
//
// 用法：
//
//	go run ./apps/api/cmd/admindiag -email admin@example.com -password 'xxx' -dsn "$DATABASE_URL"
package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/yourname/stationery-shop/apps/api/internal/pkg/password"
)

func main() {
	email := flag.String("email", "", "管理员邮箱")
	pwd := flag.String("password", "", "待验证密码")
	dsn := flag.String("dsn", os.Getenv("DATABASE_URL"), "数据库连接串")
	flag.Parse()

	if *pwd == "" {
		log.Fatal("必须提供 -password")
	}

	// 1) 进程内自检：Hash 后立刻 Verify
	hash, err := password.Hash(*pwd)
	if err != nil {
		log.Fatalf("Hash 失败: %v", err)
	}
	fmt.Println("[1] 新生成哈希 :", hash)
	fmt.Println("[2] 进程内自检 Verify :", password.Verify(*pwd, hash))

	if *dsn == "" || *email == "" {
		fmt.Println("（未传 -dsn/-email，跳过数据库检查）")
		return
	}

	conn := sqlx.NewSqlConn("postgres", *dsn)
	ctx := context.Background()

	var row struct {
		PasswordHash string `db:"password_hash"`
		Status       string `db:"status"`
	}
	if err := conn.QueryRowCtx(ctx, &row,
		`SELECT password_hash, status FROM admin.admin_users WHERE lower(email) = lower($1) LIMIT 1`,
		*email); err != nil {
		log.Fatalf("查询管理员失败: %v", err)
	}

	fmt.Println("[3] 库中哈希     :", row.PasswordHash)
	fmt.Println("[4] 账号状态     :", row.Status)
	fmt.Println("[5] 库中哈希 Verify :", password.Verify(*pwd, row.PasswordHash))

	// 拆解，便于定位格式/编码问题
	parts := strings.Split(row.PasswordHash, "$")
	fmt.Printf("[6] 分段数=%d 前缀=%q\n", len(parts), parts[0])
	if len(parts) == 4 {
		var iter int
		if _, err := fmt.Sscanf(parts[1], "%d", &iter); err != nil {
			fmt.Println("[6] 迭代次数解析失败:", err)
		} else {
			fmt.Println("[6] 迭代次数:", iter)
		}
		salt, errSalt := base64.RawStdEncoding.DecodeString(parts[2])
		want, errWant := base64.RawStdEncoding.DecodeString(parts[3])
		fmt.Printf("[6] 盐=%d字节(err=%v) 密钥=%d字节(err=%v)\n",
			len(salt), errSalt, len(want), errWant)
	}
}
