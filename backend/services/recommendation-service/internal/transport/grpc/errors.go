package grpctransport

import (
	"context"
	"errors"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func grpcError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrInvalidRecommendationRequest),
		errors.Is(err, domain.ErrInvalidLimit),
		errors.Is(err, domain.ErrUnsupportedContext),
		errors.Is(err, domain.ErrUnsupportedRecommendationType),
		errors.Is(err, domain.ErrUnsupportedStrategy),
		errors.Is(err, domain.ErrInvalidCacheKey),
		errors.Is(err, domain.ErrInvalidRecommendationSet):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrDefinitionNotFound):
		return status.Error(codes.NotFound, "recommendation definition not found")
	case errors.Is(err, domain.ErrRecommendationStorage):
		return status.Error(codes.Unavailable, "recommendations temporarily unavailable")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "recommendation deadline exceeded")
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "recommendation request canceled")
	default:
		return status.Error(codes.Internal, "internal recommendation error")
	}
}
