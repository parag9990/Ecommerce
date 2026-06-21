package events

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Envelope struct {
	ID            string          `json:"event_id"`
	Type          string          `json:"event_type"`
	Source        string          `json:"producer"`
	AggregateID   string          `json:"aggregate_id,omitempty"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	TraceID       string          `json:"trace_id,omitempty"`
	RequestID     string          `json:"request_id,omitempty"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Version       int             `json:"version"`
	Data          json.RawMessage `json:"payload"`
}

func (e Envelope) Validate() error {
	var errs []error
	if strings.TrimSpace(e.ID) == "" {
		errs = append(errs, errors.New("event_id is required"))
	}
	if strings.TrimSpace(e.Type) == "" {
		errs = append(errs, errors.New("event_type is required"))
	}
	if strings.TrimSpace(e.Source) == "" {
		errs = append(errs, errors.New("producer is required"))
	}
	if e.Version < 1 {
		errs = append(errs, errors.New("version must be positive"))
	}
	if e.OccurredAt.IsZero() {
		errs = append(errs, errors.New("occurred_at is required"))
	}
	if !json.Valid(e.Data) {
		errs = append(errs, errors.New("payload must be valid JSON"))
	}
	return errors.Join(errs...)
}

type Publisher interface {
	Publish(context.Context, string, Envelope) error
}
type Consumer interface {
	Handle(context.Context, Envelope) error
}
