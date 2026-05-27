package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoNotificationRepository struct {
	templates      *mongo.Collection
	deliveries     *mongo.Collection
	preferences    *mongo.Collection
	providerEvents *mongo.Collection
}

var (
	_ TemplateRepository      = (*MongoNotificationRepository)(nil)
	_ DeliveryRepository      = (*MongoNotificationRepository)(nil)
	_ EventDeliveryRepository = (*MongoNotificationRepository)(nil)
	_ RetryDeliveryRepository = (*MongoNotificationRepository)(nil)
	_ PreferenceRepository    = (*MongoNotificationRepository)(nil)
	_ AnalyticsRepository     = (*MongoNotificationRepository)(nil)
)

// OpenMongoNotificationRepository creates the runtime Mongo client and repository.
// Callers own the returned client and must disconnect it during shutdown.
func OpenMongoNotificationRepository(
	ctx context.Context,
	uri string,
	databaseName string,
	templatesCollection string,
	deliveriesCollection string,
	preferencesCollection string,
	providerEventsCollection string,
) (*MongoNotificationRepository, *mongo.Client, error) {
	if err := validContext(ctx); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(uri) == "" {
		return nil, nil, errors.New("mongo URI is required")
	}
	if strings.TrimSpace(databaseName) == "" {
		return nil, nil, errors.New("mongo database is required")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(strings.TrimSpace(uri)))
	if err != nil {
		return nil, nil, fmt.Errorf("connect notification mongo: %w", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, fmt.Errorf("ping notification mongo: %w", err)
	}

	repo, err := NewMongoNotificationRepository(
		client.Database(strings.TrimSpace(databaseName)),
		templatesCollection,
		deliveriesCollection,
		preferencesCollection,
		providerEventsCollection,
	)
	if err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, err
	}
	return repo, client, nil
}

// NewMongoNotificationRepository accepts an existing database so tests and
// application wiring control connection lifetime and connection options.
func NewMongoNotificationRepository(
	database *mongo.Database,
	templatesCollection string,
	deliveriesCollection string,
	preferencesCollection string,
	providerEventsCollection string,
) (*MongoNotificationRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	templatesCollection = strings.TrimSpace(templatesCollection)
	deliveriesCollection = strings.TrimSpace(deliveriesCollection)
	preferencesCollection = strings.TrimSpace(preferencesCollection)
	providerEventsCollection = strings.TrimSpace(providerEventsCollection)
	if templatesCollection == "" || deliveriesCollection == "" || preferencesCollection == "" ||
		providerEventsCollection == "" {
		return nil, errors.New("mongo notification collections are required")
	}
	if err := validCollectionName(templatesCollection); err != nil {
		return nil, fmt.Errorf("mongo templates collection: %w", err)
	}
	if err := validCollectionName(deliveriesCollection); err != nil {
		return nil, fmt.Errorf("mongo deliveries collection: %w", err)
	}
	if err := validCollectionName(preferencesCollection); err != nil {
		return nil, fmt.Errorf("mongo preferences collection: %w", err)
	}
	if err := validCollectionName(providerEventsCollection); err != nil {
		return nil, fmt.Errorf("mongo provider events collection: %w", err)
	}
	names := map[string]struct{}{}
	for _, collection := range []string{templatesCollection, deliveriesCollection, preferencesCollection, providerEventsCollection} {
		if _, exists := names[collection]; exists {
			return nil, errors.New("mongo notification collections must be different")
		}
		names[collection] = struct{}{}
	}
	return &MongoNotificationRepository{
		templates:      database.Collection(templatesCollection),
		deliveries:     database.Collection(deliveriesCollection),
		preferences:    database.Collection(preferencesCollection),
		providerEvents: database.Collection(providerEventsCollection),
	}, nil
}

func (r *MongoNotificationRepository) FindPreferenceByUserID(
	ctx context.Context,
	userID string,
) (domain.Preference, error) {
	if err := validContext(ctx); err != nil {
		return domain.Preference{}, err
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return domain.Preference{}, fmt.Errorf("%w: user_id is required", domain.ErrInvalidPreference)
	}
	var preference domain.Preference
	if err := r.preferences.FindOne(ctx, bson.D{{Key: "user_id", Value: userID}}).Decode(&preference); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Preference{}, fmt.Errorf("%w: user_id %q", domain.ErrPreferenceNotFound, userID)
		}
		return domain.Preference{}, fmt.Errorf("find notification preference: %w", err)
	}
	if err := preference.Validate(); err != nil {
		return domain.Preference{}, fmt.Errorf("validate stored notification preference: %w", err)
	}
	return preference, nil
}

func (r *MongoNotificationRepository) UpsertPreference(
	ctx context.Context,
	preference domain.Preference,
) (domain.Preference, error) {
	if err := validContext(ctx); err != nil {
		return domain.Preference{}, err
	}
	if err := preference.Validate(); err != nil {
		return domain.Preference{}, err
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "email_enabled", Value: preference.EmailEnabled},
			{Key: "sms_enabled", Value: preference.SMSEnabled},
			{Key: "push_enabled", Value: preference.PushEnabled},
			{Key: "marketing_enabled", Value: preference.MarketingEnabled},
			{Key: "consent_source", Value: preference.ConsentSource},
			{Key: "updated_at", Value: preference.UpdatedAt},
		}},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: preference.ID},
			{Key: "user_id", Value: preference.UserID},
			{Key: "created_at", Value: preference.CreatedAt},
		}},
	}
	var saved domain.Preference
	err := r.preferences.FindOneAndUpdate(
		ctx,
		bson.D{{Key: "user_id", Value: preference.UserID}},
		update,
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&saved)
	if mongo.IsDuplicateKeyError(err) {
		// A concurrent first write may win the unique user_id insert. Apply this
		// update to that row rather than exposing a spurious conflict to the user.
		err = r.preferences.FindOneAndUpdate(
			ctx,
			bson.D{{Key: "user_id", Value: preference.UserID}},
			update,
			options.FindOneAndUpdate().SetReturnDocument(options.After),
		).Decode(&saved)
	}
	if err != nil {
		return domain.Preference{}, fmt.Errorf("upsert notification preference: %w", err)
	}
	if err := saved.Validate(); err != nil {
		return domain.Preference{}, fmt.Errorf("validate saved notification preference: %w", err)
	}
	return saved, nil
}

func (r *MongoNotificationRepository) FindLatestActive(
	ctx context.Context,
	templateKey string,
	channel domain.Channel,
) (domain.Template, error) {
	if err := validContext(ctx); err != nil {
		return domain.Template{}, err
	}
	if strings.TrimSpace(templateKey) == "" {
		return domain.Template{}, fmt.Errorf("%w: template key is required", domain.ErrInvalidTemplate)
	}
	if !channel.IsSupported() {
		return domain.Template{}, fmt.Errorf("%w: %w: %q", domain.ErrInvalidTemplate, domain.ErrUnsupportedChannel, channel)
	}

	filter := bson.D{
		{Key: "template_key", Value: strings.TrimSpace(templateKey)},
		{Key: "channel", Value: channel},
		{Key: "status", Value: domain.TemplateStatusActive},
	}
	findOptions := options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}})

	var template domain.Template
	if err := r.templates.FindOne(ctx, filter, findOptions).Decode(&template); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Template{}, fmt.Errorf("%w: key %q channel %q", domain.ErrTemplateNotFound, templateKey, channel)
		}
		return domain.Template{}, fmt.Errorf("find latest active notification template: %w", err)
	}
	if err := template.Validate(); err != nil {
		return domain.Template{}, fmt.Errorf("validate stored notification template: %w", err)
	}
	return template, nil
}

func (r *MongoNotificationRepository) InsertTemplate(ctx context.Context, template domain.Template) error {
	if err := validContext(ctx); err != nil {
		return err
	}
	if err := template.Validate(); err != nil {
		return err
	}
	if _, err := r.templates.InsertOne(ctx, template); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: id %q or template revision exists", domain.ErrTemplateConflict, template.ID)
		}
		return fmt.Errorf("insert notification template: %w", err)
	}
	return nil
}

func (r *MongoNotificationRepository) FindDeliveryByID(ctx context.Context, id string) (domain.Delivery, error) {
	if err := validContext(ctx); err != nil {
		return domain.Delivery{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.Delivery{}, fmt.Errorf("%w: id is required", domain.ErrInvalidDelivery)
	}

	var delivery domain.Delivery
	if err := r.deliveries.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&delivery); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Delivery{}, fmt.Errorf("%w: id %q", domain.ErrDeliveryNotFound, id)
		}
		return domain.Delivery{}, fmt.Errorf("find notification delivery: %w", err)
	}
	if err := delivery.Validate(); err != nil {
		return domain.Delivery{}, fmt.Errorf("validate stored notification delivery: %w", err)
	}
	return delivery, nil
}

func (r *MongoNotificationRepository) InsertDelivery(ctx context.Context, delivery domain.Delivery) error {
	if err := validContext(ctx); err != nil {
		return err
	}
	if err := delivery.Validate(); err != nil {
		return err
	}
	if _, err := r.deliveries.InsertOne(ctx, delivery); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: id %q", domain.ErrDeliveryConflict, delivery.ID)
		}
		return fmt.Errorf("insert notification delivery: %w", err)
	}
	return nil
}

func (r *MongoNotificationRepository) FindDeliveryByProviderMessageID(
	ctx context.Context,
	provider string,
	messageID string,
	channel domain.Channel,
) (domain.Delivery, error) {
	if err := validContext(ctx); err != nil {
		return domain.Delivery{}, err
	}
	provider = strings.TrimSpace(provider)
	messageID = strings.TrimSpace(messageID)
	if provider == "" || messageID == "" || !channel.IsSupported() {
		return domain.Delivery{}, fmt.Errorf("%w: provider message reference is required", domain.ErrInvalidProviderEvent)
	}
	var delivery domain.Delivery
	err := r.deliveries.FindOne(ctx, bson.D{
		{Key: "provider", Value: provider},
		{Key: "provider_message_id", Value: messageID},
		{Key: "channel", Value: channel},
	}).Decode(&delivery)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Delivery{}, fmt.Errorf("%w: unmatched provider message", domain.ErrProviderEventNotFound)
		}
		return domain.Delivery{}, fmt.Errorf("find delivery by provider message id: %w", err)
	}
	if err := delivery.Validate(); err != nil {
		return domain.Delivery{}, fmt.Errorf("validate callback delivery: %w", err)
	}
	return delivery, nil
}

func (r *MongoNotificationRepository) RecordEventAndApplyMilestone(
	ctx context.Context,
	event domain.ProviderEvent,
) (domain.DeliveryEventApplyResult, error) {
	if err := validContext(ctx); err != nil {
		return domain.DeliveryEventApplyResult{}, err
	}
	if err := event.Validate(); err != nil {
		return domain.DeliveryEventApplyResult{}, err
	}
	if r.providerEvents == nil || r.deliveries == nil {
		return domain.DeliveryEventApplyResult{}, errors.New("analytics repository collections are required")
	}
	session, err := r.providerEvents.Database().Client().StartSession()
	if err != nil {
		return domain.DeliveryEventApplyResult{}, fmt.Errorf("start provider event transaction: %w", err)
	}
	defer session.EndSession(ctx)

	var result domain.DeliveryEventApplyResult
	_, err = session.WithTransaction(ctx, func(txCtx context.Context) (any, error) {
		if _, insertErr := r.providerEvents.InsertOne(txCtx, event); insertErr != nil {
			return nil, insertErr
		}
		changed, updateErr := r.applyDeliveryMilestone(txCtx, event)
		if updateErr != nil {
			return nil, updateErr
		}
		result.MilestoneChanged = changed
		return nil, nil
	})
	if mongo.IsDuplicateKeyError(err) {
		return domain.DeliveryEventApplyResult{Duplicate: true}, nil
	}
	if err != nil {
		return domain.DeliveryEventApplyResult{}, fmt.Errorf("record provider delivery event: %w", err)
	}
	return result, nil
}

func (r *MongoNotificationRepository) applyDeliveryMilestone(
	ctx context.Context,
	event domain.ProviderEvent,
) (bool, error) {
	field := ""
	set := bson.D{{Key: "updated_at", Value: event.ReceivedAt}}
	switch event.Type {
	case domain.DeliveryEventSent:
		field = "sent_at"
		set = append(set,
			bson.E{Key: field, Value: event.OccurredAt},
			bson.E{Key: "provider", Value: event.Provider},
		)
		if strings.TrimSpace(event.ProviderMessageID) != "" {
			set = append(set, bson.E{Key: "provider_message_id", Value: event.ProviderMessageID})
		}
	case domain.DeliveryEventDelivered:
		field = "delivered_at"
		set = append(set, bson.E{Key: field, Value: event.OccurredAt})
	case domain.DeliveryEventFailed:
		field = "failed_at"
		set = append(set,
			bson.E{Key: field, Value: event.OccurredAt},
			bson.E{Key: "failure_code", Value: event.FailureCode},
		)
	case domain.DeliveryEventOpened:
		field = "opened_at"
		set = append(set, bson.E{Key: field, Value: event.OccurredAt})
	default:
		return false, fmt.Errorf("%w: unsupported event type %q", domain.ErrInvalidProviderEvent, event.Type)
	}
	update, err := r.deliveries.UpdateOne(ctx,
		bson.D{
			{Key: "_id", Value: event.DeliveryID},
			{Key: field, Value: bson.D{{Key: "$exists", Value: false}}},
		},
		bson.D{{Key: "$set", Value: set}},
	)
	if err != nil {
		return false, fmt.Errorf("apply notification delivery milestone: %w", err)
	}
	if update.MatchedCount == 0 {
		var exists struct {
			ID string `bson:"_id"`
		}
		if err := r.deliveries.FindOne(ctx, bson.D{{Key: "_id", Value: event.DeliveryID}}).Decode(&exists); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return false, fmt.Errorf("%w: delivery %q", domain.ErrProviderEventNotFound, event.DeliveryID)
			}
			return false, fmt.Errorf("find event delivery after milestone update: %w", err)
		}
	}
	return update.ModifiedCount == 1, nil
}

func (r *MongoNotificationRepository) ClaimEventDelivery(ctx context.Context, delivery domain.Delivery) (bool, error) {
	if err := validContext(ctx); err != nil {
		return false, err
	}
	if (delivery.Status != domain.DeliveryStatusProcessing && delivery.Status != domain.DeliveryStatusSuppressed) ||
		strings.TrimSpace(delivery.IdempotencyKey) == "" {
		return false, fmt.Errorf("%w: event claim must have processing or suppressed status and an idempotency key",
			domain.ErrInvalidDelivery)
	}
	if err := delivery.Validate(); err != nil {
		return false, err
	}

	if _, err := r.deliveries.InsertOne(ctx, delivery); err == nil {
		return true, nil
	} else if !mongo.IsDuplicateKeyError(err) {
		return false, fmt.Errorf("insert event notification delivery claim: %w", err)
	}
	if delivery.Status == domain.DeliveryStatusSuppressed {
		return false, nil
	}

	result, err := r.deliveries.UpdateOne(ctx,
		bson.D{
			{Key: "idempotency_key", Value: delivery.IdempotencyKey},
			{Key: "status", Value: domain.DeliveryStatusFailed},
		},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "status", Value: domain.DeliveryStatusProcessing},
			{Key: "updated_at", Value: delivery.UpdatedAt},
		}}},
	)
	if err != nil {
		return false, fmt.Errorf("reclaim failed event notification delivery: %w", err)
	}
	return result.ModifiedCount == 1, nil
}

func (r *MongoNotificationRepository) RecordEventDeliveryOutcome(
	ctx context.Context,
	idempotencyKey string,
	outcome domain.EventDeliveryOutcome,
) error {
	if err := validContext(ctx); err != nil {
		return err
	}
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" {
		return fmt.Errorf("%w: idempotency key is required", domain.ErrInvalidDelivery)
	}
	if err := outcome.Validate(); err != nil {
		return err
	}
	result, err := r.deliveries.UpdateOne(ctx,
		bson.D{
			{Key: "idempotency_key", Value: idempotencyKey},
			{Key: "status", Value: domain.DeliveryStatusProcessing},
		},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "status", Value: outcome.Status},
				{Key: "provider", Value: outcome.Provider},
				{Key: "provider_message_id", Value: outcome.ProviderMessageID},
				{Key: "updated_at", Value: outcome.UpdatedAt},
			}},
			{Key: "$inc", Value: bson.D{{Key: "attempts", Value: 1}}},
		},
	)
	if err != nil {
		return fmt.Errorf("update event notification delivery outcome: %w", err)
	}
	if result.MatchedCount != 1 {
		return fmt.Errorf("%w: no processing event delivery for idempotency key", domain.ErrDeliveryNotFound)
	}
	return nil
}

func (r *MongoNotificationRepository) PrepareEventDelivery(ctx context.Context, delivery domain.Delivery) (bool, error) {
	if err := validContext(ctx); err != nil {
		return false, err
	}
	if delivery.Status != domain.DeliveryStatusPending || delivery.MaxAttempts < 1 ||
		strings.TrimSpace(delivery.IdempotencyKey) == "" {
		return false, fmt.Errorf("%w: queued event delivery must be pending and retryable", domain.ErrInvalidDelivery)
	}
	if err := delivery.Validate(); err != nil {
		return false, err
	}
	if _, err := r.deliveries.InsertOne(ctx, delivery); err == nil {
		return true, nil
	} else if !mongo.IsDuplicateKeyError(err) {
		return false, fmt.Errorf("insert queued event notification delivery: %w", err)
	}

	var existing struct {
		Status domain.DeliveryStatus `bson:"status"`
	}
	if err := r.deliveries.FindOne(ctx,
		bson.D{{Key: "idempotency_key", Value: delivery.IdempotencyKey}}).Decode(&existing); err != nil {
		return false, fmt.Errorf("find duplicate queued event notification delivery: %w", err)
	}
	// A source event may be redelivered after its first publish was interrupted.
	// Republishing pending attempt 1 is safe because the worker claim is idempotent.
	return existing.Status == domain.DeliveryStatusPending, nil
}

func (r *MongoNotificationRepository) BeginAttempt(
	ctx context.Context,
	deliveryID string,
	idempotencyKey string,
	attempt int,
	leaseUntil time.Time,
	updatedAt time.Time,
) (domain.AttemptClaim, error) {
	if err := validContext(ctx); err != nil {
		return "", err
	}
	deliveryID = strings.TrimSpace(deliveryID)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if deliveryID == "" || idempotencyKey == "" || attempt < 1 ||
		leaseUntil.IsZero() || updatedAt.IsZero() || !leaseUntil.After(updatedAt) {
		return "", fmt.Errorf("%w: invalid retry attempt claim", domain.ErrInvalidDelivery)
	}
	eligible := bson.A{
		bson.D{
			{Key: "status", Value: domain.DeliveryStatusRetryScheduled},
			{Key: "attempts", Value: attempt - 1},
		},
		bson.D{
			{Key: "status", Value: domain.DeliveryStatusProcessing},
			{Key: "processing_attempt", Value: attempt},
			{Key: "processing_lease_until", Value: bson.D{{Key: "$lte", Value: updatedAt}}},
		},
	}
	if attempt == 1 {
		eligible = append(eligible, bson.D{
			{Key: "status", Value: domain.DeliveryStatusPending},
			{Key: "attempts", Value: 0},
		})
	}
	result, err := r.deliveries.UpdateOne(ctx,
		bson.D{
			{Key: "_id", Value: deliveryID},
			{Key: "idempotency_key", Value: idempotencyKey},
			{Key: "max_attempts", Value: bson.D{{Key: "$gte", Value: attempt}}},
			{Key: "$or", Value: eligible},
		},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "status", Value: domain.DeliveryStatusProcessing},
				{Key: "processing_attempt", Value: attempt},
				{Key: "processing_lease_until", Value: leaseUntil},
				{Key: "updated_at", Value: updatedAt},
			}},
			{Key: "$unset", Value: bson.D{{Key: "next_retry_at", Value: ""}}},
		},
	)
	if err != nil {
		return "", fmt.Errorf("claim notification retry attempt: %w", err)
	}
	if result.ModifiedCount == 1 {
		return domain.AttemptAcquired, nil
	}

	var current domain.Delivery
	if err := r.deliveries.FindOne(ctx, bson.D{
		{Key: "_id", Value: deliveryID},
		{Key: "idempotency_key", Value: idempotencyKey},
	}).Decode(&current); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", fmt.Errorf("%w: delivery identity does not match", domain.ErrRetryDeliveryReference)
		}
		return "", fmt.Errorf("read unclaimed notification retry delivery: %w", err)
	}
	if current.Status == domain.DeliveryStatusAccepted ||
		current.Status == domain.DeliveryStatusSuppressed ||
		current.Status == domain.DeliveryStatusDeadLettered ||
		current.Attempts >= attempt {
		return domain.AttemptCompleted, nil
	}
	if current.Status == domain.DeliveryStatusProcessing &&
		current.ProcessingAttempt == attempt &&
		current.ProcessingLeaseUntil != nil &&
		current.ProcessingLeaseUntil.After(updatedAt) {
		return domain.AttemptBusy, nil
	}
	return "", fmt.Errorf("%w: delivery %q is not eligible for attempt %d", domain.ErrInvalidDelivery, deliveryID, attempt)
}

func (r *MongoNotificationRepository) MarkAccepted(
	ctx context.Context,
	deliveryID string,
	attempt int,
	providerName string,
	providerMessageID string,
	updatedAt time.Time,
) error {
	if strings.TrimSpace(providerName) == "" {
		return fmt.Errorf("%w: accepted retry delivery requires provider", domain.ErrInvalidDelivery)
	}
	return r.finishAttempt(ctx, deliveryID, attempt, bson.D{
		{Key: "status", Value: domain.DeliveryStatusAccepted},
		{Key: "provider", Value: strings.TrimSpace(providerName)},
		{Key: "provider_message_id", Value: strings.TrimSpace(providerMessageID)},
		{Key: "attempts", Value: attempt},
		{Key: "updated_at", Value: updatedAt},
	}, bson.D{
		{Key: "last_failure_code", Value: ""},
		{Key: "next_retry_at", Value: ""},
		{Key: "dead_lettered_at", Value: ""},
	})
}

func (r *MongoNotificationRepository) MarkSuppressed(
	ctx context.Context,
	deliveryID string,
	attempt int,
	reason domain.ConsentReason,
	updatedAt time.Time,
) error {
	if !reason.IsSuppressionReason() || updatedAt.IsZero() {
		return fmt.Errorf("%w: invalid suppressed retry outcome", domain.ErrInvalidDelivery)
	}
	return r.finishAttempt(ctx, deliveryID, attempt, bson.D{
		{Key: "status", Value: domain.DeliveryStatusSuppressed},
		{Key: "suppression_reason", Value: reason},
		{Key: "updated_at", Value: updatedAt},
	}, bson.D{
		{Key: "recipient_ciphertext", Value: ""},
		{Key: "max_attempts", Value: ""},
		{Key: "payload", Value: ""},
		{Key: "last_failure_code", Value: ""},
		{Key: "next_retry_at", Value: ""},
		{Key: "dead_lettered_at", Value: ""},
	})
}

func (r *MongoNotificationRepository) MarkRetryScheduled(
	ctx context.Context,
	deliveryID string,
	attempt int,
	nextRetryAt time.Time,
	failureCode string,
	updatedAt time.Time,
) error {
	if strings.TrimSpace(failureCode) == "" || !nextRetryAt.After(updatedAt) {
		return fmt.Errorf("%w: invalid scheduled retry outcome", domain.ErrInvalidDelivery)
	}
	return r.finishAttempt(ctx, deliveryID, attempt, bson.D{
		{Key: "status", Value: domain.DeliveryStatusRetryScheduled},
		{Key: "attempts", Value: attempt},
		{Key: "last_failure_code", Value: strings.TrimSpace(failureCode)},
		{Key: "next_retry_at", Value: nextRetryAt},
		{Key: "updated_at", Value: updatedAt},
	}, bson.D{{Key: "dead_lettered_at", Value: ""}})
}

func (r *MongoNotificationRepository) MarkDeadLettered(
	ctx context.Context,
	deliveryID string,
	attempt int,
	failureCode string,
	deadLetteredAt time.Time,
) error {
	if strings.TrimSpace(failureCode) == "" || deadLetteredAt.IsZero() {
		return fmt.Errorf("%w: invalid dead-letter outcome", domain.ErrInvalidDelivery)
	}
	return r.finishAttempt(ctx, deliveryID, attempt, bson.D{
		{Key: "status", Value: domain.DeliveryStatusDeadLettered},
		{Key: "attempts", Value: attempt},
		{Key: "last_failure_code", Value: strings.TrimSpace(failureCode)},
		{Key: "dead_lettered_at", Value: deadLetteredAt},
		{Key: "updated_at", Value: deadLetteredAt},
	}, bson.D{{Key: "next_retry_at", Value: ""}})
}

func (r *MongoNotificationRepository) finishAttempt(
	ctx context.Context,
	deliveryID string,
	attempt int,
	values bson.D,
	extraUnset bson.D,
) error {
	if err := validContext(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(deliveryID) == "" || attempt < 1 {
		return fmt.Errorf("%w: invalid retry attempt outcome", domain.ErrInvalidDelivery)
	}
	unset := append(bson.D{
		{Key: "processing_attempt", Value: ""},
		{Key: "processing_lease_until", Value: ""},
	}, extraUnset...)
	result, err := r.deliveries.UpdateOne(ctx,
		bson.D{
			{Key: "_id", Value: strings.TrimSpace(deliveryID)},
			{Key: "status", Value: domain.DeliveryStatusProcessing},
			{Key: "processing_attempt", Value: attempt},
		},
		bson.D{
			{Key: "$set", Value: values},
			{Key: "$unset", Value: unset},
		},
	)
	if err != nil {
		return fmt.Errorf("finish notification retry attempt: %w", err)
	}
	if result.MatchedCount != 1 {
		return fmt.Errorf("%w: active attempt %d for delivery %q", domain.ErrDeliveryNotFound, attempt, deliveryID)
	}
	return nil
}

func validContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func validCollectionName(name string) error {
	if strings.ContainsRune(name, '$') || strings.HasPrefix(name, "system.") {
		return errors.New("collection name is reserved or contains an invalid character")
	}
	return nil
}
