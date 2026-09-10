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
	Subtitle    string                 `json:"subtitle"`
	Description string                 `json:"description"`
	CategoryId  string                 `json:"category_id"`
	Status      string                 `json:"status"`
	Currency    string                 `json:"currency"`
	Attributes  map[string]interface{} `json:"attributes"`
	Tags        []string               `json:"tags"`
	SeoTitle    string                 `json:"seo_title"`
	SeoDesc     string                 `json:"seo_description"`
	Variants    []AdminVariantReq      `json:"variants"`
	Images      []AdminImageReq        `json:"images"`
	// key 为 locale（如 "zh"），value 为该语言内容
	Translations map[string]TranslationInput `json:"translations"`
}

type AdminVariantReq struct {
	SkuCode        string                 `json:"sku_code"`
	Title          string                 `json:"title"`
	Options        map[string]interface{} `json:"options"`
	PriceCents     int64                  `json:"price_cents"`
	CompareAtCents int64                  `json:"compare_at_cents"`
	WeightG        int                    `json:"weight_g"`
	ImageKey       string                 `json:"image_key"`
	Stock          int                    `json:"stock"`
	SortOrder      int                    `json:"sort_order"`
}

type AdminImageReq struct {
	ObjectKey string `json:"object_key"`
	Alt       string `json:"alt"`
	SortOrder int    `json:"sort_order"`
}

// TranslationInput 商品的多语言内容（locale → 翻译）
// 缺失字段留空即回落到默认语言内容
type TranslationInput struct {
	Title          string `json:"title"`
	Subtitle       string `json:"subtitle"`
	Description    string `json:"description"`
	SeoTitle       string `json:"seo_title"`
	SeoDescription string `json:"seo_description"`
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
	Size        int64  `json:"size"`
}

type PresignResp struct {
	UploadUrl string `json:"upload_url"`
	ObjectKey string `json:"object_key"`
	PublicUrl string `json:"public_url"`
}

// ---------- 后台：类目管理 ----------
type AdminCategoryReq struct {
	ParentId    string `json:"parent_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ImageKey    string `json:"image_key"`
	SortOrder   int    `json:"sort_order"`
	Status      string `json:"status"`
	// locale -> 翻译内容（中文覆盖 name/description）
	Translations map[string]CategoryTranslationInput `json:"translations"`
}

// CategoryTranslationInput 类目翻译内容
type CategoryTranslationInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ---------- 下单 ----------
type CheckoutItemReq struct {
	SkuCode string `json:"sku_code"`
	Qty     int    `json:"qty"`
}

type CreateOrderReq struct {
	Email           string                 `json:"email"`
	Currency        string                 `json:"currency"`
	Items           []CheckoutItemReq      `json:"items"`
	ShippingAddress map[string]interface{} `json:"shipping_address"`
	BillingAddress  map[string]interface{} `json:"billing_address"`
	CustomerNote    string                 `json:"customer_note"`
	DiscountCents   int64                  `json:"discount_cents"`
	CouponCode      string                 `json:"coupon_code"`
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
	CreatedAt     string                 `json:"created_at"`
}

type ShipOrderReq struct {
	Carrier     string `json:"carrier"`
	TrackingNo  string `json:"tracking_no"`
	TrackingUrl string `json:"tracking_url"`
}

type CancelOrderReq struct {
	Reason string `json:"reason"`
}

// ---------- 支付回调 ----------
type PayNotifyReq struct {
	Provider     string `json:"provider"`
	EventId      string `json:"event_id"`
	OrderNo      string `json:"order_no"`
	Status       string `json:"status"`
	AmountCents  int64  `json:"amount_cents"`
}
