-- Remove audit logs
DROP TABLE IF EXISTS audit_logs;

-- Remove API rate limits
DROP TABLE IF EXISTS api_rate_limits;

-- Remove indexes
DROP INDEX IF EXISTS idx_product_barcodes_product_id;
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_orders_deleted_at;
DROP INDEX IF EXISTS idx_products_deleted_at;

-- Remove soft delete columns
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE orders DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE products DROP COLUMN IF EXISTS deleted_at;

-- Remove created_at from product_barcodes
ALTER TABLE product_barcodes DROP COLUMN IF EXISTS created_at;