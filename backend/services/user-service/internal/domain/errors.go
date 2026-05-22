package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrValidation        = errors.New("validation failed")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrDeletedResource   = errors.New("resource is deleted")

	ErrUserNotFound         = errors.New("user not found")
	ErrDuplicateUser        = errors.New("user already exists")
	ErrAddressNotFound      = errors.New("address not found")
	ErrDuplicateAddress     = errors.New("address already exists")
	ErrSellerNotFound       = errors.New("seller profile not found")
	ErrDuplicateSeller      = errors.New("seller profile already exists")
	ErrKYCDocumentNotFound  = errors.New("kyc document not found")
	ErrDuplicateKYCDocument = errors.New("kyc document already exists")
)

type FieldError struct {
	Field   string
	Message string
}

type ValidationError struct {
	Fields []FieldError
}

func (e ValidationError) Error() string {
	if len(e.Fields) == 0 {
		return ErrValidation.Error()
	}

	parts := make([]string, 0, len(e.Fields))
	for _, field := range e.Fields {
		parts = append(parts, fmt.Sprintf("%s: %s", field.Field, field.Message))
	}

	return fmt.Sprintf("%s: %s", ErrValidation.Error(), strings.Join(parts, "; "))
}

func (e ValidationError) Unwrap() error {
	return ErrValidation
}

type validationCollector struct {
	fields []FieldError
}

func (v *validationCollector) add(field string, message string) {
	v.fields = append(v.fields, FieldError{Field: field, Message: message})
}

func (v *validationCollector) err() error {
	if len(v.fields) == 0 {
		return nil
	}

	return ValidationError{Fields: v.fields}
}

func invalidTransition(entity string, from string, to string) error {
	return fmt.Errorf("%w: %s cannot move from %q to %q", ErrInvalidTransition, entity, from, to)
}
