-- ===========================================
-- 初始化：手账独立站数据库
-- 金额一律 bigint 存「分」，时间一律 timestamptz (UTC)
-- 按业务域分 schema，后期可原地拆库
-- ===========================================

CREATE SCHEMA IF NOT EXISTS admin;
CREATE SCHEMA IF NOT EXISTS catalog;
CREATE SCHEMA IF NOT EXISTS inventory;
CREATE SCHEMA IF NOT EXISTS orders;
CREATE SCHEMA IF NOT EXISTS payments;
CREATE SCHEMA IF NOT EXISTS fulfillment;
CREATE SCHEMA IF NOT EXISTS users;
CREATE SCHEMA IF NOT EXISTS promotion;

-- 通用更新时间触发器
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ===========================================
-- admin：后台账号
-- ===========================================
CREATE TABLE admin.admin_users (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email         text        NOT NULL,
  password_hash text        NOT NULL,
  name          text        NOT NULL DEFAULT '',
  role          text        NOT NULL DEFAULT 'operator', -- admin | operator
  status        text        NOT NULL DEFAULT 'active',   -- active | disabled
  last_login_at timestamptz,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT admin_users_email_lower UNIQUE (email)
);
CREATE UNIQUE INDEX idx_admin_users_email ON admin.admin_users (lower(email));
CREATE TRIGGER trg_admin_users_updated BEFORE UPDATE ON admin.admin_users
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ===========================================
-- catalog：商品域
-- ===========================================
CREATE TABLE catalog.categories (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  parent_id   uuid REFERENCES catalog.categories(id) ON DELETE SET NULL,
  name        text NOT NULL,
  slug        text NOT NULL UNIQUE,
  description text NOT NULL DEFAULT '',
  image_key   text NOT NULL DEFAULT '',
  sort_order  int  NOT NULL DEFAULT 0,
  status      text NOT NULL DEFAULT 'active', -- active | hidden
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_categories_parent ON catalog.categories (parent_id);
CREATE INDEX idx_categories_status_sort ON catalog.categories (status, sort_order, id);
CREATE TRIGGER trg_categories_updated BEFORE UPDATE ON catalog.categories
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE catalog.products (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  title        text NOT NULL,
  slug         text NOT NULL UNIQUE,
  subtitle     text NOT NULL DEFAULT '',
  description  text NOT NULL DEFAULT '',
  category_id  uuid REFERENCES catalog.categories(id) ON DELETE SET NULL,
  status       text NOT NULL DEFAULT 'draft', -- draft | published | archived
  price_cents  bigint NOT NULL DEFAULT 0,   -- 展示用最低价，实际价格以 variant 为准
  currency     char(3) NOT NULL DEFAULT 'USD',
  -- 手账特有属性：cover_material / size / page_count / paper_weight / binding 等
  attributes   jsonb   NOT NULL DEFAULT '{}'::jsonb,
  tags         text[]  NOT NULL DEFAULT '{}',
  seo_title       text NOT NULL DEFAULT '',
  seo_description text NOT NULL DEFAULT '',
  published_at timestamptz,
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now(),
  -- 英文全文检索（面向海外），避免早期引入 ES
  search_tsv   tsvector GENERATED ALWAYS AS (
    setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(subtitle, '')), 'B') ||
    setweight(to_tsvector('english', coalesce(description, '')), 'C')
  ) STORED
);
CREATE INDEX idx_products_status_time ON catalog.products (status, published_at DESC, id DESC);
CREATE INDEX idx_products_category   ON catalog.products (category_id) WHERE category_id IS NOT NULL;
CREATE INDEX idx_products_attributes ON catalog.products USING GIN (attributes);
CREATE INDEX idx_products_tags       ON catalog.products USING GIN (tags);
CREATE INDEX idx_products_search     ON catalog.products USING GIN (search_tsv);
CREATE TRIGGER trg_products_updated BEFORE UPDATE ON catalog.products
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- SKU 变体：封面/尺寸/纸张等规格组合，价格与库存在此
CREATE TABLE catalog.product_variants (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  product_id       uuid NOT NULL REFERENCES catalog.products(id) ON DELETE CASCADE,
  sku_code         text NOT NULL UNIQUE,
  title            text NOT NULL DEFAULT '',
  options          jsonb NOT NULL DEFAULT '{}'::jsonb, -- {"cover":"布面","size":"A5"}
  price_cents      bigint NOT NULL DEFAULT 0,
  compare_at_cents bigint NOT NULL DEFAULT 0,  -- 划线价
  weight_g         int    NOT NULL DEFAULT 0,  -- 计重用
  image_key        text   NOT NULL DEFAULT '',
  status           text   NOT NULL DEFAULT 'active', -- active | disabled
  sort_order       int    NOT NULL DEFAULT 0,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_variants_product ON catalog.product_variants (product_id, sort_order, id);
CREATE TRIGGER trg_variants_updated BEFORE UPDATE ON catalog.product_variants
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE catalog.product_images (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  product_id uuid NOT NULL REFERENCES catalog.products(id) ON DELETE CASCADE,
  variant_id uuid REFERENCES catalog.product_variants(id) ON DELETE CASCADE,
  object_key text NOT NULL,      -- R2 中的 key
  alt        text NOT NULL DEFAULT '',
  sort_order int  NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_product_images_product ON catalog.product_images (product_id, sort_order, id);

-- ===========================================
-- inventory：库存域（独立，便于后期优先拆出）
-- ===========================================
CREATE TABLE inventory.inventory_items (
  variant_id  uuid PRIMARY KEY REFERENCES catalog.product_variants(id) ON DELETE CASCADE,
  available   int NOT NULL DEFAULT 0 CHECK (available >= 0),
  reserved    int NOT NULL DEFAULT 0 CHECK (reserved >= 0),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

-- 库存预占（下单时冻结，支付成功扣减，超时释放）
CREATE TABLE inventory.stock_reservations (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  variant_id uuid NOT NULL REFERENCES catalog.product_variants(id) ON DELETE CASCADE,
  qty        int  NOT NULL CHECK (qty > 0),
  ref_type   text NOT NULL DEFAULT 'cart', -- cart | order
  ref_id     text NOT NULL DEFAULT '',
  status     text NOT NULL DEFAULT 'held', -- held | released | committed
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_reservations_variant ON inventory.stock_reservations (variant_id, status);
CREATE INDEX idx_reservations_expire  ON inventory.stock_reservations (expires_at) WHERE status = 'held';

-- ===========================================
-- users：前台会员（支持访客下单，故大量字段可空）
-- ===========================================
CREATE TABLE users.users (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email         text NOT NULL,
  password_hash text NOT NULL DEFAULT '',
  name          text NOT NULL DEFAULT '',
  status        text NOT NULL DEFAULT 'active',
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_users_email ON users.users (lower(email));
CREATE TRIGGER trg_users_updated BEFORE UPDATE ON users.users
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE users.addresses (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     uuid NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
  recipient   text NOT NULL DEFAULT '',
  phone       text NOT NULL DEFAULT '',
  country     char(2) NOT NULL DEFAULT '',  -- ISO 3166-1 alpha-2
  state       text NOT NULL DEFAULT '',
  city        text NOT NULL DEFAULT '',
  address1    text NOT NULL DEFAULT '',
  address2    text NOT NULL DEFAULT '',
  postal_code text NOT NULL DEFAULT '',
  is_default  boolean NOT NULL DEFAULT false,
  created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_addresses_user ON users.addresses (user_id);

-- ===========================================
-- orders：交易域（价格全部快照，改价不影响历史订单）
-- ===========================================
CREATE TABLE orders.orders (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  order_no          text NOT NULL UNIQUE,
  user_id           uuid REFERENCES users.users(id) ON DELETE SET NULL,
  email             text NOT NULL DEFAULT '',
  status            text NOT NULL DEFAULT 'pending', -- pending|paid|fulfilled|cancelled|refunded
  currency          char(3)  NOT NULL DEFAULT 'USD',
  subtotal_cents    bigint   NOT NULL DEFAULT 0,
  shipping_cents    bigint   NOT NULL DEFAULT 0,
  discount_cents    bigint   NOT NULL DEFAULT 0,
  tax_cents         bigint   NOT NULL DEFAULT 0,
  total_cents       bigint   NOT NULL DEFAULT 0,
  shipping_address  jsonb    NOT NULL DEFAULT '{}'::jsonb,
  billing_address   jsonb    NOT NULL DEFAULT '{}'::jsonb,
  customer_note     text     NOT NULL DEFAULT '',
  paid_at           timestamptz,
  fulfilled_at      timestamptz,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_orders_status_time ON orders.orders (status, created_at DESC);
CREATE INDEX idx_orders_email       ON orders.orders (lower(email));
CREATE INDEX idx_orders_user        ON orders.orders (user_id) WHERE user_id IS NOT NULL;
CREATE TRIGGER trg_orders_updated BEFORE UPDATE ON orders.orders
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE orders.order_items (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id         uuid NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
  product_id       uuid,            -- 不设外键：商品可删，订单需保留
  variant_id       uuid,
  sku_code         text NOT NULL DEFAULT '',
  title            text NOT NULL DEFAULT '',       -- 快照
  options          jsonb NOT NULL DEFAULT '{}'::jsonb, -- 快照
  image_url        text  NOT NULL DEFAULT '',      -- 快照
  unit_price_cents bigint NOT NULL DEFAULT 0,      -- 快照
  qty              int    NOT NULL DEFAULT 1 CHECK (qty > 0),
  total_cents      bigint NOT NULL DEFAULT 0
);
CREATE INDEX idx_order_items_order ON orders.order_items (order_id);

CREATE TABLE orders.order_status_logs (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id    uuid NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
  from_status text NOT NULL DEFAULT '',
  to_status   text NOT NULL,
  operator    text NOT NULL DEFAULT 'system',
  note        text NOT NULL DEFAULT '',
  created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_order_logs_order ON orders.order_status_logs (order_id, created_at);

-- ===========================================
-- payments：支付域
-- ===========================================
CREATE TABLE payments.payments (
  id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id           uuid NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
  provider           text NOT NULL DEFAULT 'stripe', -- stripe | paypal
  provider_intent_id text NOT NULL DEFAULT '',
  status             text NOT NULL DEFAULT 'pending', -- pending|succeeded|failed|refunded
  amount_cents       bigint NOT NULL DEFAULT 0,
  currency           char(3) NOT NULL DEFAULT 'USD',
  raw                jsonb  NOT NULL DEFAULT '{}'::jsonb,
  created_at         timestamptz NOT NULL DEFAULT now(),
  updated_at         timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_payments_order ON payments.payments (order_id);
CREATE UNIQUE INDEX idx_payments_intent ON payments.payments (provider, provider_intent_id)
  WHERE provider_intent_id <> '';
CREATE TRIGGER trg_payments_updated BEFORE UPDATE ON payments.payments
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 支付回调幂等：同一 event_id 只处理一次
CREATE TABLE payments.webhook_events (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  provider     text NOT NULL,
  event_id     text NOT NULL,
  payload      jsonb NOT NULL DEFAULT '{}'::jsonb,
  processed_at timestamptz,
  created_at   timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT webhook_events_unique UNIQUE (provider, event_id)
);

-- ===========================================
-- fulfillment：履约域
-- ===========================================
CREATE TABLE fulfillment.shipping_rules (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name            text NOT NULL DEFAULT '',
  country_codes   text[] NOT NULL DEFAULT '{}', -- 空数组表示其余所有国家
  min_amount_cents bigint NOT NULL DEFAULT 0,
  max_weight_g    int  NOT NULL DEFAULT 0,      -- 0 表示不限制
  price_cents     bigint NOT NULL DEFAULT 0,
  free_threshold_cents bigint NOT NULL DEFAULT 0, -- 满额包邮阈值，0 表示不包邮
  sort_order      int  NOT NULL DEFAULT 0,
  status          text NOT NULL DEFAULT 'active',
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE fulfillment.shipments (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id     uuid NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
  carrier      text NOT NULL DEFAULT '',  -- USPS | UPS | YunExpress ...
  tracking_no  text NOT NULL DEFAULT '',
  tracking_url text NOT NULL DEFAULT '',
  status       text NOT NULL DEFAULT 'pending', -- pending|shipped|delivered
  shipped_at   timestamptz,
  delivered_at timestamptz,
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_shipments_order ON fulfillment.shipments (order_id);
CREATE TRIGGER trg_shipments_updated BEFORE UPDATE ON fulfillment.shipments
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ===========================================
-- promotion：营销域
-- ===========================================
CREATE TABLE promotion.coupons (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code             text NOT NULL UNIQUE,
  type             text NOT NULL DEFAULT 'percent', -- percent | fixed
  value            bigint NOT NULL DEFAULT 0,  -- percent: 百分比(10=9折) fixed: 分
  min_amount_cents bigint NOT NULL DEFAULT 0,
  max_uses         int NOT NULL DEFAULT 0,     -- 0 表示不限
  used_count       int NOT NULL DEFAULT 0,
  starts_at        timestamptz,
  ends_at          timestamptz,
  status           text NOT NULL DEFAULT 'active',
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_coupons_updated BEFORE UPDATE ON promotion.coupons
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
