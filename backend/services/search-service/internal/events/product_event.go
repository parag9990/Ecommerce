package events

import (
	"errors"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

type ProductIndexPayload = domain.ProductIndexPayload

type ProductIndexResult struct {
	EventID   string
	EventType string
	ProductID string
	Action    string
}

func IsPermanentProductIndexError(err error) bool {
	return errors.Is(err, domain.ErrInvalidProductEvent) ||
		errors.Is(err, domain.ErrUnsupportedProductEvent) ||
		errors.Is(err, domain.ErrInvalidProductDocument)
}
