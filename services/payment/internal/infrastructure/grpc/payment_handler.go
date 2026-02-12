package grpc

import (
	"context"

	pb "github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/gen/proto/payment/v1"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/application/dto"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/application/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentHandler struct {
	pb.UnimplementedPaymentServiceServer
	processUC *usecase.ProcessPaymentUseCase
}

func NewPaymentHandler(processUC *usecase.ProcessPaymentUseCase) *PaymentHandler {
	return &PaymentHandler{processUC: processUC}
}

func (h *PaymentHandler) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	payment, err := h.processUC.Execute(ctx, dto.CreatePaymentRequest{
		IdempotencyKey: req.IdempotencyKey,
		CustomerID:     req.CustomerId,
		OrderID:        req.OrderId,
		Amount:         req.Amount,
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ProcessPaymentResponse{
		PaymentId: payment.ID,
		Status:    payment.Status,
	}, nil
}
