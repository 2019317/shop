package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest"

	"github.com/yourname/stationery-shop/apps/api/internal/pkg/jwt"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/response"
)

type ctxKey string

const (
	CtxKeyAdminId   ctxKey = "adminId"
	CtxKeyAdminRole ctxKey = "adminRole"
)

// AdminAuth 校验后台 JWT，要求 scope=admin
func AdminAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r)
			if token == "" {
				response.Unauthorized(w, "missing token")
				return
			}
			claims, err := jwt.Parse(token, secret)
			if err != nil {
				if err == jwt.ErrExpired {
					response.Unauthorized(w, "token expired")
					return
				}
				response.Unauthorized(w, "invalid token")
				return
			}
			if claims.Scope != "admin" {
				response.Forbidden(w, "forbidden")
				return
			}

			ctx := context.WithValue(r.Context(), CtxKeyAdminId, claims.Sub)
			ctx = context.WithValue(ctx, CtxKeyAdminRole, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminAuthHandlerFunc 返回 go-zero 的 rest.Middleware 版本，供 rest.WithMiddlewares 使用
func AdminAuthHandlerFunc(secret string) rest.Middleware {
	next := AdminAuth(secret)
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return next(handler).ServeHTTP
	}
}

func extractBearer(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	return ""
}

// GetAdminId 从请求上下文读取管理员 ID
func GetAdminId(r *http.Request) string {
	if v, ok := r.Context().Value(CtxKeyAdminId).(string); ok {
		return v
	}
	return ""
}
