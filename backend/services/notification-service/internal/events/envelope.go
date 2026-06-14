package events

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

var (
	ErrInvalidEvent           = errors.New("invalid notification source event")
	ErrNonRetryableProcessing = errors.New("non-retryable notification event processing failure")
	ErrTemporaryProcessing    = errors.New("temporary notification event processing failure")
)

// Envelope is the shared versioned boundary accepted from business-event exchanges.
type Envelope struct {
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	Version    int             `json:"version"`
	OccurredAt time.Time       `json:"occurred_at"`
	Producer   string          `json:"producer"`
	TraceID    string          `json:"trace_id"`
	Payload    json.RawMessage `json:"payload"`
}

func DecodeEnvelope(body []byte) (Envelope, error) {
	var envelope Envelope
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		return Envelope{}, fmt.Errorf("%w: decode envelope", ErrInvalidEvent)
	}
	if err := ensureJSONEnd(decoder); err != nil {
		return Envelope{}, fmt.Errorf("%w: decode envelope", ErrInvalidEvent)
	}
	for value, field := range map[string]string{
		envelope.EventID:   "event_id",
		envelope.EventType: "event_type",
		envelope.Producer:  "producer",
	} {
		if strings.TrimSpace(value) == "" {
			return Envelope{}, fmt.Errorf("%w: %s is required", ErrInvalidEvent, field)
		}
	}
	if envelope.OccurredAt.IsZero() {
		return Envelope{}, fmt.Errorf("%w: occurred_at is required", ErrInvalidEvent)
	}
	if envelope.Version != 1 {
		return Envelope{}, fmt.Errorf("%w: unsupported version %d", ErrInvalidEvent, envelope.Version)
	}
	if len(envelope.Payload) == 0 || bytes.Equal(bytes.TrimSpace(envelope.Payload), []byte("null")) {
		return Envelope{}, fmt.Errorf("%w: payload is required", ErrInvalidEvent)
	}
	return envelope, nil
}

func decodePayload(payload json.RawMessage, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return fmt.Errorf("%w: decode event payload", ErrInvalidEvent)
	}
	if err := ensureJSONEnd(decoder); err != nil {
		return fmt.Errorf("%w: decode event payload", ErrInvalidEvent)
	}
	return nil
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var extra struct{}
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("additional JSON value")
		}
		return err
	}
	return nil
}
