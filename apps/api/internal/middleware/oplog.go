package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/rest"

	"github.com/yourname/stationery-shop/apps/api/internal/pkg/oplog"
)

// statusRecorder 包装 ResponseWriter 以捕获响应状态码与错误信息
type statusRecorder struct {
	http.ResponseWriter
	status  int
	written bool
	body    []byte // 仅截取少量响应体，用于记录错误信息
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.written {
		r.status = code
		r.written = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.written {
		r.status = http.StatusOK
		r.written = true
	}
	return r.ResponseWriter.Write(b)
}

// NewOpLog 记录每个请求到内存日志存储中，供后台「日志管理」页面查看
func NewOpLog(store *oplog.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 只记录业务接口，跳过健康检查与探测，避免刷屏
			if r.URL.Path == "/api/v1/healthz" || r.URL.Path == "/api/v1/readyz" {
				next.ServeHTTP(w, r)
				return
			}

			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			// 挂载错误信息载体：handler 通过 serverError/badRequest 写入真实原因，
			// 中间件在请求结束后读取并落日志
			var reason string
			r = r.WithContext(oplog.WithReason(r.Context(), &reason))

			start := time.Now()
			next.ServeHTTP(rec, r)
			elapsed := time.Since(start)

			// 注意：本项目的统一响应体把业务错误也以 HTTP 200 返回（错误码在 body 里），
			// 因此除 HTTP 状态码外，还需结合 handler 写入的真实错误原因来判定级别。
			level := "info"
			status := rec.status
			switch {
			case rec.status >= 500:
				level = "error"
			case rec.status >= 400:
				level = "warn"
			case reason != "":
				// 业务层出错（HTTP 200 + 错误码）：标记为 error 便于筛选
				level = "error"
				status = 500
			}

			entry := oplog.Entry{
				Level:      level,
				Method:     r.Method,
				Path:       r.URL.Path,
				Status:     status,
				DurationMs: elapsed.Milliseconds(),
				Ip:         clientIP(r),
				Message:    statusText(status),
			}
			if q := r.URL.RawQuery; q != "" {
				entry.Path = r.URL.Path + "?" + q
			}
			if reason != "" {
				entry.Error = reason
				entry.Message = "business error"
			}
			store.Add(entry)
		})
	}
}

// NewOpLogHandlerFunc 返回 go-zero rest.Middleware 版本
func NewOpLogHandlerFunc(store *oplog.Store) rest.Middleware {
	next := NewOpLog(store)
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return next(handler).ServeHTTP
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xrip := r.Header.Get("X-Real-Ip"); xrip != "" {
		return xrip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func statusText(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "ok"
	case code == http.StatusBadRequest:
		return "bad request"
	case code == http.StatusUnauthorized:
		return "unauthorized"
	case code == http.StatusForbidden:
		return "forbidden"
	case code == http.StatusNotFound:
		return "not found"
	case code == http.StatusTooManyRequests:
		return "rate limited"
	case code >= 500:
		return "server error"
	default:
		return http.StatusText(code)
	}
}
