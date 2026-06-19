package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidInput    = errors.New("invalid input")
	ErrUnauthenticated = errors.New("authentication required")
	ErrForbidden       = errors.New("permission denied")
	ErrReasonRequired  = errors.New("audit reason is required")
	ErrConfirmation    = errors.New("confirmation is required")
)

type FieldError struct {
	Field   string
	Message string
}

func (e FieldError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e FieldError) Is(target error) bool {
	return target == ErrInvalidInput
}

func NewFieldError(field string, message string) error {
	return FieldError{Field: field, Message: message}
}
