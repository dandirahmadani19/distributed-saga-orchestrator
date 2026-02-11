package repository

import (
	"context"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/domain/entity"
)

type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order, idempotencyKey string) (*entity.Order, error)
	FindByID(ctx context.Context, id string) (*entity.Order, error)
	Update(ctx context.Context, order *entity.Order) error
	CheckIdempotency(ctx context.Context, key string) (*entity.Order, error)
}
