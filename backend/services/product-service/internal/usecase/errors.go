package usecase

import (
	"fmt"

	"product-service/internal/domain"
)

type ErrorKind string

const (
	ErrorKindInvalidArgument    ErrorKind = "invalid_argument"
	ErrorKindUnauthenticated    ErrorKind = "unauthenticated"
	ErrorKindPermissionDenied   ErrorKind = "permission_denied"
	ErrorKindNotFound           ErrorKind = "not_found"
	ErrorKindAlreadyExists      ErrorKind = "already_exists"
	ErrorKindFailedPrecondition ErrorKind = "failed_precondition"
	ErrorKindUnavailable        ErrorKind = "unavailable"
	ErrorKindConflict           ErrorKind = "conflict"
)

const (
	ErrorCodeValidation              = "VALIDATION_ERROR"
	ErrorCodeUnauthenticated         = "UNAUTHENTICATED"
	ErrorCodePermissionDenied        = "PERMISSION_DENIED"
	ErrorCodeProductNotFound         = "PRODUCT_NOT_FOUND"
	ErrorCodeProductOwnership        = "PRODUCT_OWNERSHIP_MISMATCH"
	ErrorCodeDuplicateSKU            = "DUPLICATE_SKU"
	ErrorCodeSellerCatalogDisabled   = "SELLER_CATALOG_DISABLED"
	ErrorCodeProductNotEditable      = "PRODUCT_NOT_EDITABLE"
	ErrorCodeInvalidStatusTransition = "INVALID_STATUS_TRANSITION"
	ErrorCodeCMSUnavailable          = "CMS_UNAVAILABLE"
)

type ServiceError struct {
	Kind    ErrorKind
	Code    string
	Message string
	Report  domain.ValidationReport
	Err     error
}

func (e *ServiceError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *ServiceError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func validationFailed(report domain.ValidationReport) error {
	return &ServiceError{
		Kind:    ErrorKindInvalidArgument,
		Code:    ErrorCodeValidation,
		Message: "product validation failed",
		Report:  report,
	}
}

func serviceError(kind ErrorKind, code string, message string, err error) error {
	return &ServiceError{
		Kind:    kind,
		Code:    code,
		Message: message,
		Err:     err,
	}
}
