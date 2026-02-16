package grpc

import (
	"errors"

	derr "github.com/dandirahmadani19/distributed-saga-orchestrator/platform/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ToStatus(err error) error {
	if err == nil {
		return nil
	}

	var e *derr.Error
	if !errors.As(err, &e) {
		return status.Error(codes.Internal, "internal server error")
	}

	switch e.Code {
	case derr.NotFound:
		return status.Error(codes.NotFound, e.Message)

	case derr.Conflict:
		return status.Error(codes.AlreadyExists, e.Message)

	case derr.Invalid:
		return status.Error(codes.InvalidArgument, e.Message)

	case derr.Forbidden:
		return status.Error(codes.PermissionDenied, e.Message)

	case derr.Unauthorized:
		return status.Error(codes.Unauthenticated, e.Message)

	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
