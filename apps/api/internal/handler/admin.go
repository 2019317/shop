package handler

import (
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/yourname/stationery-shop/apps/api/internal/logic/admin"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/response"
	"github.com/yourname/stationery-shop/apps/api/internal/svc"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

// ---------- 登录 ----------
func AdminLogin(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminLoginReq
		if err := httpx.Parse(r, &req); err != nil {
			response.BadRequest(w, "invalid params")
			return
		}
		resp, err := svcCtx.AdminAuth.Login(r.Context(), req)
		if err != nil {
			// 区分「凭据错误」与「系统错误」：
			// 后者若也返回 401，会把真实的数据库/解析故障伪装成密码错误，极难排查
			if !errors.Is(err, admin.ErrInvalidCredentials) {
				logx.Errorf("admin login error: %v", err)
				response.ServerError(w, "login failed")
				return
			}
			response.Unauthorized(w, "invalid email or password")
			return
		}
		response.OK(w, resp)
	}
}

// ---------- 商品 ----------
func AdminProductList(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := parseInt(r, "page", 1)
		pageSize := parseInt(r, "page_size", 20)
		keyword := r.URL.Query().Get("keyword")

		data, err := svcCtx.AdminProduct.List(r.Context(), keyword, page, pageSize)
		if err != nil {
			response.ServerError(w, "query failed")
			return
		}
		response.OK(w, data)
	}
}

func AdminProductDetail(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			id = pathParam(r, "id")
		}
		data, err := svcCtx.AdminProduct.Detail(r.Context(), id)
		if err != nil {
			response.ServerError(w, "query failed")
			return
		}
		if data == nil {
			response.NotFound(w, "product not found")
			return
		}
		response.OK(w, data)
	}
}

func AdminProductCreate(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminProductReq
		if err := httpx.Parse(r, &req); err != nil {
			response.BadRequest(w, "invalid params")
			return
		}
		if req.Title == "" || req.Slug == "" {
			response.BadRequest(w, "title and slug are required")
			return
		}
		id, err := svcCtx.AdminProduct.Create(r.Context(), req)
		if err != nil {
			logx.Errorf("product create error: %v", err)
			response.ServerError(w, "create failed")
			return
		}
		response.OK(w, map[string]string{"id": id})
	}
}

func AdminProductUpdate(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			id = pathParam(r, "id")
		}
		var req types.AdminProductReq
		if err := httpx.Parse(r, &req); err != nil {
			response.BadRequest(w, "invalid params")
			return
		}
		if err := svcCtx.AdminProduct.Update(r.Context(), id, req); err != nil {
			logx.Errorf("product update error: %v", err)
			response.ServerError(w, "update failed")
			return
		}
		response.OK(w, map[string]string{"id": id})
	}
}

func AdminProductDelete(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			id = pathParam(r, "id")
		}
		if err := svcCtx.AdminProduct.Delete(r.Context(), id); err != nil {
			response.ServerError(w, "delete failed")
			return
		}
		response.OK(w, map[string]string{"id": id})
	}
}

// ---------- 类目 ----------
func AdminCategoryList(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := svcCtx.AdminCategory.List(r.Context())
		if err != nil {
			response.ServerError(w, "query failed")
			return
		}
		response.OK(w, list)
	}
}

func AdminCategoryCreate(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminCategoryReq
		if err := httpx.Parse(r, &req); err != nil {
			response.BadRequest(w, "invalid params")
			return
		}
		if req.Name == "" || req.Slug == "" {
			response.BadRequest(w, "name and slug are required")
			return
		}
		id, err := svcCtx.AdminCategory.Create(r.Context(), req)
		if err != nil {
			logx.Errorf("category create error: %v", err)
			response.ServerError(w, "create failed")
			return
		}
		response.OK(w, map[string]string{"id": id})
	}
}

func AdminCategoryUpdate(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			id = pathParam(r, "id")
		}
		var req types.AdminCategoryReq
		if err := httpx.Parse(r, &req); err != nil {
			response.BadRequest(w, "invalid params")
			return
		}
		if err := svcCtx.AdminCategory.Update(r.Context(), id, req); err != nil {
			response.ServerError(w, "update failed")
			return
		}
		response.OK(w, map[string]string{"id": id})
	}
}

func AdminCategoryDelete(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			id = pathParam(r, "id")
		}
		if err := svcCtx.AdminCategory.Delete(r.Context(), id); err != nil {
			response.ServerError(w, "delete failed")
			return
		}
		response.OK(w, map[string]string{"id": id})
	}
}

// ---------- 图片直传 ----------
func AdminAssetPresign(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PresignReq
		if err := httpx.Parse(r, &req); err != nil {
			response.BadRequest(w, "invalid params")
			return
		}
		resp, err := svcCtx.AdminAsset.Presign(r.Context(), req)
		if err != nil {
			if err == admin.ErrInvalidFileType {
				response.BadRequest(w, "unsupported file type")
				return
			}
			logx.Errorf("asset presign error: %v", err)
			response.ServerError(w, "presign failed")
			return
		}
		response.OK(w, resp)
	}
}
