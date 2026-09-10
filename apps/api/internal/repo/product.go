package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/yourname/stationery-shop/apps/api/internal/model"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/i18n"
)

// ErrSkuConflict 变体 SKU 已被其他商品占用
var ErrSkuConflict = errors.New("sku code already used by another product")

// ErrSlugConflict slug 已被占用
var ErrSlugConflict = errors.New("slug already exists")

// isUniqueViolation 判断是否为 PostgreSQL 唯一约束冲突（SQLSTATE 23505）
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}

type ProductRepo struct {
	conn sqlx.SqlConn
}

func NewProductRepo(conn sqlx.SqlConn) *ProductRepo {
	return &ProductRepo{conn: conn}
}

const productColumns = `id, title, slug, subtitle, description, category_id, status,
	price_cents, currency, attributes::text as attributes, tags::text as tags,
	seo_title, seo_description, published_at, created_at, updated_at`

// productColumnsLocalized 带翻译覆盖的列：
// 翻译存在且非空则使用翻译，否则回落主表（默认语言）内容
const productColumnsLocalized = `p.id,
	COALESCE(NULLIF(t.title,''), p.title) as title,
	p.slug,
	COALESCE(NULLIF(t.subtitle,''), p.subtitle) as subtitle,
	COALESCE(NULLIF(t.description,''), p.description) as description,
	p.category_id,
	p.status,
	p.price_cents,
	p.currency,
	p.attributes::text as attributes,
	p.tags::text as tags,
	COALESCE(NULLIF(t.seo_title,''), p.seo_title) as seo_title,
	COALESCE(NULLIF(t.seo_description,''), p.seo_description) as seo_description,
	p.published_at,
	p.created_at,
	p.updated_at`

// translationJoin 按 locale 关联翻译表
const translationJoin = `LEFT JOIN catalog.product_translations t
	ON t.product_id = p.id AND t.locale = %s`

// ---------- 前台 ----------

type ProductListFilter struct {
	CategorySlug string
	Keyword      string
	Status       string
	Sort         string // newest | price_asc | price_desc
	Locale       string // en | zh
	Page         int
	PageSize     int
}

func (r *ProductRepo) List(ctx context.Context, f ProductListFilter) ([]model.Product, int64, error) {
	if f.Locale == "" {
		f.Locale = i18n.DefaultLocale
	}

	// $1 = locale，$2 = status，后续参数依次递增
	where := []string{"p.status = $2"}
	args := []interface{}{f.Locale, f.Status}
	idx := 3

	if f.CategorySlug != "" {
		where = append(where, fmt.Sprintf("c.slug = $%d", idx))
		args = append(args, f.CategorySlug)
		idx++
	}
	if f.Keyword != "" {
		if f.Locale == i18n.ZhLocale {
			// 英文全文检索对中文无效，中文改用 ILIKE 模糊匹配（同时匹配主表与翻译）
			where = append(where, fmt.Sprintf(
				"(p.title ILIKE $%d OR COALESCE(t.title,'') ILIKE $%d)", idx, idx))
			args = append(args, "%"+f.Keyword+"%")
		} else {
			where = append(where, fmt.Sprintf(
				"p.search_tsv @@ websearch_to_tsquery('english', $%d)", idx))
			args = append(args, f.Keyword)
		}
		idx++
	}

	whereSQL := strings.Join(where, " AND ")
	orderSQL := "p.published_at DESC NULLS LAST, p.id DESC"
	switch f.Sort {
	case "price_asc":
		orderSQL = "p.price_cents ASC, p.id DESC"
	case "price_desc":
		orderSQL = "p.price_cents DESC, p.id DESC"
	}

	joinT := fmt.Sprintf(translationJoin, "$1")

	var total int64
	countQuery := fmt.Sprintf(
		`SELECT COUNT(*) FROM catalog.products p
		 %s
		 LEFT JOIN catalog.categories c ON c.id = p.category_id
		 WHERE %s`, joinT, whereSQL)
	if err := r.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.PageSize
	listQuery := fmt.Sprintf(
		`SELECT %s
		 FROM catalog.products p
		 %s
		 LEFT JOIN catalog.categories c ON c.id = p.category_id
		 WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		productColumnsLocalized, joinT, whereSQL, orderSQL, idx, idx+1)

	args = append(args, f.PageSize, offset)

	var list []model.Product
	if err := r.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// FindVariantBySku 按 SKU 查询变体（下单时用于服务端取价）
func (r *ProductRepo) FindVariantBySku(ctx context.Context, skuCode string) (*model.ProductVariant, error) {
	query := `SELECT id, product_id, sku_code, title, options::text as options,
		price_cents, compare_at_cents, weight_g, image_key, status, sort_order, created_at, updated_at
		FROM catalog.product_variants WHERE sku_code = $1 LIMIT 1`

	var v model.ProductVariant
	if err := r.conn.QueryRowCtx(ctx, &v, query, skuCode); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

// BatchFirstImages 批量取商品首图，避免列表查询出现 N+1
func (r *ProductRepo) BatchFirstImages(ctx context.Context, productIds []string) (map[string]string, error) {
	result := make(map[string]string, len(productIds))
	if len(productIds) == 0 {
		return result, nil
	}

	query := `SELECT DISTINCT ON (product_id) product_id, object_key
		FROM catalog.product_images
		WHERE product_id = ANY($1)
		ORDER BY product_id, sort_order, id`

	var rows []struct {
		ProductId string `db:"product_id"`
		ObjectKey string `db:"object_key"`
	}
	if err := r.conn.QueryRowsCtx(ctx, &rows, query, pq.Array(productIds)); err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ProductId] = row.ObjectKey
	}
	return result, nil
}

func (r *ProductRepo) FindBySlug(ctx context.Context, slug, locale string) (*model.ProductDetail, error) {
	if locale == "" {
		locale = i18n.DefaultLocale
	}
	joinT := fmt.Sprintf(translationJoin, "$1")

	query := fmt.Sprintf(
		`SELECT %s, COALESCE(c.slug,'') as category_slug,
			COALESCE(NULLIF(ct.name,''), c.name, '') as category_name,
			COALESCE((SELECT available FROM inventory.inventory_items i WHERE i.variant_id IN
				(SELECT id FROM catalog.product_variants v WHERE v.product_id = p.id) LIMIT 1), 0) as available_stock
		 FROM catalog.products p
		 %s
		 LEFT JOIN catalog.categories c ON c.id = p.category_id
		 LEFT JOIN catalog.category_translations ct ON ct.category_id = c.id AND ct.locale = $1
		 WHERE p.slug = $2 LIMIT 1`, productColumnsLocalized, joinT)

	var detail model.ProductDetail
	if err := r.conn.QueryRowCtx(ctx, &detail, query, locale, slug); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	variants, err := r.listVariants(ctx, detail.Id)
	if err != nil {
		return nil, err
	}
	detail.Variants = variants

	images, err := r.listImages(ctx, detail.Id)
	if err != nil {
		return nil, err
	}
	detail.Images = images

	return &detail, nil
}

func (r *ProductRepo) listVariants(ctx context.Context, productId string) ([]model.ProductVariant, error) {
	query := `SELECT id, product_id, sku_code, title, options::text as options,
		price_cents, compare_at_cents, weight_g, image_key, status, sort_order, created_at, updated_at
		FROM catalog.product_variants WHERE product_id = $1 ORDER BY sort_order, id`
	var list []model.ProductVariant
	if err := r.conn.QueryRowsCtx(ctx, &list, query, productId); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *ProductRepo) listImages(ctx context.Context, productId string) ([]model.ProductImage, error) {
	query := `SELECT id, product_id, variant_id, object_key, alt, sort_order, created_at
		FROM catalog.product_images WHERE product_id = $1 ORDER BY sort_order, id`
	var list []model.ProductImage
	if err := r.conn.QueryRowsCtx(ctx, &list, query, productId); err != nil {
		return nil, err
	}
	return list, nil
}

// ---------- 后台 ----------

func (r *ProductRepo) AdminList(ctx context.Context, keyword string, page, pageSize int) ([]model.Product, int64, error) {
	where := "1=1"
	args := []interface{}{}
	idx := 1
	if keyword != "" {
		where = fmt.Sprintf("p.title ILIKE $%d OR p.slug ILIKE $%d", idx, idx)
		args = append(args, "%"+keyword+"%")
		idx++
	}

	var total int64
	if err := r.conn.QueryRowCtx(ctx, &total,
		fmt.Sprintf(`SELECT COUNT(*) FROM catalog.products p WHERE %s`, where), args...); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(
		`SELECT p.id, p.title, p.slug, p.subtitle, p.description, p.category_id, p.status,
			p.price_cents, p.currency, p.attributes::text as attributes, p.tags::text as tags,
			p.seo_title, p.seo_description, p.published_at, p.created_at, p.updated_at
		 FROM catalog.products p WHERE %s
		 ORDER BY p.updated_at DESC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1)

	args = append(args, pageSize, (page-1)*pageSize)
	var list []model.Product
	if err := r.conn.QueryRowsCtx(ctx, &list, query, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *ProductRepo) FindById(ctx context.Context, id string) (*model.ProductDetail, error) {
	query := fmt.Sprintf(
		`SELECT %s, COALESCE(c.slug,'') as category_slug, COALESCE(c.name,'') as category_name, 0 as available_stock
		 FROM catalog.products p LEFT JOIN catalog.categories c ON c.id = p.category_id
		 WHERE p.id = $1 LIMIT 1`, productColumns)

	var detail model.ProductDetail
	if err := r.conn.QueryRowCtx(ctx, &detail, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	detail.Variants, _ = r.listVariants(ctx, id)
	detail.Images, _ = r.listImages(ctx, id)
	detail.Translations, _ = r.listTranslations(ctx, id)
	return &detail, nil
}

// listTranslations 加载商品的全部翻译，供后台编辑
func (r *ProductRepo) listTranslations(ctx context.Context, productId string) ([]model.ProductTranslation, error) {
	query := `SELECT product_id, locale, title, subtitle, description, seo_title, seo_description
		FROM catalog.product_translations WHERE product_id = $1 ORDER BY locale`
	var list []model.ProductTranslation
	if err := r.conn.QueryRowsCtx(ctx, &list, query, productId); err != nil {
		return nil, err
	}
	return list, nil
}

// saveTranslations 在同一事务内写入翻译（空内容则删除该语言记录）
func saveTranslations(ctx context.Context, session sqlx.Session, productId string, translations map[string]TranslationInput) error {
	for locale, tr := range translations {
		locale = i18n.Normalize(locale)
		if locale == i18n.DefaultLocale {
			continue // 默认语言内容存主表
		}
		if tr.Title == "" && tr.Subtitle == "" && tr.Description == "" &&
			tr.SeoTitle == "" && tr.SeoDescription == "" {
			delStmt, err := session.Prepare(
				`DELETE FROM catalog.product_translations WHERE product_id=$1 AND locale=$2`)
			if err != nil {
				return err
			}
			if _, err := delStmt.ExecCtx(ctx, productId, locale); err != nil {
				delStmt.Close()
				return err
			}
			delStmt.Close()
			continue
		}

		upsert := `INSERT INTO catalog.product_translations
		 (product_id, locale, title, subtitle, description, seo_title, seo_description)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT (product_id, locale) DO UPDATE SET
			title=EXCLUDED.title, subtitle=EXCLUDED.subtitle, description=EXCLUDED.description,
			seo_title=EXCLUDED.seo_title, seo_description=EXCLUDED.seo_description,
			updated_at=now()`
		stmt, err := session.Prepare(upsert)
		if err != nil {
			return err
		}
		_, err = stmt.ExecCtx(ctx, productId, locale, tr.Title, tr.Subtitle,
			tr.Description, tr.SeoTitle, tr.SeoDescription)
		stmt.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

type CreateProductInput struct {
	Title       string
	Slug        string
	Subtitle    string
	Description string
	CategoryId  *string
	Status      string
	PriceCents  int64
	Currency    string
	Attributes  map[string]interface{}
	Tags        []string
	SeoTitle    string
	SeoDesc     string
	Variants    []VariantInput
	Images      []ImageInput
	// locale -> 翻译内容
	Translations map[string]TranslationInput
}

// TranslationInput 商品翻译内容
type TranslationInput struct {
	Title          string
	Subtitle       string
	Description    string
	SeoTitle       string
	SeoDescription string
}

type VariantInput struct {
	SkuCode        string
	Title          string
	Options        map[string]interface{}
	PriceCents     int64
	CompareAtCents int64
	WeightG        int
	ImageKey       string
	Stock          int
	SortOrder      int
}

type ImageInput struct {
	ObjectKey string
	Alt       string
	SortOrder int
}

// Create 在一个事务内创建商品、变体、图片与库存
func (r *ProductRepo) Create(ctx context.Context, in CreateProductInput) (string, error) {
	attrJSON, _ := json.Marshal(in.Attributes)
	if in.Attributes == nil {
		attrJSON = []byte("{}")
	}
	// tags 是 PostgreSQL text[] 列，需用 pq.Array 编码为数组字面量（{}），
	// 不能用 json.Marshal（会得到 "[]"，pq 解析会报 malformed array literal）
	tags := in.Tags
	if tags == nil {
		tags = []string{}
	}

	var productId string
	err := r.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		publishedAt := "NULL"
		if in.Status == "published" {
			publishedAt = "now()"
		}
		insertProduct := fmt.Sprintf(
			`INSERT INTO catalog.products
			 (title, slug, subtitle, description, category_id, status, price_cents, currency,
			  attributes, tags, seo_title, seo_description, published_at)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,%s) RETURNING id`, publishedAt)

		stmt, err := session.Prepare(insertProduct)
		if err != nil {
			return err
		}
		defer stmt.Close()

		if err := stmt.QueryRowCtx(ctx, &productId,
			in.Title, in.Slug, in.Subtitle, in.Description, in.CategoryId, in.Status,
			in.PriceCents, in.Currency, string(attrJSON), pq.Array(tags), in.SeoTitle, in.SeoDesc,
		); err != nil {
			if isUniqueViolation(err) {
				return ErrSlugConflict
			}
			return err
		}

		for _, v := range in.Variants {
			optJSON, _ := json.Marshal(v.Options)
			if v.Options == nil {
				optJSON = []byte("{}")
			}
			insertVariant := `INSERT INTO catalog.product_variants
			 (product_id, sku_code, title, options, price_cents, compare_at_cents, weight_g, image_key, sort_order)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`
			vStmt, err := session.Prepare(insertVariant)
			if err != nil {
				return err
			}
			var variantId string
			err = vStmt.QueryRowCtx(ctx, &variantId, productId, v.SkuCode, v.Title, string(optJSON),
				v.PriceCents, v.CompareAtCents, v.WeightG, v.ImageKey, v.SortOrder)
			vStmt.Close()
			if err != nil {
				if isUniqueViolation(err) {
					return ErrSkuConflict
				}
				return err
			}

			invStmt, err := session.Prepare(
				`INSERT INTO inventory.inventory_items (variant_id, available, reserved) VALUES ($1,$2,0)
				 ON CONFLICT (variant_id) DO NOTHING`)
			if err != nil {
				return err
			}
			if _, err := invStmt.ExecCtx(ctx, variantId, v.Stock); err != nil {
				invStmt.Close()
				return err
			}
			invStmt.Close()
		}

		for _, img := range in.Images {
			imgStmt, err := session.Prepare(
				`INSERT INTO catalog.product_images (product_id, object_key, alt, sort_order) VALUES ($1,$2,$3,$4)`)
			if err != nil {
				return err
			}
			if _, err := imgStmt.ExecCtx(ctx, productId, img.ObjectKey, img.Alt, img.SortOrder); err != nil {
				imgStmt.Close()
				return err
			}
			imgStmt.Close()
		}

		return saveTranslations(ctx, session, productId, in.Translations)
	})

	return productId, err
}

func (r *ProductRepo) Update(ctx context.Context, id string, in CreateProductInput) error {
	attrJSON, _ := json.Marshal(in.Attributes)
	if in.Attributes == nil {
		attrJSON = []byte("{}")
	}
	// tags 为 text[]，使用 pq.Array 编码（见 Create 中的说明）
	tags := in.Tags
	if tags == nil {
		tags = []string{}
	}

	return r.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		updateProduct := `UPDATE catalog.products SET
			title=$1, slug=$2, subtitle=$3, description=$4, category_id=$5, status=$6,
			price_cents=$7, currency=$8, attributes=$9, tags=$10, seo_title=$11, seo_description=$12,
			published_at = CASE WHEN $6 = 'published' AND published_at IS NULL THEN now() ELSE published_at END
			WHERE id=$13`
		stmt, err := session.Prepare(updateProduct)
		if err != nil {
			return err
		}
		defer stmt.Close()
		if _, err := stmt.ExecCtx(ctx, in.Title, in.Slug, in.Subtitle, in.Description, in.CategoryId,
			in.Status, in.PriceCents, in.Currency, string(attrJSON), pq.Array(tags),
			in.SeoTitle, in.SeoDesc, id); err != nil {
			if isUniqueViolation(err) {
				return ErrSlugConflict
			}
			return err
		}

		// 变体：先禁用旧变体再写入新变体，保留历史订单引用
		if _, err := session.ExecCtx(ctx,
			`UPDATE catalog.product_variants SET status='disabled' WHERE product_id=$1`, id); err != nil {
			return err
		}
		for _, v := range in.Variants {
			optJSON, _ := json.Marshal(v.Options)
			if v.Options == nil {
				optJSON = []byte("{}")
			}
			// 变体 upsert：SKU 全局唯一，但只允许“命中本商品的旧变体”时更新。
			// 若同一 SKU 属于其他商品，WHERE 不成立 → 不更新任何行 → RETURNING 无结果，
			// 由下方判断返回 ErrSkuConflict，避免误改其他商品的数据。
			upsert := `INSERT INTO catalog.product_variants
			 (product_id, sku_code, title, options, price_cents, compare_at_cents, weight_g, image_key, status, sort_order)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'active',$9)
			 ON CONFLICT (sku_code) DO UPDATE SET
				title=EXCLUDED.title, options=EXCLUDED.options, price_cents=EXCLUDED.price_cents,
				compare_at_cents=EXCLUDED.compare_at_cents, weight_g=EXCLUDED.weight_g,
				image_key=EXCLUDED.image_key, status='active', sort_order=EXCLUDED.sort_order
			 WHERE catalog.product_variants.product_id = EXCLUDED.product_id
			 RETURNING catalog.product_variants.id`
			vStmt, err := session.Prepare(upsert)
			if err != nil {
				return err
			}
			var variantId string
			err = vStmt.QueryRowCtx(ctx, &variantId, id, v.SkuCode, v.Title, string(optJSON),
				v.PriceCents, v.CompareAtCents, v.WeightG, v.ImageKey, v.SortOrder)
			vStmt.Close()
			if err == sql.ErrNoRows {
				// 冲突的 SKU 属于其他商品
				return ErrSkuConflict
			}
			if err != nil {
				return err
			}

			invStmt, err := session.Prepare(
				`INSERT INTO inventory.inventory_items (variant_id, available) VALUES ($1,$2)
				 ON CONFLICT (variant_id) DO UPDATE SET available=EXCLUDED.available`)
			if err != nil {
				return err
			}
			if _, err := invStmt.ExecCtx(ctx, variantId, v.Stock); err != nil {
				invStmt.Close()
				return err
			}
			invStmt.Close()
		}

		// 图片：整体替换
		if _, err := session.ExecCtx(ctx, `DELETE FROM catalog.product_images WHERE product_id=$1`, id); err != nil {
			return err
		}
		for _, img := range in.Images {
			imgStmt, err := session.Prepare(
				`INSERT INTO catalog.product_images (product_id, object_key, alt, sort_order) VALUES ($1,$2,$3,$4)`)
			if err != nil {
				return err
			}
			if _, err := imgStmt.ExecCtx(ctx, id, img.ObjectKey, img.Alt, img.SortOrder); err != nil {
				imgStmt.Close()
				return err
			}
			imgStmt.Close()
		}

		return saveTranslations(ctx, session, id, in.Translations)
	})
}

func (r *ProductRepo) Delete(ctx context.Context, id string) error {
	_, err := r.conn.ExecCtx(ctx, `UPDATE catalog.products SET status='archived' WHERE id=$1`, id)
	return err
}
