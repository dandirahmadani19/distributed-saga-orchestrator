package usecase

import (
	"context"
	"fmt"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/application/dto"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/domain/entity"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/domain/repository"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/shared/pkg/logger"
)

type CancelOrderUseCase struct {
	repo   repository.OrderRepository
	logger *logger.Logger
}

func NewCancelOrderUseCase(repo repository.OrderRepository, logger *logger.Logger) *CancelOrderUseCase {
	return &CancelOrderUseCase{
		repo:   repo,
		logger: logger,
	}
}

func (uc *CancelOrderUseCase) Execute(ctx context.Context, orderID string, idempotencyKey string) (*dto.OrderResponse, error) {
	// Check idempotency
	existingOrder, err := uc.repo.CheckIdempotency(ctx, idempotencyKey)
	if err == nil && existingOrder != nil {
		uc.logger.Info().
			Str("order_id", orderID).
			Msg("Returning existing cancellation (idempoten)")
		return uc.toDTO(existingOrder), nil
	}

	// Find order
	order, err := uc.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Cancel it
	order.Cancel()

	// Persist
	if err := uc.repo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	uc.logger.Info().
		Str("order_id", orderID).
		Msg("Order cancelled successfully")

	return uc.toDTO(order), nil
}
func (uc *CancelOrderUseCase) toDTO(order *entity.Order) *dto.OrderResponse {
	return &dto.OrderResponse{
		ID:          order.ID,
		CustomerID:  order.CustomerID,
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount,
		CreatedAt:   order.CreatedAt,
	}
}
