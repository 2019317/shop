package admin

import (
	"context"
	"errors"
	"time"

	"github.com/yourname/stationery-shop/apps/api/internal/pkg/jwt"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/password"
	"github.com/yourname/stationery-shop/apps/api/internal/repo"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type AuthLogic struct {
	adminRepo *repo.AdminUserRepo
	secret    string
	ttl       time.Duration
}

func NewAuthLogic(adminRepo *repo.AdminUserRepo, secret string, ttl time.Duration) *AuthLogic {
	return &AuthLogic{adminRepo: adminRepo, secret: secret, ttl: ttl}
}

func (l *AuthLogic) Login(ctx context.Context, req types.AdminLoginReq) (*types.AdminLoginResp, error) {
	u, err := l.adminRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if u == nil || u.Status != "active" || !password.Verify(req.Password, u.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	token, err := jwt.Sign(jwt.Claims{
		Sub:   u.Id,
		Role:  u.Role,
		Scope: "admin",
	}, l.secret, l.ttl)
	if err != nil {
		return nil, err
	}

	_ = l.adminRepo.UpdateLastLogin(ctx, u.Id)

	return &types.AdminLoginResp{
		Token:     token,
		ExpiresIn: int64(l.ttl.Seconds()),
		Admin: types.AdminInfo{
			Id:    u.Id,
			Email: u.Email,
			Name:  u.Name,
			Role:  u.Role,
		},
	}, nil
}
