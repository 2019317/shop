package handler

import (
	"net/http"

	"github.com/yourname/stationery-shop/apps/api/internal/pkg/oplog"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/response"
	"github.com/yourname/stationery-shop/apps/api/internal/svc"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

// AdminLogList 查询请求日志（内存环形缓冲）
func AdminLogList(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		list, total := svcCtx.Logs.Query(oplog.Filter{
			Level:    q.Get("level"),
			Keyword:  q.Get("keyword"),
			Path:     q.Get("path"),
			Page:     parseInt(r, "page", 1),
			PageSize: parseInt(r, "page_size", 50),
		})
		response.OK(w, types.PageData{
			List:     list,
			Total:    int64(total),
			Page:     parseInt(r, "page", 1),
			PageSize: parseInt(r, "page_size", 50),
		})
	}
}

// AdminLogClear 清空请求日志
func AdminLogClear(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		svcCtx.Logs.Clear()
		response.OK(w, map[string]bool{"cleared": true})
	}
}
