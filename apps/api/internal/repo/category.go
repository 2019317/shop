package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/yourname/stationery-shop/apps/api/internal/model"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/i18n"
)

type CategoryRepo struct {
	conn sqlx.SqlConn
}

func NewCategoryRepo(conn sqlx.SqlConn) *CategoryRepo {
	return &CategoryRepo{conn: conn}
}

const categoryColumns = `id, parent_id, name, slug, description, image_key, sort_order, status, created_at, updated_at`

// categoryColumnsLocalized 按 locale 取翻译覆盖（中文时名称/描述取自翻译表）
const categoryColumnsLocalized = `c.id, c.parent_id,
	COALESCE(NULLIF(ct.name, ''), c.name) AS name,
	c.slug,
	COALESCE(NULLIF(ct.description, ''), c.description) AS description,
	c.image_key, c.sort_order, c.status, c.created_at, c.updated_at`

func (r *CategoryRepo) ListActive(ctx context.Context, locale string) ([]model.Category, error) {
	locale = i18n.Normalize(locale)
	join := ""
	if i18n.Supported(locale) {
		join = fmt.Sprintf(`LEFT JOIN catalog.category_translations ct
			ON ct.category_id = c.id AND ct.locale = '%s'`, locale)
	}
	query := fmt.Sprintf(
		`SELECT %s FROM catalog.categories c %s
		 WHERE c.status='active' ORDER BY c.sort_order, c.name`,
		categoryColumnsLocalized, join)
	var list []model.Category
	if err := r.conn.QueryRowsCtx(ctx, &list, query); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *CategoryRepo) AdminList(ctx context.Context) ([]model.Category, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM catalog.categories ORDER BY sort_order, name`, categoryColumns)
	var list []model.Category
	if err := r.conn.QueryRowsCtx(ctx, &list, query); err != nil {
		return nil, err
	}
	return list, nil
}

type CategoryInput struct {
	ParentId    *string
	Name        string
	Slug        string
	Description string
	ImageKey    string
	SortOrder   int
	Status      string
	Translations map[string]CategoryTranslationInput
}

// CategoryTranslationInput 类目翻译内容
type CategoryTranslationInput struct {
	Name        string
	Description string
}

func (r *CategoryRepo) Create(ctx context.Context, in CategoryInput) (string, error) {
	query := `INSERT INTO catalog.categories
	 (parent_id, name, slug, description, image_key, sort_order, status)
	 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`
	var id string
	if err := r.conn.QueryRowCtx(ctx, &id, query, in.ParentId, in.Name, in.Slug,
		in.Description, in.ImageKey, in.SortOrder, in.Status); err != nil {
		return "", err
	}
	if err := r.saveTranslations(ctx, id, in.Translations); err != nil {
		return id, err
	}
	return id, nil
}

func (r *CategoryRepo) Update(ctx context.Context, id string, in CategoryInput) error {
	query := `UPDATE catalog.categories SET
	 parent_id=$1, name=$2, slug=$3, description=$4, image_key=$5, sort_order=$6, status=$7
	 WHERE id=$8`
	if _, err := r.conn.ExecCtx(ctx, query, in.ParentId, in.Name, in.Slug,
		in.Description, in.ImageKey, in.SortOrder, in.Status, id); err != nil {
		return err
	}
	return r.saveTranslations(ctx, id, in.Translations)
}

// saveTranslations 写入类目翻译（空内容则删除该语言记录）
func (r *CategoryRepo) saveTranslations(ctx context.Context, categoryId string, translations map[string]TranslationInput) error {
	for locale, tr := range translations {
		locale = i18n.Normalize(locale)
		if locale == i18n.DefaultLocale {
			continue
		}
		if tr.Name == "" && tr.Description == "" {
			if _, err := r.conn.ExecCtx(ctx,
				`DELETE FROM catalog.category_translations WHERE category_id=$1 AND locale=$2`,
				categoryId, locale); err != nil {
				return err
			}
			continue
		}
		if _, err := r.conn.ExecCtx(ctx,
			`INSERT INTO catalog.category_translations (category_id, locale, name, description)
			 VALUES ($1,$2,$3,$4)
			 ON CONFLICT (category_id, locale) DO UPDATE SET
				name=EXCLUDED.name, description=EXCLUDED.description, updated_at=now()`,
			categoryId, locale, tr.Name, tr.Description); err != nil {
			return err
		}
	}
	return nil
}

func (r *CategoryRepo) Delete(ctx context.Context, id string) error {
	_, err := r.conn.ExecCtx(ctx, `UPDATE catalog.categories SET status='hidden' WHERE id=$1`, id)
	return err
}

func (r *CategoryRepo) SlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM catalog.categories WHERE slug=$1)`
	if err := r.conn.QueryRowCtx(ctx, &exists, query, slug); err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}
