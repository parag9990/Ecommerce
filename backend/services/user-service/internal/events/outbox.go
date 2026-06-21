package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
)

const (
	AggregateUser    = "user"
	AggregateSeller  = "seller"
	AggregateAddress = "address"

	StatusPending   = "pending"
	StatusPublished = "published"
	StatusFailed    = "failed"
	StatusDead      = "dead"
)

type Clock interface {
	Now() time.Time
}

type Metrics interface {
	RecordEnqueued(eventType string)
	ObservePublish(eventType string, result string, duration time.Duration)
}

type OutboxEvent struct {
	ID            int64
	EventID       string
	EventType     string
	EventVersion  int
	Topic         string
	AggregateType string
	AggregateID   string
	Payload       []byte
	RequestID     string
	TraceID       string
	Status        string
	Attempts      int
	OccurredAt    time.Time
}

type OutboxFailure struct {
	LastError     string
	NextAttemptAt time.Time
	Dead          bool
	FailedAt      time.Time
}

type OutboxStats struct {
	Pending          int64
	OldestPendingAge time.Duration
}

type OutboxStatsRepository interface {
	Stats(ctx context.Context) (OutboxStats, error)
}

type OutboxRepository interface {
	Insert(ctx context.Context, event OutboxEvent) error
	LockPending(ctx context.Context, batchSize int, lockTTL time.Duration) ([]OutboxEvent, error)
	MarkPublished(ctx context.Context, eventID string, publishedAt time.Time) error
	MarkFailed(ctx context.Context, eventID string, failure OutboxFailure) error
}

type RecorderConfig struct {
	Enabled bool
	Topic   string
	Clock   Clock
	Logger  *slog.Logger
	Metrics Metrics
}

type OutboxRecorder struct {
	repo    OutboxRepository
	enabled bool
	topic   string
	clock   Clock
	logger  *slog.Logger
	metrics Metrics
}

func NewOutboxRecorder(repo OutboxRepository, cfg RecorderConfig) (*OutboxRecorder, error) {
	if repo == nil {
		return nil, errors.New("outbox repository is required")
	}
	topic := strings.TrimSpace(cfg.Topic)
	if topic == "" {
		return nil, errors.New("events topic is required")
	}
	clock := cfg.Clock
	if clock == nil {
		clock = systemClock{}
	}
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	metrics := cfg.Metrics
	if metrics == nil {
		metrics = noopMetrics{}
	}

	return &OutboxRecorder{
		repo:    repo,
		enabled: cfg.Enabled,
		topic:   topic,
		clock:   clock,
		logger:  logger,
		metrics: metrics,
	}, nil
}

func (r *OutboxRecorder) RecordUserCreated(ctx context.Context, payload domain.UserCreatedPayload) error {
	return r.record(ctx, domain.EventUserCreated, AggregateUser, payload.UserID, payload)
}

func (r *OutboxRecorder) RecordSellerApproved(ctx context.Context, payload domain.SellerApprovedPayload) error {
	return r.record(ctx, domain.EventSellerApproved, AggregateSeller, payload.SellerID, payload)
}

func (r *OutboxRecorder) RecordAddressUpdated(ctx context.Context, payload domain.AddressUpdatedPayload) error {
	return r.record(ctx, domain.EventAddressUpdated, AggregateAddress, payload.AddressID, payload)
}

func (r *OutboxRecorder) record(ctx context.Context, eventType string, aggregateType string, aggregateID string, payload any) error {
	if !r.enabled {
		return nil
	}

	occurredAt := r.clock.Now().UTC()
	envelope, err := NewEnvelope(eventType, aggregateID, requestIDFromContext(ctx), traceIDFromContext(ctx), occurredAt, payload)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal event envelope: %w", err)
	}

	row := OutboxEvent{
		EventID:       envelope.EventID,
		EventType:     envelope.EventType,
		EventVersion:  envelope.Version,
		Topic:         r.topic,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		Payload:       raw,
		RequestID:     envelope.RequestID,
		TraceID:       envelope.TraceID,
		Status:        StatusPending,
		OccurredAt:    envelope.OccurredAt,
	}
	if err := r.repo.Insert(ctx, row); err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}
	r.metrics.RecordEnqueued(envelope.EventType)

	r.logger.InfoContext(ctx, "user_event_enqueued",
		slog.String("event_type", envelope.EventType),
		slog.String("event_id", envelope.EventID),
		slog.String("aggregate_type", aggregateType),
		slog.String("aggregate_id", aggregateID),
		slog.String("request_id", envelope.RequestID),
		slog.String("trace_id", envelope.TraceID),
	)
	return nil
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now().UTC()
}

type noopMetrics struct{}

func (noopMetrics) RecordEnqueued(string) {}

func (noopMetrics) ObservePublish(string, string, time.Duration) {}
