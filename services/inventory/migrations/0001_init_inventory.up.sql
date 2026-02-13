-- Reservations table
CREATE TABLE reservations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        VARCHAR(100) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'RESERVED',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_reservation_status CHECK (status IN ('RESERVED', 'RELEASED', 'CONFIRMED'))
);

-- Reservation items table
CREATE TABLE reservation_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reservation_id  UUID NOT NULL REFERENCES reservations(id),
    product_id      VARCHAR(100) NOT NULL,
    quantity        INT NOT NULL,

    CONSTRAINT positive_quantity CHECK (quantity > 0)
);

-- Idempotency for reservation operations
CREATE TABLE reservation_idempotency (
    key             VARCHAR(255) PRIMARY KEY,
    reservation_id  UUID NOT NULL REFERENCES reservations(id),
    operation       VARCHAR(50) NOT NULL,
    response        JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_reservations_order ON reservations(order_id);
CREATE INDEX idx_reservations_status ON reservations(status);
CREATE INDEX idx_reservation_items_reservation ON reservation_items(reservation_id);
CREATE INDEX idx_reservation_idempotency_expires ON reservation_idempotency(expires_at);
