package main

import (
	"time"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/app"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/config"
	grpcServer "github.com/dandirahmadani19/distributed-saga-orchestrator/platform/grpc"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/logger"
	grpcHandler "github.com/dandirahmadani19/distributed-saga-orchestrator/services/inventory/internal/infrastructure/grpc"
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

	handler := grpcHandler.NewInventoryHandler()
	handler.RegisterInventoryServiceServer(app.GRPC.Instance())

	if err := app.Run(); err != nil {
		app.Log.Fatal().Err(err).Msg("failed to run app")
	}

}
