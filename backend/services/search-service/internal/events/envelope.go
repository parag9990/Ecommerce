package events

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

const SupportedProductEventVersion = 1

type Envelope[T any] struct {
	EventID       string    `json:"event_id"`
	EventType     string    `json:"event_type"`
	Version       int       `json:"version"`
	Producer      string    `json:"producer"`
	Source        string    `json:"source,omitempty"`
	RequestID     string    `json:"request_id"`
	TraceID       string    `json:"trace_id"`
	CorrelationID string    `json:"correlation_id"`
	OccurredAt    time.Time `json:"occurred_at"`
	Payload       T         `json:"payload"`
}

func DecodeProductEnvelope(raw []byte) (Envelope[ProductIndexPayload], error) {
	var envelope Envelope[ProductIndexPayload]
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return Envelope[ProductIndexPayload]{}, fmt.Errorf("%w: decode product envelope: %v", domain.ErrInvalidProductEvent, err)
	}
	return envelope, nil
}

func (e Envelope[T]) ProducerName() string {
	if strings.TrimSpace(e.Producer) != "" {
		return strings.TrimSpace(e.Producer)
	}
	return strings.TrimSpace(e.Source)
}

func (e Envelope[T]) ValidateMetadata() error {
	if strings.TrimSpace(e.EventID) == "" {
		return fmt.Errorf("%w: event_id is required", domain.ErrInvalidProductEvent)
	}
	if strings.TrimSpace(e.EventType) == "" {
		return fmt.Errorf("%w: event_type is required", domain.ErrInvalidProductEvent)
	}
	if !domain.IsSupportedProductEvent(e.EventType) {
		return fmt.Errorf("%w: %s", domain.ErrUnsupportedProductEvent, e.EventType)
	}
	if e.Version != SupportedProductEventVersion {
		return fmt.Errorf("%w: version %d is not supported", domain.ErrInvalidProductEvent, e.Version)
	}
	if e.ProducerName() == "" {
		return fmt.Errorf("%w: producer is required", domain.ErrInvalidProductEvent)
	}
	if strings.TrimSpace(e.RequestID) == "" {
		return fmt.Errorf("%w: request_id is required", domain.ErrInvalidProductEvent)
	}
	if strings.TrimSpace(e.TraceID) == "" {
		return fmt.Errorf("%w: trace_id is required", domain.ErrInvalidProductEvent)
	}
	if strings.TrimSpace(e.CorrelationID) == "" {
		return fmt.Errorf("%w: correlation_id is required", domain.ErrInvalidProductEvent)
	}
	if e.OccurredAt.IsZero() {
		return fmt.Errorf("%w: occurred_at is required", domain.ErrInvalidProductEvent)
	}
	return nil
}
