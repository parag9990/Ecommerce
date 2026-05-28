package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

const DefaultSupportedEventVersion = 1

type Envelope struct {
	EventID    string                 `json:"event_id"`
	EventType  domain.SourceEventType `json:"event_type"`
	Version    int                    `json:"version"`
	OccurredAt time.Time              `json:"occurred_at"`
	Producer   string                 `json:"producer"`
	TraceID    string                 `json:"trace_id,omitempty"`
	Payload    json.RawMessage        `json:"payload"`
}

func DecodeEnvelope(raw []byte) (Envelope, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()

	var envelope Envelope
	if err := decoder.Decode(&envelope); err != nil {
		return Envelope{}, fmt.Errorf("%w: invalid json: %v", domain.ErrInvalidEventEnvelope, err)
	}
	var extra struct{}
	if err := decoder.Decode(&extra); err != io.EOF {
		return Envelope{}, fmt.Errorf("%w: multiple json values are not allowed", domain.ErrInvalidEventEnvelope)
	}
	return envelope.Normalize(), nil
}

func (e Envelope) Normalize() Envelope {
	e.EventID = strings.TrimSpace(e.EventID)
	e.EventType = domain.SourceEventType(strings.TrimSpace(string(e.EventType)))
	e.Producer = strings.TrimSpace(e.Producer)
	e.TraceID = strings.TrimSpace(e.TraceID)
	return e
}

func (e Envelope) Validate(supportedVersion int) error {
	e = e.Normalize()
	if supportedVersion <= 0 {
		supportedVersion = DefaultSupportedEventVersion
	}
	if e.EventID == "" {
		return fmt.Errorf("%w: event_id is required", domain.ErrInvalidEventEnvelope)
	}
	if e.EventType == "" {
		return fmt.Errorf("%w: event_type is required", domain.ErrInvalidEventEnvelope)
	}
	if e.Version != supportedVersion {
		return fmt.Errorf("%w: unsupported event version %d", domain.ErrInvalidEventEnvelope, e.Version)
	}
	if e.OccurredAt.IsZero() {
		return fmt.Errorf("%w: occurred_at is required", domain.ErrInvalidEventEnvelope)
	}
	if e.Producer == "" {
		return fmt.Errorf("%w: producer is required", domain.ErrInvalidEventEnvelope)
	}
	if len(e.Payload) == 0 || bytes.Equal(bytes.TrimSpace(e.Payload), []byte("null")) {
		return fmt.Errorf("%w: payload is required", domain.ErrInvalidEventEnvelope)
	}
	if !e.EventType.IsSupported() {
		return fmt.Errorf("%w: %s", domain.ErrUnsupportedEventType, e.EventType)
	}
	return nil
}
