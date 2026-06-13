package indexer

import (
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

func MapProductToDocument(payload domain.ProductIndexPayload, occurredAt time.Time) domain.ProductDocument {
	payload = payload.Normalized()

	createdAt := payload.UpdatedAt
	if payload.CreatedAt != nil && !payload.CreatedAt.IsZero() {
		createdAt = payload.CreatedAt.UTC()
	}
	if createdAt.IsZero() {
		createdAt = occurredAt.UTC()
	}
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	doc := domain.ProductDocument{
		ID:              payload.ProductID,
		Title:           payload.Title,
		Description:     optionalString(payload.Description),
		Brand:           optionalString(payload.Brand),
		CategoryIDs:     payload.CategoryIDs,
		SellerID:        payload.SellerID,
		Price:           derefFloat(payload.Price),
		Rating:          payload.Rating,
		PopularityScore: derefInt32(payload.PopularityScore),
		InStock:         derefBool(payload.InStock),
		CreatedAt:       createdAt.Unix(),
	}
	return doc
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func derefFloat(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func derefInt32(value *int32) int32 {
	if value == nil {
		return 0
	}
	return *value
}

func derefBool(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}
