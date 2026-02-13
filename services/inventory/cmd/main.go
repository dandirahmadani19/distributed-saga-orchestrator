package main

import (
	"time"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/app"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/config"
	grpcServer "github.com/dandirahmadani19/distributed-saga-orchestrator/platform/grpc"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/logger"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/internal/application/usecase"
	grpcHandler "github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/internal/infrastructure/grpc"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/internal/infrastructure/repository"
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
		app.Log.Fatal().Err(err).Msg("failed to create app")
	}

	repo := repository.NewPostgresReservationRepository(app.DB)
	uc := usecase.NewReserveInventoryUseCase(repo, app.Log)
	handler := grpcHandler.NewInventoryHandler(uc)
	handler.RegisterInventoryServiceServer(app.GRPC.Instance())

	if err := app.Run(); err != nil {
		app.Log.Fatal().Err(err).Msg("failed to run app")
	}

}
