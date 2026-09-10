package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/yourname/stationery-shop/apps/api/internal/model"
	"github.com/yourname/stationery-shop/apps/api/internal/repo"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

var (
	// ErrCouponCodeExists 券码已存在
	ErrCouponCodeExists = errors.New("coupon code already exists")
	// ErrInvalidCouponInput 优惠券参数非法
	ErrInvalidCouponInput = errors.New("invalid coupon input")
)

// CouponLogic 后台优惠券管理
type CouponLogic struct {
	couponRepo *repo.CouponRepo
}

func NewCouponLogic(couponRepo *repo.CouponRepo) *CouponLogic {
	return &CouponLogic{couponRepo: couponRepo}
}

func (l *CouponLogic) List(ctx context.Context, keyword, status string, page, pageSize int) (*types.PageData, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	list, total, err := l.couponRepo.AdminList(ctx, repo.CouponFilter{
		Keyword:  keyword,
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]types.AdminCouponItem, 0, len(list))
	for _, c := range list {
		items = append(items, toCouponItem(c))
	}
	return &types.PageData{List: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (l *CouponLogic) Detail(ctx context.Context, id string) (*types.AdminCouponItem, error) {
	c, err := l.couponRepo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, nil
	}
	item := toCouponItem(*c)
	return &item, nil
}

func (l *CouponLogic) Create(ctx context.Context, req types.AdminCouponReq) (string, error) {
	in, err := normalizeCouponInput(req)
	if err != nil {
		return "", err
	}
	exists, err := l.couponRepo.CodeExists(ctx, in.Code, "")
	if err != nil {
		return "", err
	}
	if exists {
		return "", ErrCouponCodeExists
	}
	return l.couponRepo.Create(ctx, in)
}

func (l *CouponLogic) Update(ctx context.Context, id string, req types.AdminCouponReq) error {
	in, err := normalizeCouponInput(req)
	if err != nil {
		return err
	}
	return l.couponRepo.Update(ctx, id, in)
}

// SetStatus 启用/停用
func (l *CouponLogic) SetStatus(ctx context.Context, id, status string) error {
	if status != "active" && status != "disabled" {
		return ErrInvalidCouponInput
	}
	return l.couponRepo.SetStatus(ctx, id, status)
}

func normalizeCouponInput(req types.AdminCouponReq) (repo.CouponInput, error) {
	in := repo.CouponInput{
		Code:           strings.TrimSpace(req.Code),
		Type:           req.Type,
		Value:          req.Value,
		MinAmountCents: req.MinAmountCents,
		MaxUses:        req.MaxUses,
		Status:         req.Status,
	}
	if in.Code == "" {
		return in, ErrInvalidCouponInput
	}
	if in.Type != "percent" && in.Type != "fixed" {
		return in, ErrInvalidCouponInput
	}
	if in.Value <= 0 {
		return in, ErrInvalidCouponInput
	}
	if in.Type == "percent" && in.Value > 100 {
		return in, ErrInvalidCouponInput
	}
	if in.Status == "" {
		in.Status = "active"
	}
	if in.Status != "active" && in.Status != "disabled" {
		return in, ErrInvalidCouponInput
	}
	if in.MaxUses < 0 {
		return in, ErrInvalidCouponInput
	}

	starts, err := normalizeTime(req.StartsAt, false)
	if err != nil {
		return in, ErrInvalidCouponInput
	}
	ends, err := normalizeTime(req.EndsAt, true)
	if err != nil {
		return in, ErrInvalidCouponInput
	}
	if starts != nil && ends != nil && *ends < *starts {
		return in, ErrInvalidCouponInput
	}
	in.StartsAt = starts
	in.EndsAt = ends
	return in, nil
}

// normalizeTime 将前端传来的时间字符串规整为 RFC3339；
// endOfDay 为 true 时把纯日期补成当日 23:59:59，避免「当天就失效」。
func normalizeTime(raw string, endOfDay bool) (*string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		s := t.UTC().Format(time.RFC3339)
		return &s, nil
	}
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		if endOfDay {
			t = t.Add(24*time.Hour - time.Second)
		}
		s := t.UTC().Format(time.RFC3339)
		return &s, nil
	}
	return nil, errors.New("invalid time format")
}

func toCouponItem(c model.Coupon) types.AdminCouponItem {
	item := types.AdminCouponItem{
		Id:             c.Id,
		Code:           c.Code,
		Type:           c.Type,
		Value:          c.Value,
		MinAmountCents: c.MinAmountCents,
		MaxUses:        c.MaxUses,
		UsedCount:      c.UsedCount,
		Status:         c.Status,
		CreatedAt:      c.CreatedAt.Format("2006-01-02 15:04"),
	}
	if c.StartsAt.Valid {
		item.StartsAt = c.StartsAt.Time.Format("2006-01-02 15:04")
	}
	if c.EndsAt.Valid {
		item.EndsAt = c.EndsAt.Time.Format("2006-01-02 15:04")
	}
	return item
}
