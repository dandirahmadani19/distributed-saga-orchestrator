package grpc

import (
	"context"

	pb "github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/gen/proto/payment/v1"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/application/dto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProccessPayment interface {
	Execute(ctx context.Context, req dto.CreatePaymentRequest) (*dto.PaymentResponse, error)
}
type RefundPayment interface {
	Execute(ctx context.Context, req dto.RefundPaymentRequest) (*dto.PaymentResponse, error)
}

type PaymentHandler struct {
	pb.UnimplementedPaymentServiceServer
	processUC ProccessPayment
	refundUC  RefundPayment
}

func (h *PaymentHandler) RegisterPaymentServiceServer(s *grpc.Server) {
	pb.RegisterPaymentServiceServer(s, h)
}

func NewPaymentHandler(processUC ProccessPayment, refundUC RefundPayment) *PaymentHandler {
	return &PaymentHandler{processUC: processUC, refundUC: refundUC}
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

func (h *PaymentHandler) RefundPayment(ctx context.Context, req *pb.RefundPaymentRequest) (*pb.RefundPaymentResponse, error) {
	payment, err := h.refundUC.Execute(ctx, dto.RefundPaymentRequest{
		IdempotencyKey: req.IdempotencyKey,
		PaymentID:      req.PaymentId,
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RefundPaymentResponse{
		PaymentId: payment.ID,
		Status:    payment.Status,
	}, nil
}
