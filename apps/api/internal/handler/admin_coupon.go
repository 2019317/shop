package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/yourname/stationery-shop/apps/api/internal/logic/admin"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/response"
	"github.com/yourname/stationery-shop/apps/api/internal/svc"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

// AdminCouponList 优惠券列表
func AdminCouponList(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		data, err := svcCtx.AdminCoupon.List(
			r.Context(),
			q.Get("keyword"),
			q.Get("status"),
			parseInt(r, "page", 1),
			parseInt(r, "page_size", 20),
		)
		if err != nil {
			serverError(w, r, err, "query failed")
			return
		}
		response.OK(w, data)
	}
}

// AdminCouponDetail 优惠券详情
func AdminCouponDetail(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			id = pathParam(r, "id")
		}
		data, err := svcCtx.AdminCoupon.Detail(r.Context(), id)
		if err != nil {
			serverError(w, r, err, "query failed")
			return
		}
		if data == nil {
			response.NotFound(w, "coupon not found")
			return
		}
		response.OK(w, data)
	}
}

// AdminCouponCreate 新建优惠券
func AdminCouponCreate(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminCouponReq
		if err := httpx.Parse(r, &req); err != nil {
			badRequest(w, r, err, "invalid params")
			return
		}
		id, err := svcCtx.AdminCoupon.Create(r.Context(), req)
		if err != nil {
			if err == admin.ErrCouponCodeExists {
				badRequestMsg(w, r, "coupon code already exists", "coupon code already exists")
				return
			}
			if err == admin.ErrInvalidCouponInput {
				badRequestMsg(w, r, "invalid coupon input", "invalid coupon input")
				return
			}
			serverError(w, r, err, "create failed")
			return
		}
		response.OK(w, map[string]string{"id": id})
	}
}

// AdminCouponUpdate 更新优惠券
func AdminCouponUpdate(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			id = pathParam(r, "id")
		}
		var req types.AdminCouponReq
		if err := httpx.Parse(r, &req); err != nil {
			badRequest(w, r, err, "invalid params")
			return
		}
		if err := svcCtx.AdminCoupon.Update(r.Context(), id, req); err != nil {
			if err == admin.ErrInvalidCouponInput {
				badRequestMsg(w, r, "invalid coupon input", "invalid coupon input")
				return
			}
			serverError(w, r, err, "update failed")
			return
		}
		response.OK(w, map[string]string{"id": id})
	}
}

// AdminCouponSetStatus 启用/停用优惠券
func AdminCouponSetStatus(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			id = pathParam(r, "id")
		}
		var req types.SetCouponStatusReq
		if err := httpx.Parse(r, &req); err != nil {
			badRequest(w, r, err, "invalid params")
			return
		}
		if err := svcCtx.AdminCoupon.SetStatus(r.Context(), id, req.Status); err != nil {
			if err == admin.ErrInvalidCouponInput {
				badRequestMsg(w, r, "invalid status", "invalid status")
				return
			}
			serverError(w, r, err, "update failed")
			return
		}
		response.OK(w, map[string]string{"id": id, "status": req.Status})
	}
}

// ---------- 前台 ----------

// CouponValidate 前台试算优惠券
func CouponValidate(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ValidateCouponReq
		if err := httpx.Parse(r, &req); err != nil {
			badRequest(w, r, err, "invalid params")
			return
		}
		data, err := svcCtx.ShopOrder.ValidateCoupon(r.Context(), req.Code, req.Subtotal)
		if err != nil {
			serverError(w, r, err, "validate failed")
			return
		}
		response.OK(w, data)
	}
}

// CouponActive 前台可用券列表
func CouponActive(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := svcCtx.ShopOrder.ActiveCoupons(r.Context())
		if err != nil {
			serverError(w, r, err, "query failed")
			return
		}
		response.OK(w, list)
	}
}
