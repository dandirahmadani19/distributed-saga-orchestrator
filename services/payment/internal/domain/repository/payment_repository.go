package repository

import (
	"context"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/domain/entity"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *entity.Payment, idempotencyKey string) error
	// GetByID(ctx context.Context, id string) (*entity.Payment, error)
	// Update(ctx context.Context, payment *entity.Payment) error
	CheckIdempotency(ctx context.Context, key string) (*entity.Payment, error)
}
