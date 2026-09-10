-- ===========================================
-- 优惠券核销：订单关联优惠券，便于审计与取消时返还名额
-- ===========================================

ALTER TABLE orders.orders
  ADD COLUMN coupon_id   uuid REFERENCES promotion.coupons(id) ON DELETE SET NULL,
  ADD COLUMN coupon_code text NOT NULL DEFAULT '';

CREATE INDEX idx_orders_coupon ON orders.orders (coupon_id) WHERE coupon_id IS NOT NULL;

-- 示例优惠券：下单可直接测试（10% off，满 $20 可用，不限次数）
INSERT INTO promotion.coupons (code, type, value, min_amount_cents, max_uses, status)
VALUES ('WELCOME10', 'percent', 10, 2000, 0, 'active')
ON CONFLICT (code) DO NOTHING;
