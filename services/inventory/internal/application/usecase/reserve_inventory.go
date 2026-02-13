package usecase

import (
	"context"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/logger"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/internal/application/dto"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/internal/domain/entity"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/internal/domain/repository"
)

type ReserveInventoryUseCase struct {
	repo   repository.ReservationRepository
	logger *logger.Logger
}

func NewReserveInventoryUseCase(repo repository.ReservationRepository, log *logger.Logger) *ReserveInventoryUseCase {
	return &ReserveInventoryUseCase{repo: repo, logger: log}
}

func (uc *ReserveInventoryUseCase) Execute(ctx context.Context, req dto.ReserveInventoryRequest) (*dto.ReservationResponse, error) {
	// 1. Check idempotency
	// Analogy: "Did we already reserve items for this exact request?"
	// Like a bouncer checking if you already have a wristband — no need to give another one.
	existing, err := uc.repo.CheckIdempotency(ctx, req.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		uc.logger.Info().Str("key", req.IdempotencyKey).Msg("Returning idempotent response")
		return uc.toDto(existing), nil
	}

	// 2. Convert DTO items to domain items
	items := make([]entity.ReservationItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = entity.ReservationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// 3. Create new reservation
	reservation := entity.NewReservation(req.OrderID, items)

	// 4. Save to DB
	if err := uc.repo.Create(ctx, reservation, req.IdempotencyKey); err != nil {
		return nil, err
	}
	return uc.toDto(reservation), nil
}

func (uc *ReserveInventoryUseCase) toDto(reservation *entity.Reservation) *dto.ReservationResponse {
	return &dto.ReservationResponse{
		ID:        reservation.ID,
		OrderID:   reservation.OrderID,
		Status:    string(reservation.Status),
		CreatedAt: reservation.CreatedAt,
		UpdatedAt: reservation.UpdatedAt,
	}
}
