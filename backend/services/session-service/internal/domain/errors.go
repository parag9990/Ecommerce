package domain

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidSession         = errors.New("invalid session model")
	ErrInvalidSessionEvent    = errors.New("invalid session event")
	ErrInvalidJourney         = errors.New("invalid journey model")
	ErrInvalidHeatmap         = errors.New("invalid heatmap model")
	ErrInvalidAnalytics       = errors.New("invalid analytics model")
	ErrInvalidRetentionPolicy = errors.New("invalid retention policy")
	ErrInvalidActiveSession   = errors.New("invalid active session")
	ErrSessionNotFound        = errors.New("session not found")
	ErrSessionEventNotFound   = errors.New("session event not found")
	ErrJourneyNotFound        = errors.New("journey not found")
	ErrHeatmapNotFound        = errors.New("heatmap not found")
	ErrNilSession             = errors.New("session is nil")
	ErrNotFound               = errors.New("not found")
	ErrInvalidInput           = errors.New("invalid input")
	ErrUnauthenticated        = errors.New("authentication required")
	ErrForbidden              = errors.New("permission denied")
	ErrReasonRequired         = errors.New("audit reason is required")
	ErrConfirmation           = errors.New("confirmation is required")
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

func (e FieldError) Is(target error) bool { return target == ErrInvalidInput }

func NewFieldError(field string, message string) error {
	return FieldError{Field: field, Message: message}
}
