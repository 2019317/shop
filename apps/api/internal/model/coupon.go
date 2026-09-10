package model

import "time"

// Coupon 营销域优惠券
// type: percent（value 为百分比，10 表示 9 折）| fixed（value 为分）
type Coupon struct {
	Id             string     `db:"id"`
	Code           string     `db:"code"`
	Type           string     `db:"type"`
	Value          int64      `db:"value"`
	MinAmountCents int64      `db:"min_amount_cents"`
	MaxUses        int        `db:"max_uses"`
	UsedCount      int        `db:"used_count"`
	StartsAt       *time.Time `db:"starts_at"`
	EndsAt         *time.Time `db:"ends_at"`
	Status         string     `db:"status"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
}
