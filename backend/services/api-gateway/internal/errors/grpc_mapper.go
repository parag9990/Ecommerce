package gatewayerrors

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapGRPCError(err error) (MappedError, bool) {
	st, ok := status.FromError(err)
	if !ok {
		return MappedError{}, false
	}

	details, headers := extractDetails(st)
	code, httpStatus, message := mapGRPCCode(st.Code(), st.Message())
	return MappedError{
		Status:   httpStatus,
		Code:     code,
		Message:  message,
		Details:  details,
		Headers:  headers,
		GRPCCode: st.Code().String(),
		Internal: err,
	}, true
}

func mapGRPCCode(code codes.Code, msg string) (string, int, string) {
	switch code {
	case codes.InvalidArgument:
		return CodeValidation, http.StatusBadRequest, messageOrDefault(msg, "Invalid request")
	case codes.Unauthenticated:
		return CodeUnauthorized, http.StatusUnauthorized, "Authentication required"
	case codes.PermissionDenied:
		return CodeForbidden, http.StatusForbidden, "You do not have permission to perform this action"
	case codes.NotFound:
		return CodeNotFound, http.StatusNotFound, messageOrDefault(msg, "Resource not found")
	case codes.AlreadyExists, codes.Aborted:
		return CodeConflict, http.StatusConflict, messageOrDefault(msg, "Request conflicts with current state")
	case codes.FailedPrecondition:
		return CodeFailedPrecondition, http.StatusUnprocessableEntity, messageOrDefault(msg, "Request cannot be processed in the current state")
	case codes.OutOfRange:
		return CodeValidation, http.StatusBadRequest, messageOrDefault(msg, "Value is out of allowed range")
	case codes.ResourceExhausted:
		return CodeRateLimited, http.StatusTooManyRequests, messageOrDefault(msg, "Too many requests")
	case codes.DeadlineExceeded:
		return CodeTimeout, http.StatusGatewayTimeout, "Request timed out"
	case codes.Unavailable:
		return CodeServiceUnavailable, http.StatusServiceUnavailable, "Service temporarily unavailable"
	case codes.Unimplemented:
		return CodeNotImplemented, http.StatusNotImplemented, "Feature is not implemented yet"
	case codes.Canceled:
		return CodeRequestCancelled, StatusClientClosedRequest, "Request was cancelled"
	case codes.Internal, codes.Unknown, codes.DataLoss:
		fallthrough
	default:
		return CodeInternal, http.StatusInternalServerError, "Internal server error"
	}
}

func messageOrDefault(msg string, fallback string) string {
	return sanitizePublicMessage(msg, fallback)
}
