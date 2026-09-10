package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/yourname/stationery-shop/apps/api/internal/pkg/i18n"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/response"
	"github.com/yourname/stationery-shop/apps/api/internal/svc"
)

// ---------- 前台：商品 ----------
func ProductList(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		data, err := svcCtx.ShopProduct.List(
			r.Context(),
			q.Get("category"),
			q.Get("q"),
			q.Get("sort"),
			i18n.FromRequest(r),
			parseInt(r, "page", 1),
			parseInt(r, "page_size", 24),
		)
		if err != nil {
			serverError(w, r, err, "query failed")
			return
		}
		response.OK(w, data)
	}
}

func ProductDetail(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		if slug == "" {
			slug = pathParam(r, "slug")
		}
		data, err := svcCtx.ShopProduct.Detail(r.Context(), slug, i18n.FromRequest(r))
		if err != nil {
			serverError(w, r, err, "query failed")
			return
		}
		if data == nil {
			response.NotFound(w, "product not found")
			return
		}
		response.OK(w, data)
	}
}

func CategoryList(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := svcCtx.ShopProduct.Categories(r.Context(), i18n.FromRequest(r))
		if err != nil {
			serverError(w, r, err, "query failed")
			return
		}
		response.OK(w, list)
	}
}

// ---------- 工具 ----------
func parseInt(r *http.Request, key string, def int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return def
	}
	return v
}

// pathParam 兼容不同 Go 版本的路径参数获取
func pathParam(r *http.Request, key string) string {
	segs := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	for i, seg := range segs {
		if seg == key && i+1 < len(segs) {
			return segs[i+1]
		}
	}
	return ""
}
