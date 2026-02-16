package usecase

import (
	"context"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/logger"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/application/dto"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/domain/entity"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/domain/repository"
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
		return nil, err
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
		ID:         payment.ID,
		OrderID:    payment.OrderID,
		CustomerID: payment.CustomerID,
		Amount:     payment.Amount,
		Status:     string(payment.Status),
		CreatedAt:  payment.CreatedAt,
		UpdatedAt:  payment.UpdatedAt,
	}
}
