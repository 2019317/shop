package repo

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/yourname/stationery-shop/apps/api/internal/model"
)

type ShippingRepo struct {
	conn sqlx.SqlConn
}

func NewShippingRepo(conn sqlx.SqlConn) *ShippingRepo {
	return &ShippingRepo{conn: conn}
}

// Match 按目的地国家、订单金额与重量匹配运费规则
// 规则优先级：指定国家的规则优先，其次通用规则（country_codes 为空）
func (r *ShippingRepo) Match(ctx context.Context, countryCode string, amountCents int64, weightG int) (int64, string, error) {
	query := `SELECT id, name, country_codes::text as country_codes, min_amount_cents, max_weight_g,
		price_cents, free_threshold_cents, sort_order, status
		FROM fulfillment.shipping_rules
		WHERE status='active'
		ORDER BY sort_order, id`

	var rules []model.ShippingRule
	if err := r.conn.QueryRowsCtx(ctx, &rules, query); err != nil {
		return 0, "", err
	}

	country := strings.ToUpper(countryCode)
	var fallback *model.ShippingRule
	for i := range rules {
		rule := rules[i]
		if rule.MaxWeightG > 0 && weightG > rule.MaxWeightG {
			continue
		}
		if amountCents < rule.MinAmountCents {
			continue
		}

		codes := parseArray(rule.CountryCodes)
		matched := len(codes) == 0 // 空数组表示通用规则
		for _, c := range codes {
			if strings.EqualFold(c, country) {
				matched = true
				break
			}
		}

		if !matched {
			continue
		}
		if len(codes) == 0 {
			if fallback == nil {
				fallback = &rule
			}
			continue
		}

		// 命中指定国家规则：立即返回
		return applyRule(rule, amountCents), rule.Name, nil
	}

	if fallback != nil {
		return applyRule(*fallback, amountCents), fallback.Name, nil
	}
	return 0, "", nil
}

func applyRule(rule model.ShippingRule, amountCents int64) int64 {
	if rule.FreeThreshold > 0 && amountCents >= rule.FreeThreshold {
		return 0
	}
	return rule.PriceCents
}

func parseArray(raw string) []string {
	raw = strings.Trim(raw, "{}")
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.Trim(p, `"`))
	}
	return out
}
