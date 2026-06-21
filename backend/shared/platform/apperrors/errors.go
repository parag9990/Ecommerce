package apperrors

import (
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Code string

const (
	CodeValidation   Code = "VALIDATION_ERROR"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeForbidden    Code = "FORBIDDEN"
	CodeNotFound     Code = "NOT_FOUND"
	CodeConflict     Code = "CONFLICT"
	CodeRateLimited  Code = "RATE_LIMITED"
	CodeUnavailable  Code = "SERVICE_UNAVAILABLE"
	CodeInternal     Code = "INTERNAL_ERROR"
)

type Error struct {
	Code       Code              `json:"code"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
	HTTPStatus int               `json:"-"`
	Cause      error             `json:"-"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.Cause }

func New(code Code, message string, httpStatus ...int) *Error {
	statusCode := HTTPStatus(code)
	if len(httpStatus) > 0 && httpStatus[0] != 0 {
		statusCode = httpStatus[0]
	}
	if code == "" {
		code = CodeInternal
	}
	if message == "" {
		message = safeMessage(code)
	}
	return &Error{Code: code, Message: message, HTTPStatus: statusCode}
}

func WithDetails(code Code, message string, details map[string]string) *Error {
	err := New(code, message)
	err.Details = cloneDetails(details)
	return err
}

func Wrap(err error, code Code, message string, httpStatus ...int) error {
	if err == nil {
		return nil
	}
	appErr := New(code, message, httpStatus...)
	appErr.Cause = err
	return appErr
}

func As(err error) (*Error, bool) {
	var appErr *Error
	return appErr, errors.As(err, &appErr)
}

func HTTPStatus(code Code) int {
	switch code {
	case CodeValidation:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodeUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func GRPCCode(code Code) codes.Code {
	switch code {
	case CodeValidation:
		return codes.InvalidArgument
	case CodeUnauthorized:
		return codes.Unauthenticated
	case CodeForbidden:
		return codes.PermissionDenied
	case CodeNotFound:
		return codes.NotFound
	case CodeConflict:
		return codes.AlreadyExists
	case CodeRateLimited:
		return codes.ResourceExhausted
	case CodeUnavailable:
		return codes.Unavailable
	default:
		return codes.Internal
	}
}

func ToGRPC(err error) error {
	if err == nil {
		return nil
	}
	appErr, ok := As(err)
	if !ok {
		return status.Error(codes.Internal, safeMessage(CodeInternal))
	}
	return status.Error(GRPCCode(appErr.Code), appErr.Message)
}

func FromGRPC(err error) *Error {
	if err == nil {
		return nil
	}
	grpcStatus, ok := status.FromError(err)
	if !ok {
		return New(CodeInternal, safeMessage(CodeInternal))
	}
	code := codeFromGRPC(grpcStatus.Code())
	return New(code, grpcStatus.Message())
}

func codeFromGRPC(code codes.Code) Code {
	switch code {
	case codes.InvalidArgument:
		return CodeValidation
	case codes.Unauthenticated:
		return CodeUnauthorized
	case codes.PermissionDenied:
		return CodeForbidden
	case codes.NotFound:
		return CodeNotFound
	case codes.AlreadyExists, codes.FailedPrecondition, codes.Aborted:
		return CodeConflict
	case codes.ResourceExhausted:
		return CodeRateLimited
	case codes.Unavailable, codes.DeadlineExceeded:
		return CodeUnavailable
	default:
		return CodeInternal
	}
}

func safeMessage(code Code) string {
	if code == CodeInternal {
		return "internal server error"
	}
	return "request failed"
}

func cloneDetails(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
