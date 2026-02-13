package grpc

import (
	"time"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/interceptor"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/logger"
	"google.golang.org/grpc"
)

func DefaultUnaryInterceptors(
	log *logger.Logger,
	timeout time.Duration,
) []grpc.UnaryServerInterceptor {

	return []grpc.UnaryServerInterceptor{
		interceptor.RecoveryInterceptor(log),
		interceptor.LoggingInterceptor(log),
		interceptor.TimeoutInterceptor(timeout),
	}
}
