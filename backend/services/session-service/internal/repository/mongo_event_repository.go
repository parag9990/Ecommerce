package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	sessionEventsCollectionName = "session_events"
	defaultRawEventTTLDays      = 90
)

type MongoEventRepository struct {
	collection      *mongo.Collection
	validation      domain.SessionEventValidationConfig
	rawEventTTLDays int
	logger          *slog.Logger
}

type mongoSessionEventDocument struct {
	ID                  string `bson:"_id"`
	domain.SessionEvent `bson:",inline"`
}

func NewMongoEventRepository(database *mongo.Database, validation domain.SessionEventValidationConfig, rawEventTTLDays int, logger *slog.Logger) (*MongoEventRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	if rawEventTTLDays <= 0 {
		rawEventTTLDays = defaultRawEventTTLDays
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &MongoEventRepository{
		collection:      database.Collection(sessionEventsCollectionName),
		validation:      validation.WithDefaults(),
		rawEventTTLDays: rawEventTTLDays,
		logger:          logger,
	}, nil
}

func (r *MongoEventRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}, {Key: "occurred_at", Value: 1}},
			Options: options.Index().SetName("idx_session_timeline"),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "occurred_at", Value: -1}},
			Options: options.Index().
				SetName("idx_user_event_history").
				SetPartialFilterExpression(bson.D{{Key: "user_id", Value: bson.D{{Key: "$type", Value: "string"}}}}),
		},
		{
			Keys:    bson.D{{Key: "event_type", Value: 1}, {Key: "occurred_at", Value: -1}},
			Options: options.Index().SetName("idx_event_type_time"),
		},
		{
			Keys:    bson.D{{Key: "event_type", Value: 1}, {Key: "occurred_at", Value: 1}, {Key: "path", Value: 1}},
			Options: options.Index().SetName("idx_events_heatmap_scan"),
		},
		{
			Keys:    bson.D{{Key: "device.type", Value: 1}, {Key: "event_type", Value: 1}, {Key: "occurred_at", Value: -1}},
			Options: options.Index().SetName("idx_events_device_event_time"),
		},
		{
			Keys: bson.D{{Key: "request_id", Value: 1}},
			Options: options.Index().
				SetName("idx_session_events_request_id").
				SetPartialFilterExpression(bson.D{{Key: "request_id", Value: bson.D{{Key: "$type", Value: "string"}}}}),
		},
		{
			Keys: bson.D{{Key: "occurred_at", Value: 1}},
			Options: options.Index().
				SetName(fmt.Sprintf("ttl_raw_events_%d_days", r.rawEventTTLDays)).
				SetExpireAfterSeconds(int32(r.rawEventTTLDays * 24 * 60 * 60)),
		},
	}
	if _, err := r.collection.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("create session event indexes: %w", err)
	}
	r.logger.InfoContext(ctx, "session.mongo.event_indexes_ready",
		slog.Int("raw_event_ttl_days", r.rawEventTTLDays),
	)
	return nil
}

func (r *MongoEventRepository) InsertEvent(ctx context.Context, event domain.SessionEvent) error {
	normalized := event.Normalize()
	if normalized.SchemaVersion == 0 {
		normalized.SchemaVersion = domain.CurrentSessionEventSchemaVersion
	}
	if normalized.RetentionClass == "" {
		normalized.RetentionClass = domain.RetentionClassRawEvent
	}
	if normalized.RetainUntil == nil && !normalized.OccurredAt.IsZero() {
		retainUntil := normalized.OccurredAt.AddDate(0, 0, r.rawEventTTLDays)
		normalized.RetainUntil = &retainUntil
	}
	if err := normalized.ValidateWithConfig(r.validation); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInvalidSessionEvent, err)
	}

	doc := mongoSessionEventDocument{
		ID:           normalized.EventID,
		SessionEvent: normalized,
	}
	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: duplicate event_id", domain.ErrInvalidSessionEvent)
		}
		return fmt.Errorf("insert session event: %w", err)
	}
	return nil
}

func (r *MongoEventRepository) ListEventsBySession(ctx context.Context, sessionID string, limit int) ([]domain.SessionEvent, error) {
	return r.ListEventsBySessionTimeline(ctx, sessionID, limit, nil)
}

func (r *MongoEventRepository) ListEventsBySessionTimeline(ctx context.Context, sessionID string, limit int, cursorOccurredAt *time.Time) ([]domain.SessionEvent, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("%w: session_id is required", domain.ErrInvalidSessionEvent)
	}
	if limit <= 0 {
		limit = 100
	}

	filter := bson.D{{Key: "session_id", Value: sessionID}}
	if cursorOccurredAt != nil {
		cursor := cursorOccurredAt.UTC()
		if !cursor.IsZero() {
			filter = append(filter, bson.E{Key: "occurred_at", Value: bson.D{{Key: "$gt", Value: cursor}}})
		}
	}
	cursor, err := r.collection.Find(
		ctx,
		filter,
		options.Find().
			SetSort(bson.D{
				{Key: "occurred_at", Value: 1},
				{Key: "received_at", Value: 1},
				{Key: "_id", Value: 1},
			}).
			SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, fmt.Errorf("list session events: %w", err)
	}
	defer cursor.Close(ctx)

	events := make([]domain.SessionEvent, 0)
	for cursor.Next(ctx) {
		var doc mongoSessionEventDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode session event: %w", err)
		}
		events = append(events, doc.SessionEvent.Normalize())
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate session events: %w", err)
	}
	return events, nil
}

func (r *MongoEventRepository) ListHeatmapEvents(ctx context.Context, from time.Time, to time.Time, limit int, cursorAfter *domain.HeatmapEventCursor) ([]domain.SessionEvent, error) {
	from = from.UTC()
	to = to.UTC()
	if from.IsZero() || to.IsZero() || !from.Before(to) {
		return nil, fmt.Errorf("%w: invalid heatmap event window", domain.ErrInvalidHeatmap)
	}
	if limit <= 0 {
		limit = 1000
	}

	rangeFilter := bson.D{
		{Key: "event_type", Value: bson.D{{Key: "$in", Value: bson.A{string(domain.EventClick), string(domain.EventScroll)}}}},
		{Key: "occurred_at", Value: bson.D{{Key: "$gte", Value: from}, {Key: "$lt", Value: to}}},
	}
	if cursorAfter != nil && !cursorAfter.OccurredAt.IsZero() {
		cursorOccurredAt := cursorAfter.OccurredAt.UTC()
		cursorEventID := strings.TrimSpace(cursorAfter.EventID)
		rangeFilter = append(rangeFilter, bson.E{Key: "$or", Value: bson.A{
			bson.D{{Key: "occurred_at", Value: bson.D{{Key: "$gt", Value: cursorOccurredAt}}}},
			bson.D{
				{Key: "occurred_at", Value: cursorOccurredAt},
				{Key: "event_id", Value: bson.D{{Key: "$gt", Value: cursorEventID}}},
			},
		}})
	}

	cursor, err := r.collection.Find(
		ctx,
		rangeFilter,
		options.Find().
			SetSort(bson.D{{Key: "occurred_at", Value: 1}, {Key: "event_id", Value: 1}}).
			SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, fmt.Errorf("list heatmap events: %w", err)
	}
	defer cursor.Close(ctx)

	events := make([]domain.SessionEvent, 0)
	for cursor.Next(ctx) {
		var doc mongoSessionEventDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode heatmap event: %w", err)
		}
		events = append(events, doc.SessionEvent.Normalize())
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate heatmap events: %w", err)
	}
	return events, nil
}
