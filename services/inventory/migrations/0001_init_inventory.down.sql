DROP INDEX IF EXISTS idx_reservation_idempotency_expires;
DROP INDEX IF EXISTS idx_reservation_items_reservation;
DROP INDEX IF EXISTS idx_reservations_status;
DROP INDEX IF EXISTS idx_reservations_order;
DROP TABLE IF EXISTS reservation_idempotency;
DROP TABLE IF EXISTS reservation_items;
DROP TABLE IF EXISTS reservations;
