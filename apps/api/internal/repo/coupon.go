package repo

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/yourname/stationery-shop/apps/api/internal/model"
)

// CouponRepo 优惠券数据访问（promotion 域）
type CouponRepo struct {
	conn sqlx.SqlConn
}

func NewCouponRepo(conn sqlx.SqlConn) *CouponRepo {
	return &CouponRepo{conn: conn}
}

// FindActiveByCode 按券码查询（不区分大小写）。不存在时返回 (nil, nil)
func (r *CouponRepo) FindActiveByCode(ctx context.Context, code string) (*model.Coupon, error) {
	query := `SELECT id, code, type, value, min_amount_cents, max_uses, used_count,
		starts_at, ends_at, status, created_at, updated_at
		FROM promotion.coupons WHERE lower(code) = lower($1) LIMIT 1`
	var c model.Coupon
	if err := r.conn.QueryRowCtx(ctx, &c, query, code); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// IncrementUsage 原子占用一个使用名额，防止并发超发。
// 返回 true 表示占用成功；false 表示名额已用尽（并发或额度不足）。
func (r *CouponRepo) IncrementUsage(ctx context.Context, id string) (bool, error) {
	res, err := r.conn.ExecCtx(ctx,
		`UPDATE promotion.coupons
		 SET used_count = used_count + 1, updated_at = now()
		 WHERE id = $1 AND (max_uses = 0 OR used_count < max_uses)`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

// DecrementUsage 释放一个使用名额（订单取消时调用），不使计数小于 0
func (r *CouponRepo) DecrementUsage(ctx context.Context, id string) error {
	_, err := r.conn.ExecCtx(ctx,
		`UPDATE promotion.coupons
		 SET used_count = GREATEST(used_count - 1, 0), updated_at = now()
		 WHERE id = $1`, id)
	return err
}
