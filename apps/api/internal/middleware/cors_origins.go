package middleware

import (
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest"
)

// NewCorsOrigins 按配置生成 CORS 中间件。
// 配置为空或 "*" 时允许全部来源（回写请求头中的 Origin，以便携带凭证）。
func NewCorsOrigins(corsOrigins string) func(http.Handler) http.Handler {
	origins := parseOrigins(corsOrigins)
	allowAll := len(origins) == 0 || origins["*"]

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				if allowAll || origins[origin] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

			// 预检请求直接返回，不再向下传递
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// NewCorsHandlerFunc 返回 go-zero 的 rest.Middleware 版本，供 rest.WithMiddlewares 使用
func NewCorsHandlerFunc(corsOrigins string) rest.Middleware {
	next := NewCorsOrigins(corsOrigins)
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return next(handler).ServeHTTP
	}
}

// parseOrigins 解析逗号分隔的来源列表，并去除尾部斜杠
func parseOrigins(raw string) map[string]bool {
	out := make(map[string]bool)
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(o), "/"))
		if o != "" {
			out[o] = true
		}
	}
	return out
}
