package admin

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/yourname/stationery-shop/apps/api/internal/repo"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

type ProductLogic struct {
	productRepo *repo.ProductRepo
}

func NewProductLogic(productRepo *repo.ProductRepo) *ProductLogic {
	return &ProductLogic{productRepo: productRepo}
}

func (l *ProductLogic) List(ctx context.Context, keyword string, page, pageSize int) (*types.PageData, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	list, total, err := l.productRepo.AdminList(ctx, keyword, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]types.AdminProductItem, 0, len(list))
	for _, p := range list {
		items = append(items, types.AdminProductItem{
			Id:         p.Id,
			Title:      p.Title,
			Slug:       p.Slug,
			Status:     p.Status,
			PriceCents: p.PriceCents,
			Currency:   p.Currency,
			UpdatedAt:  p.UpdatedAt.Format(time.RFC3339),
		})
	}
	return &types.PageData{List: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (l *ProductLogic) Detail(ctx context.Context, id string) (*types.AdminProductReq, error) {
	detail, err := l.productRepo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, nil
	}

	attrs := map[string]interface{}{}
	_ = json.Unmarshal([]byte(detail.Attributes), &attrs)
	tags := parseStringArray(detail.Tags)

	req := &types.AdminProductReq{
		Title:       detail.Title,
		Slug:        detail.Slug,
		Subtitle:    detail.Subtitle,
		Description: detail.Description,
		Status:      detail.Status,
		Currency:    detail.Currency,
		Attributes:  attrs,
		Tags:        tags,
		SeoTitle:    detail.SeoTitle,
		SeoDesc:     detail.SeoDescription,
	}
	if detail.CategoryId.Valid {
		req.CategoryId = detail.CategoryId.String
	}

	for _, v := range detail.Variants {
		opts := map[string]interface{}{}
		_ = json.Unmarshal([]byte(v.Options), &opts)
		req.Variants = append(req.Variants, types.AdminVariantReq{
			SkuCode:        v.SkuCode,
			Title:          v.Title,
			Options:        opts,
			PriceCents:     v.PriceCents,
			CompareAtCents: v.CompareAtCents,
			WeightG:        v.WeightG,
			ImageKey:       v.ImageKey,
			SortOrder:      v.SortOrder,
		})
	}
	for _, img := range detail.Images {
		req.Images = append(req.Images, types.AdminImageReq{
			ObjectKey: img.ObjectKey,
			Alt:       img.Alt,
			SortOrder: img.SortOrder,
		})
	}
	req.Translations = make(map[string]types.TranslationInput)
	for _, tr := range detail.Translations {
		req.Translations[tr.Locale] = types.TranslationInput{
			Title:       tr.Title,
			Subtitle:    tr.Subtitle,
			Description: tr.Description,
			SeoTitle:    tr.SeoTitle,
			SeoDescription: tr.SeoDescription,
		}
	}
	return req, nil
}

func (l *ProductLogic) Create(ctx context.Context, req types.AdminProductReq) (string, error) {
	if req.Currency == "" {
		req.Currency = "USD"
	}
	if req.Status == "" {
		req.Status = "draft"
	}

	in := repo.CreateProductInput{
		Title:       req.Title,
		Slug:        req.Slug,
		Subtitle:    req.Subtitle,
		Description: req.Description,
		Status:      req.Status,
		PriceCents:  computeMinPrice(req.Variants),
		Currency:    req.Currency,
		Attributes:  req.Attributes,
		Tags:        req.Tags,
		SeoTitle:    req.SeoTitle,
		SeoDesc:     req.SeoDesc,
	}
	if req.CategoryId != "" {
		in.CategoryId = &req.CategoryId
	}
	for _, v := range req.Variants {
		in.Variants = append(in.Variants, repo.VariantInput{
			SkuCode:        v.SkuCode,
			Title:          v.Title,
			Options:        v.Options,
			PriceCents:     v.PriceCents,
			CompareAtCents: v.CompareAtCents,
			WeightG:        v.WeightG,
			ImageKey:       v.ImageKey,
			Stock:          v.Stock,
			SortOrder:      v.SortOrder,
		})
	}
	for _, img := range req.Images {
		in.Images = append(in.Images, repo.ImageInput{
			ObjectKey: img.ObjectKey, Alt: img.Alt, SortOrder: img.SortOrder,
		})
	}
	in.Translations = toTranslations(req.Translations)

	return l.productRepo.Create(ctx, in)
}

func (l *ProductLogic) Update(ctx context.Context, id string, req types.AdminProductReq) error {
	in := repo.CreateProductInput{
		Title:       req.Title,
		Slug:        req.Slug,
		Subtitle:    req.Subtitle,
		Description: req.Description,
		Status:      req.Status,
		PriceCents:  computeMinPrice(req.Variants),
		Currency:    req.Currency,
		Attributes:  req.Attributes,
		Tags:        req.Tags,
		SeoTitle:    req.SeoTitle,
		SeoDesc:     req.SeoDesc,
	}
	if req.CategoryId != "" {
		in.CategoryId = &req.CategoryId
	}
	for _, v := range req.Variants {
		in.Variants = append(in.Variants, repo.VariantInput{
			SkuCode:        v.SkuCode,
			Title:          v.Title,
			Options:        v.Options,
			PriceCents:     v.PriceCents,
			CompareAtCents: v.CompareAtCents,
			WeightG:        v.WeightG,
			ImageKey:       v.ImageKey,
			Stock:          v.Stock,
			SortOrder:      v.SortOrder,
		})
	}
	for _, img := range req.Images {
		in.Images = append(in.Images, repo.ImageInput{
			ObjectKey: img.ObjectKey, Alt: img.Alt, SortOrder: img.SortOrder,
		})
	}
	in.Translations = toTranslations(req.Translations)
	return l.productRepo.Update(ctx, id, in)
}

func (l *ProductLogic) Delete(ctx context.Context, id string) error {
	return l.productRepo.Delete(ctx, id)
}

// computeMinPrice 取变体最低价作为列表展示价
func computeMinPrice(variants []types.AdminVariantReq) int64 {
	if len(variants) == 0 {
		return 0
	}
	min := variants[0].PriceCents
	for _, v := range variants[1:] {
		if v.PriceCents < min {
			min = v.PriceCents
		}
	}
	return min
}

func parseStringArray(raw string) []string {
	raw = strings.Trim(raw, "{}")
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.Trim(p, `"`))
	}
	return out
}

// toTranslations 将 API 层翻译结构转换为 repo 层结构
func toTranslations(src map[string]types.TranslationInput) map[string]repo.TranslationInput {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]repo.TranslationInput, len(src))
	for locale, tr := range src {
		out[locale] = repo.TranslationInput{
			Title:          tr.Title,
			Subtitle:       tr.Subtitle,
			Description:    tr.Description,
			SeoTitle:       tr.SeoTitle,
			SeoDescription: tr.SeoDescription,
		}
	}
	return out
}
