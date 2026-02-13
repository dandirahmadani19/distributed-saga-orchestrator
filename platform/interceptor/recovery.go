package interceptor

import (
	"context"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RecoveryInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {

		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Str("method", info.FullMethod).
					Msg("panic recovered")

				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}
