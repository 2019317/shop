package admin

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/yourname/stationery-shop/apps/api/internal/pkg/jwt"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/password"
	"github.com/yourname/stationery-shop/apps/api/internal/repo"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

// dummyPasswordHash 用于账号不存在时执行等成本哈希比对，避免通过响应时间枚举账号
var dummyPasswordHash = func() string {
	h, _ := password.Hash("stationery-shop-timing-dummy")
	return h
}()

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
		logx.Errorf("admin login lookup error: %v", err)
		return nil, err
	}

	// 无论账号是否存在都执行一次 PBKDF2 比对，抹平时间侧信道
	valid := false
	if u != nil {
		valid = password.Verify(req.Password, u.PasswordHash)
	} else {
		password.Verify(req.Password, dummyPasswordHash)
	}

	if u == nil || u.Status != "active" || !valid {
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
