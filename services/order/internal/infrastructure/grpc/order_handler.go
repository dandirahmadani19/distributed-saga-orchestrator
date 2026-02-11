package grpc

import (
	"context"

	pb "github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/gen/proto/order/v1"
)

type OrderHandler struct {
	pb.UnimplementedOrderServiceServer
}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	return &pb.CreateOrderResponse{}, nil
}
