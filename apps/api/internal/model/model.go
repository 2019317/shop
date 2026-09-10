package model

import (
	"time"
)

// ---------- admin ----------
type AdminUser struct {
	Id           string     `db:"id"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	Name         string     `db:"name"`
	Role         string     `db:"role"`
	Status       string     `db:"status"`
	LastLoginAt  *time.Time `db:"last_login_at"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

// ---------- catalog ----------
type Category struct {
	Id          string    `db:"id"`
	ParentId    *string   `db:"parent_id"`
	Name        string    `db:"name"`
	Slug        string    `db:"slug"`
	Description string    `db:"description"`
	ImageKey    string    `db:"image_key"`
	SortOrder   int       `db:"sort_order"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type Product struct {
	Id             string    `db:"id"`
	Title          string    `db:"title"`
	Slug           string    `db:"slug"`
	Subtitle       string    `db:"subtitle"`
	Description    string    `db:"description"`
	CategoryId     *string   `db:"category_id"`
	Status         string    `db:"status"`
	PriceCents     int64     `db:"price_cents"`
	Currency       string    `db:"currency"`
	Attributes     string    `db:"attributes"` // jsonb 以文本读取
	Tags           string    `db:"tags"`       // text[] 以文本读取
	SeoTitle       string    `db:"seo_title"`
	SeoDescription string    `db:"seo_description"`
	PublishedAt    *time.Time `db:"published_at"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type ProductVariant struct {
	Id             string    `db:"id"`
	ProductId      string    `db:"product_id"`
	SkuCode        string    `db:"sku_code"`
	Title          string    `db:"title"`
	Options        string    `db:"options"` // jsonb
	PriceCents     int64     `db:"price_cents"`
	CompareAtCents int64     `db:"compare_at_cents"`
	WeightG        int       `db:"weight_g"`
	ImageKey       string    `db:"image_key"`
	Status         string    `db:"status"`
	SortOrder      int       `db:"sort_order"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type ProductImage struct {
	Id        string    `db:"id"`
	ProductId string    `db:"product_id"`
	VariantId *string   `db:"variant_id"`
	ObjectKey string    `db:"object_key"`
	Alt       string    `db:"alt"`
	SortOrder int       `db:"sort_order"`
	CreatedAt time.Time `db:"created_at"`
}

// 商品详情聚合（含变体与图片）
type ProductDetail struct {
	Product
	CategorySlug   string           `db:"category_slug"`
	CategoryName   string           `db:"category_name"`
	Variants       []ProductVariant `db:"-"`
	Images         []ProductImage   `db:"-"`
	AvailableStock int              `db:"available_stock"`
	// 后台编辑用：各语言的翻译内容
	Translations []ProductTranslation `db:"-"`
}

// ProductTranslation 商品翻译（locale 覆盖主表内容）
type ProductTranslation struct {
	ProductId      string `db:"product_id"`
	Locale         string `db:"locale"`
	Title          string `db:"title"`
	Subtitle       string `db:"subtitle"`
	Description    string `db:"description"`
	SeoTitle       string `db:"seo_title"`
	SeoDescription string `db:"seo_description"`
}
