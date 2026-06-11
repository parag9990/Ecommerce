package events

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"product-service/internal/domain"
	"product-service/internal/repository"
)

func TestOutboxRelayPublishesPendingEvent(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	outbox := newRelayMemoryOutbox(validRelayOutboxEvent(now))
	publisher := &recordingPublisher{}
	relay, err := NewOutboxRelay(
		outbox,
		publisher,
		relayTestClock{now: now},
		slog.Default(),
		OutboxRelayOptions{BatchSize: 10, MaxAttempts: 5},
	)
	if err != nil {
		t.Fatalf("NewOutboxRelay returned error: %v", err)
	}

	if err := relay.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("published event count = %d, want 1", len(publisher.events))
	}
	if outbox.events["evt_1"].Status != domain.OutboxStatusPublished {
		t.Fatalf("status = %s, want published", outbox.events["evt_1"].Status)
	}
	if outbox.events["evt_1"].Attempts != 1 {
		t.Fatalf("attempts = %d, want 1", outbox.events["evt_1"].Attempts)
	}
}

func TestOutboxRelayDeadLettersAfterMaxAttempts(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	event := validRelayOutboxEvent(now)
	event.Attempts = 4
	outbox := newRelayMemoryOutbox(event)
	relay, err := NewOutboxRelay(
		outbox,
		failingPublisher{err: errors.New("broker down")},
		relayTestClock{now: now},
		slog.Default(),
		OutboxRelayOptions{BatchSize: 10, MaxAttempts: 5},
	)
	if err != nil {
		t.Fatalf("NewOutboxRelay returned error: %v", err)
	}

	if err := relay.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	got := outbox.events["evt_1"]
	if got.Status != domain.OutboxStatusDeadLettered {
		t.Fatalf("status = %s, want dead_lettered", got.Status)
	}
	if got.Attempts != 5 {
		t.Fatalf("attempts = %d, want 5", got.Attempts)
	}
}

type relayTestClock struct {
	now time.Time
}

func (c relayTestClock) Now() time.Time { return c.now }

type recordingPublisher struct {
	events []domain.EventEnvelope
}

func (p *recordingPublisher) Publish(ctx context.Context, topic string, envelope domain.EventEnvelope) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	p.events = append(p.events, envelope)
	return nil
}

type failingPublisher struct {
	err error
}

func (p failingPublisher) Publish(ctx context.Context, topic string, envelope domain.EventEnvelope) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return p.err
}

type relayMemoryOutbox struct {
	events map[string]domain.ProductOutboxEvent
}

func newRelayMemoryOutbox(events ...domain.ProductOutboxEvent) *relayMemoryOutbox {
	outbox := &relayMemoryOutbox{events: make(map[string]domain.ProductOutboxEvent, len(events))}
	for _, event := range events {
		outbox.events[event.ID] = event
	}
	return outbox
}

func (o *relayMemoryOutbox) InsertOutboxEvent(ctx context.Context, event *domain.ProductOutboxEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	o.events[event.ID] = *event
	return nil
}

func (o *relayMemoryOutbox) ListPendingOutboxEvents(
	ctx context.Context,
	limit int,
	now time.Time,
) ([]domain.ProductOutboxEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var result []domain.ProductOutboxEvent
	for _, event := range o.events {
		if len(result) == limit {
			break
		}
		if event.Status == domain.OutboxStatusPending && !event.NextAttemptAt.After(now) {
			result = append(result, event)
		}
	}
	return result, nil
}

func (o *relayMemoryOutbox) MarkOutboxEventPublishing(ctx context.Context, eventID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	event, ok := o.events[eventID]
	if !ok || event.Status != domain.OutboxStatusPending {
		return repository.ErrWriteConflict
	}
	event.Status = domain.OutboxStatusPublishing
	event.Attempts++
	o.events[eventID] = event
	return nil
}

func (o *relayMemoryOutbox) MarkOutboxEventPublished(
	ctx context.Context,
	eventID string,
	publishedAt time.Time,
) error {
	return o.setStatus(ctx, eventID, domain.OutboxStatusPublished)
}

func (o *relayMemoryOutbox) MarkOutboxEventFailed(
	ctx context.Context,
	eventID string,
	nextAttemptAt time.Time,
	lastError string,
) error {
	if err := o.setStatus(ctx, eventID, domain.OutboxStatusPending); err != nil {
		return err
	}
	event := o.events[eventID]
	event.NextAttemptAt = nextAttemptAt
	event.LastError = lastError
	o.events[eventID] = event
	return nil
}

func (o *relayMemoryOutbox) MarkOutboxEventDeadLettered(
	ctx context.Context,
	eventID string,
	lastError string,
) error {
	if err := o.setStatus(ctx, eventID, domain.OutboxStatusDeadLettered); err != nil {
		return err
	}
	event := o.events[eventID]
	event.LastError = lastError
	o.events[eventID] = event
	return nil
}

func (o *relayMemoryOutbox) setStatus(
	ctx context.Context,
	eventID string,
	status domain.ProductOutboxStatus,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	event, ok := o.events[eventID]
	if !ok {
		return repository.ErrNotFound
	}
	event.Status = status
	o.events[eventID] = event
	return nil
}

func validRelayOutboxEvent(now time.Time) domain.ProductOutboxEvent {
	return domain.ProductOutboxEvent{
		ID:            "evt_1",
		Topic:         domain.ProductEventDefaultTopic,
		EventType:     string(domain.ProductEventUpdated),
		Version:       domain.ProductEventSchemaVersion,
		Source:        domain.ProductEventDefaultSource,
		RequestID:     "req_1",
		TraceID:       "trace_1",
		Payload:       map[string]any{"product_id": "prod_1", "search_action": string(domain.SearchActionDelete), "updated_at": now},
		Status:        domain.OutboxStatusPending,
		NextAttemptAt: now,
		OccurredAt:    now,
	}
}
