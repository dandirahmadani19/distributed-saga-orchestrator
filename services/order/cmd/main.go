package main

import (
	"log"
	"time"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/app"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/config"
	grpcServer "github.com/dandirahmadani19/distributed-saga-orchestrator/platform/grpc"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/logger"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/application/usecase"
	grpcHandler "github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/infrastructure/grpc"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/infrastructure/repository"
	"google.golang.org/grpc"
)

func main() {
	app, err := app.New(
		app.WithGRPCOptions(func(cfg *config.Config, log *logger.Logger) []grpc.ServerOption {
			return []grpc.ServerOption{
				grpc.ChainUnaryInterceptor(
					grpcServer.DefaultUnaryInterceptors(log, 5*time.Minute)...,
				),
			}
		}),
	)
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}

	repo := repository.NewPostgresOrderRepository(app.DB)
	ucReserve := usecase.NewCreateOrderUseCase(repo, app.Log)
	ucRelease := usecase.NewCancelOrderUseCase(repo, app.Log)
	handler := grpcHandler.NewOrderHandler(ucReserve, ucRelease)
	handler.RegisterOrderServiceServer(app.GRPC.Instance())

	if err := app.Run(); err != nil {
		app.Log.Fatal().Err(err).Msg("failed to run app")
	}

}
