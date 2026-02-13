package grpc

import (
	"context"

	pb "github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/gen/proto/inventory/v1"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/internal/application/dto"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/internal/application/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InventoryHandler struct {
	pb.UnimplementedInventoryServiceServer
	ucReserve *usecase.ReserveInventoryUseCase
}

func NewInventoryHandler(ucReserve *usecase.ReserveInventoryUseCase) *InventoryHandler {
	return &InventoryHandler{
		ucReserve: ucReserve,
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
	return &pb.ReleaseInventoryResponse{}, nil
}
