package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/yourname/stationery-shop/apps/api/internal/logic/admin"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/response"
	"github.com/yourname/stationery-shop/apps/api/internal/repo"
	"github.com/yourname/stationery-shop/apps/api/internal/svc"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

// AdminShippingRuleList 运费规则列表
func AdminShippingRuleList(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		data, err := svcCtx.AdminShipping.ListRules(
			r.Context(),
			q.Get("status"),
			parseInt(r, "page", 1),
			parseInt(r, "page_size", 50),
		)
		if err != nil {
			serverError(w, r, err, "query failed")
			return
		}
		response.OK(w, data)
	}
}

// AdminShippingRuleCreate 新建运费规则
func AdminShippingRuleCreate(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminShippingRuleReq
		if err := httpx.Parse(r, &req); err != nil {
			response.BadRequest(w, "invalid params")
			return
		}
		id, err := svcCtx.AdminShipping.CreateRule(r.Context(), req)
		if err != nil {
			if err == admin.ErrInvalidShippingRule {
				response.BadRequest(w, "invalid shipping rule")
				return
			}
			serverError(w, r, err, "create failed")
			return
		}
		response.OK(w, map[string]string{"id": id})
	}
}

// AdminShippingRuleUpdate 更新运费规则
func AdminShippingRuleUpdate(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			id = pathParam(r, "id")
		}
		var req types.AdminShippingRuleReq
		if err := httpx.Parse(r, &req); err != nil {
			response.BadRequest(w, "invalid params")
			return
		}
		if err := svcCtx.AdminShipping.UpdateRule(r.Context(), id, req); err != nil {
			if err == admin.ErrInvalidShippingRule {
				response.BadRequest(w, "invalid shipping rule")
				return
			}
			serverError(w, r, err, "update failed")
			return
		}
		response.OK(w, map[string]string{"id": id})
	}
}

// AdminShippingRuleDelete 删除（停用）运费规则
func AdminShippingRuleDelete(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			id = pathParam(r, "id")
		}
		if err := svcCtx.AdminShipping.DeleteRule(r.Context(), id); err != nil {
			serverError(w, r, err, "delete failed")
			return
		}
		response.OK(w, map[string]string{"id": id})
	}
}

// AdminOrderMarkDelivered 将订单运单标记为已送达
func AdminOrderMarkDelivered(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			id = pathParam(r, "id")
		}
		if err := svcCtx.AdminShipping.MarkDelivered(r.Context(), id); err != nil {
			if err == repo.ErrOrderMissing {
				response.NotFound(w, "order not found")
				return
			}
			if err == admin.ErrNoShipment {
				response.BadRequest(w, "order has no shipment")
				return
			}
			serverError(w, r, err, "update failed")
			return
		}
		response.OK(w, map[string]string{"id": id, "shipment_status": "delivered"})
	}
}
