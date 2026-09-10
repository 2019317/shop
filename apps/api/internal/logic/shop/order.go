package shop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yourname/stationery-shop/apps/api/internal/model"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/event"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/payment"
	"github.com/yourname/stationery-shop/apps/api/internal/repo"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

var (
	ErrEmptyCart         = errors.New("cart is empty")
	ErrVariantNotFound   = errors.New("variant not found")
	ErrVariantDisabled   = errors.New("variant is not available")
	ErrOrderNotFound     = errors.New("order not found")
	ErrInvalidOrderState = errors.New("invalid order state")
	ErrAmountMismatch    = errors.New("payment amount mismatch")
	ErrInvalidCoupon     = errors.New("invalid coupon code")
	ErrCouponExpired     = errors.New("coupon has expired")
	ErrCouponNotStarted  = errors.New("coupon is not yet active")
	ErrCouponExhausted   = errors.New("coupon has reached its usage limit")
	ErrCouponMinAmount   = errors.New("order does not meet the coupon minimum amount")
	ErrNoShippingRule    = errors.New("no shipping method available for the destination")
)

type OrderLogic struct {
	orderRepo     *repo.OrderRepo
	productRepo   *repo.ProductRepo
	inventoryRepo *repo.InventoryRepo
	shippingRepo  *repo.ShippingRepo
	couponRepo    *repo.CouponRepo
	gateway       payment.Gateway
	bus           *event.Bus
	publicURL     func(objectKey string) string
}

func NewOrderLogic(
	orderRepo *repo.OrderRepo,
	productRepo *repo.ProductRepo,
	inventoryRepo *repo.InventoryRepo,
	shippingRepo *repo.ShippingRepo,
	couponRepo *repo.CouponRepo,
	gateway payment.Gateway,
	bus *event.Bus,
	publicURL func(objectKey string) string,
) *OrderLogic {
	return &OrderLogic{
		orderRepo:     orderRepo,
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
		shippingRepo:  shippingRepo,
		couponRepo:    couponRepo,
		gateway:       gateway,
		bus:           bus,
		publicURL:     publicURL,
	}
}

// CreateOrder 下单主流程
// 关键：商品价格一律以服务端数据库为准，绝不信任前端传入的金额
func (l *OrderLogic) CreateOrder(ctx context.Context, req types.CreateOrderReq) (*types.OrderVO, error) {
	if len(req.Items) == 0 {
		return nil, ErrEmptyCart
	}

	// 1. 逐项校验并计算小计（价格来自数据库）
	var subtotal int64
	var weightG int
	items := make([]repo.CreateOrderItem, 0, len(req.Items))

	for _, item := range req.Items {
		if item.Qty <= 0 || item.Qty > 99 {
			return nil, fmt.Errorf("invalid quantity for sku %s", item.SkuCode)
		}

		variant, err := l.productRepo.FindVariantBySku(ctx, item.SkuCode)
		if err != nil {
			return nil, err
		}
		if variant == nil {
			return nil, ErrVariantNotFound
		}
		if variant.Status != "active" {
			return nil, ErrVariantDisabled
		}

		lineTotal := variant.PriceCents * int64(item.Qty)
		subtotal += lineTotal
		weightG += variant.WeightG * item.Qty

		opts := map[string]interface{}{}
		_ = json.Unmarshal([]byte(variant.Options), &opts)

		imageUrl := ""
		if variant.ImageKey != "" {
			imageUrl = l.publicURL(variant.ImageKey)
		}

		items = append(items, repo.CreateOrderItem{
			VariantId:      variant.Id,
			SkuCode:        variant.SkuCode,
			Title:          variant.Title,
			Options:        opts,
			ImageUrl:       imageUrl,
			UnitPriceCents: variant.PriceCents,
			Qty:            item.Qty,
		})
	}

	// 2. 计算优惠：服务端校验券码，绝不信任前端传入的 discount_cents
	coupon, err := l.applyCoupon(ctx, req.CouponCode, subtotal)
	if err != nil {
		return nil, err
	}
	discountCents := int64(0)
	var couponId *string
	couponCode := ""
	if coupon != nil {
		discountCents = coupon.DiscountCents
		couponId = &coupon.Id
		couponCode = coupon.Code
	}

	// 3. 计算运费
	country := ""
	if v, ok := req.ShippingAddress["country"].(string); ok {
		country = v
	}
	shippingCents, shippingName, err := l.shippingRepo.Match(ctx, country, subtotal, weightG)
	if err != nil {
		return nil, err
	}
	if shippingName == "" {
		// 未命中任何运费规则时不静默免运费，避免资损
		return nil, ErrNoShippingRule
	}

	total := subtotal + shippingCents - discountCents
	if total < 0 {
		total = 0
	}

	// 4. 创建订单（pending）
	order, err := l.orderRepo.Create(ctx, repo.CreateOrderInput{
		Email:           req.Email,
		Currency:        req.Currency,
		SubtotalCents:   subtotal,
		ShippingCents:   shippingCents,
		DiscountCents:   discountCents,
		TotalCents:      total,
		CouponId:        couponId,
		CouponCode:      couponCode,
		ShippingAddress: req.ShippingAddress,
		BillingAddress:  req.BillingAddress,
		CustomerNote:    req.CustomerNote,
		Items:           items,
	})
	if err != nil {
		return nil, err
	}

	// 5. 预占库存（按 SKU 聚合数量）
	aggregated := map[string]int{}
	for _, it := range items {
		aggregated[it.VariantId] += it.Qty
	}
	for variantId, qty := range aggregated {
		if _, err := l.inventoryRepo.Reserve(ctx, variantId, qty, "order", order.Id); err != nil {
			// 预占失败：取消订单并释放已预占的部分
			_ = l.inventoryRepo.Release(ctx, "order", order.Id)
			_ = l.orderRepo.UpdateStatus(ctx, order.Id, "cancelled", "system", "stock reservation failed")
			return nil, err
		}
	}

	// 6. 创建支付意图
	intent, err := l.gateway.Create(ctx, payment.CreateInput{
		OrderId:     order.Id,
		OrderNo:     order.OrderNo,
		AmountCents: total,
		Currency:    req.Currency,
		Email:       req.Email,
		Description: fmt.Sprintf("Order %s", order.OrderNo),
	})
	if err != nil {
		_ = l.inventoryRepo.Release(ctx, "order", order.Id)
		_ = l.orderRepo.UpdateStatus(ctx, order.Id, "cancelled", "system", "payment creation failed")
		return nil, err
	}

	if err := l.orderRepo.CreatePayment(ctx, order.Id, l.gateway.Name(),
		intent.ProviderIntentId, total, req.Currency); err != nil {
		return nil, err
	}

	// 7. 原子占用优惠券名额，防止并发超发；失败则回滚本次下单
	if couponId != nil {
		ok, err := l.couponRepo.IncrementUsage(ctx, *couponId)
		if err != nil {
			_ = l.inventoryRepo.Release(ctx, "order", order.Id)
			_ = l.orderRepo.UpdateStatus(ctx, order.Id, "cancelled", "system", "coupon reservation failed")
			return nil, err
		}
		if !ok {
			_ = l.inventoryRepo.Release(ctx, "order", order.Id)
			_ = l.orderRepo.UpdateStatus(ctx, order.Id, "cancelled", "system", "coupon exhausted")
			return nil, ErrCouponExhausted
		}
	}

	// 8. 发布领域事件（后续可挂接：发送确认邮件、库存同步等）
	l.bus.Publish(ctx, event.Event{
		Type: event.OrderCreated,
		Payload: map[string]interface{}{
			"order_id":  order.Id,
			"order_no":  order.OrderNo,
			"email":     order.Email,
			"total":     total,
			"shipping":  shippingName,
		},
	})

	// 重新加载带明细的订单（含 items / payment），再转换为 VO
	detail, err := l.orderRepo.FindById(ctx, order.Id)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, ErrOrderNotFound
	}
	return l.toVO(ctx, detail, intent)
}

// MarkPaid 支付成功回调处理（幂等）
func (l *OrderLogic) MarkPaid(ctx context.Context, orderNo string, provider, eventId string, amountCents int64) error {
	if strings.TrimSpace(orderNo) == "" || strings.TrimSpace(provider) == "" || strings.TrimSpace(eventId) == "" {
		return ErrInvalidOrderState
	}

	order, err := l.orderRepo.FindByOrderNo(ctx, orderNo)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}

	// 幂等：同一回调事件只处理一次
	ok, err := l.orderRepo.TryMarkWebhook(ctx, provider, eventId, fmt.Sprintf(`{"order_no":%q}`, orderNo))
	if err != nil {
		return err
	}
	if !ok {
		return nil // 已处理，直接返回
	}

	if order.Status == "paid" || order.Status == "fulfilled" {
		return nil
	}
	if order.Status != "pending" {
		return ErrInvalidOrderState
	}

	// 金额严格校验，防止伪造回调（包含 0 金额的免费订单，也必须与订单总额一致）
	if amountCents != order.TotalCents {
		return ErrAmountMismatch
	}

	// 1. 提交库存预占（真正扣减）
	if err := l.inventoryRepo.Commit(ctx, "order", order.Id); err != nil {
		return err
	}

	// 2. 更新订单状态
	if err := l.orderRepo.UpdateStatus(ctx, order.Id, "paid", "system", "payment succeeded"); err != nil {
		return err
	}

	// 3. 更新支付记录
	if order.Payment != nil {
		_ = l.orderRepo.UpdatePaymentStatus(ctx, provider, order.Payment.ProviderIntentId, "succeeded")
	}

	l.bus.Publish(ctx, event.Event{
		Type: event.OrderPaid,
		Payload: map[string]interface{}{
			"order_id": order.Id,
			"order_no": order.OrderNo,
			"email":    order.Email,
			"total":    order.TotalCents,
		},
	})
	return nil
}

// Cancel 取消订单并回滚库存
func (l *OrderLogic) Cancel(ctx context.Context, orderNo, reason string) error {
	order, err := l.orderRepo.FindByOrderNo(ctx, orderNo)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}
	if order.Status != "pending" {
		return ErrInvalidOrderState
	}

	if err := l.inventoryRepo.Release(ctx, "order", order.Id); err != nil {
		return err
	}
	if err := l.orderRepo.UpdateStatus(ctx, order.Id, "cancelled", "system", reason); err != nil {
		return err
	}
	if order.CouponId.Valid {
		_ = l.couponRepo.DecrementUsage(ctx, order.CouponId.String)
	}

	l.bus.Publish(ctx, event.Event{
		Type:    event.OrderCanceled,
		Payload: map[string]interface{}{"order_id": order.Id, "order_no": order.OrderNo},
	})
	return nil
}

// DetailByNo 供用户查询订单：必须同时提供订单号与下单邮箱，防止订单号被枚举泄露隐私
func (l *OrderLogic) DetailByNo(ctx context.Context, orderNo, email string) (*types.OrderVO, error) {
	if strings.TrimSpace(orderNo) == "" || strings.TrimSpace(email) == "" {
		return nil, nil
	}
	order, err := l.orderRepo.FindByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, nil
	}
	if !strings.EqualFold(strings.TrimSpace(email), order.Email) {
		return nil, nil
	}
	return l.toVO(ctx, order, nil)
}

// ValidateCoupon 前台试算优惠券：返回是否可用与可抵扣金额（不占用名额）
func (l *OrderLogic) ValidateCoupon(ctx context.Context, code string, subtotal int64) (*types.ValidateCouponResp, error) {
	resp := &types.ValidateCouponResp{Code: strings.TrimSpace(code)}
	if resp.Code == "" {
		resp.Message = "coupon code is required"
		return resp, nil
	}
	discount, err := l.applyCoupon(ctx, resp.Code, subtotal)
	if err != nil {
		resp.Message = err.Error()
		return resp, nil
	}
	if discount == nil {
		resp.Message = "invalid coupon code"
		return resp, nil
	}
	resp.Valid = true
	resp.Code = discount.Code
	resp.DiscountCents = discount.DiscountCents
	resp.Message = "ok"
	return resp, nil
}

// ActiveCoupons 前台展示的可用券（只暴露券码与优惠信息，供用户自行填写）
func (l *OrderLogic) ActiveCoupons(ctx context.Context) ([]types.PublicCoupon, error) {
	list, err := l.couponRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]types.PublicCoupon, 0, len(list))
	for _, c := range list {
		item := types.PublicCoupon{
			Code:           c.Code,
			Type:           c.Type,
			Value:          c.Value,
			MinAmountCents: c.MinAmountCents,
		}
		if c.EndsAt.Valid {
			item.EndsAt = c.EndsAt.Time.Format("2006-01-02")
		}
		out = append(out, item)
	}
	return out, nil
}

func (l *OrderLogic) toVO(ctx context.Context, order *model.OrderDetail, intent *payment.IntentResult) (*types.OrderVO, error) {
	items := make([]types.OrderItemVO, 0, len(order.Items))
	for _, it := range order.Items {
		opts := map[string]interface{}{}
		_ = json.Unmarshal([]byte(it.Options), &opts)
		items = append(items, types.OrderItemVO{
			SkuCode:   it.SkuCode,
			Title:     it.Title,
			Options:   opts,
			ImageUrl:  it.ImageUrl,
			UnitPrice: it.UnitPriceCents,
			Qty:       it.Qty,
			Total:     it.TotalCents,
		})
	}

	vo := &types.OrderVO{
		OrderNo:        order.OrderNo,
		Email:          order.Email,
		Status:         order.Status,
		Currency:       order.Currency,
		SubtotalCents:  order.SubtotalCents,
		ShippingCents:  order.ShippingCents,
		DiscountCents:  order.DiscountCents,
		TotalCents:     order.TotalCents,
		CustomerNote:   order.CustomerNote,
		ShippingAddr:   decodeMap(order.ShippingAddress),
		Items:          items,
		CreatedAt:      order.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if order.Payment != nil {
		vo.PaymentStatus = order.Payment.Status
		vo.Provider = order.Payment.Provider
	}
	if order.Shipment != nil {
		vo.TrackingNo = order.Shipment.TrackingNo
		vo.Carrier = order.Shipment.Carrier
		vo.TrackingUrl = order.Shipment.TrackingUrl
	}
	if intent != nil {
		vo.ClientSecret = intent.ClientSecret
		vo.CheckoutUrl = intent.CheckoutUrl
	}
	return vo, nil
}

func decodeMap(raw string) map[string]interface{} {
	out := map[string]interface{}{}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

// couponDiscount 优惠券计算结果
type couponDiscount struct {
	Id            string
	Code          string
	DiscountCents int64
}

// applyCoupon 服务端校验并计算优惠金额。
// code 为空返回 (nil, nil)；无效券返回对应错误。折扣金额绝不信任客户端。
func (l *OrderLogic) applyCoupon(ctx context.Context, code string, subtotal int64) (*couponDiscount, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, nil
	}

	coupon, err := l.couponRepo.FindActiveByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if coupon == nil || coupon.Status != "active" {
		return nil, ErrInvalidCoupon
	}

	now := time.Now().UTC()
	if coupon.StartsAt.Valid && now.Before(coupon.StartsAt.Time) {
		return nil, ErrCouponNotStarted
	}
	if coupon.EndsAt.Valid && now.After(coupon.EndsAt.Time) {
		return nil, ErrCouponExpired
	}
	if coupon.MinAmountCents > 0 && subtotal < coupon.MinAmountCents {
		return nil, ErrCouponMinAmount
	}
	if coupon.MaxUses > 0 && coupon.UsedCount >= coupon.MaxUses {
		return nil, ErrCouponExhausted
	}

	var discount int64
	switch coupon.Type {
	case "percent":
		discount = subtotal * coupon.Value / 100
	case "fixed":
		discount = coupon.Value
	default:
		return nil, ErrInvalidCoupon
	}
	if discount < 0 {
		discount = 0
	}
	if discount > subtotal {
		discount = subtotal
	}

	return &couponDiscount{Id: coupon.Id, Code: coupon.Code, DiscountCents: discount}, nil
}
