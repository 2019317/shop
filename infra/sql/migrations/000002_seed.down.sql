-- 回滚：000002_seed
DELETE FROM fulfillment.shipping_rules WHERE name = 'Standard International';
DELETE FROM catalog.categories WHERE slug IN (
  'journals-notebooks','planner-inserts','stickers-washi','pens-writing','paper-goods','accessories'
);
