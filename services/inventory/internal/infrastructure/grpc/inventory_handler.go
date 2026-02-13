package grpc

import (
	"context"

	pb "github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/gen/proto/inventory/v1"
	"google.golang.org/grpc"
)

type InventoryHandler struct {
	pb.UnimplementedInventoryServiceServer
}

func NewInventoryHandler() *InventoryHandler {
	return &InventoryHandler{}
}

func (h *InventoryHandler) RegisterInventoryServiceServer(s *grpc.Server) {
	pb.RegisterInventoryServiceServer(s, h)
}

func (h *InventoryHandler) ReserveInventory(ctx context.Context, req *pb.ReserveInventoryRequest) (*pb.ReserveInventoryResponse, error) {
	return &pb.ReserveInventoryResponse{}, nil
}

func (h *InventoryHandler) ReleaseInventory(ctx context.Context, req *pb.ReleaseInventoryRequest) (*pb.ReleaseInventoryResponse, error) {
	return &pb.ReleaseInventoryResponse{}, nil
}
