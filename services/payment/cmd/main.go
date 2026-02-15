package main

import (
	"time"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/app"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/config"
	grpcServer "github.com/dandirahmadani19/distributed-saga-orchestrator/platform/grpc"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/logger"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/application/usecase"
	grpcHandler "github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/infrastructure/grpc"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/infrastructure/repository"
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

	repo := repository.NewPostgresPaymentRepository(app.DB)
	ucProcess := usecase.NewProcessPaymentUseCase(repo, app.Log)
	ucRefund := usecase.NewRefundPaymentUseCase(repo, app.Log)
	handler := grpcHandler.NewPaymentHandler(ucProcess, ucRefund)
	handler.RegisterPaymentServiceServer(app.GRPC.Instance())

	if err := app.Run(); err != nil {
		app.Log.Fatal().Err(err).Msg("failed to run app")
	}

}
