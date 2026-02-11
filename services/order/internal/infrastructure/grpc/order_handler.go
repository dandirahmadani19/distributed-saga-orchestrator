package grpc

import (
	"context"

	pb "github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/gen/proto/order/v1"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/application/dto"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/application/usecase"
)

type OrderHandler struct {
	pb.UnimplementedOrderServiceServer
	createUC *usecase.CreateOrderUseCase
}

func NewOrderHandler(createUC *usecase.CreateOrderUseCase) *OrderHandler {
	return &OrderHandler{
		createUC: createUC,
	}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	// Convert protobuf to DTO
	items := make([]dto.OrderItemDTO, len(req.Items))
	for i, item := range req.Items {
		items[i] = dto.OrderItemDTO{
			ProductID: item.ProductId,
			Quantity:  int(item.Quantity),
			Price:     item.Price,
		}
	}

	dtoReq := dto.CreateOrderRequest{
		IdempotencyKey: req.IdempotencyKey,
		CustomerID:     req.CustomerId,
		Items:          items,
		TotalAmount:    req.TotalAmount,
	}

	// Execute use case
	result, err := h.createUC.Execute(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	// Convert DTO to protobuf
	return &pb.CreateOrderResponse{
		OrderId:   result.ID,
		Status:    result.Status,
		CreatedAt: result.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
