package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/yourname/stationery-shop/apps/api/internal/model"
)

// ErrCouponCodeConflict 券码已被占用
var ErrCouponCodeConflict = errors.New("coupon code already exists")

// CouponRepo 优惠券数据访问（promotion 域）
type CouponRepo struct {
	conn sqlx.SqlConn
}

func NewCouponRepo(conn sqlx.SqlConn) *CouponRepo {
	return &CouponRepo{conn: conn}
}

const couponColumns = `id, code, type, value, min_amount_cents, max_uses, used_count,
	starts_at, ends_at, status, created_at, updated_at`

// FindActiveByCode 按券码查询（不区分大小写）。不存在时返回 (nil, nil)
func (r *CouponRepo) FindActiveByCode(ctx context.Context, code string) (*model.Coupon, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM promotion.coupons WHERE lower(code) = lower($1) LIMIT 1`, couponColumns)
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

// ---------- 后台管理 ----------

// CouponFilter 后台列表筛选条件
type CouponFilter struct {
	Keyword  string
	Status   string
	Page     int
	PageSize int
}

// AdminList 后台分页查询优惠券
func (r *CouponRepo) AdminList(ctx context.Context, f CouponFilter) ([]model.Coupon, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	idx := 1
	if f.Keyword != "" {
		where = append(where, fmt.Sprintf("code ILIKE $%d", idx))
		args = append(args, "%"+f.Keyword+"%")
		idx++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, f.Status)
		idx++
	}
	whereSQL := strings.Join(where, " AND ")

	var total int64
	if err := r.conn.QueryRowCtx(ctx, &total,
		fmt.Sprintf(`SELECT COUNT(*) FROM promotion.coupons WHERE %s`, whereSQL), args...); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(
		`SELECT %s FROM promotion.coupons WHERE %s
		 ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		couponColumns, whereSQL, idx, idx+1)
	var list []model.Coupon
	if err := r.conn.QueryRowsCtx(ctx, &list, query,
		append(args, f.PageSize, (f.Page-1)*f.PageSize)...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// CouponInput 创建/更新优惠券的输入
type CouponInput struct {
	Code           string
	Type           string
	Value          int64
	MinAmountCents int64
	MaxUses        int
	StartsAt       *string
	EndsAt         *string
	Status         string
}

// FindById 按主键查询
func (r *CouponRepo) FindById(ctx context.Context, id string) (*model.Coupon, error) {
	query := fmt.Sprintf(`SELECT %s FROM promotion.coupons WHERE id=$1 LIMIT 1`, couponColumns)
	var c model.Coupon
	if err := r.conn.QueryRowCtx(ctx, &c, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// CodeExists 判断券码是否已被占用（可排除某个 id，排除时传合法 uuid，创建时传空串）
func (r *CouponRepo) CodeExists(ctx context.Context, code, excludeId string) (bool, error) {
	// 注意：不能用 id <> $2::uuid 直接比较，空串转 uuid 会报错；
	// 用 NULLIF 将空串转为 NULL，比较结果为 NULL（视作不排除），避免类型转换异常。
	query := `SELECT EXISTS(
		SELECT 1 FROM promotion.coupons
		WHERE lower(code)=lower($1)
		  AND ($2 = '' OR id <> NULLIF($2, '')::uuid)
	)`
	if err := r.conn.QueryRowCtx(ctx, &exists, query, code, excludeId); err != nil {
		return false, err
	}
	return exists, nil
}

// Create 新建优惠券，返回 id
func (r *CouponRepo) Create(ctx context.Context, in CouponInput) (string, error) {
	query := `INSERT INTO promotion.coupons
	 (code, type, value, min_amount_cents, max_uses, starts_at, ends_at, status)
	 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`
	var id string
	if err := r.conn.QueryRowCtx(ctx, &id, query, in.Code, in.Type, in.Value,
		in.MinAmountCents, in.MaxUses, in.StartsAt, in.EndsAt, in.Status); err != nil {
		if isUniqueViolation(err) {
			return "", ErrCouponCodeConflict
		}
		return "", err
	}
	return id, nil
}

// Update 更新优惠券（券码与已用量不可改，避免破坏核销记录）
func (r *CouponRepo) Update(ctx context.Context, id string, in CouponInput) error {
	query := `UPDATE promotion.coupons SET
	 type=$1, value=$2, min_amount_cents=$3, max_uses=$4,
	 starts_at=$5, ends_at=$6, status=$7, updated_at=now()
	 WHERE id=$8`
	_, err := r.conn.ExecCtx(ctx, query, in.Type, in.Value, in.MinAmountCents,
		in.MaxUses, in.StartsAt, in.EndsAt, in.Status, id)
	return err
}

// SetStatus 启用/停用优惠券
func (r *CouponRepo) SetStatus(ctx context.Context, id, status string) error {
	_, err := r.conn.ExecCtx(ctx,
		`UPDATE promotion.coupons SET status=$1, updated_at=now() WHERE id=$2`, status, id)
	return err
}

// ListActive 前台可用券列表（未过期、已开始、状态 active）
func (r *CouponRepo) ListActive(ctx context.Context) ([]model.Coupon, error) {
	query := fmt.Sprintf(`SELECT %s FROM promotion.coupons
		WHERE status='active'
		  AND (starts_at IS NULL OR starts_at <= now())
		  AND (ends_at IS NULL OR ends_at >= now())
		  AND (max_uses = 0 OR used_count < max_uses)
		ORDER BY created_at DESC`, couponColumns)
	var list []model.Coupon
	if err := r.conn.QueryRowsCtx(ctx, &list, query); err != nil {
		return nil, err
	}
	return list, nil
}
