-- Payments table
CREATE TABLE payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        VARCHAR(100) NOT NULL,
    customer_id     VARCHAR(100) NOT NULL,
    amount          DECIMAL(10, 2) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_status CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'REFUNDED'))
);

-- Idempotency for payment operations
CREATE TABLE payment_idempotency (
    key             VARCHAR(255) PRIMARY KEY,
    payment_id      UUID NOT NULL REFERENCES payments(id),
    operation       VARCHAR(50) NOT NULL,
    response        JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_payments_order ON payments(order_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payment_idempotency_expires ON payment_idempotency(expires_at);
