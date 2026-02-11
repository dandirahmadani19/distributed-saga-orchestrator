package server

import (
	"database/sql"
	"fmt"
	"net"
	"time"

	pb "github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/gen/proto/order/v1"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/application/usecase"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/config"
	grpcHandler "github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/infrastructure/grpc"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/infrastructure/repository"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/shared/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	logger     *logger.Logger
}

// New creates a new server (no config parameter needed!)
func New(db *sql.DB, log *logger.Logger) (*Server, error) {
	repo := repository.NewPostgresOrderRepository(db)
	createUC := usecase.NewCreateOrderUseCase(repo, log)
	cancelUC := usecase.NewCancelOrderUseCase(repo, log)
	handler := grpcHandler.NewOrderHandler(createUC, cancelUC)

	// Setup gRPC server
	grpcServer := grpc.NewServer(grpc.ConnectionTimeout(5 * time.Second))
	pb.RegisterOrderServiceServer(grpcServer, handler)

	reflection.Register(grpcServer)

	// Use global config for port
	port := config.Server().Port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	return &Server{
		grpcServer: grpcServer,
		listener:   lis,
		logger:     log,
	}, nil
}

func (s *Server) Start() error {
	s.logger.Info().
		Str("port", s.listener.Addr().String()).
		Msg("🚀 Order Service started")
	return s.grpcServer.Serve(s.listener)
}

func (s *Server) GracefulStop() {
	s.logger.Info().Msg("Shutting down server...")
	s.grpcServer.GracefulStop()
	s.logger.Info().Msg("✅ Server stopped")
}
