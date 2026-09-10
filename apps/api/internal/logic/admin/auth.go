package admin

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/yourname/stationery-shop/apps/api/internal/model"
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

// safeStatus / safeHashLen 供调试日志安全取值（u 可能为 nil）
func safeStatus(u *model.AdminUser) string {
	if u == nil {
		return "<nil>"
	}
	return u.Status
}

func safeHashLen(u *model.AdminUser) int {
	if u == nil {
		return -1
	}
	return len(u.PasswordHash)
}

func (l *AuthLogic) Login(ctx context.Context, req types.AdminLoginReq) (*types.AdminLoginResp, error) {
	u, err := l.adminRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		logx.Errorf("[login-debug] FindByEmail error: %v", err)
		return nil, err
	}
	// 临时调试：定位登录失败原因（排查完毕后移除）
	logx.Infof("[login-debug] email=%q pwdLen=%d u==nil:%v status=%q hashLen=%d verify=%v",
		req.Email, len(req.Password), u == nil, safeStatus(u), safeHashLen(u),
		u != nil && password.Verify(req.Password, u.PasswordHash))

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
