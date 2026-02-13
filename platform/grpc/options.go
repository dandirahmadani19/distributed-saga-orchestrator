package grpc

import (
	"time"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/logger"
	"google.golang.org/grpc"
)

func DefaultUnaryServerOptions(
	log *logger.Logger,
	timeout time.Duration,
) []grpc.ServerOption {

	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			DefaultUnaryInterceptors(log, timeout)...,
		),
	}
}
