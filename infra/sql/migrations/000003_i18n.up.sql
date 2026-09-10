-- ===========================================
-- 多语言支持：商品与类目翻译表
-- 主表字段保留为默认语言（en）内容，翻译表按 locale 覆盖
-- 查询时用 COALESCE(NULLIF(t.title,''), p.title) 实现「缺失回落」
-- ===========================================

CREATE TABLE catalog.product_translations (
  product_id       uuid NOT NULL REFERENCES catalog.products(id) ON DELETE CASCADE,
  locale           varchar(8) NOT NULL,
  title            text NOT NULL DEFAULT '',
  subtitle         text NOT NULL DEFAULT '',
  description      text NOT NULL DEFAULT '',
  seo_title        text NOT NULL DEFAULT '',
  seo_description  text NOT NULL DEFAULT '',
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (product_id, locale)
);
CREATE INDEX idx_product_translations_locale ON catalog.product_translations (locale);
CREATE TRIGGER trg_product_translations_updated BEFORE UPDATE ON catalog.product_translations
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE catalog.category_translations (
  category_id  uuid NOT NULL REFERENCES catalog.categories(id) ON DELETE CASCADE,
  locale       varchar(8) NOT NULL,
  name         text NOT NULL DEFAULT '',
  description  text NOT NULL DEFAULT '',
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (category_id, locale)
);
CREATE INDEX idx_category_translations_locale ON catalog.category_translations (locale);

-- 中文检索：英文全文检索对中文无效，增加标题的 trigram 索引以支持模糊匹配
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_products_title_trgm ON catalog.products USING GIN (title gin_trgm_ops);
