package domain

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	CodeUnauthenticated         ErrorCode = "UNAUTHENTICATED"
	CodeAdminContextMissing     ErrorCode = "ADMIN_CONTEXT_MISSING"
	CodeForbidden               ErrorCode = "FORBIDDEN"
	CodeMFARequired             ErrorCode = "MFA_REQUIRED"
	CodeReasonRequired          ErrorCode = "REASON_REQUIRED"
	CodeRequestContextMissing   ErrorCode = "REQUEST_CONTEXT_MISSING"
	CodeInvalidStatus           ErrorCode = "INVALID_STATUS"
	CodeInvalidStatusTransition ErrorCode = "INVALID_STATUS_TRANSITION"
	CodeOrderNotFound           ErrorCode = "ORDER_NOT_FOUND"
	CodePaymentNotFound         ErrorCode = "PAYMENT_NOT_FOUND"
	CodeRefundNotFound          ErrorCode = "REFUND_NOT_FOUND"
	CodeRefundNotReviewable     ErrorCode = "REFUND_NOT_REVIEWABLE"
	CodeReviewTaskAlreadyExists ErrorCode = "REVIEW_TASK_ALREADY_EXISTS"
	CodePlatformSettingNotFound ErrorCode = "PLATFORM_SETTING_NOT_FOUND"
	CodeSettingVersionConflict  ErrorCode = "SETTING_VERSION_CONFLICT"
	CodeUserNotFound            ErrorCode = "USER_NOT_FOUND"
	CodeSellerNotFound          ErrorCode = "SELLER_NOT_FOUND"
	CodeDownstreamUnavailable   ErrorCode = "DOWNSTREAM_UNAVAILABLE"
	CodeAdminDisabled           ErrorCode = "ADMIN_DISABLED"
	CodeValidationFailed        ErrorCode = "VALIDATION_FAILED"
	CodeInternal                ErrorCode = "INTERNAL"
)

type AppError struct {
	Code               ErrorCode
	Message            string
	RequiredPermission Permission
	Cause              error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.RequiredPermission != "" {
		return fmt.Sprintf("%s: %s: %s", e.Code, e.Message, e.RequiredPermission)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

func NewUnauthenticated(message string) *AppError {
	return &AppError{Code: CodeUnauthenticated, Message: message}
}

func NewAdminContextMissing(message string) *AppError {
	return &AppError{Code: CodeAdminContextMissing, Message: message}
}

func NewForbidden(permission Permission) *AppError {
	return &AppError{
		Code:               CodeForbidden,
		Message:            "admin does not have required permission",
		RequiredPermission: permission,
	}
}

func NewMFARequired() *AppError {
	return &AppError{Code: CodeMFARequired, Message: "admin MFA verification is required for this action"}
}

func NewReasonRequired() *AppError {
	return &AppError{Code: CodeReasonRequired, Message: "action reason is required"}
}

func NewRequestContextMissing() *AppError {
	return &AppError{Code: CodeRequestContextMissing, Message: "request id is required for admin audit context"}
}

func NewInvalidStatus(resource string, status string) *AppError {
	return &AppError{
		Code:    CodeInvalidStatus,
		Message: fmt.Sprintf("invalid %s status %q", resource, status),
	}
}

func NewInvalidStatusTransition(resource string, current string, next string) *AppError {
	return &AppError{
		Code:    CodeInvalidStatusTransition,
		Message: fmt.Sprintf("%s status cannot change from %s to %s through admin status API", resource, current, next),
	}
}

func NewOrderNotFound(orderID string) *AppError {
	return &AppError{Code: CodeOrderNotFound, Message: fmt.Sprintf("order %q was not found", orderID)}
}

func NewPaymentNotFound(paymentID string) *AppError {
	return &AppError{Code: CodePaymentNotFound, Message: fmt.Sprintf("payment %q was not found", paymentID)}
}

func NewRefundNotFound(refundID string) *AppError {
	return &AppError{Code: CodeRefundNotFound, Message: fmt.Sprintf("refund %q was not found", refundID)}
}

func NewRefundNotReviewable(status string) *AppError {
	return &AppError{
		Code:    CodeRefundNotReviewable,
		Message: fmt.Sprintf("refund is in %s state and cannot be reviewed", status),
	}
}

func NewReviewTaskAlreadyExists(resourceType string, resourceID string) *AppError {
	return &AppError{
		Code:    CodeReviewTaskAlreadyExists,
		Message: fmt.Sprintf("open review task already exists for %s %q", resourceType, resourceID),
	}
}

func NewPlatformSettingNotFound(key PlatformSettingKey) *AppError {
	return &AppError{Code: CodePlatformSettingNotFound, Message: fmt.Sprintf("platform setting %q was not found", key)}
}

func NewSettingVersionConflict(key PlatformSettingKey, expected uint64, current uint64) *AppError {
	return &AppError{
		Code:    CodeSettingVersionConflict,
		Message: fmt.Sprintf("platform setting %q version conflict: expected %d, current %d", key, expected, current),
	}
}

func NewUserNotFound(userID string) *AppError {
	return &AppError{Code: CodeUserNotFound, Message: fmt.Sprintf("user %q was not found", userID)}
}

func NewSellerNotFound(sellerID string) *AppError {
	return &AppError{Code: CodeSellerNotFound, Message: fmt.Sprintf("seller %q was not found", sellerID)}
}

func NewDownstreamUnavailable(message string, cause error) *AppError {
	return &AppError{Code: CodeDownstreamUnavailable, Message: message, Cause: cause}
}

func NewValidationError(message string) *AppError {
	return &AppError{Code: CodeValidationFailed, Message: message}
}

func NewInternal(message string, cause error) *AppError {
	return &AppError{Code: CodeInternal, Message: message, Cause: cause}
}
