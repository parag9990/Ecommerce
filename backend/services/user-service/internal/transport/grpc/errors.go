package grpc

import (
	"context"
	"errors"

	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := status.FromError(err); ok {
		return err
	}

	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")
	case errors.Is(err, domain.ErrSellerNotFound):
		return status.Error(codes.NotFound, "seller profile not found")
	case errors.Is(err, domain.ErrDuplicateUser):
		return status.Error(codes.AlreadyExists, "user already exists")
	case errors.Is(err, domain.ErrDuplicateSeller):
		return status.Error(codes.AlreadyExists, "seller profile already exists")
	case errors.Is(err, domain.ErrForbidden):
		return status.Error(codes.PermissionDenied, "permission denied")
	case errors.Is(err, domain.ErrDeletedResource):
		return status.Error(codes.FailedPrecondition, "resource is deleted")
	case errors.Is(err, domain.ErrValidation), errors.Is(err, domain.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, "invalid request")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
