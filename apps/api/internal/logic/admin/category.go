package admin

import (
	"context"

	"github.com/yourname/stationery-shop/apps/api/internal/repo"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

type CategoryLogic struct {
	categoryRepo *repo.CategoryRepo
	storage      PublicURLer
}

type PublicURLer interface {
	PublicURL(objectKey string) string
}

func NewCategoryLogic(categoryRepo *repo.CategoryRepo, storage PublicURLer) *CategoryLogic {
	return &CategoryLogic{categoryRepo: categoryRepo, storage: storage}
}

func (l *CategoryLogic) List(ctx context.Context) ([]types.CategoryVO, error) {
	list, err := l.categoryRepo.AdminList(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]types.CategoryVO, 0, len(list))
	for _, c := range list {
		out = append(out, types.CategoryVO{
			Id:        c.Id,
			Name:      c.Name,
			Slug:      c.Slug,
			Desc:      c.Description,
			ImageUrl:  l.storage.PublicURL(c.ImageKey),
			SortOrder: c.SortOrder,
		})
	}
	return out, nil
}

func (l *CategoryLogic) Create(ctx context.Context, req types.AdminCategoryReq) (string, error) {
	in := repo.CategoryInput{
		Name:         req.Name,
		Slug:         req.Slug,
		Description:  req.Description,
		ImageKey:     req.ImageKey,
		SortOrder:    req.SortOrder,
		Status:       req.Status,
		Translations: toCategoryTranslations(req.Translations),
	}
	if req.ParentId != "" {
		in.ParentId = &req.ParentId
	}
	if in.Status == "" {
		in.Status = "active"
	}
	return l.categoryRepo.Create(ctx, in)
}

func (l *CategoryLogic) Update(ctx context.Context, id string, req types.AdminCategoryReq) error {
	in := repo.CategoryInput{
		Name:         req.Name,
		Slug:         req.Slug,
		Description:  req.Description,
		ImageKey:     req.ImageKey,
		SortOrder:    req.SortOrder,
		Status:       req.Status,
		Translations: toCategoryTranslations(req.Translations),
	}
	if req.ParentId != "" {
		in.ParentId = &req.ParentId
	}
	return l.categoryRepo.Update(ctx, id, in)
}

// toCategoryTranslations 将 API 层翻译结构转换为 repo 层结构
func toCategoryTranslations(src map[string]types.CategoryTranslationInput) map[string]repo.CategoryTranslationInput {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]repo.CategoryTranslationInput, len(src))
	for locale, tr := range src {
		out[locale] = repo.CategoryTranslationInput{
			Name:        tr.Name,
			Description: tr.Description,
		}
	}
	return out
}

func (l *CategoryLogic) Delete(ctx context.Context, id string) error {
	return l.categoryRepo.Delete(ctx, id)
}
