package shop

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/yourname/stationery-shop/apps/api/internal/repo"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

type ProductLogic struct {
	productRepo  *repo.ProductRepo
	categoryRepo *repo.CategoryRepo
	publicURL    func(objectKey string) string
}

func NewProductLogic(
	productRepo *repo.ProductRepo,
	categoryRepo *repo.CategoryRepo,
	publicURL func(objectKey string) string,
) *ProductLogic {
	return &ProductLogic{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		publicURL:    publicURL,
	}
}

func (l *ProductLogic) List(ctx context.Context, categorySlug, keyword, sort, locale string, page, pageSize int) (*types.PageData, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 60 {
		pageSize = 24
	}

	f := repo.ProductListFilter{
		CategorySlug: categorySlug,
		Keyword:      keyword,
		Status:       "published",
		Sort:         sort,
		Locale:       locale,
		Page:         page,
		PageSize:     pageSize,
	}

	list, total, err := l.productRepo.List(ctx, f)
	if err != nil {
		return nil, err
	}

	// 批量取封面图，避免 N+1
	covers := map[string]string{}
	ids := make([]string, 0, len(list))
	for _, p := range list {
		ids = append(ids, p.Id)
	}
	if len(ids) > 0 {
		if m, err := l.productRepo.BatchFirstImages(ctx, ids); err == nil {
			covers = m
		}
	}

	items := make([]types.ProductListItem, 0, len(list))
	for _, p := range list {
		cover := covers[p.Id]
		if cover == "" {
			cover = ""
		}
		items = append(items, types.ProductListItem{
			Id:         p.Id,
			Title:      p.Title,
			Slug:       p.Slug,
			Subtitle:   p.Subtitle,
			PriceCents: p.PriceCents,
			Currency:   p.Currency,
			CoverUrl:   cover,
		})
	}
	return &types.PageData{List: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (l *ProductLogic) Detail(ctx context.Context, slug, locale string) (*types.ProductDetailVO, error) {
	detail, err := l.productRepo.FindBySlug(ctx, slug, locale)
	if err != nil {
		return nil, err
	}
	if detail == nil || detail.Status != "published" {
		return nil, nil
	}

	attrs := map[string]interface{}{}
	_ = json.Unmarshal([]byte(detail.Attributes), &attrs)

	variants := make([]types.VariantVO, 0, len(detail.Variants))
	for _, v := range detail.Variants {
		if v.Status != "active" {
			continue
		}
		opts := map[string]interface{}{}
		_ = json.Unmarshal([]byte(v.Options), &opts)
		imageUrl := ""
		if v.ImageKey != "" {
			imageUrl = l.publicURL(v.ImageKey)
		}
		variants = append(variants, types.VariantVO{
			Id:             v.Id,
			SkuCode:        v.SkuCode,
			Title:          v.Title,
			Options:        opts,
			PriceCents:     v.PriceCents,
			CompareAtCents: v.CompareAtCents,
			WeightG:        v.WeightG,
			ImageUrl:       imageUrl,
			InStock:        detail.AvailableStock > 0,
		})
	}

	images := make([]types.ImageVO, 0, len(detail.Images))
	for _, img := range detail.Images {
		images = append(images, types.ImageVO{
			Url:  l.publicURL(img.ObjectKey),
			Alt:  img.Alt,
			Sort: img.SortOrder,
		})
	}

	return &types.ProductDetailVO{
		Id:             detail.Id,
		Title:          detail.Title,
		Slug:           detail.Slug,
		Subtitle:       detail.Subtitle,
		Description:    detail.Description,
		PriceCents:     detail.PriceCents,
		Currency:       detail.Currency,
		CategorySlug:   detail.CategorySlug,
		CategoryName:   detail.CategoryName,
		Attributes:     attrs,
		Tags:           parseTags(detail.Tags),
		SeoTitle:       detail.SeoTitle,
		SeoDescription: detail.SeoDescription,
		Images:         images,
		Variants:       variants,
	}, nil
}

func (l *ProductLogic) Categories(ctx context.Context, locale string) ([]types.CategoryVO, error) {
	list, err := l.categoryRepo.ListActive(ctx, locale)
	if err != nil {
		return nil, err
	}
	out := make([]types.CategoryVO, 0, len(list))
	for _, c := range list {
		imageUrl := ""
		if c.ImageKey != "" {
			imageUrl = l.publicURL(c.ImageKey)
		}
		out = append(out, types.CategoryVO{
			Id:        c.Id,
			Name:      c.Name,
			Slug:      c.Slug,
			Desc:      c.Description,
			ImageUrl:  imageUrl,
			SortOrder: c.SortOrder,
		})
	}
	return out, nil
}

func parseTags(raw string) []string {
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
