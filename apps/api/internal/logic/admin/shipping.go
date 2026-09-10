package admin

import (
	"context"
	"errors"
	"strings"

	"github.com/yourname/stationery-shop/apps/api/internal/model"
	"github.com/yourname/stationery-shop/apps/api/internal/repo"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

var (
	// ErrInvalidShippingRule 运费规则参数非法
	ErrInvalidShippingRule = errors.New("invalid shipping rule")
	// ErrNoShipment 订单尚未发货，无法更新物流状态
	ErrNoShipment = errors.New("order has no shipment")
)

// ShippingLogic 后台运费规则/物流管理
type ShippingLogic struct {
	shippingRepo *repo.ShippingRepo
	orderRepo    *repo.OrderRepo
}

func NewShippingLogic(shippingRepo *repo.ShippingRepo, orderRepo *repo.OrderRepo) *ShippingLogic {
	return &ShippingLogic{shippingRepo: shippingRepo, orderRepo: orderRepo}
}

func (l *ShippingLogic) ListRules(ctx context.Context, status string, page, pageSize int) (*types.PageData, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	list, total, err := l.shippingRepo.AdminList(ctx, repo.ShippingRuleFilter{
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.AdminShippingRule, 0, len(list))
	for _, r := range list {
		items = append(items, toShippingRule(r))
	}
	return &types.PageData{List: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (l *ShippingLogic) CreateRule(ctx context.Context, req types.AdminShippingRuleReq) (string, error) {
	in, err := normalizeShippingRule(req)
	if err != nil {
		return "", err
	}
	return l.shippingRepo.CreateRule(ctx, in)
}

func (l *ShippingLogic) UpdateRule(ctx context.Context, id string, req types.AdminShippingRuleReq) error {
	in, err := normalizeShippingRule(req)
	if err != nil {
		return err
	}
	return l.shippingRepo.UpdateRule(ctx, id, in)
}

func (l *ShippingLogic) DeleteRule(ctx context.Context, id string) error {
	return l.shippingRepo.DeleteRule(ctx, id)
}

// MarkDelivered 将运单标记为已送达（订单保持 fulfilled，但物流状态更新）
func (l *ShippingLogic) MarkDelivered(ctx context.Context, orderId string) error {
	detail, err := l.orderRepo.FindById(ctx, orderId)
	if err != nil {
		return err
	}
	if detail == nil {
		return repo.ErrOrderMissing
	}
	if detail.Shipment == nil {
		return ErrNoShipment
	}
	return l.shippingRepo.MarkDelivered(ctx, orderId)
}

func normalizeShippingRule(req types.AdminShippingRuleReq) (repo.ShippingRuleInput, error) {
	in := repo.ShippingRuleInput{
		Name:               strings.TrimSpace(req.Name),
		CountryCodes:       req.CountryCodes,
		MinAmountCents:     req.MinAmountCents,
		MaxWeightG:         req.MaxWeightG,
		PriceCents:         req.PriceCents,
		FreeThresholdCents: req.FreeThresholdCents,
		SortOrder:          req.SortOrder,
		Status:             req.Status,
	}
	if in.Name == "" {
		return in, ErrInvalidShippingRule
	}
	if in.MinAmountCents < 0 || in.MaxWeightG < 0 || in.PriceCents < 0 || in.FreeThresholdCents < 0 {
		return in, ErrInvalidShippingRule
	}
	if in.Status == "" {
		in.Status = "active"
	}
	if in.Status != "active" && in.Status != "inactive" {
		return in, ErrInvalidShippingRule
	}
	return in, nil
}

func toShippingRule(r model.ShippingRule) types.AdminShippingRule {
	return types.AdminShippingRule{
		Id:                 r.Id,
		Name:               r.Name,
		CountryCodes:       parseCodes(r.CountryCodes),
		MinAmountCents:     r.MinAmountCents,
		MaxWeightG:         r.MaxWeightG,
		PriceCents:         r.PriceCents,
		FreeThresholdCents: r.FreeThreshold,
		SortOrder:          r.SortOrder,
		Status:             r.Status,
	}
}

func parseCodes(raw string) []string {
	raw = strings.Trim(raw, "{}")
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.Trim(strings.TrimSpace(p), `"`)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
