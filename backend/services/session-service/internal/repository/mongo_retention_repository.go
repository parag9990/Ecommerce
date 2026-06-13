package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const retentionDeletionAuditsCollectionName = "retention_deletion_audits"
const anonymizedIdentityValue = "anonymized"

type MongoRetentionRepository struct {
	sessions       *mongo.Collection
	events         *mongo.Collection
	journeys       *mongo.Collection
	heatmapPoints  *mongo.Collection
	heatmapMarkers *mongo.Collection
	aggregates     *mongo.Collection
	audits         *mongo.Collection
	logger         *slog.Logger
}

func NewMongoRetentionRepository(database *mongo.Database, logger *slog.Logger) (*MongoRetentionRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &MongoRetentionRepository{
		sessions:       database.Collection(sessionsCollectionName),
		events:         database.Collection(sessionEventsCollectionName),
		journeys:       database.Collection(journeySummariesCollectionName),
		heatmapPoints:  database.Collection(heatmapPointsCollectionName),
		heatmapMarkers: database.Collection(heatmapBucketSessionsCollectionName),
		aggregates:     database.Collection(analyticsAggregatesCollectionName),
		audits:         database.Collection(retentionDeletionAuditsCollectionName),
		logger:         logger,
	}, nil
}

func (r *MongoRetentionRepository) EnsureIndexes(ctx context.Context) error {
	sessionIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "retain_until", Value: 1}, {Key: "legal_hold", Value: 1}, {Key: "anonymized_at", Value: 1}},
			Options: options.Index().SetName("idx_sessions_retention_due"),
		},
		{
			Keys:    bson.D{{Key: "retention_class", Value: 1}, {Key: "last_seen_at", Value: -1}},
			Options: options.Index().SetName("idx_sessions_retention_class_last_seen"),
		},
	}
	if _, err := r.sessions.Indexes().CreateMany(ctx, sessionIndexes); err != nil {
		return fmt.Errorf("create session retention indexes: %w", err)
	}

	journeyIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "retain_until", Value: 1}, {Key: "legal_hold", Value: 1}, {Key: "anonymized_at", Value: 1}},
			Options: options.Index().SetName("idx_journey_retention_due"),
		},
		{
			Keys:    bson.D{{Key: "retention_class", Value: 1}, {Key: "last_event_at", Value: -1}},
			Options: options.Index().SetName("idx_journey_retention_class_recent"),
		},
	}
	if _, err := r.journeys.Indexes().CreateMany(ctx, journeyIndexes); err != nil {
		return fmt.Errorf("create journey retention indexes: %w", err)
	}

	heatmapIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "retain_until", Value: 1}},
			Options: options.Index().SetName("idx_heatmap_retention_due"),
		},
	}
	if _, err := r.heatmapPoints.Indexes().CreateMany(ctx, heatmapIndexes); err != nil {
		return fmt.Errorf("create heatmap retention indexes: %w", err)
	}

	markerIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "first_seen_at", Value: 1}},
			Options: options.Index().SetName("idx_heatmap_markers_first_seen"),
		},
	}
	if _, err := r.heatmapMarkers.Indexes().CreateMany(ctx, markerIndexes); err != nil {
		return fmt.Errorf("create heatmap marker retention indexes: %w", err)
	}

	aggregateIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "contains_pii", Value: 1}, {Key: "updated_at", Value: -1}},
			Options: options.Index().SetName("idx_aggregates_pii_policy"),
		},
		{
			Keys:    bson.D{{Key: "bucket_end", Value: 1}, {Key: "calculated_at", Value: 1}},
			Options: options.Index().SetName("idx_aggregates_retention_due"),
		},
	}
	if _, err := r.aggregates.Indexes().CreateMany(ctx, aggregateIndexes); err != nil {
		return fmt.Errorf("create aggregate retention indexes: %w", err)
	}

	auditIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "request_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_retention_deletion_request"),
		},
		{
			Keys:    bson.D{{Key: "completed_at", Value: -1}},
			Options: options.Index().SetName("idx_retention_deletion_completed"),
		},
		{
			Keys:    bson.D{{Key: "user_id_hash", Value: 1}, {Key: "completed_at", Value: -1}},
			Options: options.Index().SetName("idx_retention_deletion_user_hash"),
		},
	}
	if _, err := r.audits.Indexes().CreateMany(ctx, auditIndexes); err != nil {
		return fmt.Errorf("create deletion audit indexes: %w", err)
	}
	r.logger.InfoContext(ctx, "session.mongo.retention_indexes_ready")
	return nil
}

func (r *MongoRetentionRepository) AnonymizeExpiredSessions(ctx context.Context, now time.Time, retentionDays int, limit int, dryRun bool) (int64, error) {
	now = normalizeRepositoryTime(now)
	filter := expiredSessionFilter(now, retentionDays)
	ids, err := collectIDs(ctx, r.sessions, filter, limit)
	if err != nil {
		return 0, fmt.Errorf("find expired sessions: %w", err)
	}
	if dryRun || len(ids) == 0 {
		return int64(len(ids)), nil
	}
	result, err := r.sessions.UpdateMany(ctx, idsFilter(ids), anonymizeSessionUpdate(now))
	if err != nil {
		return 0, fmt.Errorf("anonymize expired sessions: %w", err)
	}
	return result.ModifiedCount, nil
}

func (r *MongoRetentionRepository) AnonymizeExpiredJourneys(ctx context.Context, now time.Time, retentionDays int, limit int, dryRun bool) (int64, error) {
	now = normalizeRepositoryTime(now)
	filter := expiredJourneyFilter(now, retentionDays)
	ids, err := collectIDs(ctx, r.journeys, filter, limit)
	if err != nil {
		return 0, fmt.Errorf("find expired journey summaries: %w", err)
	}
	if dryRun || len(ids) == 0 {
		return int64(len(ids)), nil
	}
	result, err := r.journeys.UpdateMany(ctx, idsFilter(ids), anonymizeJourneyUpdate(now))
	if err != nil {
		return 0, fmt.Errorf("anonymize expired journey summaries: %w", err)
	}
	return result.ModifiedCount, nil
}

func (r *MongoRetentionRepository) PurgeExpiredHeatmap(ctx context.Context, now time.Time, retentionDays int, limit int, dryRun bool) (int64, int64, error) {
	now = normalizeRepositoryTime(now)
	cutoff := now.AddDate(0, 0, -retentionDays)
	filter := bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "retain_until", Value: bson.D{{Key: "$lte", Value: now}}}},
		bson.D{
			{Key: "retain_until", Value: bson.D{{Key: "$exists", Value: false}}},
			{Key: "day", Value: bson.D{{Key: "$lt", Value: cutoff.Format("2006-01-02")}}},
		},
	}}}
	ids, err := collectIDs(ctx, r.heatmapPoints, filter, limit)
	if err != nil {
		return 0, 0, fmt.Errorf("find expired heatmap points: %w", err)
	}
	if len(ids) == 0 {
		return 0, 0, nil
	}
	markerFilter := bson.D{{Key: "point_id", Value: bson.D{{Key: "$in", Value: ids}}}}
	markerCount, err := r.heatmapMarkers.CountDocuments(ctx, markerFilter)
	if err != nil {
		return 0, 0, fmt.Errorf("count expired heatmap markers: %w", err)
	}
	if dryRun {
		return int64(len(ids)), markerCount, nil
	}
	markerResult, err := r.heatmapMarkers.DeleteMany(ctx, markerFilter)
	if err != nil {
		return 0, 0, fmt.Errorf("purge heatmap markers: %w", err)
	}
	pointResult, err := r.heatmapPoints.DeleteMany(ctx, idsFilter(ids))
	if err != nil {
		return 0, 0, fmt.Errorf("purge heatmap points: %w", err)
	}
	return pointResult.DeletedCount, markerResult.DeletedCount, nil
}

func (r *MongoRetentionRepository) PurgeExpiredAggregates(ctx context.Context, now time.Time, retentionYears int, limit int, dryRun bool) (int64, error) {
	now = normalizeRepositoryTime(now)
	cutoff := now.AddDate(-retentionYears, 0, 0)
	filter := bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "bucket_end", Value: bson.D{{Key: "$lt", Value: cutoff}}}},
		bson.D{{Key: "cohort_end", Value: bson.D{{Key: "$lt", Value: cutoff}}}},
		bson.D{
			{Key: "bucket_end", Value: bson.D{{Key: "$exists", Value: false}}},
			{Key: "cohort_end", Value: bson.D{{Key: "$exists", Value: false}}},
			{Key: "calculated_at", Value: bson.D{{Key: "$lt", Value: cutoff}}},
		},
	}}}
	ids, err := collectIDs(ctx, r.aggregates, filter, limit)
	if err != nil {
		return 0, fmt.Errorf("find expired analytics aggregates: %w", err)
	}
	if dryRun || len(ids) == 0 {
		return int64(len(ids)), nil
	}
	result, err := r.aggregates.DeleteMany(ctx, idsFilter(ids))
	if err != nil {
		return 0, fmt.Errorf("purge expired analytics aggregates: %w", err)
	}
	return result.DeletedCount, nil
}

func (r *MongoRetentionRepository) CountAggregatesWithPII(ctx context.Context) (int64, error) {
	count, err := r.aggregates.CountDocuments(ctx, aggregatePIIFilter())
	if err != nil {
		return 0, fmt.Errorf("count aggregates with pii: %w", err)
	}
	return count, nil
}

func (r *MongoRetentionRepository) CountLegalHoldRetentionDue(ctx context.Context, now time.Time, sessionDays int, journeyDays int) (int64, error) {
	now = normalizeRepositoryTime(now)
	sessionCount, err := r.sessions.CountDocuments(ctx, legalHoldDueFilter(expiredSessionDueClause(now, sessionDays)))
	if err != nil {
		return 0, fmt.Errorf("count legal hold sessions: %w", err)
	}
	journeyCount, err := r.journeys.CountDocuments(ctx, legalHoldDueFilter(expiredJourneyDueClause(now, journeyDays)))
	if err != nil {
		return 0, fmt.Errorf("count legal hold journeys: %w", err)
	}
	return sessionCount + journeyCount, nil
}

func (r *MongoRetentionRepository) DeleteRawEventsByIdentity(ctx context.Context, identity domain.SessionDataIdentity, dryRun bool) (int64, error) {
	filter, err := identityFilter(identity)
	if err != nil {
		return 0, err
	}
	if dryRun {
		count, err := r.events.CountDocuments(ctx, filter)
		if err != nil {
			return 0, fmt.Errorf("count raw events for deletion: %w", err)
		}
		return count, nil
	}
	result, err := r.events.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("delete raw events by identity: %w", err)
	}
	return result.DeletedCount, nil
}

func (r *MongoRetentionRepository) DeleteSessionsByIdentity(ctx context.Context, identity domain.SessionDataIdentity, dryRun bool) (int64, error) {
	filter, err := identityFilter(identity)
	if err != nil {
		return 0, err
	}
	if dryRun {
		count, err := r.sessions.CountDocuments(ctx, filter)
		if err != nil {
			return 0, fmt.Errorf("count sessions for deletion: %w", err)
		}
		return count, nil
	}
	result, err := r.sessions.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("delete sessions by identity: %w", err)
	}
	return result.DeletedCount, nil
}

func (r *MongoRetentionRepository) AnonymizeSessionsByIdentity(ctx context.Context, identity domain.SessionDataIdentity, at time.Time, dryRun bool) (int64, error) {
	filter, err := identityFilter(identity)
	if err != nil {
		return 0, err
	}
	at = normalizeRepositoryTime(at)
	if dryRun {
		count, err := r.sessions.CountDocuments(ctx, filter)
		if err != nil {
			return 0, fmt.Errorf("count sessions for anonymization: %w", err)
		}
		return count, nil
	}
	result, err := r.sessions.UpdateMany(ctx, filter, anonymizeSessionForDeletionUpdate(at))
	if err != nil {
		return 0, fmt.Errorf("anonymize sessions by identity: %w", err)
	}
	return result.ModifiedCount, nil
}

func (r *MongoRetentionRepository) DeleteJourneysByIdentity(ctx context.Context, identity domain.SessionDataIdentity, dryRun bool) (int64, error) {
	filter, err := identityFilter(identity)
	if err != nil {
		return 0, err
	}
	if dryRun {
		count, err := r.journeys.CountDocuments(ctx, filter)
		if err != nil {
			return 0, fmt.Errorf("count journeys for deletion: %w", err)
		}
		return count, nil
	}
	result, err := r.journeys.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("delete journeys by identity: %w", err)
	}
	return result.DeletedCount, nil
}

func (r *MongoRetentionRepository) AnonymizeJourneysByIdentity(ctx context.Context, identity domain.SessionDataIdentity, at time.Time, dryRun bool) (int64, error) {
	filter, err := identityFilter(identity)
	if err != nil {
		return 0, err
	}
	at = normalizeRepositoryTime(at)
	if dryRun {
		count, err := r.journeys.CountDocuments(ctx, filter)
		if err != nil {
			return 0, fmt.Errorf("count journeys for anonymization: %w", err)
		}
		return count, nil
	}
	result, err := r.journeys.UpdateMany(ctx, filter, anonymizeJourneyForDeletionUpdate(at))
	if err != nil {
		return 0, fmt.Errorf("anonymize journeys by identity: %w", err)
	}
	return result.ModifiedCount, nil
}

func (r *MongoRetentionRepository) WriteDeletionAudit(ctx context.Context, request domain.DeleteUserSessionDataRequest, result domain.DeleteUserSessionDataResult) error {
	req := request.Normalize()
	doc := bson.D{
		{Key: "_id", Value: result.RequestID},
		{Key: "request_id", Value: result.RequestID},
		{Key: "user_id_hash", Value: domain.HashRetentionIdentifier(req.Identity.UserID)},
		{Key: "anonymous_id_hashes", Value: hashStrings(req.Identity.AnonymousIDs)},
		{Key: "reason", Value: req.Reason},
		{Key: "hard_delete", Value: req.HardDelete},
		{Key: "requested_by", Value: req.RequestedBy},
		{Key: "requested_at", Value: req.RequestedAt},
		{Key: "raw_events_deleted", Value: result.RawEventsDeleted},
		{Key: "sessions_anonymized", Value: result.SessionsAnonymized},
		{Key: "sessions_deleted", Value: result.SessionsDeleted},
		{Key: "journeys_anonymized", Value: result.JourneysAnonymized},
		{Key: "journeys_deleted", Value: result.JourneysDeleted},
		{Key: "redis_keys_deleted", Value: result.RedisKeysDeleted},
		{Key: "redis_errors", Value: result.RedisErrors},
		{Key: "completed_at", Value: result.CompletedAt},
		{Key: "retention_class", Value: string(domain.RetentionClassDeletionAudit)},
		{Key: "created_at", Value: time.Now().UTC()},
	}
	if _, err := r.audits.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: duplicate deletion request_id", domain.ErrInvalidRetentionPolicy)
		}
		return fmt.Errorf("write deletion audit: %w", err)
	}
	return nil
}

func (r *MongoRetentionRepository) SetLegalHold(ctx context.Context, request domain.SetLegalHoldRequest) (domain.SetLegalHoldResult, error) {
	req := request.Normalize()
	if err := req.Validate(); err != nil {
		return domain.SetLegalHoldResult{}, err
	}
	now := normalizeRepositoryTime(req.SetAt)
	sessionFilter := bson.D{{Key: "session_id", Value: req.SessionID}}
	update := legalHoldUpdate(req, now)
	sessionResult, err := r.sessions.UpdateOne(ctx, sessionFilter, update)
	if err != nil {
		return domain.SetLegalHoldResult{}, fmt.Errorf("set session legal hold: %w", err)
	}
	if sessionResult.MatchedCount == 0 {
		return domain.SetLegalHoldResult{}, domain.ErrSessionNotFound
	}
	journeyResult, err := r.journeys.UpdateMany(ctx, sessionFilter, update)
	if err != nil {
		return domain.SetLegalHoldResult{}, fmt.Errorf("set journey legal hold: %w", err)
	}
	return domain.SetLegalHoldResult{
		SessionID:       req.SessionID,
		LegalHold:       req.Hold,
		SessionsMatched: sessionResult.MatchedCount,
		JourneysMatched: journeyResult.MatchedCount,
		UpdatedAt:       now,
	}, nil
}

func collectIDs(ctx context.Context, collection *mongo.Collection, filter any, limit int) ([]any, error) {
	if limit <= 0 {
		limit = domain.DefaultRetentionWorkerBatchSize
	}
	cursor, err := collection.Find(ctx, filter, options.Find().
		SetProjection(bson.D{{Key: "_id", Value: 1}}).
		SetSort(bson.D{{Key: "_id", Value: 1}}).
		SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	ids := make([]any, 0)
	for cursor.Next(ctx) {
		var row struct {
			ID any `bson:"_id"`
		}
		if err := cursor.Decode(&row); err != nil {
			return nil, err
		}
		ids = append(ids, row.ID)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

func idsFilter(ids []any) bson.D {
	return bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}}
}

func expiredSessionFilter(now time.Time, retentionDays int) bson.D {
	return retentionFilter(false, expiredSessionDueClause(now, retentionDays))
}

func expiredJourneyFilter(now time.Time, retentionDays int) bson.D {
	return retentionFilter(false, expiredJourneyDueClause(now, retentionDays))
}

func retentionFilter(legalHold bool, dueClause bson.D) bson.D {
	holdValue := bson.D{{Key: "$ne", Value: true}}
	if legalHold {
		holdValue = bson.D{{Key: "$eq", Value: true}}
	}
	return bson.D{{Key: "$and", Value: bson.A{
		bson.D{{Key: "legal_hold", Value: holdValue}},
		bson.D{{Key: "$or", Value: bson.A{
			bson.D{{Key: "anonymized_at", Value: bson.D{{Key: "$exists", Value: false}}}},
			bson.D{{Key: "anonymized_at", Value: nil}},
		}}},
		dueClause,
	}}}
}

func legalHoldDueFilter(dueClause bson.D) bson.D {
	return retentionFilter(true, dueClause)
}

func expiredSessionDueClause(now time.Time, retentionDays int) bson.D {
	cutoff := now.AddDate(0, 0, -retentionDays)
	return bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "retain_until", Value: bson.D{{Key: "$lte", Value: now}}}},
		bson.D{
			{Key: "retain_until", Value: bson.D{{Key: "$exists", Value: false}}},
			{Key: "last_seen_at", Value: bson.D{{Key: "$lte", Value: cutoff}}},
		},
	}}}
}

func expiredJourneyDueClause(now time.Time, retentionDays int) bson.D {
	cutoff := now.AddDate(0, 0, -retentionDays)
	return bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "retain_until", Value: bson.D{{Key: "$lte", Value: now}}}},
		bson.D{
			{Key: "retain_until", Value: bson.D{{Key: "$exists", Value: false}}},
			{Key: "last_event_at", Value: bson.D{{Key: "$lte", Value: cutoff}}},
		},
	}}}
}

func anonymizeSessionUpdate(now time.Time) bson.D {
	return bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "anonymous_id", Value: anonymizedIdentityValue},
			{Key: "user_agent", Value: anonymizedIdentityValue},
			{Key: "ip_hash", Value: anonymizedIdentityValue},
			{Key: "geo.source", Value: domain.GeoSourceUnknown},
			{Key: "anonymized_at", Value: now},
			{Key: "updated_at", Value: now},
			{Key: "retention_class", Value: string(domain.RetentionClassSessionMetadata)},
		}},
		{Key: "$unset", Value: bson.D{
			{Key: "user_id", Value: ""},
			{Key: "referrer", Value: ""},
			{Key: "user_agent_hash", Value: ""},
			{Key: "device_fingerprint_hash", Value: ""},
			{Key: "auth_session_id", Value: ""},
			{Key: "geo.region", Value: ""},
			{Key: "geo.city", Value: ""},
			{Key: "geo.timezone", Value: ""},
		}},
	}
}

func anonymizeSessionForDeletionUpdate(now time.Time) bson.D {
	update := anonymizeSessionUpdate(now)
	update[0].Value = append(update[0].Value.(bson.D), bson.E{Key: "deletion_requested_at", Value: now})
	return update
}

func anonymizeJourneyUpdate(now time.Time) bson.D {
	return bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "anonymous_id", Value: anonymizedIdentityValue},
			{Key: "anonymized_at", Value: now},
			{Key: "updated_at", Value: now},
			{Key: "retention_class", Value: string(domain.RetentionClassDerivedSummary)},
		}},
		{Key: "$unset", Value: bson.D{
			{Key: "user_id", Value: ""},
		}},
	}
}

func anonymizeJourneyForDeletionUpdate(now time.Time) bson.D {
	return anonymizeJourneyUpdate(now)
}

func identityFilter(identity domain.SessionDataIdentity) (bson.D, error) {
	normalized := identity.Normalize()
	if err := normalized.Validate(); err != nil {
		return nil, err
	}
	ors := make(bson.A, 0, 2)
	if normalized.UserID != "" {
		ors = append(ors, bson.D{{Key: "user_id", Value: normalized.UserID}})
	}
	if len(normalized.AnonymousIDs) > 0 {
		ors = append(ors, bson.D{{Key: "anonymous_id", Value: bson.D{{Key: "$in", Value: normalized.AnonymousIDs}}}})
	}
	return bson.D{{Key: "$or", Value: ors}}, nil
}

func aggregatePIIFilter() bson.D {
	return bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "contains_pii", Value: true}},
		bson.D{{Key: "user_id", Value: bson.D{{Key: "$exists", Value: true}}}},
		bson.D{{Key: "anonymous_id", Value: bson.D{{Key: "$exists", Value: true}}}},
		bson.D{{Key: "ip_hash", Value: bson.D{{Key: "$exists", Value: true}}}},
		bson.D{{Key: "segment.user_id", Value: bson.D{{Key: "$exists", Value: true}}}},
		bson.D{{Key: "segment.anonymous_id", Value: bson.D{{Key: "$exists", Value: true}}}},
		bson.D{{Key: "segment.ip_hash", Value: bson.D{{Key: "$exists", Value: true}}}},
	}}}
}

func legalHoldUpdate(request domain.SetLegalHoldRequest, now time.Time) bson.D {
	if request.Hold {
		return bson.D{{Key: "$set", Value: bson.D{
			{Key: "legal_hold", Value: true},
			{Key: "legal_hold_reason", Value: request.Reason},
			{Key: "legal_hold_set_by", Value: request.ActorID},
			{Key: "legal_hold_set_at", Value: now},
			{Key: "updated_at", Value: now},
		}}}
	}
	return bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "legal_hold", Value: false},
			{Key: "updated_at", Value: now},
		}},
		{Key: "$unset", Value: bson.D{
			{Key: "legal_hold_reason", Value: ""},
			{Key: "legal_hold_set_by", Value: ""},
			{Key: "legal_hold_set_at", Value: ""},
		}},
	}
}

func normalizeRepositoryTime(value time.Time) time.Time {
	value = value.UTC()
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value
}

func hashStrings(values []string) []string {
	values = (domain.SessionDataIdentity{AnonymousIDs: values}).Normalize().AnonymousIDs
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if hashed := domain.HashRetentionIdentifier(value); hashed != "" {
			out = append(out, hashed)
		}
	}
	return out
}
