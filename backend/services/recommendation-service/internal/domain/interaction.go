package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type UserInteraction struct {
	ID                  string          `json:"id" bson:"_id"`
	DedupeKey           string          `json:"dedupe_key" bson:"dedupe_key"`
	EventID             string          `json:"event_id" bson:"event_id"`
	SourceEventType     SourceEventType `json:"event_type" bson:"event_type"`
	NormalizedEventType InteractionType `json:"normalized_event_type" bson:"normalized_event_type"`
	Version             int             `json:"version" bson:"version"`
	Producer            string          `json:"producer" bson:"producer"`
	TraceID             string          `json:"trace_id,omitempty" bson:"trace_id,omitempty"`
	UserID              string          `json:"user_id,omitempty" bson:"user_id,omitempty"`
	AnonymousID         string          `json:"anonymous_id,omitempty" bson:"anonymous_id,omitempty"`
	SessionID           string          `json:"session_id,omitempty" bson:"session_id,omitempty"`
	ProductID           string          `json:"product_id" bson:"product_id"`
	VariantID           string          `json:"variant_id,omitempty" bson:"variant_id,omitempty"`
	CategoryID          string          `json:"category_id,omitempty" bson:"category_id,omitempty"`
	SellerID            string          `json:"seller_id,omitempty" bson:"seller_id,omitempty"`
	BrandID             string          `json:"brand_id,omitempty" bson:"brand_id,omitempty"`
	Weight              int             `json:"weight" bson:"weight"`
	Quantity            int             `json:"quantity" bson:"quantity"`
	Metadata            map[string]any  `json:"metadata,omitempty" bson:"metadata,omitempty"`
	OccurredAt          time.Time       `json:"occurred_at" bson:"occurred_at"`
	ReceivedAt          time.Time       `json:"received_at" bson:"received_at"`
}

type InteractionInsertResult struct {
	Duplicate bool
}

func NewInteractionDedupeKey(eventID string, eventType InteractionType, productID string, variantID string) string {
	return strings.Join([]string{
		strings.TrimSpace(eventID),
		string(eventType),
		strings.TrimSpace(productID),
		strings.TrimSpace(variantID),
	}, "#")
}

func NewInteractionID(dedupeKey string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(dedupeKey)))
	return "interaction_" + hex.EncodeToString(sum[:16])
}

func (i UserInteraction) Validate() error {
	if strings.TrimSpace(i.ID) == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidInteraction)
	}
	if strings.TrimSpace(i.DedupeKey) == "" {
		return fmt.Errorf("%w: dedupe_key is required", ErrInvalidInteraction)
	}
	if strings.TrimSpace(i.EventID) == "" {
		return fmt.Errorf("%w: event_id is required", ErrInvalidInteraction)
	}
	if !i.SourceEventType.IsSupported() {
		return fmt.Errorf("%w: event_type is unsupported", ErrInvalidInteraction)
	}
	if !i.NormalizedEventType.IsValid() {
		return fmt.Errorf("%w: normalized_event_type is unsupported", ErrInvalidInteraction)
	}
	if i.Version <= 0 {
		return fmt.Errorf("%w: version must be greater than zero", ErrInvalidInteraction)
	}
	if strings.TrimSpace(i.Producer) == "" {
		return fmt.Errorf("%w: producer is required", ErrInvalidInteraction)
	}
	if strings.TrimSpace(i.UserID) == "" && strings.TrimSpace(i.AnonymousID) == "" {
		return fmt.Errorf("%w: user_id or anonymous_id is required", ErrInvalidInteraction)
	}
	if strings.TrimSpace(i.ProductID) == "" {
		return fmt.Errorf("%w: product_id is required", ErrInvalidInteraction)
	}
	if i.Quantity <= 0 {
		return fmt.Errorf("%w: quantity must be greater than zero", ErrInvalidInteraction)
	}
	if i.OccurredAt.IsZero() {
		return fmt.Errorf("%w: occurred_at is required", ErrInvalidInteraction)
	}
	if i.ReceivedAt.IsZero() {
		return fmt.Errorf("%w: received_at is required", ErrInvalidInteraction)
	}
	return nil
}
