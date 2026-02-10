-- Orders table
CREATE TABLE orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id     VARCHAR(100) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'CREATED',
    total_amount    DECIMAL(10, 2) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_status CHECK (status IN ('CREATED', 'CONFIRMED', 'CANCELLED'))
);

CREATE TABLE order_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id      VARCHAR(100) NOT NULL,
    quantity        INT NOT NULL CHECK (quantity > 0),
    price           DECIMAL(10, 2) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Idempotency for order operations
CREATE TABLE order_idempotency (
    key             VARCHAR(255) PRIMARY KEY,
    order_id        UUID NOT NULL REFERENCES orders(id),
    operation       VARCHAR(50) NOT NULL,
    response        JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_order_idempotency_expires ON order_idempotency(expires_at);
