package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// Healthz 健康检查（供 Nginx / Docker / 监控探针使用）
func Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.OkJson(w, map[string]any{
			"status": "ok",
			"service": "shop-api",
		})
	}
}
