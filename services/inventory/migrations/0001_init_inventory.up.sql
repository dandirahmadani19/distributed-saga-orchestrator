-- Inventory items
CREATE TABLE inventory (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      VARCHAR(100) NOT NULL UNIQUE,
    quantity        INT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    reserved        INT NOT NULL DEFAULT 0 CHECK (reserved >= 0),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT quantity_check CHECK (reserved <= quantity)
);

-- Reservations
CREATE TABLE reservations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        VARCHAR(100) NOT NULL,
    product_id      VARCHAR(100) NOT NULL,
    quantity        INT NOT NULL CHECK (quantity > 0),
    status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    released_at     TIMESTAMPTZ,

    CONSTRAINT valid_status CHECK (status IN ('ACTIVE', 'RELEASED'))
);

-- Idempotency for inventory operations
CREATE TABLE inventory_idempotency (
    key             VARCHAR(255) PRIMARY KEY,
    reservation_id  UUID REFERENCES reservations(id),
    operation       VARCHAR(50) NOT NULL,
    response        JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL
);

-- Insert sample inventory items
INSERT INTO inventory (product_id, quantity) VALUES
    ('PROD-001', 100),
    ('PROD-002', 50),
    ('PROD-003', 200);

CREATE INDEX idx_reservations_order ON reservations(order_id);
CREATE INDEX idx_reservations_status ON reservations(status);
CREATE INDEX idx_inventory_idempotency_expires ON inventory_idempotency(expires_at);
