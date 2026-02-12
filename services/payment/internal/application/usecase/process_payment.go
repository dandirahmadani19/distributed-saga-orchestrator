package usecase

import (
	"context"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/application/dto"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/domain/entity"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/domain/repository"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/shared/pkg/logger"
)

type ProcessPaymentUseCase struct {
	repo   repository.PaymentRepository
	logger *logger.Logger
}

func NewProcessPaymentUseCase(repo repository.PaymentRepository, log *logger.Logger) *ProcessPaymentUseCase {
	return &ProcessPaymentUseCase{repo: repo, logger: log}
}

func (uc *ProcessPaymentUseCase) Execute(ctx context.Context, req dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
	// 1. Check idempotency
	existing, err := uc.repo.CheckIdempotency(ctx, req.IdempotencyKey)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		uc.logger.Info().Str("key", req.IdempotencyKey).Msg("Returning idempotent response")
		return uc.toDto(existing), nil
	}

	// 2. Create new payment
	payment := entity.NewPayment(req.OrderID, req.CustomerID, req.Amount)

	// 3. Save to DB
	if err := uc.repo.Create(ctx, payment, req.IdempotencyKey); err != nil {
		return nil, err
	}

	return uc.toDto(payment), nil
}

func (uc *ProcessPaymentUseCase) toDto(payment *entity.Payment) *dto.PaymentResponse {
	return &dto.PaymentResponse{
		ID:         payment.ID,
		OrderID:    payment.OrderID,
		CustomerID: payment.CustomerID,
		Amount:     payment.Amount,
		Status:     string(payment.Status),
		CreatedAt:  payment.CreatedAt,
		UpdatedAt:  payment.UpdatedAt,
	}
}
