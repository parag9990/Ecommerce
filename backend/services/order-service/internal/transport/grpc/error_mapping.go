package ordergrpc

import (
	"context"
	"errors"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toStatusError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "request deadline exceeded")
	case errors.Is(err, domain.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, "authentication required")
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrOrderForbidden), errors.Is(err, domain.ErrCartForbidden):
		return status.Error(codes.PermissionDenied, "operation not allowed")
	case errors.Is(err, domain.ErrOrderNotFound), errors.Is(err, domain.ErrCartNotFound):
		return status.Error(codes.NotFound, "order not found")
	case errors.Is(err, domain.ErrInvalidRequest), errors.Is(err, domain.ErrInvalidPageToken),
		errors.Is(err, domain.ErrInvalidCheckoutCommand), errors.Is(err, domain.ErrInvalidAddress),
		errors.Is(err, domain.ErrTrackingRequired), errors.Is(err, domain.ErrInvalidPaymentResult),
		errors.Is(err, domain.ErrIdempotencyKeyRequired), errors.Is(err, domain.ErrIdempotencyKeyInvalid):
		return status.Error(codes.InvalidArgument, "invalid request")
	case errors.Is(err, domain.ErrIdempotencyConflict):
		return status.Error(codes.AlreadyExists, "idempotency key conflicts with an earlier checkout request")
	case errors.Is(err, domain.ErrCheckoutInProgress):
		return status.Error(codes.Aborted, "checkout is already processing; retry with the same key")
	case errors.Is(err, domain.ErrInventoryUnavailable):
		return status.Error(codes.ResourceExhausted, "inventory is unavailable")
	case errors.Is(err, domain.ErrInvalidOrderStatusTransition), errors.Is(err, domain.ErrOrderStatusUnchanged),
		errors.Is(err, domain.ErrMultiSellerFulfillmentPending), errors.Is(err, domain.ErrCartEmpty),
		errors.Is(err, domain.ErrProductUnavailable), errors.Is(err, domain.ErrVariantUnavailable),
		errors.Is(err, domain.ErrOrderNotPayable), errors.Is(err, domain.ErrInventoryReservationExpired),
		errors.Is(err, domain.ErrInvalidPaymentTransition), errors.Is(err, domain.ErrCheckoutFailed):
		return status.Error(codes.FailedPrecondition, "operation cannot be applied in the current state")
	case errors.Is(err, domain.ErrPaymentIntentPendingResolution):
		return status.Error(codes.Unavailable, "payment state is being resolved")
	case errors.Is(err, domain.ErrOrderCreateOutcomeUnknown):
		return status.Error(codes.Unavailable, "order state is being resolved")
	case errors.Is(err, domain.ErrTemporarilyUnavailable):
		return status.Error(codes.Unavailable, "order service temporarily unavailable")
	default:
		return status.Error(codes.Internal, "internal order service error")
	}
}
