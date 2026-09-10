package types

// ---------- 通用 ----------
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// ---------- 后台：登录 ----------
type AdminLoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AdminLoginResp struct {
	Token     string    `json:"token"`
	ExpiresIn int64     `json:"expires_in"`
	Admin     AdminInfo `json:"admin"`
}

type AdminInfo struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// ---------- 商品：前台 ----------
type VariantVO struct {
	Id             string                 `json:"id"`
	SkuCode        string                 `json:"sku_code"`
	Title          string                 `json:"title"`
	Options        map[string]interface{} `json:"options"`
	PriceCents     int64                  `json:"price_cents"`
	CompareAtCents int64                  `json:"compare_at_cents"`
	WeightG        int                    `json:"weight_g"`
	ImageUrl       string                 `json:"image_url"`
	InStock        bool                   `json:"in_stock"`
}

type ProductListItem struct {
	Id         string `json:"id"`
	Title      string `json:"title"`
	Slug       string `json:"slug"`
	Subtitle   string `json:"subtitle"`
	PriceCents int64  `json:"price_cents"`
	Currency   string `json:"currency"`
	CoverUrl   string `json:"cover_url"`
	CategorySlug string `json:"category_slug,omitempty"`
}

type ProductDetailVO struct {
	Id             string                 `json:"id"`
	Title          string                 `json:"title"`
	Slug           string                 `json:"slug"`
	Subtitle       string                 `json:"subtitle"`
	Description    string                 `json:"description"`
	PriceCents     int64                  `json:"price_cents"`
	Currency       string                 `json:"currency"`
	CategorySlug   string                 `json:"category_slug"`
	CategoryName   string                 `json:"category_name"`
	Attributes     map[string]interface{} `json:"attributes"`
	Tags           []string               `json:"tags"`
	SeoTitle       string                 `json:"seo_title"`
	SeoDescription string                 `json:"seo_description"`
	Images         []ImageVO               `json:"images"`
	Variants       []VariantVO             `json:"variants"`
}

type ImageVO struct {
	Url  string `json:"url"`
	Alt  string `json:"alt"`
	Sort int    `json:"sort"`
}

// ---------- 类目 ----------
type CategoryVO struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Desc      string `json:"description"`
	ImageUrl  string `json:"image_url"`
	SortOrder int    `json:"sort_order"`
}

// ---------- 后台：商品管理 ----------
type AdminProductReq struct {
	Title       string                 `json:"title"`
	Slug        string                 `json:"slug"`
	Subtitle    string                 `json:"subtitle,optional"`
	Description string                 `json:"description,optional"`
	CategoryId  string                 `json:"category_id,optional"`
	Status      string                 `json:"status,optional"`
	Currency    string                 `json:"currency,optional"`
	Attributes  map[string]interface{} `json:"attributes,optional"`
	Tags        []string               `json:"tags,optional"`
	SeoTitle    string                 `json:"seo_title,optional"`
	SeoDesc     string                 `json:"seo_description,optional"`
	Variants    []AdminVariantReq      `json:"variants,optional"`
	Images      []AdminImageReq        `json:"images,optional"`
	// key 为 locale（如 "zh"），value 为该语言内容
	Translations map[string]TranslationInput `json:"translations,optional"`
}

type AdminVariantReq struct {
	SkuCode        string                 `json:"sku_code,optional"`
	Title          string                 `json:"title,optional"`
	Options        map[string]interface{} `json:"options,optional"`
	PriceCents     int64                  `json:"price_cents,optional"`
	CompareAtCents int64                  `json:"compare_at_cents,optional"`
	WeightG        int                    `json:"weight_g,optional"`
	ImageKey       string                 `json:"image_key,optional"`
	Stock          int                    `json:"stock,optional"`
	SortOrder      int                    `json:"sort_order,optional"`
}

type AdminImageReq struct {
	ObjectKey string `json:"object_key,optional"`
	Alt       string `json:"alt,optional"`
	SortOrder int    `json:"sort_order,optional"`
}

// TranslationInput 商品的多语言内容（locale → 翻译）
// 缺失字段留空即回落到默认语言内容
type TranslationInput struct {
	Title          string `json:"title,optional"`
	Subtitle       string `json:"subtitle,optional"`
	Description    string `json:"description,optional"`
	SeoTitle       string `json:"seo_title,optional"`
	SeoDescription string `json:"seo_description,optional"`
}

type AdminProductItem struct {
	Id         string `json:"id"`
	Title      string `json:"title"`
	Slug       string `json:"slug"`
	Status     string `json:"status"`
	PriceCents int64  `json:"price_cents"`
	Currency   string `json:"currency"`
	Stock      int    `json:"stock"`
	UpdatedAt  string `json:"updated_at"`
}

// ---------- 后台：文件上传 ----------
type PresignReq struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size,optional"`
}

type PresignResp struct {
	UploadUrl string `json:"upload_url"`
	ObjectKey string `json:"object_key"`
	PublicUrl string `json:"public_url"`
}

// ---------- 后台：类目管理 ----------
type AdminCategoryReq struct {
	ParentId    string `json:"parent_id,optional"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,optional"`
	ImageKey    string `json:"image_key,optional"`
	SortOrder   int    `json:"sort_order,optional"`
	Status      string `json:"status,optional"`
	// locale -> 翻译内容（中文覆盖 name/description）
	Translations map[string]CategoryTranslationInput `json:"translations,optional"`
}

// CategoryTranslationInput 类目翻译内容
type CategoryTranslationInput struct {
	Name        string `json:"name,optional"`
	Description string `json:"description,optional"`
}

// ---------- 下单 ----------
type CheckoutItemReq struct {
	SkuCode string `json:"sku_code"`
	Qty     int    `json:"qty,optional"`
}

type CreateOrderReq struct {
	Email           string                 `json:"email"`
	Currency        string                 `json:"currency,optional"`
	Items           []CheckoutItemReq      `json:"items"`
	ShippingAddress map[string]interface{} `json:"shipping_address"`
	BillingAddress  map[string]interface{} `json:"billing_address,optional"`
	CustomerNote    string                 `json:"customer_note,optional"`
	// 注意：discount_cents 仅作兼容保留，实际折扣一律由服务端依据 coupon_code 计算，
	// 客户端传入值不会被信任
	DiscountCents int64  `json:"discount_cents,optional"`
	CouponCode    string `json:"coupon_code,optional"`
}

type OrderItemVO struct {
	SkuCode   string                 `json:"sku_code"`
	Title     string                 `json:"title"`
	Options   map[string]interface{} `json:"options"`
	ImageUrl  string                 `json:"image_url"`
	UnitPrice int64                  `json:"unit_price_cents"`
	Qty       int                    `json:"qty"`
	Total     int64                  `json:"total_cents"`
}

type OrderVO struct {
	OrderNo        string                 `json:"order_no"`
	Email          string                 `json:"email"`
	Status         string                 `json:"status"`
	Currency       string                 `json:"currency"`
	SubtotalCents  int64                  `json:"subtotal_cents"`
	ShippingCents  int64                  `json:"shipping_cents"`
	DiscountCents  int64                  `json:"discount_cents"`
	TotalCents     int64                  `json:"total_cents"`
	CustomerNote   string                 `json:"customer_note"`
	ShippingAddr   map[string]interface{} `json:"shipping_address"`
	Items          []OrderItemVO          `json:"items"`
	PaymentStatus  string                 `json:"payment_status,omitempty"`
	Provider       string                 `json:"provider,omitempty"`
	ClientSecret   string                 `json:"client_secret,omitempty"`
	CheckoutUrl    string                 `json:"checkout_url,omitempty"`
	Carrier        string                 `json:"carrier,omitempty"`
	TrackingNo     string                 `json:"tracking_no,omitempty"`
	TrackingUrl    string                 `json:"tracking_url,omitempty"`
	CreatedAt      string                 `json:"created_at"`
}

// ---------- 后台：订单管理 ----------
type AdminOrderItem struct {
	Id         string `json:"id"`
	OrderNo    string `json:"order_no"`
	Email      string `json:"email"`
	Status     string `json:"status"`
	TotalCents int64  `json:"total_cents"`
	Currency   string `json:"currency"`
	CreatedAt  string `json:"created_at"`
}

type AdminOrderDetail struct {
	Id            string                 `json:"id"`
	OrderNo       string                 `json:"order_no"`
	Email         string                 `json:"email"`
	Status        string                 `json:"status"`
	Currency      string                 `json:"currency"`
	SubtotalCents int64                  `json:"subtotal_cents"`
	ShippingCents int64                  `json:"shipping_cents"`
	DiscountCents int64                  `json:"discount_cents"`
	TotalCents    int64                  `json:"total_cents"`
	CustomerNote  string                 `json:"customer_note"`
	ShippingAddr  map[string]interface{} `json:"shipping_address"`
	Items         []OrderItemVO          `json:"items"`
	PaymentStatus string                 `json:"payment_status,omitempty"`
	Provider      string                 `json:"provider,omitempty"`
	Carrier       string                 `json:"carrier,omitempty"`
	TrackingNo    string                 `json:"tracking_no,omitempty"`
	TrackingUrl   string                 `json:"tracking_url,omitempty"`
	ShipmentStatus string                `json:"shipment_status,omitempty"`
	CouponCode    string                 `json:"coupon_code,omitempty"`
	CreatedAt     string                 `json:"created_at"`
}

type ShipOrderReq struct {
	Carrier     string `json:"carrier"`
	TrackingNo  string `json:"tracking_no"`
	TrackingUrl string `json:"tracking_url,optional"`
}

type CancelOrderReq struct {
	Reason string `json:"reason,optional"`
}

// ---------- 后台：优惠券管理 ----------
type AdminCouponItem struct {
	Id             string `json:"id"`
	Code           string `json:"code"`
	Type           string `json:"type"`  // percent | fixed
	Value          int64  `json:"value"` // percent: 百分比(10=9折) fixed: 分
	MinAmountCents int64  `json:"min_amount_cents"`
	MaxUses        int    `json:"max_uses"` // 0 表示不限
	UsedCount      int    `json:"used_count"`
	StartsAt       string `json:"starts_at,omitempty"`
	EndsAt         string `json:"ends_at,omitempty"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
}

type AdminCouponReq struct {
	Code           string `json:"code"`
	Type           string `json:"type"`
	Value          int64  `json:"value"`
	MinAmountCents int64  `json:"min_amount_cents,optional"`
	MaxUses        int    `json:"max_uses,optional"`
	StartsAt       string `json:"starts_at,optional"`
	EndsAt         string `json:"ends_at,optional"`
	Status         string `json:"status,optional"`
}

type SetCouponStatusReq struct {
	Status string `json:"status"`
}

// ---------- 前台：优惠券试算 ----------
type ValidateCouponReq struct {
	Code     string `json:"code"`
	Subtotal int64  `json:"subtotal_cents,optional"`
}

type ValidateCouponResp struct {
	Valid         bool   `json:"valid"`
	Code          string `json:"code"`
	DiscountCents int64  `json:"discount_cents"`
	Message       string `json:"message"`
}

type PublicCoupon struct {
	Code           string `json:"code"`
	Type           string `json:"type"`
	Value          int64  `json:"value"`
	MinAmountCents int64  `json:"min_amount_cents"`
	EndsAt         string `json:"ends_at,omitempty"`
}

// ---------- 后台：运费规则 ----------
type AdminShippingRule struct {
	Id                 string   `json:"id"`
	Name               string   `json:"name"`
	CountryCodes       []string `json:"country_codes"` // 空表示其余所有国家
	MinAmountCents     int64    `json:"min_amount_cents"`
	MaxWeightG         int      `json:"max_weight_g"`
	PriceCents         int64    `json:"price_cents"`
	FreeThresholdCents int64    `json:"free_threshold_cents"`
	SortOrder          int      `json:"sort_order"`
	Status             string   `json:"status"`
}

type AdminShippingRuleReq struct {
	Name               string   `json:"name"`
	CountryCodes       []string `json:"country_codes,optional"`
	MinAmountCents     int64    `json:"min_amount_cents,optional"`
	MaxWeightG         int      `json:"max_weight_g,optional"`
	PriceCents         int64    `json:"price_cents,optional"`
	FreeThresholdCents int64    `json:"free_threshold_cents,optional"`
	SortOrder          int      `json:"sort_order,optional"`
	Status             string   `json:"status,optional"`
}

// ---------- 支付回调 ----------
type PayNotifyReq struct {
	Provider     string `json:"provider"`
	EventId      string `json:"event_id"`
	OrderNo      string `json:"order_no"`
	Status       string `json:"status"`
	AmountCents  int64  `json:"amount_cents"`
}
