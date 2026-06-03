package handlers

import (
	"log/slog"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *UserHandler) writeGRPCError(w http.ResponseWriter, r *http.Request, err error) {
	st, ok := status.FromError(err)
	if !ok {
		h.logger.ErrorContext(r.Context(), "gateway_user_http_error",
			slog.String("error_type", "non_grpc"),
		)
		writeAPIError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	switch st.Code() {
	case codes.InvalidArgument:
		writeAPIError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", st.Message())
	case codes.Unauthenticated:
		writeAPIError(w, r, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication required")
	case codes.PermissionDenied:
		writeAPIError(w, r, http.StatusForbidden, "PERMISSION_DENIED", "Permission denied")
	case codes.NotFound:
		writeAPIError(w, r, http.StatusNotFound, "NOT_FOUND", "Resource not found")
	case codes.AlreadyExists:
		writeAPIError(w, r, http.StatusConflict, "CONFLICT", st.Message())
	case codes.FailedPrecondition:
		writeAPIError(w, r, http.StatusConflict, "FAILED_PRECONDITION", st.Message())
	case codes.ResourceExhausted:
		writeAPIError(w, r, http.StatusTooManyRequests, "RATE_LIMITED", st.Message())
	case codes.DeadlineExceeded:
		writeAPIError(w, r, http.StatusGatewayTimeout, "UPSTREAM_TIMEOUT", "User service timed out")
	case codes.Unavailable:
		writeAPIError(w, r, http.StatusServiceUnavailable, "UPSTREAM_UNAVAILABLE", "User service is temporarily unavailable")
	default:
		h.logger.ErrorContext(r.Context(), "gateway_user_grpc_error",
			slog.String("code", st.Code().String()),
		)
		writeAPIError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}
