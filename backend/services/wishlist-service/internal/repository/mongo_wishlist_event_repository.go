package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	DefaultWishlistEventCollectionName = "wishlist_events"
	WishlistEventPendingRetryIndexName = "idx_wishlist_events_pending_retry"
	WishlistEventTypeTimeIndexName     = "idx_wishlist_events_type_time"
	WishlistEventPublishedTTLIndexName = "ttl_wishlist_events_published_at"
	WishlistEventClaimLeaseIndexName   = "idx_wishlist_events_claim_lease"
	publishedWishlistEventTTLSeconds   = int32(30 * 24 * 60 * 60)
	defaultWishlistEventClaimBatchSize = 50
	maxWishlistEventLastErrorLength    = 2048
	defaultWishlistEventClaimLease     = 30 * time.Second
)

type MongoWishlistEventRepository struct {
	database       *mongo.Database
	collection     *mongo.Collection
	collectionName string
	logger         *slog.Logger
	workerID       string
	claimLease     time.Duration
}

func NewMongoWishlistEventRepository(database *mongo.Database, collectionName string, logger *slog.Logger) (*MongoWishlistEventRepository, error) {
	if database == nil {
		return nil, errors.New("wishlist event mongo database is required")
	}
	collectionName = strings.TrimSpace(collectionName)
	if collectionName == "" {
		collectionName = DefaultWishlistEventCollectionName
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &MongoWishlistEventRepository{
		database:       database,
		collection:     database.Collection(collectionName),
		collectionName: collectionName,
		logger:         logger,
		workerID:       newWishlistEventWorkerID(),
		claimLease:     defaultWishlistEventClaimLease,
	}, nil
}

func (r *MongoWishlistEventRepository) SetClaimLease(lease time.Duration) {
	if r != nil && lease > 0 {
		r.claimLease = lease
	}
}

func (r *MongoWishlistEventRepository) CollectionName() string {
	if r == nil {
		return ""
	}
	return r.collectionName
}

func (r *MongoWishlistEventRepository) Ping(ctx context.Context) error {
	if r == nil || r.database == nil {
		return errors.New("wishlist event repository is not initialized")
	}
	if err := r.database.RunCommand(contextOrBackground(ctx), bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		return fmt.Errorf("ping wishlist event mongo database: %w", err)
	}
	return nil
}

func (r *MongoWishlistEventRepository) EnsureCollection(ctx context.Context) error {
	if r == nil || r.database == nil || r.collection == nil {
		return errors.New("wishlist event repository is not initialized")
	}
	ctx = contextOrBackground(ctx)
	if err := r.ensureCollectionValidator(ctx); err != nil {
		return err
	}
	if err := r.ensureIndexes(ctx); err != nil {
		return err
	}
	return nil
}

func (r *MongoWishlistEventRepository) Enqueue(ctx context.Context, event domain.WishlistAnalyticsEvent) error {
	if r == nil || r.collection == nil {
		return errors.New("wishlist event repository is not initialized")
	}
	document, err := NewWishlistEventDocument(event)
	if err != nil {
		return err
	}
	_, err = r.collection.InsertOne(contextOrBackground(ctx), document)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			r.logger.Info("wishlist analytics event already enqueued", "event_id", document.ID, "event_type", document.EventType)
			return nil
		}
		return fmt.Errorf("enqueue wishlist analytics event %q: %w", document.ID, err)
	}
	return nil
}

func (r *MongoWishlistEventRepository) ClaimPending(ctx context.Context, limit int, now time.Time) ([]domain.WishlistAnalyticsEvent, error) {
	if r == nil || r.collection == nil {
		return nil, errors.New("wishlist event repository is not initialized")
	}
	if limit <= 0 {
		limit = defaultWishlistEventClaimBatchSize
	}
	now = normalizeTime(now)
	ctx = contextOrBackground(ctx)
	lockedUntil := now.Add(r.claimLease)
	filter := bson.M{"$or": bson.A{
		bson.M{"status": domain.WishlistEventPending, "next_retry_at": bson.M{"$lte": now}},
		bson.M{"status": domain.WishlistEventPublishing, "locked_until": bson.M{"$lte": now}},
		bson.M{"status": domain.WishlistEventPublishing, "locked_until": bson.M{"$exists": false}},
	}}
	update := bson.M{"$set": bson.M{
		"status":       domain.WishlistEventPublishing,
		"locked_by":    r.workerID,
		"locked_until": lockedUntil,
		"updated_at":   now,
	}}
	events := make([]domain.WishlistAnalyticsEvent, 0, limit)
	for len(events) < limit {
		var document WishlistEventDocument
		err := r.collection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().
			SetSort(bson.D{{Key: "created_at", Value: 1}}).
			SetReturnDocument(options.After)).Decode(&document)
		if errors.Is(err, mongo.ErrNoDocuments) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("claim wishlist analytics event: %w", err)
		}
		event, err := document.ToDomain()
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (r *MongoWishlistEventRepository) MarkPublished(ctx context.Context, eventID string, publishedAt time.Time) error {
	if r == nil || r.collection == nil {
		return errors.New("wishlist event repository is not initialized")
	}
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return fmt.Errorf("%w: event_id is required", domain.ErrInvalidWishlist)
	}
	publishedAt = normalizeTime(publishedAt)
	result, err := r.collection.UpdateOne(
		contextOrBackground(ctx),
		bson.M{"_id": eventID, "status": domain.WishlistEventPublishing, "locked_by": r.workerID},
		bson.M{"$set": bson.M{
			"status":       domain.WishlistEventPublished,
			"published_at": publishedAt,
			"last_error":   "",
			"updated_at":   publishedAt,
		}, "$unset": bson.M{"locked_by": "", "locked_until": ""}},
	)
	if err != nil {
		return fmt.Errorf("mark wishlist analytics event %q published: %w", eventID, err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("%w: %s", domain.ErrWishlistEventNotFound, eventID)
	}
	return nil
}

func (r *MongoWishlistEventRepository) MarkRetry(ctx context.Context, eventID string, attempts int, nextRetryAt time.Time, lastError string) error {
	if r == nil || r.collection == nil {
		return errors.New("wishlist event repository is not initialized")
	}
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return fmt.Errorf("%w: event_id is required", domain.ErrInvalidWishlist)
	}
	if attempts < 0 {
		return fmt.Errorf("%w: attempts must be greater than or equal to zero", domain.ErrInvalidWishlist)
	}
	nextRetryAt = normalizeTime(nextRetryAt)
	result, err := r.collection.UpdateOne(
		contextOrBackground(ctx),
		bson.M{"_id": eventID, "status": domain.WishlistEventPublishing, "locked_by": r.workerID},
		bson.M{"$set": bson.M{
			"status":        domain.WishlistEventPending,
			"attempts":      int32(attempts),
			"next_retry_at": nextRetryAt,
			"last_error":    truncateWishlistEventError(lastError),
			"updated_at":    time.Now().UTC(),
		}, "$unset": bson.M{"locked_by": "", "locked_until": ""}},
	)
	if err != nil {
		return fmt.Errorf("schedule wishlist analytics event %q retry: %w", eventID, err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("%w: %s", domain.ErrWishlistEventNotFound, eventID)
	}
	return nil
}

func (r *MongoWishlistEventRepository) MarkFailed(ctx context.Context, eventID string, attempts int, lastError string, failedAt time.Time) error {
	if r == nil || r.collection == nil {
		return errors.New("wishlist event repository is not initialized")
	}
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return fmt.Errorf("%w: event_id is required", domain.ErrInvalidWishlist)
	}
	if attempts < 0 {
		return fmt.Errorf("%w: attempts must be greater than or equal to zero", domain.ErrInvalidWishlist)
	}
	failedAt = normalizeTime(failedAt)
	result, err := r.collection.UpdateOne(
		contextOrBackground(ctx),
		bson.M{"_id": eventID, "status": domain.WishlistEventPublishing, "locked_by": r.workerID},
		bson.M{"$set": bson.M{
			"status":     domain.WishlistEventFailed,
			"attempts":   int32(attempts),
			"last_error": truncateWishlistEventError(lastError),
			"updated_at": failedAt,
		}, "$unset": bson.M{"locked_by": "", "locked_until": ""}},
	)
	if err != nil {
		return fmt.Errorf("mark wishlist analytics event %q failed: %w", eventID, err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("%w: %s", domain.ErrWishlistEventNotFound, eventID)
	}
	return nil
}

func (r *MongoWishlistEventRepository) ensureCollectionValidator(ctx context.Context) error {
	exists, err := r.collectionExists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		createOptions := options.CreateCollection().
			SetValidator(WishlistEventCollectionValidator()).
			SetValidationLevel(validationLevelStrict).
			SetValidationAction(validationActionError)

		err = r.database.CreateCollection(ctx, r.collectionName, createOptions)
		if err == nil {
			r.logger.Info("wishlist event collection created", "collection", r.collectionName)
			return nil
		}
		if !isNamespaceExists(err) {
			return fmt.Errorf("create wishlist event collection %q: %w", r.collectionName, err)
		}
	}

	command := bson.D{
		{Key: "collMod", Value: r.collectionName},
		{Key: "validator", Value: WishlistEventCollectionValidator()},
		{Key: "validationLevel", Value: validationLevelStrict},
		{Key: "validationAction", Value: validationActionError},
	}
	if err := r.database.RunCommand(ctx, command).Err(); err != nil {
		return fmt.Errorf("update wishlist event collection validator %q: %w", r.collectionName, err)
	}
	r.logger.Info("wishlist event collection validator ensured", "collection", r.collectionName)
	return nil
}

func (r *MongoWishlistEventRepository) collectionExists(ctx context.Context) (bool, error) {
	names, err := r.database.ListCollectionNames(ctx, bson.D{{Key: wishlistCollectionFieldName, Value: r.collectionName}})
	if err != nil {
		return false, fmt.Errorf("list wishlist event collections: %w", err)
	}
	return len(names) > 0, nil
}

func (r *MongoWishlistEventRepository) ensureIndexes(ctx context.Context) error {
	names, err := r.collection.Indexes().CreateMany(ctx, WishlistEventIndexModels())
	if err != nil {
		return fmt.Errorf("ensure wishlist event indexes: %w", err)
	}
	r.logger.Info("wishlist event indexes ensured", "collection", r.collectionName, "indexes", names)
	return nil
}

func WishlistEventIndexModels() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "locked_until", Value: 1},
				{Key: "next_retry_at", Value: 1},
				{Key: "created_at", Value: 1},
			},
			Options: options.Index().SetName(WishlistEventClaimLeaseIndexName),
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "next_retry_at", Value: 1},
				{Key: "created_at", Value: 1},
			},
			Options: options.Index().
				SetName(WishlistEventPendingRetryIndexName),
		},
		{
			Keys: bson.D{
				{Key: "event_type", Value: 1},
				{Key: "occurred_at", Value: -1},
			},
			Options: options.Index().
				SetName(WishlistEventTypeTimeIndexName),
		},
		{
			Keys: bson.D{{Key: "published_at", Value: 1}},
			Options: options.Index().
				SetName(WishlistEventPublishedTTLIndexName).
				SetExpireAfterSeconds(publishedWishlistEventTTLSeconds),
		},
	}
}

func WishlistEventCollectionValidator() bson.D {
	return bson.D{
		{Key: "$jsonSchema", Value: bson.D{
			{Key: "bsonType", Value: "object"},
			{Key: "title", Value: "WishlistAnalyticsEvent"},
			{Key: "required", Value: bson.A{"_id", "event_type", "version", "topic", "payload", "status", "attempts", "next_retry_at", "occurred_at", "created_at", "updated_at"}},
			{Key: "properties", Value: bson.D{
				{Key: "_id", Value: bson.D{
					{Key: "bsonType", Value: "string"},
					{Key: "description", Value: "Globally unique wishlist analytics event id"},
				}},
				{Key: "event_type", Value: bson.D{
					{Key: "enum", Value: bson.A{string(domain.WishlistEventItemAdded), string(domain.WishlistEventItemRemoved)}},
					{Key: "description", Value: "Wishlist analytics event type"},
				}},
				{Key: "version", Value: bson.D{
					{Key: "bsonType", Value: "int"},
					{Key: "minimum", Value: 1},
					{Key: "description", Value: "Event schema version"},
				}},
				{Key: "topic", Value: bson.D{
					{Key: "bsonType", Value: "string"},
					{Key: "description", Value: "Destination message topic"},
				}},
				{Key: "payload", Value: bson.D{
					{Key: "bsonType", Value: "object"},
					{Key: "required", Value: bson.A{"user_id", "product_id", "action", "source"}},
					{Key: "properties", Value: bson.D{
						{Key: "user_id", Value: bson.D{{Key: "bsonType", Value: "string"}}},
						{Key: "product_id", Value: bson.D{{Key: "bsonType", Value: "string"}}},
						{Key: "variant_id", Value: bson.D{{Key: "bsonType", Value: "string"}}},
						{Key: "action", Value: bson.D{
							{Key: "enum", Value: bson.A{string(domain.WishlistEventActionAdd), string(domain.WishlistEventActionRemove)}},
						}},
						{Key: "source", Value: bson.D{
							{Key: "enum", Value: bson.A{domain.WishlistEventSource}},
						}},
						{Key: "availability", Value: bson.D{
							{Key: "enum", Value: bson.A{"unknown", "in_stock", "out_of_stock", "deleted"}},
						}},
						{Key: "last_known_price", Value: bson.D{
							{Key: "bsonType", Value: "object"},
							{Key: "required", Value: bson.A{"amount", "currency"}},
							{Key: "properties", Value: bson.D{
								{Key: "amount", Value: bson.D{{Key: "bsonType", Value: "long"}}},
								{Key: "currency", Value: bson.D{
									{Key: "bsonType", Value: "string"},
									{Key: "pattern", Value: "^[A-Z]{3}$"},
								}},
							}},
						}},
					}},
				}},
				{Key: "trace_id", Value: bson.D{{Key: "bsonType", Value: "string"}}},
				{Key: "status", Value: bson.D{
					{Key: "enum", Value: bson.A{
						string(domain.WishlistEventPending),
						string(domain.WishlistEventPublishing),
						string(domain.WishlistEventPublished),
						string(domain.WishlistEventFailed),
					}},
				}},
				{Key: "attempts", Value: bson.D{
					{Key: "bsonType", Value: "int"},
					{Key: "minimum", Value: 0},
				}},
				{Key: "next_retry_at", Value: bson.D{{Key: "bsonType", Value: "date"}}},
				{Key: "last_error", Value: bson.D{{Key: "bsonType", Value: "string"}}},
				{Key: "occurred_at", Value: bson.D{{Key: "bsonType", Value: "date"}}},
				{Key: "published_at", Value: bson.D{{Key: "bsonType", Value: bson.A{"date", "null"}}}},
				{Key: "locked_by", Value: bson.D{{Key: "bsonType", Value: "string"}}},
				{Key: "locked_until", Value: bson.D{{Key: "bsonType", Value: bson.A{"date", "null"}}}},
				{Key: "created_at", Value: bson.D{{Key: "bsonType", Value: "date"}}},
				{Key: "updated_at", Value: bson.D{{Key: "bsonType", Value: "date"}}},
			}},
		}},
	}
}

func newWishlistEventWorkerID() string {
	var value [12]byte
	if _, err := rand.Read(value[:]); err == nil {
		return "wishlist-worker-" + hex.EncodeToString(value[:])
	}
	return fmt.Sprintf("wishlist-worker-%d", time.Now().UTC().UnixNano())
}
