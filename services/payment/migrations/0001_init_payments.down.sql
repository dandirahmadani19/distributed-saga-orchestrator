DROP INDEX IF EXISTS idx_payment_idempotency_expires;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_order;
DROP TABLE IF EXISTS payment_idempotency;
DROP TABLE IF EXISTS payments;
