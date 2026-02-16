package grpc

import (
	"context"

	grpcPlatform "github.com/dandirahmadani19/distributed-saga-orchestrator/platform/grpc"
	pb "github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/gen/proto/order/v1"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/application/dto"
	"google.golang.org/grpc"
)

type OrderCreator interface {
	Execute(ctx context.Context, req dto.CreateOrderRequest) (*dto.OrderResponse, error)
}
type OrderCanceller interface {
	Execute(ctx context.Context, orderID string, idempotencyKey string) (*dto.OrderResponse, error)
}

type OrderHandler struct {
	pb.UnimplementedOrderServiceServer
	createUC OrderCreator
	cancelUC OrderCanceller
}

func NewOrderHandler(createUC OrderCreator, cancelUC OrderCanceller) *OrderHandler {
	return &OrderHandler{
		createUC: createUC,
		cancelUC: cancelUC,
	}
}

func (h *OrderHandler) RegisterOrderServiceServer(s *grpc.Server) {
	pb.RegisterOrderServiceServer(s, h)
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
		return nil, grpcPlatform.ToStatus(err)
	}

	// Convert DTO to protobuf
	return &pb.CreateOrderResponse{
		OrderId:   result.ID,
		Status:    result.Status,
		CreatedAt: result.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	result, err := h.cancelUC.Execute(ctx, req.OrderId, req.IdempotencyKey)
	if err != nil {
		return nil, grpcPlatform.ToStatus(err)
	}
	return &pb.CancelOrderResponse{
		OrderId: result.ID,
		Status:  result.Status,
	}, nil
}
