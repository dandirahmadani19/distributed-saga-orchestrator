package app

import (
	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/config"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/logger"
	"google.golang.org/grpc"
)

type Option func(*App)

type GRPCOptionFactory func(cfg *config.Config, log *logger.Logger) []grpc.ServerOption

func WithGRPCOptions(factory GRPCOptionFactory) Option {
	return func(a *App) {
		a.grpcFactory = factory
	}
}
