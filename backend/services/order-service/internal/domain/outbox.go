package domain

import (
	"fmt"
	"strings"
	"time"
)

type OutboxStatus string

const (
	OutboxStatusPending    OutboxStatus = "pending"
	OutboxStatusProcessing OutboxStatus = "processing"
	OutboxStatusPublished  OutboxStatus = "published"
	OutboxStatusFailed     OutboxStatus = "failed"
	OutboxStatusDeadLetter OutboxStatus = "dead_letter"
)

type OutboxEvent struct {
	EventID          string
	EventType        string
	Version          int
	AggregateType    string
	AggregateID      string
	RoutingKey       string
	DeduplicationKey string
	SourceHistoryID  string
	Payload          []byte
	TraceID          string
	Status           OutboxStatus
	Attempts         int
	NextAttemptAt    *time.Time
	PublishedAt      *time.Time
	LockedAt         *time.Time
	LockedBy         string
	LastError        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (e OutboxEvent) ValidateForInsert() error {
	if strings.TrimSpace(e.EventID) == "" || len(strings.TrimSpace(e.EventID)) > maxPublicIDLength {
		return fmt.Errorf("%w: invalid event_id", ErrInvalidOrder)
	}
	if strings.TrimSpace(e.EventType) == "" || len(strings.TrimSpace(e.EventType)) > 80 {
		return fmt.Errorf("%w: invalid event_type", ErrInvalidOrder)
	}
	if e.Version <= 0 {
		return fmt.Errorf("%w: invalid event version", ErrInvalidOrder)
	}
	if strings.TrimSpace(e.AggregateType) == "" || len(strings.TrimSpace(e.AggregateType)) > 80 {
		return fmt.Errorf("%w: invalid aggregate_type", ErrInvalidOrder)
	}
	if strings.TrimSpace(e.AggregateID) == "" || len(strings.TrimSpace(e.AggregateID)) > maxPublicIDLength {
		return fmt.Errorf("%w: invalid aggregate_id", ErrInvalidOrder)
	}
	if strings.TrimSpace(e.RoutingKey) == "" || len(strings.TrimSpace(e.RoutingKey)) > 120 {
		return fmt.Errorf("%w: invalid routing_key", ErrInvalidOrder)
	}
	if strings.TrimSpace(e.DeduplicationKey) == "" || len(strings.TrimSpace(e.DeduplicationKey)) > 160 {
		return fmt.Errorf("%w: invalid deduplication_key", ErrInvalidOrder)
	}
	if len(strings.TrimSpace(e.SourceHistoryID)) > maxPublicIDLength {
		return fmt.Errorf("%w: invalid source_history_id", ErrInvalidOrder)
	}
	if len(e.Payload) == 0 {
		return fmt.Errorf("%w: outbox payload is required", ErrInvalidOrder)
	}
	if len(strings.TrimSpace(e.TraceID)) > 128 {
		return fmt.Errorf("%w: invalid trace_id", ErrInvalidOrder)
	}
	if e.Status != "" && !IsKnownOutboxStatus(e.Status) {
		return fmt.Errorf("%w: unknown outbox status", ErrInvalidOrder)
	}
	if e.CreatedAt.IsZero() {
		return fmt.Errorf("%w: outbox created_at is required", ErrInvalidOrder)
	}
	return nil
}

func IsKnownOutboxStatus(status OutboxStatus) bool {
	switch status {
	case OutboxStatusPending,
		OutboxStatusProcessing,
		OutboxStatusPublished,
		OutboxStatusFailed,
		OutboxStatusDeadLetter:
		return true
	default:
		return false
	}
}
