DROP INDEX IF EXISTS idx_order_idempotency_expires;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_customer;
DROP TABLE IF EXISTS order_idempotency;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
