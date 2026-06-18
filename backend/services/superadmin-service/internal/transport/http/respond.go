package http

import (
	"encoding/json"
	"net/http"

	"ecommerce/superadmin-service/internal/domain"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	response := errorResponse{
		Code:      string(domain.CodeInternal),
		Message:   "internal error",
		RequestID: requestIDFromRequest(r),
	}

	if appErr, ok := domain.AsAppError(err); ok {
		status = statusForCode(appErr.Code)
		response.Code = string(appErr.Code)
		response.Message = appErr.Message
		response.RequiredPermission = string(appErr.RequiredPermission)
	}

	writeJSON(w, status, response)
}

func statusForCode(code domain.ErrorCode) int {
	switch code {
	case domain.CodeUnauthenticated, domain.CodeAdminContextMissing:
		return http.StatusUnauthorized
	case domain.CodeForbidden, domain.CodeMFARequired, domain.CodeAdminDisabled:
		return http.StatusForbidden
	case domain.CodeReasonRequired, domain.CodeRequestContextMissing, domain.CodeValidationFailed,
		domain.CodeInvalidStatus, domain.CodeInvalidStatusTransition:
		return http.StatusBadRequest
	case domain.CodeUserNotFound, domain.CodeSellerNotFound,
		domain.CodeOrderNotFound, domain.CodePaymentNotFound, domain.CodeRefundNotFound,
		domain.CodePlatformSettingNotFound:
		return http.StatusNotFound
	case domain.CodeRefundNotReviewable, domain.CodeReviewTaskAlreadyExists, domain.CodeSettingVersionConflict:
		return http.StatusConflict
	case domain.CodeDownstreamUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
