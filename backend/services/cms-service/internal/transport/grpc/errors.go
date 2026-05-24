package grpctransport

import (
	"context"
	"errors"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func grpcError(err error) error {
	if err == nil {
		return nil
	}
	var validationErr domain.ValidationError
	if errors.As(err, &validationErr) {
		return status.Error(codes.InvalidArgument, validationErr.Error())
	}
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	case errors.Is(err, domain.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, "authentication required")
	case errors.Is(err, domain.ErrForbidden),
		errors.Is(err, domain.ErrInactiveStaff),
		errors.Is(err, domain.ErrCouponOwnershipMismatch),
		errors.Is(err, domain.ErrCampaignOwnershipMismatch):
		return status.Error(codes.PermissionDenied, "permission denied")
	case errors.Is(err, domain.ErrSellerContextRequired),
		errors.Is(err, domain.ErrInvalidSellerScope),
		errors.Is(err, domain.ErrInvalidAuditLogFilter):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrCouponNotFound),
		errors.Is(err, domain.ErrCampaignNotFound),
		errors.Is(err, domain.ErrSellerSettingsNotFound):
		return status.Error(codes.NotFound, "resource not found")
	case errors.Is(err, domain.ErrStaffLookupFailed),
		errors.Is(err, domain.ErrAnalyticsUnavailable),
		errors.Is(err, domain.ErrAuditLogsUnavailable):
		return status.Error(codes.Unavailable, "dependency temporarily unavailable")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
