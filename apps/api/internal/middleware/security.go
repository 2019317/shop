package middleware

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

// NewSecurityHeaders 返回统一的安全响应头中间件。
// 设置常见的防护头，降低 XSS / 点击劫持 / MIME 嗅探等风险。
func NewSecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "no-referrer")
			h.Set("X-XSS-Protection", "1; mode=block")
			h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			next.ServeHTTP(w, r)
		})
	}
}

// NewSecurityHeadersHandlerFunc 返回 go-zero 的 rest.Middleware 版本
func NewSecurityHeadersHandlerFunc() rest.Middleware {
	next := NewSecurityHeaders()
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return next(handler).ServeHTTP
	}
}
