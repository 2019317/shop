package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/yourname/stationery-shop/apps/api/internal/model"
)

type AdminUserRepo struct {
	conn sqlx.SqlConn
}

func NewAdminUserRepo(conn sqlx.SqlConn) *AdminUserRepo {
	return &AdminUserRepo{conn: conn}
}

func (r *AdminUserRepo) FindByEmail(ctx context.Context, email string) (*model.AdminUser, error) {
	query := `SELECT id, email, password_hash, name, role, status, last_login_at, created_at, updated_at
		FROM admin.admin_users WHERE lower(email) = lower($1) LIMIT 1`
	var u model.AdminUser
	if err := r.conn.QueryRowCtx(ctx, &u, query, email); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *AdminUserRepo) FindById(ctx context.Context, id string) (*model.AdminUser, error) {
	query := `SELECT id, email, password_hash, name, role, status, last_login_at, created_at, updated_at
		FROM admin.admin_users WHERE id = $1 LIMIT 1`
	var u model.AdminUser
	if err := r.conn.QueryRowCtx(ctx, &u, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *AdminUserRepo) UpdateLastLogin(ctx context.Context, id string) error {
	_, err := r.conn.ExecCtx(ctx,
		`UPDATE admin.admin_users SET last_login_at=$1 WHERE id=$2`, time.Now(), id)
	return err
}

func (r *AdminUserRepo) Create(ctx context.Context, email, passwordHash, name, role string) (string, error) {
	query := `INSERT INTO admin.admin_users (email, password_hash, name, role, status)
		VALUES ($1,$2,$3,$4,'active') RETURNING id`
	var id string
	if err := r.conn.QueryRowCtx(ctx, &id, query, email, passwordHash, name, role); err != nil {
		return "", err
	}
	return id, nil
}
