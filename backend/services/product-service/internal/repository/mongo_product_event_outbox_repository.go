package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"product-service/internal/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	defaultOutboxListLimit       = 100
	defaultOutboxPublishingLease = 5 * time.Minute
	maxStoredOutboxErrorLength   = 2000
)

func (r *MongoProductRepository) InsertOutboxEvent(ctx context.Context, event *domain.ProductOutboxEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.insertOutboxEventDocument(ctx, event)
}

func (r *MongoProductRepository) ListPendingOutboxEvents(
	ctx context.Context,
	limit int,
	now time.Time,
) ([]domain.ProductOutboxEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = defaultOutboxListLimit
	}
	now = now.UTC()
	cursor, err := r.outbox.Find(
		ctx,
		bson.D{e("$or", bson.A{
			bson.D{
				e("status", string(domain.OutboxStatusPending)),
				e("next_attempt_at", bson.D{e("$lte", now)}),
			},
			bson.D{
				e("status", string(domain.OutboxStatusPublishing)),
				e("next_attempt_at", bson.D{e("$lte", now)}),
			},
		})},
		options.Find().
			SetSort(bson.D{e("next_attempt_at", 1), e("occurred_at", 1), e("_id", 1)}).
			SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, fmt.Errorf("list pending product outbox events: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []productOutboxEventDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode pending product outbox events: %w", err)
	}
	events := make([]domain.ProductOutboxEvent, 0, len(docs))
	for _, doc := range docs {
		events = append(events, doc.toDomain())
	}
	return events, nil
}

func (r *MongoProductRepository) MarkOutboxEventPublishing(ctx context.Context, eventID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	now := time.Now().UTC()
	return r.markOutboxEvent(
		ctx,
		bson.D{
			e("_id", strings.TrimSpace(eventID)),
			e("$or", bson.A{
				bson.D{e("status", string(domain.OutboxStatusPending))},
				bson.D{
					e("status", string(domain.OutboxStatusPublishing)),
					e("next_attempt_at", bson.D{e("$lte", now)}),
				},
			}),
		},
		bson.D{
			e("$set", bson.D{
				e("status", string(domain.OutboxStatusPublishing)),
				e("next_attempt_at", now.Add(defaultOutboxPublishingLease)),
				e("last_error", ""),
			}),
			e("$inc", bson.D{e("attempts", 1)}),
		},
		"mark product outbox event publishing",
	)
}

func (r *MongoProductRepository) MarkOutboxEventPublished(
	ctx context.Context,
	eventID string,
	publishedAt time.Time,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	publishedAt = publishedAt.UTC()
	return r.markOutboxEvent(
		ctx,
		bson.D{
			e("_id", strings.TrimSpace(eventID)),
			e("status", string(domain.OutboxStatusPublishing)),
		},
		bson.D{e("$set", bson.D{
			e("status", string(domain.OutboxStatusPublished)),
			e("published_at", publishedAt),
			e("next_attempt_at", publishedAt),
			e("last_error", ""),
		})},
		"mark product outbox event published",
	)
}

func (r *MongoProductRepository) MarkOutboxEventFailed(
	ctx context.Context,
	eventID string,
	nextAttemptAt time.Time,
	lastError string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.markOutboxEvent(
		ctx,
		bson.D{
			e("_id", strings.TrimSpace(eventID)),
			e("status", string(domain.OutboxStatusPublishing)),
		},
		bson.D{e("$set", bson.D{
			e("status", string(domain.OutboxStatusPending)),
			e("next_attempt_at", nextAttemptAt.UTC()),
			e("last_error", truncateOutboxError(lastError)),
		})},
		"mark product outbox event failed",
	)
}

func (r *MongoProductRepository) MarkOutboxEventDeadLettered(
	ctx context.Context,
	eventID string,
	lastError string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.markOutboxEvent(
		ctx,
		bson.D{
			e("_id", strings.TrimSpace(eventID)),
			e("status", string(domain.OutboxStatusPublishing)),
		},
		bson.D{e("$set", bson.D{
			e("status", string(domain.OutboxStatusDeadLettered)),
			e("last_error", truncateOutboxError(lastError)),
		})},
		"mark product outbox event dead-lettered",
	)
}

func (r *MongoProductRepository) insertOutboxEventDocument(
	ctx context.Context,
	event *domain.ProductOutboxEvent,
) error {
	if event == nil {
		return fmt.Errorf("outbox event is required")
	}
	if report := event.Validate(); report.HasErrors() {
		return domain.ValidationError{Report: report}
	}
	_, err := r.outbox.InsertOne(ctx, productOutboxEventDocumentFromDomain(*event))
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicateKey
	}
	if err != nil {
		return fmt.Errorf("insert product outbox event %q: %w", event.ID, err)
	}
	return nil
}

func (r *MongoProductRepository) markOutboxEvent(
	ctx context.Context,
	filter bson.D,
	update bson.D,
	operation string,
) error {
	result, err := r.outbox.UpdateOne(ctx, filter, update)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	if result.MatchedCount == 0 {
		return ErrWriteConflict
	}
	return nil
}

type productOutboxEventDocument struct {
	ID            string         `bson:"_id"`
	Topic         string         `bson:"topic"`
	EventType     string         `bson:"event_type"`
	Version       int            `bson:"version"`
	Source        string         `bson:"source"`
	RequestID     string         `bson:"request_id"`
	TraceID       string         `bson:"trace_id"`
	Payload       map[string]any `bson:"payload"`
	Status        string         `bson:"status"`
	Attempts      int            `bson:"attempts"`
	NextAttemptAt time.Time      `bson:"next_attempt_at"`
	OccurredAt    time.Time      `bson:"occurred_at"`
	PublishedAt   *time.Time     `bson:"published_at,omitempty"`
	LastError     string         `bson:"last_error,omitempty"`
}

func productOutboxEventDocumentFromDomain(event domain.ProductOutboxEvent) productOutboxEventDocument {
	payload := make(map[string]any, len(event.Payload))
	for key, value := range event.Payload {
		payload[key] = value
	}
	return productOutboxEventDocument{
		ID:            event.ID,
		Topic:         event.Topic,
		EventType:     event.EventType,
		Version:       event.Version,
		Source:        event.Source,
		RequestID:     event.RequestID,
		TraceID:       event.TraceID,
		Payload:       payload,
		Status:        string(event.Status),
		Attempts:      event.Attempts,
		NextAttemptAt: event.NextAttemptAt.UTC(),
		OccurredAt:    event.OccurredAt.UTC(),
		PublishedAt:   event.PublishedAt,
		LastError:     event.LastError,
	}
}

func (d productOutboxEventDocument) toDomain() domain.ProductOutboxEvent {
	payload := make(map[string]any, len(d.Payload))
	for key, value := range d.Payload {
		payload[key] = value
	}
	return domain.ProductOutboxEvent{
		ID:            d.ID,
		Topic:         d.Topic,
		EventType:     d.EventType,
		Version:       d.Version,
		Source:        d.Source,
		RequestID:     d.RequestID,
		TraceID:       d.TraceID,
		Payload:       payload,
		Status:        domain.ProductOutboxStatus(d.Status),
		Attempts:      d.Attempts,
		NextAttemptAt: d.NextAttemptAt,
		OccurredAt:    d.OccurredAt,
		PublishedAt:   d.PublishedAt,
		LastError:     d.LastError,
	}
}

func truncateOutboxError(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= maxStoredOutboxErrorLength {
		return value
	}
	return value[:maxStoredOutboxErrorLength]
}
