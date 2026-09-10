DROP INDEX IF EXISTS idx_orders_coupon;
ALTER TABLE orders.orders DROP COLUMN IF EXISTS coupon_id;
ALTER TABLE orders.orders DROP COLUMN IF EXISTS coupon_code;
