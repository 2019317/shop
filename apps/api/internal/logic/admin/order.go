package admin

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/yourname/stationery-shop/apps/api/internal/pkg/event"
	"github.com/yourname/stationery-shop/apps/api/internal/repo"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

var (
	ErrOrderNotPaid   = errors.New("order is not paid")
	ErrOrderNotPending = errors.New("only pending orders can be cancelled")
)

type OrderLogic struct {
	orderRepo     *repo.OrderRepo
	inventoryRepo *repo.InventoryRepo
	couponRepo    *repo.CouponRepo
	bus           *event.Bus
}

func NewOrderLogic(orderRepo *repo.OrderRepo, inventoryRepo *repo.InventoryRepo, couponRepo *repo.CouponRepo, bus *event.Bus) *OrderLogic {
	return &OrderLogic{orderRepo: orderRepo, inventoryRepo: inventoryRepo, couponRepo: couponRepo, bus: bus}
}

func (l *OrderLogic) List(ctx context.Context, status, keyword string, page, pageSize int) (*types.PageData, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	list, total, err := l.orderRepo.AdminList(ctx, repo.AdminOrderFilter{
		Status:   status,
		Keyword:  keyword,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]types.AdminOrderItem, 0, len(list))
	for _, o := range list {
		items = append(items, types.AdminOrderItem{
			Id:         o.Id,
			OrderNo:    o.OrderNo,
			Email:      o.Email,
			Status:     o.Status,
			TotalCents: o.TotalCents,
			Currency:   o.Currency,
			CreatedAt:  o.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	return &types.PageData{List: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (l *OrderLogic) Detail(ctx context.Context, id string) (*types.AdminOrderDetail, error) {
	detail, err := l.orderRepo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, nil
	}

	items := make([]types.OrderItemVO, 0, len(detail.Items))
	for _, it := range detail.Items {
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

	out := &types.AdminOrderDetail{
		Id:            detail.Id,
		OrderNo:       detail.OrderNo,
		Email:         detail.Email,
		Status:        detail.Status,
		Currency:      detail.Currency,
		SubtotalCents: detail.SubtotalCents,
		ShippingCents: detail.ShippingCents,
		DiscountCents: detail.DiscountCents,
		TotalCents:    detail.TotalCents,
		CustomerNote:  detail.CustomerNote,
		ShippingAddr:  decodeMap(detail.ShippingAddress),
		Items:         items,
		CreatedAt:     detail.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if detail.Payment != nil {
		out.PaymentStatus = detail.Payment.Status
		out.Provider = detail.Payment.Provider
	}
	if detail.Shipment != nil {
		out.Carrier = detail.Shipment.Carrier
		out.TrackingNo = detail.Shipment.TrackingNo
		out.TrackingUrl = detail.Shipment.TrackingUrl
	}
	return out, nil
}

// Ship 发货：填写物流单号并流转订单状态
func (l *OrderLogic) Ship(ctx context.Context, id string, req types.ShipOrderReq) error {
	detail, err := l.orderRepo.FindById(ctx, id)
	if err != nil {
		return err
	}
	if detail == nil {
		return repo.ErrOrderMissing
	}
	if detail.Status != "paid" {
		return ErrOrderNotPaid
	}

	if err := l.orderRepo.CreateShipment(ctx, id, req.Carrier, req.TrackingNo, req.TrackingUrl); err != nil {
		return err
	}
	if err := l.orderRepo.UpdateStatus(ctx, id, "fulfilled", "admin", "shipped"); err != nil {
		return err
	}

	l.bus.Publish(ctx, event.Event{
		Type: event.OrderShipped,
		Payload: map[string]interface{}{
			"order_id":    id,
			"order_no":    detail.OrderNo,
			"email":       detail.Email,
			"carrier":     req.Carrier,
			"tracking_no": req.TrackingNo,
		},
	})
	return nil
}

// Cancel 取消订单并回滚库存
func (l *OrderLogic) Cancel(ctx context.Context, id, reason string) error {
	detail, err := l.orderRepo.FindById(ctx, id)
	if err != nil {
		return err
	}
	if detail == nil {
		return repo.ErrOrderMissing
	}
	if detail.Status != "pending" {
		return ErrOrderNotPending
	}

	if err := l.inventoryRepo.Release(ctx, "order", id); err != nil {
		return err
	}
	if err := l.orderRepo.UpdateStatus(ctx, id, "cancelled", "admin", reason); err != nil {
		return err
	}
	if detail.CouponId != nil {
		_ = l.couponRepo.DecrementUsage(ctx, *detail.CouponId)
	}

	l.bus.Publish(ctx, event.Event{
		Type:    event.OrderCanceled,
		Payload: map[string]interface{}{"order_id": id, "order_no": detail.OrderNo},
	})
	return nil
}

func decodeMap(raw string) map[string]interface{} {
	out := map[string]interface{}{}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}
