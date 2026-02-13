package repository

import (
	"context"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/internal/domain/entity"
)

// ReservationRepository defines WHAT we need (not HOW)
// Analogy: This is like a job description — "We need someone who can:
//   - Create reservations
//   - Check if a request was already processed"
//   - Get reservations by ID
//   - Get reservations by Order ID
//   - Update reservations
//
// But we don't specify if they use PostgreSQL, MongoDB, or a notebook.
type ReservationRepository interface {
	Create(ctx context.Context, reservation *entity.Reservation, idempotencyKey string) error
	CheckIdempotency(ctx context.Context, key string) (*entity.Reservation, error)
	GetByOrderID(ctx context.Context, orderID string) (*entity.Reservation, error)
	Update(ctx context.Context, reservation *entity.Reservation) error
}
