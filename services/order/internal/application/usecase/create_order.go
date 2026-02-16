package usecase

import (
	"context"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/logger"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/application/dto"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/domain/entity"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/domain/repository"
)

// CreateOrderUseCase handles order creation logic
type CreateOrderUseCase struct {
	repo   repository.OrderRepository
	logger *logger.Logger
}

// NewCreateOrderUseCase creates a new use case
func NewCreateOrderUseCase(repo repository.OrderRepository, logger *logger.Logger) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		repo:   repo,
		logger: logger,
	}
}

// Execute runs the use case
func (uc *CreateOrderUseCase) Execute(ctx context.Context, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	// Check idempotency first
	existingOrder, err := uc.repo.CheckIdempotency(ctx, req.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	if existingOrder != nil {
		uc.logger.InfoWithTrace(ctx).
			Str("order_id", existingOrder.ID).
			Msg("Returning existing order (idempotent)")

		return uc.toDTO(existingOrder), nil
	}

	// Convert DTO items to domain items
	items := make([]entity.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = entity.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	// Create order using domain factory
	order, err := entity.NewOrder(req.CustomerID, items, req.TotalAmount)
	if err != nil {
		return nil, err
	}

	// Persist order
	createdOrder, err := uc.repo.Create(ctx, order, req.IdempotencyKey)
	if err != nil {
		return nil, err
	}

	uc.logger.InfoWithTrace(ctx).
		Str("order_id", createdOrder.ID).
		Str("customer_id", createdOrder.CustomerID).
		Msg("Order created successfully")

	return uc.toDTO(createdOrder), nil
}

// toDTO converts domain entity to DTO
func (uc *CreateOrderUseCase) toDTO(order *entity.Order) *dto.OrderResponse {
	return &dto.OrderResponse{
		ID:          order.ID,
		CustomerID:  order.CustomerID,
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount,
		CreatedAt:   order.CreatedAt,
	}
}
