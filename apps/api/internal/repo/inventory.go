package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrReservationNotFound = errors.New("reservation not found")
)

type InventoryRepo struct {
	conn sqlx.SqlConn
}

func NewInventoryRepo(conn sqlx.SqlConn) *InventoryRepo {
	return &InventoryRepo{conn: conn}
}

// 库存预留有效期：超时未支付则释放
const ReservationTTL = 30 * time.Minute

// Reserve 预占库存（行锁保证并发安全）
// 流程：available -= qty，reserved += qty，写入预占记录
func (r *InventoryRepo) Reserve(ctx context.Context, variantId string, qty int, refType, refId string) (string, error) {
	var reservationId string

	err := r.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		var available int
		// SELECT ... FOR UPDATE 锁定该行，防止并发超卖
		if err := session.QueryRowCtx(ctx, &available,
			`SELECT available FROM inventory.inventory_items WHERE variant_id=$1 FOR UPDATE`,
			variantId); err != nil {
			if err == sql.ErrNoRows {
				return ErrInsufficientStock
			}
			return err
		}
		if available < qty {
			return ErrInsufficientStock
		}

		if _, err := session.ExecCtx(ctx,
			`UPDATE inventory.inventory_items
			 SET available = available - $1, reserved = reserved + $1, updated_at = now()
			 WHERE variant_id=$2`, qty, variantId); err != nil {
			return err
		}

		expiresAt := time.Now().Add(ReservationTTL)
		stmt, err := session.Prepare(
			`INSERT INTO inventory.stock_reservations (variant_id, qty, ref_type, ref_id, status, expires_at)
			 VALUES ($1,$2,$3,$4,'held',$5) RETURNING id`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		return stmt.QueryRowCtx(ctx, &reservationId, variantId, qty, refType, refId, expiresAt)
	})

	return reservationId, err
}

// Commit 支付成功后真正扣减（reserved 减少，库存不再回滚）
func (r *InventoryRepo) Commit(ctx context.Context, refType, refId string) error {
	return r.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		query := `SELECT id, variant_id, qty FROM inventory.stock_reservations
			WHERE ref_type=$1 AND ref_id=$2 AND status='held'`
		var rows []struct {
			Id        string `db:"id"`
			VariantId string `db:"variant_id"`
			Qty       int    `db:"qty"`
		}
		if err := session.QueryRowsCtx(ctx, &rows, query, refType, refId); err != nil {
			return err
		}
		for _, row := range rows {
			if _, err := session.ExecCtx(ctx,
				`UPDATE inventory.inventory_items SET reserved = reserved - $1, updated_at = now()
				 WHERE variant_id=$2`, row.Qty, row.VariantId); err != nil {
				return err
			}
			if _, err := session.ExecCtx(ctx,
				`UPDATE inventory.stock_reservations SET status='committed' WHERE id=$1`, row.Id); err != nil {
				return err
			}
		}
		return nil
	})
}

// Release 取消订单/支付失败时回滚库存
func (r *InventoryRepo) Release(ctx context.Context, refType, refId string) error {
	return r.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		query := `SELECT id, variant_id, qty FROM inventory.stock_reservations
			WHERE ref_type=$1 AND ref_id=$2 AND status='held'`
		var rows []struct {
			Id        string `db:"id"`
			VariantId string `db:"variant_id"`
			Qty       int    `db:"qty"`
		}
		if err := session.QueryRowsCtx(ctx, &rows, query, refType, refId); err != nil {
			return err
		}
		for _, row := range rows {
			if _, err := session.ExecCtx(ctx,
				`UPDATE inventory.inventory_items
				 SET available = available + $1, reserved = reserved - $1, updated_at = now()
				 WHERE variant_id=$2`, row.Qty, row.VariantId); err != nil {
				return err
			}
			if _, err := session.ExecCtx(ctx,
				`UPDATE inventory.stock_reservations SET status='released' WHERE id=$1`, row.Id); err != nil {
				return err
			}
		}
		return nil
	})
}

// ReleaseExpired 释放超时未支付的预占（定时任务调用）
func (r *InventoryRepo) ReleaseExpired(ctx context.Context) (int, error) {
	var rows []struct {
		Id        string `db:"id"`
		VariantId string `db:"variant_id"`
		Qty       int    `db:"qty"`
	}
	query := `SELECT id, variant_id, qty FROM inventory.stock_reservations
		WHERE status='held' AND expires_at < now()`
	if err := r.conn.QueryRowsCtx(ctx, &rows, query); err != nil {
		return 0, err
	}

	count := 0
	for _, row := range rows {
		err := r.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
			if _, err := session.ExecCtx(ctx,
				`UPDATE inventory.inventory_items
				 SET available = available + $1, reserved = reserved - $1, updated_at = now()
				 WHERE variant_id=$2`, row.Qty, row.VariantId); err != nil {
				return err
			}
			_, err := session.ExecCtx(ctx,
				`UPDATE inventory.stock_reservations SET status='released' WHERE id=$1 AND status='held'`, row.Id)
			return err
		})
		if err == nil {
			count++
		}
	}
	return count, nil
}

// Available 查询可用库存
func (r *InventoryRepo) Available(ctx context.Context, variantId string) (int, error) {
	var available int
	if err := r.conn.QueryRowCtx(ctx, &available,
		`SELECT available FROM inventory.inventory_items WHERE variant_id=$1`, variantId); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return available, nil
}

// SetStock 后台直接设置库存（补货）
func (r *InventoryRepo) SetStock(ctx context.Context, variantId string, qty int) error {
	if qty < 0 {
		return fmt.Errorf("stock cannot be negative")
	}
	_, err := r.conn.ExecCtx(ctx,
		`INSERT INTO inventory.inventory_items (variant_id, available, reserved) VALUES ($1,$2,0)
		 ON CONFLICT (variant_id) DO UPDATE SET available=$2, updated_at=now()`, variantId, qty)
	return err
}
