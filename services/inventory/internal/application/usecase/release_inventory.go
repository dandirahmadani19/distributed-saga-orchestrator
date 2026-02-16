package usecase

import (
	"context"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/logger"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/internal/application/dto"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/internal/domain/entity"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/internal/domain/repository"
)

// ReleaseInventoryUseCase handles the business logic of releasing reserved items
// Analogy: This is what happens when a customer CANCELS their order.
// The warehouse manager tells the workers: "Put those items back on the shelf."
type ReleaseInventoryUseCase struct {
	repo   repository.ReservationRepository
	logger *logger.Logger
}

func NewReleaseInventoryUseCase(repo repository.ReservationRepository, log *logger.Logger) *ReleaseInventoryUseCase {
	return &ReleaseInventoryUseCase{repo: repo, logger: log}
}

func (uc *ReleaseInventoryUseCase) Execute(ctx context.Context, req dto.ReleaseInventoryRequest) (*dto.ReservationResponse, error) {
	// 1. Check idempotency
	existing, err := uc.repo.CheckIdempotency(ctx, req.IdempotencyKey)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		uc.logger.Info().Str("order_id", req.OrderID).Msg("Returning existing release response")
		return uc.toDto(existing), nil
	}

	// 2. Find reservation by order ID
	// Unlike Refund (which uses payment_id), Release uses order_id
	// because the orchestrator knows the order_id, not the reservation_id
	reservation, err := uc.repo.GetByOrderID(ctx, req.OrderID)
	if err != nil {
		return nil, err
	}

	// 3. Release the reservation
	reservation.Release()

	// 4. Save to DB
	if err := uc.repo.Update(ctx, reservation); err != nil {
		return nil, err
	}

	return uc.toDto(reservation), nil
}

func (uc *ReleaseInventoryUseCase) toDto(res *entity.Reservation) *dto.ReservationResponse {
	return &dto.ReservationResponse{
		ID:        res.ID,
		OrderID:   res.OrderID,
		Status:    string(res.Status),
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
	}
}
