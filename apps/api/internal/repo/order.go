package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/yourname/stationery-shop/apps/api/internal/model"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/password"
)

// ErrOrderMissing 订单不存在
var ErrOrderMissing = errors.New("order not found")

type OrderRepo struct {
	conn sqlx.SqlConn
}

func NewOrderRepo(conn sqlx.SqlConn) *OrderRepo {
	return &OrderRepo{conn: conn}
}

type CreateOrderItem struct {
	VariantId      string
	SkuCode        string
	Title          string
	Options        map[string]interface{}
	ImageUrl       string
	UnitPriceCents int64
	Qty            int
}

type CreateOrderInput struct {
	Email           string
	UserId          *string
	Currency        string
	SubtotalCents   int64
	ShippingCents   int64
	DiscountCents   int64
	TaxCents        int64
	TotalCents      int64
	CouponId        *string
	CouponCode      string
	ShippingAddress map[string]interface{}
	BillingAddress  map[string]interface{}
	CustomerNote    string
	Items           []CreateOrderItem
}

// Create 在一个事务内创建订单与订单项（价格以服务端计算为准）
func (r *OrderRepo) Create(ctx context.Context, in CreateOrderInput) (*model.Order, error) {
	orderNo := generateOrderNo()
	shippingJSON, _ := json.Marshal(in.ShippingAddress)
	billingJSON, _ := json.Marshal(in.BillingAddress)
	if in.ShippingAddress == nil {
		shippingJSON = []byte("{}")
	}
	if in.BillingAddress == nil {
		billingJSON = []byte("{}")
	}

	order := &model.Order{}
	err := r.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		insertOrder := `INSERT INTO orders.orders
		 (order_no, user_id, email, status, currency, subtotal_cents, shipping_cents,
		  discount_cents, tax_cents, total_cents, coupon_id, coupon_code, shipping_address, billing_address, customer_note)
		 VALUES ($1,$2,$3,'pending',$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		 RETURNING id, order_no, user_id, email, status, currency, subtotal_cents, shipping_cents,
		  discount_cents, tax_cents, total_cents, coupon_id, coupon_code, shipping_address::text, billing_address::text,
		  customer_note, paid_at, fulfilled_at, created_at, updated_at`

		stmt, err := session.Prepare(insertOrder)
		if err != nil {
			return err
		}
		defer stmt.Close()

		err = stmt.QueryRowCtx(ctx, order, orderNo, in.UserId, in.Email, in.Currency,
			in.SubtotalCents, in.ShippingCents, in.DiscountCents, in.TaxCents, in.TotalCents,
			in.CouponId, in.CouponCode, string(shippingJSON), string(billingJSON), in.CustomerNote)
		if err != nil {
			return err
		}

		for _, item := range in.Items {
			optJSON, _ := json.Marshal(item.Options)
			if item.Options == nil {
				optJSON = []byte("{}")
			}
			itemStmt, err := session.Prepare(
				`INSERT INTO orders.order_items
				 (order_id, product_id, variant_id, sku_code, title, options, image_url,
				  unit_price_cents, qty, total_cents)
				 VALUES ($1,
					(SELECT product_id FROM catalog.product_variants WHERE id=$2),
					$2,$3,$4,$5,$6,$7,$8,$9)`)
			if err != nil {
				return err
			}
			_, err = itemStmt.ExecCtx(ctx, order.Id, item.VariantId, item.SkuCode, item.Title,
				string(optJSON), item.ImageUrl, item.UnitPriceCents, item.Qty,
				item.UnitPriceCents*int64(item.Qty))
			itemStmt.Close()
			if err != nil {
				return err
			}
		}

		logStmt, err := session.Prepare(
			`INSERT INTO orders.order_status_logs (order_id, from_status, to_status, operator, note)
			 VALUES ($1,'','pending','system','order created')`)
		if err != nil {
			return err
		}
		defer logStmt.Close()
		_, err = logStmt.ExecCtx(ctx, order.Id)
		return err
	})

	if err != nil {
		return nil, err
	}
	return order, nil
}

func (r *OrderRepo) FindByOrderNo(ctx context.Context, orderNo string) (*model.OrderDetail, error) {
	query := `SELECT id, order_no, user_id, email, status, currency, subtotal_cents, shipping_cents,
		discount_cents, tax_cents, total_cents, coupon_id, coupon_code,
		shipping_address::text as shipping_address, billing_address::text as billing_address,
		customer_note, paid_at, fulfilled_at, created_at, updated_at
		FROM orders.orders WHERE order_no=$1 LIMIT 1`

	var detail model.OrderDetail
	if err := r.conn.QueryRowCtx(ctx, &detail, query, orderNo); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if err := r.loadRelations(ctx, &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

func (r *OrderRepo) FindById(ctx context.Context, id string) (*model.OrderDetail, error) {
	query := `SELECT id, order_no, user_id, email, status, currency, subtotal_cents, shipping_cents,
		discount_cents, tax_cents, total_cents, coupon_id, coupon_code,
		shipping_address::text as shipping_address, billing_address::text as billing_address,
		customer_note, paid_at, fulfilled_at, created_at, updated_at
		FROM orders.orders WHERE id=$1 LIMIT 1`

	var detail model.OrderDetail
	if err := r.conn.QueryRowCtx(ctx, &detail, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if err := r.loadRelations(ctx, &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

func (r *OrderRepo) loadRelations(ctx context.Context, detail *model.OrderDetail) error {
	items := []model.OrderItem{}
	if err := r.conn.QueryRowsCtx(ctx, &items,
		`SELECT id, order_id, product_id, variant_id, sku_code, title, options::text as options,
			image_url, unit_price_cents, qty, total_cents
		 FROM orders.order_items WHERE order_id=$1 ORDER BY id`, detail.Id); err != nil {
		return err
	}
	detail.Items = items

	var payment model.Payment
	if err := r.conn.QueryRowCtx(ctx, &payment,
		`SELECT id, order_id, provider, provider_intent_id, status, amount_cents, currency,
			raw::text as raw, created_at, updated_at
		 FROM payments.payments WHERE order_id=$1 ORDER BY created_at DESC LIMIT 1`, detail.Id); err == nil {
		detail.Payment = &payment
	}

	var shipment model.Shipment
	if err := r.conn.QueryRowCtx(ctx, &shipment,
		`SELECT id, order_id, carrier, tracking_no, tracking_url, status, shipped_at, delivered_at,
			created_at, updated_at
		 FROM fulfillment.shipments WHERE order_id=$1 ORDER BY created_at DESC LIMIT 1`, detail.Id); err == nil {
		detail.Shipment = &shipment
	}
	return nil
}

// UpdateStatus 更新订单状态并记录流转日志
func (r *OrderRepo) UpdateStatus(ctx context.Context, id, toStatus, operator, note string) error {
	return r.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		var from string
		if err := session.QueryRowCtx(ctx, &from,
			`SELECT status FROM orders.orders WHERE id=$1 FOR UPDATE`, id); err != nil {
			return err
		}
		if from == toStatus {
			return nil
		}

		extra := ""
		switch toStatus {
		case "paid":
			extra = ", paid_at = now()"
		case "fulfilled":
			extra = ", fulfilled_at = now()"
		}

		if _, err := session.ExecCtx(ctx,
			fmt.Sprintf(`UPDATE orders.orders SET status=$1, updated_at=now()%s WHERE id=$2`, extra),
			toStatus, id); err != nil {
			return err
		}

		stmt, err := session.Prepare(
			`INSERT INTO orders.order_status_logs (order_id, from_status, to_status, operator, note)
			 VALUES ($1,$2,$3,$4,$5)`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.ExecCtx(ctx, id, from, toStatus, operator, note)
		return err
	})
}

type AdminOrderFilter struct {
	Status  string
	Keyword string
	Page    int
	PageSize int
}

func (r *OrderRepo) AdminList(ctx context.Context, f AdminOrderFilter) ([]model.Order, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	idx := 1

	if f.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, f.Status)
		idx++
	}
	if f.Keyword != "" {
		where = append(where, fmt.Sprintf("(order_no ILIKE $%d OR email ILIKE $%d)", idx, idx))
		args = append(args, "%"+f.Keyword+"%")
		idx++
	}
	whereSQL := strings.Join(where, " AND ")

	var total int64
	if err := r.conn.QueryRowCtx(ctx, &total,
		fmt.Sprintf(`SELECT COUNT(*) FROM orders.orders WHERE %s`, whereSQL), args...); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT id, order_no, user_id, email, status, currency, subtotal_cents,
		shipping_cents, discount_cents, tax_cents, total_cents, coupon_id, coupon_code,
		shipping_address::text as shipping_address, billing_address::text as billing_address,
		customer_note, paid_at, fulfilled_at, created_at, updated_at
		FROM orders.orders WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		whereSQL, idx, idx+1)

	var list []model.Order
	if err := r.conn.QueryRowsCtx(ctx, &list, query,
		append(args, f.PageSize, (f.Page-1)*f.PageSize)...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ---------- 支付记录 ----------

func (r *OrderRepo) CreatePayment(ctx context.Context, orderId, provider, intentId string, amountCents int64, currency string) error {
	_, err := r.conn.ExecCtx(ctx,
		`INSERT INTO payments.payments (order_id, provider, provider_intent_id, status, amount_cents, currency)
		 VALUES ($1,$2,$3,'pending',$4,$5)
		 ON CONFLICT (provider, provider_intent_id) DO NOTHING`,
		orderId, provider, intentId, amountCents, currency)
	return err
}

func (r *OrderRepo) UpdatePaymentStatus(ctx context.Context, provider, intentId, status string) error {
	_, err := r.conn.ExecCtx(ctx,
		`UPDATE payments.payments SET status=$1, updated_at=now()
		 WHERE provider=$2 AND provider_intent_id=$3`, status, provider, intentId)
	return err
}

// TryMarkWebhook 幂等：同一 event_id 只处理一次
func (r *OrderRepo) TryMarkWebhook(ctx context.Context, provider, eventId, payload string) (bool, error) {
	var id string
	err := r.conn.QueryRowCtx(ctx, &id,
		`INSERT INTO payments.webhook_events (provider, event_id, payload)
		 VALUES ($1,$2,$3) ON CONFLICT (provider, event_id) DO NOTHING RETURNING id`,
		provider, eventId, payload)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // 已处理过
		}
		return false, err
	}
	return id != "", nil
}

// ---------- 履约 ----------

func (r *OrderRepo) CreateShipment(ctx context.Context, orderId, carrier, trackingNo, trackingUrl string) error {
	_, err := r.conn.ExecCtx(ctx,
		`INSERT INTO fulfillment.shipments (order_id, carrier, tracking_no, tracking_url, status)
		 VALUES ($1,$2,$3,$4,'shipped')`, orderId, carrier, trackingNo, trackingUrl)
	return err
}

func generateOrderNo() string {
	return time.Now().Format("20060102") + strings.ToUpper(password.RandomString(8))
}
