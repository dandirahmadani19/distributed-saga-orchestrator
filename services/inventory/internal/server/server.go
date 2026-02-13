package server

// import (
// 	"context"
// 	"fmt"
// 	"net"
// 	"time"

// 	"github.com/dandirahmadani19/distributed-saga-orchestrator/shared/pkg/interceptor"
// 	"github.com/dandirahmadani19/distributed-saga-orchestrator/shared/pkg/logger"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/keepalive"
// )

// type GRPCConfig struct {
// 	Port string
// }

// type Server struct {
// 	grpcServer *grpc.Server
// 	listener   net.Listener
// 	logger     *logger.Logger
// }

// func New(
// 	cfg GRPCConfig,
// 	register func(*grpc.Server),
// 	log *logger.Logger,
// ) (*Server, error) {

// 	opts := []grpc.ServerOption{
// 		grpc.ConnectionTimeout(5 * time.Second),

// 		// Keepalive protection
// 		grpc.KeepaliveParams(keepalive.ServerParameters{
// 			MaxConnectionIdle:     5 * time.Minute,
// 			MaxConnectionAge:      30 * time.Minute,
// 			MaxConnectionAgeGrace: 5 * time.Minute,
// 			Time:                  2 * time.Hour,
// 			Timeout:               20 * time.Second,
// 		}),

// 		// Interceptors
// 		grpc.ChainUnaryInterceptor(
// 			interceptor.LoggingInterceptor(log),
// 			interceptor.RecoveryInterceptor(log),
// 			interceptor.TimeoutInterceptor(5*time.Second),
// 		),
// 	}

// 	grpcServer := grpc.NewServer(opts...)

// 	// Register services from main
// 	register(grpcServer)

// 	lis, err := net.Listen("tcp", ":"+cfg.Port)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to listen: %w", err)
// 	}

// 	return &Server{
// 		grpcServer: grpcServer,
// 		listener:   lis,
// 		logger:     log,
// 	}, nil
// }

// func (s *Server) Start() error {
// 	s.logger.Info().
// 		Str("addr", s.listener.Addr().String()).
// 		Msg("🚀 gRPC server started")

// 	return s.grpcServer.Serve(s.listener)
// }

// func (s *Server) Shutdown(ctx context.Context) error {
// 	done := make(chan struct{})

// 	go func() {
// 		s.grpcServer.GracefulStop()
// 		close(done)
// 	}()

// 	select {
// 	case <-done:
// 		s.logger.Info().Msg("server stopped gracefully")
// 		return nil
// 	case <-ctx.Done():
// 		s.grpcServer.Stop()
// 		return ctx.Err()
// 	}
// }
