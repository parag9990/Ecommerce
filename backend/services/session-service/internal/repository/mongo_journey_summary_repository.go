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

const journeySummariesCollectionName = "journey_summaries"
const defaultJourneySummaryRetentionDays = 365

type MongoJourneySummaryRepository struct {
	collection    *mongo.Collection
	validation    domain.JourneySummaryValidationConfig
	retentionDays int
	logger        *slog.Logger
}

type mongoJourneySummaryDocument struct {
	ID                    string `bson:"_id"`
	domain.JourneySummary `bson:",inline"`
}

type MongoJourneySummaryRepositoryOption func(*MongoJourneySummaryRepository)

func WithJourneySummaryRetentionDays(days int) MongoJourneySummaryRepositoryOption {
	return func(r *MongoJourneySummaryRepository) {
		if days > 0 {
			r.retentionDays = days
		}
	}
}

func NewMongoJourneySummaryRepository(database *mongo.Database, validation domain.JourneySummaryValidationConfig, logger *slog.Logger, opts ...MongoJourneySummaryRepositoryOption) (*MongoJourneySummaryRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	repo := &MongoJourneySummaryRepository{
		collection:    database.Collection(journeySummariesCollectionName),
		validation:    validation.WithDefaults(),
		retentionDays: defaultJourneySummaryRetentionDays,
		logger:        logger,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(repo)
		}
	}
	return repo, nil
}

func (r *MongoJourneySummaryRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_journey_session"),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "last_event_at", Value: -1}},
			Options: options.Index().
				SetName("idx_journey_user_recent").
				SetPartialFilterExpression(bson.D{{Key: "user_id", Value: bson.D{{Key: "$type", Value: "string"}}}}),
		},
		{
			Keys:    bson.D{{Key: "last_event_at", Value: -1}},
			Options: options.Index().SetName("idx_journey_recent"),
		},
		{
			Keys:    bson.D{{Key: "retain_until", Value: 1}, {Key: "legal_hold", Value: 1}, {Key: "anonymized_at", Value: 1}},
			Options: options.Index().SetName("idx_journey_retention_due"),
		},
	}
	if _, err := r.collection.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("create journey summary indexes: %w", err)
	}
	r.logger.InfoContext(ctx, "session.mongo.journey_summary_indexes_ready")
	return nil
}

func (r *MongoJourneySummaryRepository) UpsertJourneySummary(ctx context.Context, summary domain.JourneySummary) error {
	normalized := summary.Normalize()
	if normalized.SchemaVersion == 0 {
		normalized.SchemaVersion = domain.CurrentJourneySummarySchemaVersion
	}
	now := time.Now().UTC()
	if normalized.CalculatedAt.IsZero() {
		normalized.CalculatedAt = now
	}
	if normalized.UpdatedAt.IsZero() {
		normalized.UpdatedAt = normalized.CalculatedAt
	}
	normalized = r.applyRetentionDefaults(normalized)
	if err := normalized.ValidateWithConfig(r.validation); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInvalidJourney, err)
	}

	set, unset := journeySummarySet(normalized)
	createdAt := normalized.CreatedAt
	if createdAt.IsZero() {
		createdAt = normalized.CalculatedAt
	}
	update := bson.D{
		{Key: "$set", Value: set},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: "journey_" + normalized.SessionID},
			{Key: "session_id", Value: normalized.SessionID},
			{Key: "created_at", Value: createdAt},
			{Key: "legal_hold", Value: false},
		}},
	}
	if len(unset) > 0 {
		update = append(update, bson.E{Key: "$unset", Value: unset})
	}

	_, err := r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "session_id", Value: normalized.SessionID}},
		update,
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: duplicate journey summary session_id", domain.ErrInvalidJourney)
		}
		return fmt.Errorf("upsert journey summary: %w", err)
	}
	return nil
}

func (r *MongoJourneySummaryRepository) applyRetentionDefaults(summary domain.JourneySummary) domain.JourneySummary {
	out := summary.Normalize()
	if out.RetentionClass == "" {
		out.RetentionClass = domain.RetentionClassDerivedSummary
	}
	if out.RetainUntil == nil {
		base := out.LastEventAt
		if base.IsZero() {
			base = out.UpdatedAt
		}
		if base.IsZero() {
			base = out.CalculatedAt
		}
		if base.IsZero() {
			base = time.Now().UTC()
		}
		retainUntil := base.UTC().AddDate(0, 0, r.retentionDays)
		out.RetainUntil = &retainUntil
	}
	return out
}

func (r *MongoJourneySummaryRepository) FindJourneySummaryBySessionID(ctx context.Context, sessionID string) (domain.JourneySummary, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return domain.JourneySummary{}, fmt.Errorf("%w: session_id is required", domain.ErrInvalidJourney)
	}

	var doc mongoJourneySummaryDocument
	err := r.collection.FindOne(ctx, bson.D{{Key: "session_id", Value: sessionID}}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.JourneySummary{}, domain.ErrJourneyNotFound
		}
		return domain.JourneySummary{}, fmt.Errorf("find journey summary by session id: %w", err)
	}
	return doc.JourneySummary.Normalize(), nil
}

func journeySummarySet(summary domain.JourneySummary) (bson.D, bson.D) {
	milestones := summary.Milestones
	if milestones == nil {
		milestones = []domain.JourneyMilestone{}
	}
	topPaths := summary.TopPaths
	if topPaths == nil {
		topPaths = []domain.JourneyPathCount{}
	}

	set := bson.D{
		{Key: "anonymous_id", Value: summary.AnonymousID},
		{Key: "duration_seconds", Value: summary.DurationSeconds},
		{Key: "total_events", Value: summary.TotalEvents},
		{Key: "products_viewed", Value: summary.ProductsViewed},
		{Key: "searches", Value: summary.Searches},
		{Key: "cart_actions", Value: summary.CartActions},
		{Key: "checkout_started", Value: summary.CheckoutStarted},
		{Key: "payment_completed", Value: summary.PaymentCompleted},
		{Key: "milestones", Value: milestones},
		{Key: "top_paths", Value: topPaths},
		{Key: "schema_version", Value: summary.SchemaVersion},
		{Key: "calculated_at", Value: summary.CalculatedAt},
		{Key: "updated_at", Value: summary.UpdatedAt},
		{Key: "retain_until", Value: summary.RetainUntil},
		{Key: "retention_class", Value: string(summary.RetentionClass)},
	}
	unset := bson.D{}

	if summary.UserID != nil {
		set = append(set, bson.E{Key: "user_id", Value: *summary.UserID})
	} else {
		unset = append(unset, bson.E{Key: "user_id", Value: ""})
	}
	if summary.EntryPage != "" {
		set = append(set, bson.E{Key: "entry_page", Value: summary.EntryPage})
	} else {
		unset = append(unset, bson.E{Key: "entry_page", Value: ""})
	}
	if summary.ExitPage != "" {
		set = append(set, bson.E{Key: "exit_page", Value: summary.ExitPage})
	} else {
		unset = append(unset, bson.E{Key: "exit_page", Value: ""})
	}
	if !summary.FirstEventAt.IsZero() {
		set = append(set, bson.E{Key: "first_event_at", Value: summary.FirstEventAt})
	} else {
		unset = append(unset, bson.E{Key: "first_event_at", Value: ""})
	}
	if !summary.LastEventAt.IsZero() {
		set = append(set, bson.E{Key: "last_event_at", Value: summary.LastEventAt})
	} else {
		unset = append(unset, bson.E{Key: "last_event_at", Value: ""})
	}
	return set, unset
}
