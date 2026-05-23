package gatewayerrors

import (
	"context"
	stderrors "errors"
	"net/http"
)

type Mapper interface {
	Map(ctx context.Context, err error) MappedError
}

type Service struct{}

func NewMapper() *Service {
	return &Service{}
}

func Map(ctx context.Context, err error) MappedError {
	return NewMapper().Map(ctx, err)
}

func (Service) Map(_ context.Context, err error) MappedError {
	if err == nil {
		return MappedError{}
	}

	var local LocalError
	if stderrors.As(err, &local) {
		return mapLocalError(local, err)
	}
	if stderrors.Is(err, context.DeadlineExceeded) {
		return timeoutError(err)
	}
	if stderrors.Is(err, context.Canceled) {
		return cancelledError(err)
	}
	if mapped, ok := mapGRPCError(err); ok {
		return mapped
	}
	return internalError(err)
}

func mapLocalError(local LocalError, err error) MappedError {
	mapped := MappedError{
		Status:   normalizeStatus(local.StatusCode()),
		Code:     normalizeCode(local.ErrorCode()),
		Message:  sanitizePublicMessage(local.PublicMessage(), fallbackMessage(local.ErrorCode())),
		Details:  sanitizeDetails(local.PublicDetails()),
		Internal: err,
	}
	if provider, ok := local.(HeaderProvider); ok {
		mapped.Headers = copyHeaders(provider.ResponseHeaders())
	}
	return mapped
}

func timeoutError(err error) MappedError {
	return MappedError{
		Status:   http.StatusGatewayTimeout,
		Code:     CodeTimeout,
		Message:  "Request timed out",
		Internal: err,
	}
}

func cancelledError(err error) MappedError {
	return MappedError{
		Status:   StatusClientClosedRequest,
		Code:     CodeRequestCancelled,
		Message:  "Request was cancelled",
		Internal: err,
	}
}

func internalError(err error) MappedError {
	return MappedError{
		Status:   http.StatusInternalServerError,
		Code:     CodeInternal,
		Message:  "Internal server error",
		Internal: err,
	}
}

func fallbackMessage(code string) string {
	switch code {
	case CodeValidation:
		return "Invalid request"
	case CodeUnauthorized:
		return "Authentication required"
	case CodeForbidden:
		return "You do not have permission to perform this action"
	case CodeNotFound:
		return "Resource not found"
	case CodeConflict:
		return "Request conflicts with current state"
	case CodeFailedPrecondition:
		return "Request cannot be processed in the current state"
	case CodeRateLimited:
		return "Too many requests"
	case CodeRequestCancelled:
		return "Request was cancelled"
	case CodeTimeout:
		return "Request timed out"
	case CodeServiceUnavailable:
		return "Service temporarily unavailable"
	case CodeBadGateway:
		return "Bad gateway"
	case CodeNotImplemented:
		return "Feature is not implemented yet"
	default:
		return "Internal server error"
	}
}
