-- Drop indexes first
DROP INDEX IF EXISTS idx_idempotency_expires;
DROP INDEX IF EXISTS idx_saga_locks_expires;
DROP INDEX IF EXISTS idx_sagas_status;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS saga_locks;
DROP TABLE IF EXISTS saga_steps;
DROP TABLE IF EXISTS sagas;