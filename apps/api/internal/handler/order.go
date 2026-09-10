package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/yourname/stationery-shop/apps/api/internal/logic/shop"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/response"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/validate"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/webhook"
	"github.com/yourname/stationery-shop/apps/api/internal/svc"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

// CreateOrder 提交订单（价格一律服务端计算）
func CreateOrder(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			badRequest(w, r, err, "invalid params")
			return
		}
		if !validate.Email(req.Email) {
			badRequestMsg(w, r, "invalid email", "invalid email")
			return
		}
		if len(req.Items) == 0 {
			badRequestMsg(w, r, "cart is empty", "cart is empty")
			return
		}
		if req.Currency == "" {
			req.Currency = "USD"
		}
		if !validate.Currency(req.Currency) {
			badRequestMsg(w, r, "invalid currency", "invalid currency")
			return
		}
		if !validate.MaxLen(req.CustomerNote, 2000) {
			badRequestMsg(w, r, "customer note is too long", "customer note is too long")
			return
		}
		if country, ok := req.ShippingAddress["country"].(string); !ok || !validate.Country(country) {
			badRequestMsg(w, r, "invalid shipping country", "invalid shipping country")
			return
		}

		order, err := svcCtx.ShopOrder.CreateOrder(r.Context(), req)
		if err != nil {
			logx.Errorf("create order error: %v", err)
			switch err {
			case shop.ErrVariantNotFound, shop.ErrVariantDisabled, shop.ErrEmptyCart:
				badRequestMsg(w, r, err.Error(), "some items are unavailable")
			case shop.ErrInvalidCoupon:
				badRequestMsg(w, r, err.Error(), "invalid coupon code")
			case shop.ErrCouponExpired:
				badRequestMsg(w, r, err.Error(), "coupon has expired")
			case shop.ErrCouponNotStarted:
				badRequestMsg(w, r, err.Error(), "coupon is not yet active")
			case shop.ErrCouponExhausted:
				badRequestMsg(w, r, err.Error(), "coupon has reached its usage limit")
			case shop.ErrCouponMinAmount:
				badRequestMsg(w, r, err.Error(), "order does not meet the coupon minimum amount")
			case shop.ErrNoShippingRule:
				badRequestMsg(w, r, err.Error(), "no shipping method available for the destination")
			default:
				if err.Error() == "insufficient stock" {
					badRequestMsg(w, r, err.Error(), "insufficient stock")
					return
				}
				serverError(w, r, err, "create order failed")
			}
			return
		}
		response.OK(w, order)
	}
}

// OrderDetail 按订单号 + 下单邮箱查询（保护用户隐私，防止订单号被枚举）
func OrderDetail(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderNo := r.PathValue("order_no")
		if orderNo == "" {
			orderNo = pathParam(r, "order_no")
		}
		email := r.URL.Query().Get("email")
		order, err := svcCtx.ShopOrder.DetailByNo(r.Context(), orderNo, email)
		if err != nil {
			serverError(w, r, err, "query failed")
			return
		}
		if order == nil {
			response.NotFound(w, "order not found")
			return
		}
		response.OK(w, order)
	}
}

// PayNotify 支付回调（mock 联调 / Stripe Webhook）。
// 幂等由 event_id 唯一约束保证；生产环境强制校验 HMAC 签名。
func PayNotify(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			badRequest(w, r, err, "invalid params")
			return
		}
		defer r.Body.Close()

		if !authorizeWebhook(svcCtx, rawBody, r) {
			response.Unauthorized(w, "invalid signature")
			return
		}

		var req types.PayNotifyReq
		if err := json.Unmarshal(rawBody, &req); err != nil {
			badRequest(w, r, err, "invalid params")
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
			logx.Errorf("mark paid error: %v", err)
			if err == shop.ErrAmountMismatch {
				badRequestMsg(w, r, err.Error(), "amount mismatch")
				return
			}
			if err == shop.ErrInvalidOrderState || err == shop.ErrOrderNotFound {
				badRequestMsg(w, r, err.Error(), "invalid order")
				return
			}
			serverError(w, r, err, "payment notify failed")
			return
		}
		response.OK(w, map[string]bool{"handled": true})
	}
}

// authorizeWebhook 校验回调签名：配置了 WebhookSecret 则强制校验；
// 未配置时生产环境直接拒绝（fail closed），开发/测试允许 mock 自通知。
func authorizeWebhook(svcCtx *svc.ServiceContext, rawBody []byte, r *http.Request) bool {
	secret := svcCtx.Config.Payment.WebhookSecret
	sig := r.Header.Get("X-Webhook-Signature")
	if secret != "" {
		return webhook.VerifySignature(secret, string(rawBody), sig)
	}
	return svcCtx.Config.Mode != "pro"
}
