package provider

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrInvalidProviderRequest  = errors.New("invalid payment provider request")
	ErrProviderNotFound        = errors.New("payment provider not found")
	ErrDuplicateProvider       = errors.New("duplicate payment provider")
	ErrUnsupportedOperation    = errors.New("payment provider operation unsupported")
	ErrInvalidWebhookSignature = errors.New("invalid payment webhook signature")
)

type ErrorCode string

const (
	ErrorCodeValidation       ErrorCode = "provider_validation_error"
	ErrorCodeAuthentication   ErrorCode = "provider_authentication_error"
	ErrorCodeRateLimited      ErrorCode = "provider_rate_limited"
	ErrorCodeTimeout          ErrorCode = "provider_timeout"
	ErrorCodeUnavailable      ErrorCode = "provider_unavailable"
	ErrorCodeDeclined         ErrorCode = "payment_declined"
	ErrorCodeUnknownStatus    ErrorCode = "provider_unknown_status"
	ErrorCodeProviderNotFound ErrorCode = "provider_not_found"
	ErrorCodeUnsupported      ErrorCode = "provider_operation_unsupported"
	ErrorCodeWebhookSignature ErrorCode = "invalid_webhook_signature"
)

type Error struct {
	Code       ErrorCode
	Message    string
	Provider   string
	Operation  string
	Retryable  bool
	StatusCode int
	Cause      error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	message := strings.TrimSpace(e.Message)
	if message == "" {
		message = string(e.Code)
	}
	provider := NormalizeProviderName(e.Provider)
	operation := strings.TrimSpace(e.Operation)
	switch {
	case provider != "" && operation != "":
		return fmt.Sprintf("%s: provider=%s operation=%s", message, provider, operation)
	case provider != "":
		return fmt.Sprintf("%s: provider=%s", message, provider)
	case operation != "":
		return fmt.Sprintf("%s: operation=%s", message, operation)
	default:
		return message
	}
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func NewError(providerName string, operation string, code ErrorCode, message string, opts ...ErrorOption) *Error {
	err := &Error{
		Code:       code,
		Message:    strings.TrimSpace(message),
		Provider:   NormalizeProviderName(providerName),
		Operation:  strings.TrimSpace(operation),
		StatusCode: http.StatusBadGateway,
	}
	for _, opt := range opts {
		opt(err)
	}
	return err
}

type ErrorOption func(*Error)

func WithRetryable(retryable bool) ErrorOption {
	return func(err *Error) {
		err.Retryable = retryable
	}
}

func WithStatusCode(statusCode int) ErrorOption {
	return func(err *Error) {
		err.StatusCode = statusCode
	}
}

func WithCause(cause error) ErrorOption {
	return func(err *Error) {
		err.Cause = cause
	}
}

func NewUnsupportedOperationError(providerName string, operation string) *Error {
	return NewError(
		providerName,
		operation,
		ErrorCodeUnsupported,
		"payment provider operation is not enabled",
		WithStatusCode(http.StatusNotImplemented),
		WithCause(ErrUnsupportedOperation),
	)
}

func NewProviderNotFoundError(providerName string) *Error {
	return NewError(
		providerName,
		"",
		ErrorCodeProviderNotFound,
		"payment provider is not registered",
		WithStatusCode(http.StatusBadRequest),
		WithCause(ErrProviderNotFound),
	)
}

func IsRetryable(err error) bool {
	var providerErr *Error
	return errors.As(err, &providerErr) && providerErr.Retryable
}

func CodeOf(err error) (ErrorCode, bool) {
	var providerErr *Error
	if !errors.As(err, &providerErr) {
		return "", false
	}
	return providerErr.Code, true
}
