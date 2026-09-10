-- ===========================================
-- 种子数据：手账制品常用类目 + 默认运费规则
-- ===========================================

INSERT INTO catalog.categories (name, slug, description, sort_order) VALUES
  ('Journals & Notebooks', 'journals-notebooks', 'Handcrafted journals, dot grid and blank notebooks', 10),
  ('Planner Inserts',      'planner-inserts',    'Refill inserts for A5/A6/personal size planners', 20),
  ('Stickers & Washi',     'stickers-washi',     'Original illustration stickers and washi tapes', 30),
  ('Pens & Writing',       'pens-writing',       'Fountain pens, gel pens and refills', 40),
  ('Paper Goods',          'paper-goods',        'Letter paper, cards and memo pads', 50),
  ('Accessories',          'accessories',        'Bookmarks, charms and storage', 60)
ON CONFLICT (slug) DO NOTHING;

-- 默认运费规则：满 59 美元包邮，否则 8.9 美元
INSERT INTO fulfillment.shipping_rules
  (name, country_codes, min_amount_cents, max_weight_g, price_cents, free_threshold_cents, sort_order)
VALUES
  ('Standard International', '{}', 0, 0, 890, 5900, 10)
ON CONFLICT DO NOTHING;
