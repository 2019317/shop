package model

import "time"

type Order struct {
	Id              string    `db:"id"`
	OrderNo         string    `db:"order_no"`
	UserId          *string   `db:"user_id"`
	Email           string    `db:"email"`
	Status          string    `db:"status"`
	Currency        string    `db:"currency"`
	SubtotalCents   int64     `db:"subtotal_cents"`
	ShippingCents   int64     `db:"shipping_cents"`
	DiscountCents   int64     `db:"discount_cents"`
	TaxCents        int64     `db:"tax_cents"`
	TotalCents      int64     `db:"total_cents"`
	ShippingAddress string    `db:"shipping_address"` // jsonb 以文本读取
	BillingAddress  string    `db:"billing_address"`
	CustomerNote    string    `db:"customer_note"`
	PaidAt          *time.Time `db:"paid_at"`
	FulfilledAt     *time.Time `db:"fulfilled_at"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

type OrderItem struct {
	Id             string    `db:"id"`
	OrderId        string    `db:"order_id"`
	ProductId      *string   `db:"product_id"`
	VariantId      *string   `db:"variant_id"`
	SkuCode        string    `db:"sku_code"`
	Title          string    `db:"title"`
	Options        string    `db:"options"` // jsonb
	ImageUrl       string    `db:"image_url"`
	UnitPriceCents int64     `db:"unit_price_cents"`
	Qty            int       `db:"qty"`
	TotalCents     int64     `db:"total_cents"`
}

// 订单聚合，含明细
type OrderDetail struct {
	Order
	Items     []OrderItem `db:"-"`
	Payment   *Payment    `db:"-"`
	Shipment  *Shipment   `db:"-"`
}

type Payment struct {
	Id               string    `db:"id"`
	OrderId          string    `db:"order_id"`
	Provider         string    `db:"provider"`
	ProviderIntentId string    `db:"provider_intent_id"`
	Status           string    `db:"status"`
	AmountCents      int64     `db:"amount_cents"`
	Currency         string    `db:"currency"`
	Raw              string    `db:"raw"` // jsonb
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

type Shipment struct {
	Id          string     `db:"id"`
	OrderId     string     `db:"order_id"`
	Carrier     string     `db:"carrier"`
	TrackingNo  string     `db:"tracking_no"`
	TrackingUrl string     `db:"tracking_url"`
	Status      string     `db:"status"`
	ShippedAt   *time.Time `db:"shipped_at"`
	DeliveredAt *time.Time `db:"delivered_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type ShippingRule struct {
	Id                string  `db:"id"`
	Name              string  `db:"name"`
	CountryCodes      string  `db:"country_codes"` // text[]
	MinAmountCents    int64   `db:"min_amount_cents"`
	MaxWeightG        int     `db:"max_weight_g"`
	PriceCents        int64   `db:"price_cents"`
	FreeThreshold     int64   `db:"free_threshold_cents"`
	SortOrder         int     `db:"sort_order"`
	Status            string  `db:"status"`
}
