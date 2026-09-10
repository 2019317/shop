-- 回滚：000003_i18n
DROP INDEX IF EXISTS catalog.idx_products_title_trgm;
DROP TABLE IF EXISTS catalog.category_translations;
DROP TABLE IF EXISTS catalog.product_translations;
