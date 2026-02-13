package interceptor

import (
	"context"
	"time"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/logger"
	"google.golang.org/grpc"
)

func LoggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		start := time.Now()

		resp, err := handler(ctx, req)

		log.Info().
			Str("method", info.FullMethod).
			Dur("duration", time.Since(start)).
			Err(err).
			Msg("grpc request")

		return resp, err
	}
}
