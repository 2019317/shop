package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/yourname/stationery-shop/apps/api/internal/pkg/response"
)

// RateLimiter 内存固定窗口限流。
// 适用于单实例部署；多实例横向扩展时应替换为 Redis 实现的分布式限流。
type RateLimiter struct {
	mu      sync.Mutex
	window  time.Duration
	limit   int
	entries map[string]*windowEntry
}

type windowEntry struct {
	count   int
	resetAt time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		window:  window,
		limit:   limit,
		entries: make(map[string]*windowEntry),
	}
}

// Allow 判断是否放行；窗口过期自动重置。
func (rl *RateLimiter) Allow(key string) bool {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, ok := rl.entries[key]
	if !ok || now.After(entry.resetAt) {
		rl.entries[key] = &windowEntry{count: 1, resetAt: now.Add(rl.window)}
		return true
	}
	if entry.count >= rl.limit {
		return false
	}
	entry.count++
	return true
}

// RateLimit 返回按客户端 IP 限流的 go-zero 中间件。
func RateLimit(limit int, window time.Duration) rest.Middleware {
	rl := NewRateLimiter(limit, window)
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !rl.Allow(clientIP(r)) {
				httpx.WriteJson(w, http.StatusTooManyRequests,
					response.Result{Code: 429, Msg: "too many requests"})
				return
			}
			handler(w, r)
		}
	}
}

// clientIP 提取客户端真实 IP（Nginx 反代场景取 X-Forwarded-For 首个地址）
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i > 0 {
			xff = xff[:i]
		}
		return strings.TrimSpace(xff)
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return strings.TrimSpace(xr)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
