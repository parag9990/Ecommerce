package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const sessionsCollectionName = "sessions"
const defaultSessionMetadataRetentionDays = 365

type MongoSessionRepository struct {
	collection            *mongo.Collection
	validation            domain.ValidationConfig
	metadataRetentionDays int
	logger                *slog.Logger
}

type mongoSessionDocument struct {
	ID             string `bson:"_id"`
	domain.Session `bson:",inline"`
}

type MongoSessionRepositoryOption func(*MongoSessionRepository)

func WithSessionMetadataRetentionDays(days int) MongoSessionRepositoryOption {
	return func(r *MongoSessionRepository) {
		if days > 0 {
			r.metadataRetentionDays = days
		}
	}
}

func NewMongoSessionRepository(database *mongo.Database, validation domain.ValidationConfig, logger *slog.Logger, opts ...MongoSessionRepositoryOption) (*MongoSessionRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	repo := &MongoSessionRepository{
		collection:            database.Collection(sessionsCollectionName),
		validation:            validation.WithDefaults(),
		metadataRetentionDays: defaultSessionMetadataRetentionDays,
		logger:                logger,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(repo)
		}
	}
	return repo, nil
}

func (r *MongoSessionRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_session_id"),
		},
		{
			Keys:    bson.D{{Key: "anonymous_id", Value: 1}, {Key: "started_at", Value: -1}},
			Options: options.Index().SetName("idx_anonymous_sessions"),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "started_at", Value: -1}},
			Options: options.Index().
				SetName("idx_user_sessions").
				SetPartialFilterExpression(bson.D{{Key: "user_id", Value: bson.D{{Key: "$type", Value: "string"}}}}),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "last_seen_at", Value: -1}},
			Options: options.Index().SetName("idx_status_last_seen"),
		},
		{
			Keys:    bson.D{{Key: "channel", Value: 1}, {Key: "started_at", Value: -1}},
			Options: options.Index().SetName("idx_sessions_channel_started"),
		},
		{
			Keys:    bson.D{{Key: "device.type", Value: 1}, {Key: "started_at", Value: -1}},
			Options: options.Index().SetName("idx_sessions_device_type_started"),
		},
		{
			Keys:    bson.D{{Key: "device.browser", Value: 1}, {Key: "device.os", Value: 1}, {Key: "started_at", Value: -1}},
			Options: options.Index().SetName("idx_sessions_browser_os_started"),
		},
		{
			Keys:    bson.D{{Key: "geo.country", Value: 1}, {Key: "started_at", Value: -1}},
			Options: options.Index().SetName("idx_sessions_country_started"),
		},
		{
			Keys:    bson.D{{Key: "ip_hash", Value: 1}, {Key: "last_seen_at", Value: -1}},
			Options: options.Index().SetName("idx_sessions_ip_hash_last_seen"),
		},
		{
			Keys: bson.D{{Key: "device_fingerprint_hash", Value: 1}, {Key: "last_seen_at", Value: -1}},
			Options: options.Index().
				SetName("idx_sessions_device_fingerprint_last_seen").
				SetPartialFilterExpression(bson.D{{Key: "device_fingerprint_hash", Value: bson.D{{Key: "$type", Value: "string"}}}}),
		},
		{
			Keys:    bson.D{{Key: "retain_until", Value: 1}, {Key: "legal_hold", Value: 1}, {Key: "anonymized_at", Value: 1}},
			Options: options.Index().SetName("idx_sessions_retention_due"),
		},
	}
	if _, err := r.collection.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("create session indexes: %w", err)
	}
	r.logger.InfoContext(ctx, "session.mongo.session_indexes_ready")
	return nil
}

func (r *MongoSessionRepository) UpsertSession(ctx context.Context, session domain.Session) error {
	normalized := session.Normalize()
	normalized = r.applyRetentionDefaults(normalized)
	if err := normalized.ValidateWithConfig(r.validation); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInvalidSession, err)
	}

	doc := mongoSessionDocument{
		ID:      normalized.SessionID,
		Session: normalized,
	}
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.D{{Key: "session_id", Value: normalized.SessionID}},
		doc,
		options.Replace().SetUpsert(true),
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: duplicate session_id", domain.ErrInvalidSession)
		}
		return fmt.Errorf("upsert session: %w", err)
	}
	return nil
}

func (r *MongoSessionRepository) TouchSession(ctx context.Context, touch domain.SessionTouch) error {
	normalized := touch.Normalize()
	if normalized.SchemaVersion == 0 {
		normalized.SchemaVersion = domain.CurrentSessionSchemaVersion
	}
	if err := normalized.Validate(); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInvalidSession, err)
	}

	entryPage := "/"
	if normalized.Path != nil {
		entryPage = *normalized.Path
	}

	setOnInsert := bson.D{
		{Key: "_id", Value: normalized.SessionID},
		{Key: "session_id", Value: normalized.SessionID},
		{Key: "anonymous_id", Value: normalized.AnonymousID},
		{Key: "schema_version", Value: domain.CurrentSessionSchemaVersion},
		{Key: "entry_page", Value: entryPage},
		{Key: "utm", Value: bson.D{}},
		{Key: "risk_level", Value: domain.RiskLevelUnknown},
		{Key: "risk_reasons", Value: bson.A{}},
		{Key: "started_at", Value: normalized.OccurredAt},
		{Key: "created_at", Value: normalized.ReceivedAt},
	}
	set := bson.D{
		{Key: "status", Value: domain.SessionStatusActive},
		{Key: "channel", Value: normalized.Channel},
		{Key: "user_agent", Value: normalized.UserAgent},
		{Key: "device", Value: normalized.Device},
		{Key: "client", Value: normalized.Client},
		{Key: "ip_hash", Value: normalized.IPHash},
		{Key: "ip_version", Value: normalized.IPVersion},
		{Key: "geo", Value: normalized.Geo},
		{Key: "last_seen_at", Value: normalized.OccurredAt},
		{Key: "updated_at", Value: normalized.ReceivedAt},
	}
	if normalized.UserAgentHash != nil {
		set = append(set, bson.E{Key: "user_agent_hash", Value: *normalized.UserAgentHash})
	}
	if normalized.UserID != nil {
		set = append(set, bson.E{Key: "user_id", Value: *normalized.UserID})
	}
	if normalized.Path != nil {
		set = append(set, bson.E{Key: "exit_page", Value: *normalized.Path})
	}
	if normalized.DeviceFingerprintHash != nil {
		set = append(set, bson.E{Key: "device_fingerprint_hash", Value: *normalized.DeviceFingerprintHash})
	}
	retainUntil := normalized.OccurredAt.AddDate(0, 0, r.metadataRetentionDays)
	set = append(set,
		bson.E{Key: "retain_until", Value: retainUntil},
		bson.E{Key: "retention_class", Value: string(domain.RetentionClassSessionMetadata)},
	)
	setOnInsert = append(setOnInsert, bson.E{Key: "legal_hold", Value: false})

	_, err := r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "session_id", Value: normalized.SessionID}},
		bson.D{{Key: "$setOnInsert", Value: setOnInsert}, {Key: "$set", Value: set}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: duplicate session_id", domain.ErrInvalidSession)
		}
		return fmt.Errorf("touch session: %w", err)
	}
	return nil
}

func (r *MongoSessionRepository) applyRetentionDefaults(session domain.Session) domain.Session {
	out := session.Normalize()
	if out.RetentionClass == "" {
		out.RetentionClass = domain.RetentionClassSessionMetadata
	}
	if out.RetainUntil == nil {
		base := out.LastSeenAt
		if base.IsZero() {
			base = out.UpdatedAt
		}
		if base.IsZero() {
			base = time.Now().UTC()
		}
		retainUntil := base.UTC().AddDate(0, 0, r.metadataRetentionDays)
		out.RetainUntil = &retainUntil
	}
	return out
}

func (r *MongoSessionRepository) FindSessionByID(ctx context.Context, sessionID string) (domain.Session, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return domain.Session{}, fmt.Errorf("%w: session_id is required", domain.ErrInvalidSession)
	}

	var doc mongoSessionDocument
	err := r.collection.FindOne(ctx, bson.D{{Key: "session_id", Value: sessionID}}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Session{}, domain.ErrSessionNotFound
		}
		return domain.Session{}, fmt.Errorf("find session by id: %w", err)
	}
	return doc.Session.Normalize(), nil
}

func (r *MongoSessionRepository) ListUserSessions(ctx context.Context, userID string, limit int) ([]domain.Session, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", domain.ErrInvalidSession)
	}
	if limit <= 0 {
		limit = 100
	}

	cursor, err := r.collection.Find(
		ctx,
		bson.D{{Key: "user_id", Value: userID}},
		options.Find().SetSort(bson.D{{Key: "started_at", Value: -1}}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, fmt.Errorf("list user sessions: %w", err)
	}
	defer cursor.Close(ctx)

	sessions := make([]domain.Session, 0)
	for cursor.Next(ctx) {
		var doc mongoSessionDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode user session: %w", err)
		}
		sessions = append(sessions, doc.Session.Normalize())
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate user sessions: %w", err)
	}
	return sessions, nil
}

func (r *MongoSessionRepository) ListSessions(ctx context.Context, filter domain.SessionListFilter) ([]domain.Session, int64, error) {
	normalized := filter.Normalize()
	if normalized.Page <= 0 {
		normalized.Page = 1
	}
	if normalized.PageSize <= 0 {
		normalized.PageSize = 100
	}

	query := bson.D{}
	if normalized.UserID != "" {
		query = append(query, bson.E{Key: "user_id", Value: normalized.UserID})
	}
	if normalized.AnonymousID != "" {
		query = append(query, bson.E{Key: "anonymous_id", Value: normalized.AnonymousID})
	}
	if !normalized.From.IsZero() || !normalized.To.IsZero() {
		rangeFilter := bson.D{}
		if !normalized.From.IsZero() {
			rangeFilter = append(rangeFilter, bson.E{Key: "$gte", Value: normalized.From})
		}
		if !normalized.To.IsZero() {
			rangeFilter = append(rangeFilter, bson.E{Key: "$lt", Value: normalized.To})
		}
		query = append(query, bson.E{Key: "started_at", Value: rangeFilter})
	}
	if normalized.DeviceType != "" {
		query = append(query, bson.E{Key: "device.type", Value: normalized.DeviceType})
	}
	if normalized.Channel != "" {
		query = append(query, bson.E{Key: "channel", Value: normalized.Channel})
	}
	if normalized.Status != "" {
		query = append(query, bson.E{Key: "status", Value: normalized.Status})
	}
	if normalized.Country != "" {
		query = append(query, bson.E{Key: "geo.country", Value: normalized.Country})
	}
	if normalized.EntryPage != "" {
		query = append(query, bson.E{Key: "entry_page", Value: normalized.EntryPage})
	}
	if normalized.Query != "" {
		pattern := primitiveRegex(normalized.Query)
		query = append(query, bson.E{Key: "$or", Value: bson.A{
			bson.D{{Key: "session_id", Value: pattern}},
			bson.D{{Key: "anonymous_id", Value: pattern}},
			bson.D{{Key: "user_id", Value: pattern}},
			bson.D{{Key: "entry_page", Value: pattern}},
			bson.D{{Key: "exit_page", Value: pattern}},
		}})
	}

	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("count sessions: %w", err)
	}
	cursor, err := r.collection.Find(
		ctx,
		query,
		options.Find().
			SetSort(bson.D{{Key: "started_at", Value: -1}, {Key: "_id", Value: 1}}).
			SetSkip(int64((normalized.Page-1)*normalized.PageSize)).
			SetLimit(int64(normalized.PageSize)),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list sessions: %w", err)
	}
	defer cursor.Close(ctx)

	sessions := make([]domain.Session, 0)
	for cursor.Next(ctx) {
		var doc mongoSessionDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, 0, fmt.Errorf("decode session: %w", err)
		}
		sessions = append(sessions, doc.Session.Normalize())
	}
	if err := cursor.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate sessions: %w", err)
	}
	return sessions, total, nil
}

func primitiveRegex(value string) bson.Regex {
	return bson.Regex{Pattern: regexp.QuoteMeta(strings.TrimSpace(value)), Options: "i"}
}

func (r *MongoSessionRepository) MarkEnded(ctx context.Context, sessionID string, endedAt time.Time, reason domain.SessionEndReason) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return fmt.Errorf("%w: session_id is required", domain.ErrInvalidSession)
	}
	reason = domain.SessionEndReason(strings.TrimSpace(string(reason)))
	if reason == "" {
		reason = domain.EndReasonUnknown
	}
	if !reason.Valid() {
		return fmt.Errorf("%w: unsupported end reason", domain.ErrInvalidSession)
	}
	timestamp := endedAt.UTC()
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}

	set := bson.D{
		{Key: "status", Value: domain.SessionStatusEnded},
		{Key: "ended_at", Value: timestamp},
		{Key: "end_reason", Value: reason},
		{Key: "last_seen_at", Value: timestamp},
		{Key: "updated_at", Value: timestamp},
		{Key: "retain_until", Value: timestamp.AddDate(0, 0, r.metadataRetentionDays)},
		{Key: "retention_class", Value: string(domain.RetentionClassSessionMetadata)},
	}
	unset := bson.D{{Key: "revoked_at", Value: ""}}

	switch reason {
	case domain.EndReasonInactivityTimeout:
		set = replaceSetValue(set, "status", domain.SessionStatusExpired)
	case domain.EndReasonSecurityRevoke, domain.EndReasonAdminRevoke:
		set = bson.D{
			{Key: "status", Value: domain.SessionStatusRevoked},
			{Key: "revoked_at", Value: timestamp},
			{Key: "end_reason", Value: reason},
			{Key: "last_seen_at", Value: timestamp},
			{Key: "updated_at", Value: timestamp},
			{Key: "retain_until", Value: timestamp.AddDate(0, 0, r.metadataRetentionDays)},
			{Key: "retention_class", Value: string(domain.RetentionClassSessionMetadata)},
		}
		unset = bson.D{{Key: "ended_at", Value: ""}}
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "session_id", Value: sessionID}},
		bson.D{{Key: "$set", Value: set}, {Key: "$unset", Value: unset}},
	)
	if err != nil {
		return fmt.Errorf("mark session ended: %w", err)
	}
	if result.MatchedCount == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func replaceSetValue(values bson.D, key string, value any) bson.D {
	for i := range values {
		if values[i].Key == key {
			values[i].Value = value
			return values
		}
	}
	return append(values, bson.E{Key: key, Value: value})
}
