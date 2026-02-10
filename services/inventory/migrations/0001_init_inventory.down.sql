DROP INDEX IF EXISTS idx_inventory_idempotency_expires;
DROP INDEX IF EXISTS idx_reservations_status;
DROP INDEX IF EXISTS idx_reservations_order;
DROP TABLE IF EXISTS inventory_idempotency;
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS inventory;