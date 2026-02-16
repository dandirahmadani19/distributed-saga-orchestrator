package grpc

import (
	"context"

	pb "github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/gen/proto/inventory/v1"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/internal/application/dto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReserveInventory interface {
	Execute(ctx context.Context, req dto.ReserveInventoryRequest) (*dto.ReservationResponse, error)
}
type ReleaseInventory interface {
	Execute(ctx context.Context, req dto.ReleaseInventoryRequest) (*dto.ReservationResponse, error)
}

type InventoryHandler struct {
	pb.UnimplementedInventoryServiceServer
	ucReserve ReserveInventory
	ucRelease ReleaseInventory
}

func NewInventoryHandler(ucReserve ReserveInventory, ucRelease ReleaseInventory) *InventoryHandler {
	return &InventoryHandler{
		ucReserve: ucReserve,
		ucRelease: ucRelease,
	}
}

func (h *InventoryHandler) RegisterInventoryServiceServer(s *grpc.Server) {
	pb.RegisterInventoryServiceServer(s, h)
}

func (h *InventoryHandler) ReserveInventory(ctx context.Context, req *pb.ReserveInventoryRequest) (*pb.ReserveInventoryResponse, error) {
	items := make([]dto.ReserveItemRequest, len(req.Items))
	for i, item := range req.Items {
		items[i] = dto.ReserveItemRequest{
			ProductID: item.ProductId,
			Quantity:  int(item.Quantity),
		}
	}

	reservation, err := h.ucReserve.Execute(ctx, dto.ReserveInventoryRequest{
		IdempotencyKey: req.IdempotencyKey,
		OrderID:        req.OrderId,
		Items:          items,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ReserveInventoryResponse{
		ReservationId: reservation.ID,
		Status:        string(reservation.Status),
	}, nil
}

func (h *InventoryHandler) ReleaseInventory(ctx context.Context, req *pb.ReleaseInventoryRequest) (*pb.ReleaseInventoryResponse, error) {
	reservation, err := h.ucRelease.Execute(ctx, dto.ReleaseInventoryRequest{
		IdempotencyKey: req.IdempotencyKey,
		OrderID:        req.OrderId,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ReleaseInventoryResponse{
		Status: reservation.Status,
	}, nil
}
