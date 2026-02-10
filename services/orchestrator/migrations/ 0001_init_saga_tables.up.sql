-- Saga definitions
CREATE TABLE sagas (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_type       VARCHAR(100) NOT NULL,          -- e.g., 'order_saga'
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    payload         JSONB NOT NULL,                 -- saga input data
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,

    CONSTRAINT valid_status CHECK (status IN (
        'PENDING', 'EXECUTING', 'COMPLETED',
        'COMPENSATING', 'COMPENSATED', 'FAILED'
    ))
);

-- Individual saga steps
CREATE TABLE saga_steps (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_id         UUID NOT NULL REFERENCES sagas(id) ON DELETE CASCADE,
    step_name       VARCHAR(100) NOT NULL,          -- e.g., 'create_order'
    step_order      INT NOT NULL,                   -- execution order (1, 2, 3...)
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    idempotency_key UUID NOT NULL,                  -- for retry safety
    request_payload JSONB,
    response_payload JSONB,
    error_message   TEXT,
    executed_at     TIMESTAMPTZ,
    compensated_at  TIMESTAMPTZ,
    retry_count     INT NOT NULL DEFAULT 0,

    CONSTRAINT valid_step_status CHECK (status IN (
        'PENDING', 'EXECUTING', 'SUCCEEDED', 'FAILED',
        'COMPENSATING', 'COMPENSATED', 'COMPENSATION_FAILED'
    )),
    UNIQUE(saga_id, step_order)
);

-- Distributed lock for saga processing
CREATE TABLE saga_locks (
    saga_id         UUID PRIMARY KEY REFERENCES sagas(id) ON DELETE CASCADE,
    owner_id        VARCHAR(100) NOT NULL,          -- orchestrator instance ID
    acquired_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,           -- lease expiration

    CONSTRAINT valid_lease CHECK (expires_at > acquired_at)
);

-- Idempotency records (to prevent duplicate saga creation)
CREATE TABLE idempotency_keys (
    key             VARCHAR(255) PRIMARY KEY,
    saga_id         UUID NOT NULL REFERENCES sagas(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL            -- cleanup after TTL
);

-- Indexes for performance
CREATE INDEX idx_sagas_status ON sagas(status) WHERE status IN ('PENDING', 'EXECUTING', 'COMPENSATING');
CREATE INDEX idx_saga_locks_expires ON saga_locks(expires_at);
CREATE INDEX idx_idempotency_expires ON idempotency_keys(expires_at);