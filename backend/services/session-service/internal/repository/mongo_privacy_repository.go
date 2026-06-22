package repository

import (
	"context"
	"errors"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const privacySettingsID = "global"

type MongoPrivacyRepository struct {
	settingsCollection         *mongo.Collection
	deletionRequestsCollection *mongo.Collection
	auditCollection            *mongo.Collection
	sessionsCollection         *mongo.Collection
	eventsCollection           *mongo.Collection
	journeySummariesCollection *mongo.Collection
}

func NewMongoPrivacyRepository(db *mongo.Database) (*MongoPrivacyRepository, error) {
	if db == nil {
		return nil, errors.New("mongo database is required")
	}
	return &MongoPrivacyRepository{
		settingsCollection:         db.Collection("privacy_settings"),
		deletionRequestsCollection: db.Collection("analytics_deletion_requests"),
		auditCollection:            db.Collection("admin_audit_events"),
		sessionsCollection:         db.Collection("sessions"),
		eventsCollection:           db.Collection("session_events"),
		journeySummariesCollection: db.Collection("journey_summaries"),
	}, nil
}

func (r *MongoPrivacyRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.deletionRequestsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("analytics_deletion_requests_status_created_at"),
		},
		{
			Keys:    bson.D{{Key: "requested_by", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("analytics_deletion_requests_requested_by_created_at"),
		},
		{
			Keys:    bson.D{{Key: "target_hash", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("analytics_deletion_requests_target_hash_created_at"),
		},
		{
			Keys: bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().
				SetExpireAfterSeconds(int32((730 * 24 * time.Hour) / time.Second)).
				SetName("analytics_deletion_requests_created_at_ttl"),
		},
	})
	if err != nil {
		return err
	}

	_, err = r.auditCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "actor_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("admin_audit_actor_created_at"),
		},
		{
			Keys:    bson.D{{Key: "action", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("admin_audit_action_created_at"),
		},
	})
	return err
}

func (r *MongoPrivacyRepository) GetPrivacySettings(ctx context.Context) (domain.PrivacySettings, error) {
	var doc privacySettingsDocument
	err := r.settingsCollection.FindOne(ctx, bson.D{{Key: "_id", Value: privacySettingsID}}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.PrivacySettings{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.PrivacySettings{}, err
	}
	return doc.toDomain(), nil
}

func (r *MongoPrivacyRepository) UpsertPrivacySettings(ctx context.Context, settings domain.PrivacySettings) error {
	doc := privacySettingsDocumentFromDomain(settings)
	_, err := r.settingsCollection.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: privacySettingsID}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "masking", Value: doc.Masking},
			{Key: "retention", Value: doc.Retention},
			{Key: "updated_at", Value: doc.UpdatedAt},
			{Key: "updated_by", Value: doc.UpdatedBy},
		}}},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (r *MongoPrivacyRepository) PreviewDeletion(ctx context.Context, target domain.DeletionTarget) (domain.DeletionPreview, error) {
	sessionCount, err := r.sessionsCollection.CountDocuments(ctx, deletionFilter(target))
	if err != nil {
		return domain.DeletionPreview{}, err
	}
	eventCount, err := r.eventsCollection.CountDocuments(ctx, deletionFilter(target))
	if err != nil {
		return domain.DeletionPreview{}, err
	}
	journeyCount, err := r.journeySummariesCollection.CountDocuments(ctx, deletionFilter(target))
	if err != nil {
		return domain.DeletionPreview{}, err
	}

	return domain.DeletionPreview{
		TargetType:              target.Type,
		TargetValueMasked:       domain.MaskIdentifier(target.Value),
		MatchedSessions:         sessionCount,
		MatchedEvents:           eventCount,
		MatchedJourneySummaries: journeyCount,
		AggregateImpact:         domain.AggregateImpactAnonymizedOrUnchanged,
	}, nil
}

func (r *MongoPrivacyRepository) ApplyDeletion(ctx context.Context, target domain.DeletionTarget) (domain.DeletionPreview, error) {
	filter := deletionFilter(target)
	now := time.Now().UTC()

	eventResult, err := r.eventsCollection.DeleteMany(ctx, filter)
	if err != nil {
		return domain.DeletionPreview{}, err
	}
	journeyResult, err := r.journeySummariesCollection.DeleteMany(ctx, filter)
	if err != nil {
		return domain.DeletionPreview{}, err
	}
	sessionResult, err := r.sessionsCollection.UpdateMany(
		ctx,
		filter,
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "privacy_deleted", Value: true},
				{Key: "privacy_deleted_at", Value: now},
				{Key: "privacy_deletion_target_type", Value: string(target.Type)},
			}},
			{Key: "$unset", Value: bson.D{
				{Key: "user_id", Value: ""},
				{Key: "anonymous_id", Value: ""},
				{Key: "ip_hash", Value: ""},
				{Key: "device.user_agent", Value: ""},
			}},
		},
	)
	if err != nil {
		return domain.DeletionPreview{}, err
	}

	return domain.DeletionPreview{
		TargetType:              target.Type,
		TargetValueMasked:       domain.MaskIdentifier(target.Value),
		MatchedSessions:         sessionResult.MatchedCount,
		MatchedEvents:           eventResult.DeletedCount,
		MatchedJourneySummaries: journeyResult.DeletedCount,
		AggregateImpact:         domain.AggregateImpactAnonymizedOrUnchanged,
	}, nil
}

func (r *MongoPrivacyRepository) CreateDeletionRequest(ctx context.Context, request domain.DeletionRequest) error {
	_, err := r.deletionRequestsCollection.InsertOne(ctx, deletionRequestDocumentFromDomain(request))
	return err
}

func (r *MongoPrivacyRepository) UpdateDeletionRequestStatus(ctx context.Context, requestID string, status domain.DeletionRequestStatus, completedAt *time.Time, message string) error {
	set := bson.D{{Key: "status", Value: string(status)}}
	if completedAt != nil {
		set = append(set, bson.E{Key: "completed_at", Value: completedAt})
	}
	update := bson.D{}
	if message != "" {
		set = append(set, bson.E{Key: "error", Value: message})
	} else {
		update = append(update, bson.E{Key: "$unset", Value: bson.D{{Key: "error", Value: ""}}})
	}
	update = append(update, bson.E{Key: "$set", Value: set})

	_, err := r.deletionRequestsCollection.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: requestID}},
		update,
	)
	return err
}

func (r *MongoPrivacyRepository) ListDeletionRequests(ctx context.Context, limit int) ([]domain.DeletionRequest, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	cursor, err := r.deletionRequestsCollection.Find(
		ctx,
		bson.D{},
		options.Find().
			SetLimit(int64(limit)).
			SetSort(bson.D{{Key: "created_at", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []deletionRequestDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	requests := make([]domain.DeletionRequest, 0, len(docs))
	for _, doc := range docs {
		requests = append(requests, doc.toDomain())
	}
	return requests, nil
}

func (r *MongoPrivacyRepository) CreateAuditEvent(ctx context.Context, event domain.AuditEvent) error {
	_, err := r.auditCollection.InsertOne(ctx, auditEventDocumentFromDomain(event))
	return err
}

func (r *MongoPrivacyRepository) ApplyRetentionSettings(ctx context.Context, settings domain.RetentionSettings) error {
	if err := domain.ValidateRetentionSettings(settings); err != nil {
		return err
	}
	database := r.settingsCollection.Database()
	commands := []bson.D{
		{{Key: "collMod", Value: "session_events"}, {Key: "index", Value: bson.D{{Key: "keyPattern", Value: bson.D{{Key: "occurred_at", Value: 1}}}, {Key: "expireAfterSeconds", Value: int64(settings.RawEventsDays) * 86400}}}},
		{{Key: "collMod", Value: "analytics_deletion_requests"}, {Key: "index", Value: bson.D{{Key: "name", Value: "analytics_deletion_requests_created_at_ttl"}, {Key: "expireAfterSeconds", Value: int64(settings.DeletionRequestLogDays) * 86400}}}},
	}
	for _, command := range commands {
		if err := database.RunCommand(ctx, command).Err(); err != nil {
			return err
		}
	}
	updates := []struct {
		collection *mongo.Collection
		timeField  string
		amount     int
		unit       string
	}{
		{r.journeySummariesCollection, "calculated_at", settings.JourneySummariesDays, "day"},
		{database.Collection("heatmap_points"), "last_seen_at", settings.HeatmapAggregatesDays, "day"},
		{database.Collection("analytics_aggregates"), "updated_at", settings.AnalyticsAggregatesMonths, "month"},
	}
	for _, update := range updates {
		pipeline := mongo.Pipeline{{{Key: "$set", Value: bson.D{{Key: "retain_until", Value: bson.D{{Key: "$dateAdd", Value: bson.D{{Key: "startDate", Value: "$" + update.timeField}, {Key: "unit", Value: update.unit}, {Key: "amount", Value: update.amount}}}}}}}}}
		if _, err := update.collection.UpdateMany(ctx, bson.D{{Key: update.timeField, Value: bson.D{{Key: "$type", Value: "date"}}}}, pipeline); err != nil {
			return err
		}
	}
	return nil
}

func deletionFilter(target domain.DeletionTarget) bson.D {
	switch target.Type {
	case domain.DeletionTargetUserID:
		return bson.D{{Key: "user_id", Value: target.Value}}
	case domain.DeletionTargetAnonymousID:
		return bson.D{{Key: "anonymous_id", Value: target.Value}}
	case domain.DeletionTargetSessionID:
		return bson.D{{Key: "$or", Value: bson.A{
			bson.D{{Key: "session_id", Value: target.Value}},
			bson.D{{Key: "_id", Value: target.Value}},
		}}}
	default:
		return bson.D{{Key: "_id", Value: "__no_match__"}}
	}
}

type privacySettingsDocument struct {
	ID        string                    `bson:"_id"`
	Masking   privacyMaskingDocument    `bson:"masking"`
	Retention retentionSettingsDocument `bson:"retention"`
	UpdatedAt time.Time                 `bson:"updated_at"`
	UpdatedBy string                    `bson:"updated_by"`
}

type privacyMaskingDocument struct {
	UserIDMode          string `bson:"user_id_mode"`
	AnonymousIDMode     string `bson:"anonymous_id_mode"`
	SessionIDMode       string `bson:"session_id_mode"`
	LocationGranularity string `bson:"location_granularity"`
	ShowSearchQueries   bool   `bson:"show_search_queries"`
	ShowIPHash          bool   `bson:"show_ip_hash"`
}

type retentionSettingsDocument struct {
	RawEventsDays             int `bson:"raw_events_days"`
	JourneySummariesDays      int `bson:"journey_summaries_days"`
	HeatmapAggregatesDays     int `bson:"heatmap_aggregates_days"`
	AnalyticsAggregatesMonths int `bson:"analytics_aggregates_months"`
	ActiveSessionTTLMinutes   int `bson:"active_session_ttl_minutes"`
	DeletionRequestLogDays    int `bson:"deletion_request_log_days"`
}

type deletionRequestDocument struct {
	ID                      string     `bson:"_id"`
	TargetType              string     `bson:"target_type"`
	TargetHash              string     `bson:"target_hash"`
	TargetValueMasked       string     `bson:"target_value_masked"`
	Status                  string     `bson:"status"`
	RequestedBy             string     `bson:"requested_by"`
	Reason                  string     `bson:"reason"`
	MatchedSessions         int64      `bson:"matched_sessions"`
	MatchedEvents           int64      `bson:"matched_events"`
	MatchedJourneySummaries int64      `bson:"matched_journey_summaries"`
	MatchedActiveSessions   int64      `bson:"matched_active_sessions"`
	CreatedAt               time.Time  `bson:"created_at"`
	CompletedAt             *time.Time `bson:"completed_at,omitempty"`
	Error                   string     `bson:"error,omitempty"`
}

type auditEventDocument struct {
	ID           string         `bson:"_id"`
	ActorID      string         `bson:"actor_id"`
	Action       string         `bson:"action"`
	ResourceType string         `bson:"resource_type"`
	ResourceID   string         `bson:"resource_id"`
	RequestID    string         `bson:"request_id,omitempty"`
	Reason       string         `bson:"reason,omitempty"`
	Metadata     map[string]any `bson:"metadata,omitempty"`
	CreatedAt    time.Time      `bson:"created_at"`
}

func privacySettingsDocumentFromDomain(settings domain.PrivacySettings) privacySettingsDocument {
	return privacySettingsDocument{
		ID: privacySettingsID,
		Masking: privacyMaskingDocument{
			UserIDMode:          string(settings.Masking.UserIDMode),
			AnonymousIDMode:     string(settings.Masking.AnonymousIDMode),
			SessionIDMode:       string(settings.Masking.SessionIDMode),
			LocationGranularity: string(settings.Masking.LocationGranularity),
			ShowSearchQueries:   settings.Masking.ShowSearchQueries,
			ShowIPHash:          settings.Masking.ShowIPHash,
		},
		Retention: retentionSettingsDocument{
			RawEventsDays:             settings.Retention.RawEventsDays,
			JourneySummariesDays:      settings.Retention.JourneySummariesDays,
			HeatmapAggregatesDays:     settings.Retention.HeatmapAggregatesDays,
			AnalyticsAggregatesMonths: settings.Retention.AnalyticsAggregatesMonths,
			ActiveSessionTTLMinutes:   settings.Retention.ActiveSessionTTLMinutes,
			DeletionRequestLogDays:    settings.Retention.DeletionRequestLogDays,
		},
		UpdatedAt: settings.UpdatedAt,
		UpdatedBy: settings.UpdatedBy,
	}
}

func (doc privacySettingsDocument) toDomain() domain.PrivacySettings {
	return domain.PrivacySettings{
		Masking: domain.PrivacyMaskingSettings{
			UserIDMode:          domain.MaskingMode(doc.Masking.UserIDMode),
			AnonymousIDMode:     domain.MaskingMode(doc.Masking.AnonymousIDMode),
			SessionIDMode:       domain.MaskingMode(doc.Masking.SessionIDMode),
			LocationGranularity: domain.LocationGranularity(doc.Masking.LocationGranularity),
			ShowSearchQueries:   doc.Masking.ShowSearchQueries,
			ShowIPHash:          doc.Masking.ShowIPHash,
		},
		Retention: domain.RetentionSettings{
			RawEventsDays:             doc.Retention.RawEventsDays,
			JourneySummariesDays:      doc.Retention.JourneySummariesDays,
			HeatmapAggregatesDays:     doc.Retention.HeatmapAggregatesDays,
			AnalyticsAggregatesMonths: doc.Retention.AnalyticsAggregatesMonths,
			ActiveSessionTTLMinutes:   doc.Retention.ActiveSessionTTLMinutes,
			DeletionRequestLogDays:    doc.Retention.DeletionRequestLogDays,
		},
		UpdatedAt: doc.UpdatedAt,
		UpdatedBy: doc.UpdatedBy,
	}
}

func deletionRequestDocumentFromDomain(request domain.DeletionRequest) deletionRequestDocument {
	return deletionRequestDocument{
		ID:                      request.RequestID,
		TargetType:              string(request.TargetType),
		TargetHash:              request.TargetHash,
		TargetValueMasked:       request.TargetValueMasked,
		Status:                  string(request.Status),
		RequestedBy:             request.RequestedBy,
		Reason:                  request.Reason,
		MatchedSessions:         request.MatchedSessions,
		MatchedEvents:           request.MatchedEvents,
		MatchedJourneySummaries: request.MatchedJourneySummaries,
		MatchedActiveSessions:   request.MatchedActiveSessions,
		CreatedAt:               request.CreatedAt,
		CompletedAt:             request.CompletedAt,
		Error:                   request.Error,
	}
}

func (doc deletionRequestDocument) toDomain() domain.DeletionRequest {
	return domain.DeletionRequest{
		RequestID:               doc.ID,
		TargetType:              domain.DeletionTargetType(doc.TargetType),
		TargetHash:              doc.TargetHash,
		TargetValueMasked:       doc.TargetValueMasked,
		Status:                  domain.DeletionRequestStatus(doc.Status),
		RequestedBy:             doc.RequestedBy,
		Reason:                  doc.Reason,
		MatchedSessions:         doc.MatchedSessions,
		MatchedEvents:           doc.MatchedEvents,
		MatchedJourneySummaries: doc.MatchedJourneySummaries,
		MatchedActiveSessions:   doc.MatchedActiveSessions,
		CreatedAt:               doc.CreatedAt,
		CompletedAt:             doc.CompletedAt,
		Error:                   doc.Error,
	}
}

func auditEventDocumentFromDomain(event domain.AuditEvent) auditEventDocument {
	return auditEventDocument{
		ID:           event.EventID,
		ActorID:      event.ActorID,
		Action:       event.Action,
		ResourceType: event.ResourceType,
		ResourceID:   event.ResourceID,
		RequestID:    event.RequestID,
		Reason:       event.Reason,
		Metadata:     event.Metadata,
		CreatedAt:    event.CreatedAt,
	}
}
