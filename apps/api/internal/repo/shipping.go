package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
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

// normalizeCodes 清洗国家代码：去空格、转大写、去空项
func normalizeCodes(codes []string) []string {
	out := make([]string, 0, len(codes))
	for _, c := range codes {
		c = strings.ToUpper(strings.TrimSpace(c))
		if c != "" {
			out = append(out, c)
		}
	}
	return out
}

// ---------- 后台管理：运费规则 ----------

type ShippingRuleFilter struct {
	Status   string
	Page     int
	PageSize int
}

func (r *ShippingRepo) AdminList(ctx context.Context, f ShippingRuleFilter) ([]model.ShippingRule, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	idx := 1
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, f.Status)
		idx++
	}
	whereSQL := strings.Join(where, " AND ")

	var total int64
	if err := r.conn.QueryRowCtx(ctx, &total,
		`SELECT COUNT(*) FROM fulfillment.shipping_rules WHERE `+whereSQL, args...); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT id, name, country_codes::text as country_codes, min_amount_cents, max_weight_g,
		price_cents, free_threshold_cents, sort_order, status
		FROM fulfillment.shipping_rules WHERE %s
		ORDER BY sort_order, id LIMIT $%d OFFSET $%d`, whereSQL, idx, idx+1)

	var list []model.ShippingRule
	if err := r.conn.QueryRowsCtx(ctx, &list, query,
		append(args, f.PageSize, (f.Page-1)*f.PageSize)...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ShippingRuleInput 创建/更新运费规则
type ShippingRuleInput struct {
	Name               string
	CountryCodes       []string
	MinAmountCents     int64
	MaxWeightG         int
	PriceCents         int64
	FreeThresholdCents int64
	SortOrder          int
	Status             string
}

func (r *ShippingRepo) FindRuleById(ctx context.Context, id string) (*model.ShippingRule, error) {
	query := `SELECT id, name, country_codes::text as country_codes, min_amount_cents, max_weight_g,
		price_cents, free_threshold_cents, sort_order, status
		FROM fulfillment.shipping_rules WHERE id=$1 LIMIT 1`
	var rule model.ShippingRule
	if err := r.conn.QueryRowCtx(ctx, &rule, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rule, nil
}

func (r *ShippingRepo) CreateRule(ctx context.Context, in ShippingRuleInput) (string, error) {
	query := `INSERT INTO fulfillment.shipping_rules
	 (name, country_codes, min_amount_cents, max_weight_g, price_cents, free_threshold_cents, sort_order, status)
	 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`
	var id string
	if err := r.conn.QueryRowCtx(ctx, &id, query, in.Name, pq.Array(normalizeCodes(in.CountryCodes)),
		in.MinAmountCents, in.MaxWeightG, in.PriceCents, in.FreeThresholdCents, in.SortOrder, in.Status); err != nil {
		return "", err
	}
	return id, nil
}

func (r *ShippingRepo) UpdateRule(ctx context.Context, id string, in ShippingRuleInput) error {
	query := `UPDATE fulfillment.shipping_rules SET
	 name=$1, country_codes=$2, min_amount_cents=$3, max_weight_g=$4,
	 price_cents=$5, free_threshold_cents=$6, sort_order=$7, status=$8, updated_at=now()
	 WHERE id=$9`
	_, err := r.conn.ExecCtx(ctx, query, in.Name, pq.Array(normalizeCodes(in.CountryCodes)),
		in.MinAmountCents, in.MaxWeightG, in.PriceCents, in.FreeThresholdCents, in.SortOrder, in.Status, id)
	return err
}

func (r *ShippingRepo) DeleteRule(ctx context.Context, id string) error {
	_, err := r.conn.ExecCtx(ctx,
		`UPDATE fulfillment.shipping_rules SET status='inactive', updated_at=now() WHERE id=$1`, id)
	return err
}

// ---------- 履约：运单状态 ----------

// MarkDelivered 将订单的最新运单标记为已送达
func (r *ShippingRepo) MarkDelivered(ctx context.Context, orderId string) error {
	_, err := r.conn.ExecCtx(ctx,
		`UPDATE fulfillment.shipments
		 SET status='delivered', delivered_at=now(), updated_at=now()
		 WHERE order_id=$1 AND status='shipped'`, orderId)
	return err
}
