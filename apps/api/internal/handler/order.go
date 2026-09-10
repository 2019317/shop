package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/yourname/stationery-shop/apps/api/internal/logic/shop"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/response"
	"github.com/yourname/stationery-shop/apps/api/internal/svc"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

// CreateOrder 提交订单
func CreateOrder(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			response.BadRequest(w, "invalid params")
			return
		}
		if req.Email == "" {
			response.BadRequest(w, "email is required")
			return
		}
		if len(req.Items) == 0 {
			response.BadRequest(w, "cart is empty")
			return
		}
		if req.Currency == "" {
			req.Currency = "USD"
		}

		order, err := svcCtx.ShopOrder.CreateOrder(r.Context(), req)
		if err != nil {
			switch err {
			case shop.ErrVariantNotFound:
				response.BadRequest(w, "some items are no longer available")
			case shop.ErrVariantDisabled:
				response.BadRequest(w, "some items are unavailable")
			case shop.ErrEmptyCart:
				response.BadRequest(w, "cart is empty")
			default:
				if err.Error() == "insufficient stock" {
					response.BadRequest(w, "insufficient stock")
					return
				}
				response.ServerError(w, "create order failed: "+err.Error())
			}
			return
		}
		response.OK(w, order)
	}
}

// OrderDetail 按订单号查询（下单成功后用户凭此查询）
func OrderDetail(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderNo := r.PathValue("order_no")
		if orderNo == "" {
			orderNo = pathParam(r, "order_no")
		}
		order, err := svcCtx.ShopOrder.DetailByNo(r.Context(), orderNo)
		if err != nil {
			response.ServerError(w, "query failed")
			return
		}
		if order == nil {
			response.NotFound(w, "order not found")
			return
		}
		response.OK(w, order)
	}
}

// PayNotify 支付回调（当前为 mock 渠道调用；接入 Stripe 后由其 Webhook 调用）
// 幂等由 event_id 唯一约束保证，重复回调不会重复发货
func PayNotify(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PayNotifyReq
		if err := httpx.Parse(r, &req); err != nil {
			response.BadRequest(w, "invalid params")
			return
		}
		if req.Provider == "" {
			req.Provider = svcCtx.Config.Payment.Driver
		}
		if req.Status != "succeeded" {
			response.OK(w, map[string]bool{"handled": false})
			return
		}

		if err := svcCtx.ShopOrder.MarkPaid(r.Context(), req.OrderNo,
			req.Provider, req.EventId, req.AmountCents); err != nil {
			response.ServerError(w, err.Error())
			return
		}
		response.OK(w, map[string]bool{"handled": true})
	}
}
