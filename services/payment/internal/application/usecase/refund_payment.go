package usecase

import (
	"context"
	"fmt"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/application/dto"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/domain/entity"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/domain/repository"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/shared/pkg/logger"
)

type RefundPaymentUseCase struct {
	repo   repository.PaymentRepository
	logger *logger.Logger
}

func NewRefundPaymentUseCase(repo repository.PaymentRepository, log *logger.Logger) *RefundPaymentUseCase {
	return &RefundPaymentUseCase{repo: repo, logger: log}
}

func (uc *RefundPaymentUseCase) Execute(ctx context.Context, req dto.RefundPaymentRequest) (*dto.PaymentResponse, error) {
	// 1. Check idempotency
	existing, err := uc.repo.CheckIdempotency(ctx, req.IdempotencyKey)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		uc.logger.Info().Str("payment_id", req.PaymentID).Msg("Returning existing refund payment")
		return uc.toDto(existing), nil
	}

	// 2. Get payment
	payment, err := uc.repo.GetByID(ctx, req.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// 3. Refund payment
	payment.Refund()

	// 4. Save to DB
	if err := uc.repo.Update(ctx, payment); err != nil {
		return nil, err
	}

	return uc.toDto(payment), nil
}

func (uc *RefundPaymentUseCase) toDto(payment *entity.Payment) *dto.PaymentResponse {
	return &dto.PaymentResponse{
		ID:     payment.ID,
		Status: string(payment.Status),
	}
}
