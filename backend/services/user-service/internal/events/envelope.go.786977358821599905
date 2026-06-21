package events

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	SourceUserService   = "user-service"
	EventSchemaVersion1 = 1
)

type Envelope[T any] struct {
	EventID     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
	Version     int       `json:"version"`
	Source      string    `json:"source"`
	RequestID   string    `json:"request_id,omitempty"`
	TraceID     string    `json:"trace_id,omitempty"`
	AggregateID string    `json:"aggregate_id"`
	OccurredAt  time.Time `json:"occurred_at"`
	Payload     T         `json:"payload"`
}

func NewEnvelope[T any](eventType string, aggregateID string, requestID string, traceID string, occurredAt time.Time, payload T) (Envelope[T], error) {
	eventID, err := newEventID()
	if err != nil {
		return Envelope[T]{}, err
	}

	return Envelope[T]{
		EventID:     eventID,
		EventType:   eventType,
		Version:     EventSchemaVersion1,
		Source:      SourceUserService,
		RequestID:   requestID,
		TraceID:     traceID,
		AggregateID: aggregateID,
		OccurredAt:  occurredAt.UTC(),
		Payload:     payload,
	}, nil
}

func newEventID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate event id: %w", err)
	}
	return "evt_" + hex.EncodeToString(raw[:]), nil
}
