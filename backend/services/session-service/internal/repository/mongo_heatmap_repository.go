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
	heatmapPointsCollectionName         = "heatmap_points"
	heatmapCheckpointsCollectionName    = "heatmap_checkpoints"
	heatmapBucketSessionsCollectionName = "heatmap_bucket_sessions"
	defaultHeatmapPointLimit            = 5000
	defaultHeatmapRetentionDays         = 730
)

type MongoHeatmapRepository struct {
	points         *mongo.Collection
	checkpoints    *mongo.Collection
	bucketSessions *mongo.Collection
	retentionDays  int
	logger         *slog.Logger
}

type heatmapBucketSessionDocument struct {
	ID          string    `bson:"_id"`
	BucketKey   string    `bson:"bucket_key"`
	PointID     string    `bson:"point_id"`
	SessionID   string    `bson:"session_id"`
	FirstSeenAt time.Time `bson:"first_seen_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

type heatmapCheckpointDocument struct {
	ID                                  string `bson:"_id"`
	domain.HeatmapAggregationCheckpoint `bson:",inline"`
}

type MongoHeatmapRepositoryOption func(*MongoHeatmapRepository)

func WithHeatmapRetentionDays(days int) MongoHeatmapRepositoryOption {
	return func(r *MongoHeatmapRepository) {
		if days > 0 {
			r.retentionDays = days
		}
	}
}

func NewMongoHeatmapRepository(database *mongo.Database, logger *slog.Logger, opts ...MongoHeatmapRepositoryOption) (*MongoHeatmapRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	repo := &MongoHeatmapRepository{
		points:         database.Collection(heatmapPointsCollectionName),
		checkpoints:    database.Collection(heatmapCheckpointsCollectionName),
		bucketSessions: database.Collection(heatmapBucketSessionsCollectionName),
		retentionDays:  defaultHeatmapRetentionDays,
		logger:         logger,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(repo)
		}
	}
	return repo, nil
}

func (r *MongoHeatmapRepository) EnsureIndexes(ctx context.Context) error {
	pointIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "path", Value: 1}, {Key: "device_type", Value: 1}, {Key: "day", Value: 1}},
			Options: options.Index().SetName("idx_heatmap_route_device_day"),
		},
		{
			Keys: bson.D{
				{Key: "heatmap_type", Value: 1},
				{Key: "path", Value: 1},
				{Key: "device_type", Value: 1},
				{Key: "viewport_bucket", Value: 1},
				{Key: "day", Value: 1},
				{Key: "x_bucket", Value: 1},
				{Key: "y_bucket", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("uniq_heatmap_click_bucket").
				SetPartialFilterExpression(bson.D{{Key: "heatmap_type", Value: string(domain.HeatmapTypeClick)}}),
		},
		{
			Keys: bson.D{
				{Key: "heatmap_type", Value: 1},
				{Key: "path", Value: 1},
				{Key: "device_type", Value: 1},
				{Key: "viewport_bucket", Value: 1},
				{Key: "day", Value: 1},
				{Key: "depth_bucket", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("uniq_heatmap_scroll_bucket").
				SetPartialFilterExpression(bson.D{{Key: "heatmap_type", Value: string(domain.HeatmapTypeScroll)}}),
		},
		{
			Keys:    bson.D{{Key: "retain_until", Value: 1}},
			Options: options.Index().SetName("idx_heatmap_retention_due"),
		},
	}
	if _, err := r.points.Indexes().CreateMany(ctx, pointIndexes); err != nil {
		return fmt.Errorf("create heatmap point indexes: %w", err)
	}

	sessionIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "bucket_key", Value: 1}, {Key: "session_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_heatmap_bucket_session"),
		},
		{
			Keys:    bson.D{{Key: "point_id", Value: 1}},
			Options: options.Index().SetName("idx_heatmap_bucket_sessions_point"),
		},
	}
	if _, err := r.bucketSessions.Indexes().CreateMany(ctx, sessionIndexes); err != nil {
		return fmt.Errorf("create heatmap bucket-session indexes: %w", err)
	}

	checkpointIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "worker_name", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_heatmap_checkpoint_worker"),
		},
	}
	if _, err := r.checkpoints.Indexes().CreateMany(ctx, checkpointIndexes); err != nil {
		return fmt.Errorf("create heatmap checkpoint indexes: %w", err)
	}

	r.logger.InfoContext(ctx, "session.mongo.heatmap_indexes_ready")
	return nil
}

func (r *MongoHeatmapRepository) UpsertHeatmapPoint(ctx context.Context, point domain.HeatmapPoint, sessionID string) error {
	normalized := point.Normalize()
	if err := normalized.Validate(); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInvalidHeatmap, err)
	}
	normalized = r.applyRetentionDefaults(normalized)
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return fmt.Errorf("%w: session_id is required", domain.ErrInvalidHeatmap)
	}

	uniqueSessionDelta, rollbackSessionMarker, err := r.insertBucketSession(ctx, normalized, sessionID)
	if err != nil {
		return err
	}
	filter := heatmapPointFilter(normalized)
	update := heatmapPointUpdate(normalized, uniqueSessionDelta)
	if _, err := r.points.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true)); err != nil {
		rollbackSessionMarker()
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: duplicate heatmap bucket", domain.ErrInvalidHeatmap)
		}
		return fmt.Errorf("upsert heatmap point: %w", err)
	}
	return nil
}

func (r *MongoHeatmapRepository) ListHeatmapPoints(ctx context.Context, filter domain.HeatmapFilter, limit int) ([]domain.HeatmapPoint, error) {
	normalized := filter.Normalize()
	if err := normalized.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidHeatmap, err)
	}
	if limit <= 0 || limit > defaultHeatmapPointLimit {
		limit = defaultHeatmapPointLimit
	}

	query := bson.D{
		{Key: "heatmap_type", Value: string(normalized.HeatmapType)},
		{Key: "path", Value: normalized.Path},
		{Key: "device_type", Value: string(normalized.DeviceType)},
		{Key: "day", Value: bson.D{{Key: "$gte", Value: normalized.FromDay}, {Key: "$lte", Value: normalized.ToDay}}},
	}
	if normalized.ViewportBucket != "" {
		query = append(query, bson.E{Key: "viewport_bucket", Value: normalized.ViewportBucket})
	}

	cursor, err := r.points.Find(ctx, query, options.Find().
		SetSort(bson.D{{Key: "y", Value: 1}, {Key: "x", Value: 1}, {Key: "day", Value: 1}}).
		SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, fmt.Errorf("list heatmap points: %w", err)
	}
	defer cursor.Close(ctx)

	points := make([]domain.HeatmapPoint, 0)
	for cursor.Next(ctx) {
		var point domain.HeatmapPoint
		if err := cursor.Decode(&point); err != nil {
			return nil, fmt.Errorf("decode heatmap point: %w", err)
		}
		points = append(points, point.Normalize())
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate heatmap points: %w", err)
	}
	return points, nil
}

func (r *MongoHeatmapRepository) FindHeatmapCheckpoint(ctx context.Context, workerName string) (domain.HeatmapAggregationCheckpoint, error) {
	workerName = strings.TrimSpace(workerName)
	if workerName == "" {
		return domain.HeatmapAggregationCheckpoint{}, fmt.Errorf("%w: worker_name is required", domain.ErrInvalidHeatmap)
	}

	var doc heatmapCheckpointDocument
	err := r.checkpoints.FindOne(ctx, bson.D{{Key: "worker_name", Value: workerName}}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.HeatmapAggregationCheckpoint{}, domain.ErrHeatmapNotFound
		}
		return domain.HeatmapAggregationCheckpoint{}, fmt.Errorf("find heatmap checkpoint: %w", err)
	}
	return doc.HeatmapAggregationCheckpoint, nil
}

func (r *MongoHeatmapRepository) SaveHeatmapCheckpoint(ctx context.Context, checkpoint domain.HeatmapAggregationCheckpoint) error {
	normalized := checkpoint
	normalized.WorkerName = strings.TrimSpace(normalized.WorkerName)
	normalized.LastProcessedAt = normalized.LastProcessedAt.UTC()
	normalized.UpdatedAt = normalized.UpdatedAt.UTC()
	if normalized.WorkerName == "" {
		return fmt.Errorf("%w: worker_name is required", domain.ErrInvalidHeatmap)
	}
	if normalized.LastProcessedAt.IsZero() {
		return fmt.Errorf("%w: last_processed_at is required", domain.ErrInvalidHeatmap)
	}
	if normalized.UpdatedAt.IsZero() {
		normalized.UpdatedAt = time.Now().UTC()
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "worker_name", Value: normalized.WorkerName},
			{Key: "last_processed_at", Value: normalized.LastProcessedAt},
			{Key: "updated_at", Value: normalized.UpdatedAt},
		}},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: "checkpoint_" + normalized.WorkerName},
		}},
	}
	if _, err := r.checkpoints.UpdateOne(ctx, bson.D{{Key: "worker_name", Value: normalized.WorkerName}}, update, options.UpdateOne().SetUpsert(true)); err != nil {
		return fmt.Errorf("save heatmap checkpoint: %w", err)
	}
	return nil
}

func (r *MongoHeatmapRepository) applyRetentionDefaults(point domain.HeatmapPoint) domain.HeatmapPoint {
	out := point.Normalize()
	if out.RetentionClass == "" {
		out.RetentionClass = domain.RetentionClassAggregateFine
	}
	if out.RetainUntil == nil {
		base := out.LastSeenAt
		if base.IsZero() {
			base = out.UpdatedAt
		}
		if base.IsZero() {
			if day, err := time.Parse("2006-01-02", out.Day); err == nil {
				base = day.UTC()
			}
		}
		if base.IsZero() {
			base = time.Now().UTC()
		}
		retainUntil := base.UTC().AddDate(0, 0, r.retentionDays)
		out.RetainUntil = &retainUntil
	}
	return out
}

func (r *MongoHeatmapRepository) insertBucketSession(ctx context.Context, point domain.HeatmapPoint, sessionID string) (int, func(), error) {
	now := time.Now().UTC()
	markerID := point.SessionMarkerID(sessionID)
	doc := heatmapBucketSessionDocument{
		ID:          markerID,
		BucketKey:   point.BucketKey(),
		PointID:     point.ID,
		SessionID:   sessionID,
		FirstSeenAt: point.FirstSeenAt,
		UpdatedAt:   now,
	}
	_, err := r.bucketSessions.InsertOne(ctx, doc)
	if err == nil {
		return 1, func() {
			_, _ = r.bucketSessions.DeleteOne(context.Background(), bson.D{{Key: "_id", Value: markerID}})
		}, nil
	}
	if mongo.IsDuplicateKeyError(err) {
		return 0, func() {}, nil
	}
	r.logger.WarnContext(ctx, "session.heatmap.unique_session_marker_failed",
		slog.String("point_id", point.ID),
		slog.String("session_id", sessionID),
		slog.String("error", err.Error()),
	)
	return 0, func() {}, fmt.Errorf("insert heatmap bucket session marker: %w", err)
}

func heatmapPointFilter(point domain.HeatmapPoint) bson.D {
	filter := bson.D{
		{Key: "heatmap_type", Value: string(point.HeatmapType)},
		{Key: "path", Value: point.Path},
		{Key: "device_type", Value: string(point.DeviceType)},
		{Key: "viewport_bucket", Value: point.ViewportBucket},
		{Key: "day", Value: point.Day},
	}
	if point.HeatmapType == domain.HeatmapTypeClick {
		filter = append(filter,
			bson.E{Key: "x_bucket", Value: *point.XBucket},
			bson.E{Key: "y_bucket", Value: *point.YBucket},
		)
	} else {
		filter = append(filter, bson.E{Key: "depth_bucket", Value: *point.DepthBucket})
	}
	return filter
}

func heatmapPointUpdate(point domain.HeatmapPoint, uniqueSessionDelta int) bson.D {
	setOnInsert := bson.D{
		{Key: "_id", Value: point.ID},
		{Key: "heatmap_type", Value: string(point.HeatmapType)},
		{Key: "path", Value: point.Path},
		{Key: "normalized_path", Value: point.NormalizedPath},
		{Key: "device_type", Value: string(point.DeviceType)},
		{Key: "viewport_bucket", Value: point.ViewportBucket},
		{Key: "day", Value: point.Day},
		{Key: "x", Value: point.X},
		{Key: "y", Value: point.Y},
		{Key: "schema_version", Value: point.SchemaVersion},
		{Key: "created_at", Value: point.CreatedAt},
		{Key: "retain_until", Value: point.RetainUntil},
		{Key: "retention_class", Value: string(point.RetentionClass)},
	}
	if point.XBucket != nil {
		setOnInsert = append(setOnInsert, bson.E{Key: "x_bucket", Value: *point.XBucket})
	}
	if point.YBucket != nil {
		setOnInsert = append(setOnInsert, bson.E{Key: "y_bucket", Value: *point.YBucket})
	}
	if point.DepthBucket != nil {
		setOnInsert = append(setOnInsert, bson.E{Key: "depth_bucket", Value: *point.DepthBucket})
	}
	return bson.D{
		{Key: "$setOnInsert", Value: setOnInsert},
		{Key: "$set", Value: bson.D{
			{Key: "normalized_path", Value: point.NormalizedPath},
			{Key: "schema_version", Value: point.SchemaVersion},
			{Key: "updated_at", Value: point.UpdatedAt},
			{Key: "retain_until", Value: point.RetainUntil},
			{Key: "retention_class", Value: string(point.RetentionClass)},
		}},
		{Key: "$inc", Value: bson.D{
			{Key: "weight", Value: 1},
			{Key: "sample_events", Value: 1},
			{Key: "unique_sessions", Value: uniqueSessionDelta},
		}},
		{Key: "$min", Value: bson.D{{Key: "first_seen_at", Value: point.FirstSeenAt}}},
		{Key: "$max", Value: bson.D{{Key: "last_seen_at", Value: point.LastSeenAt}}},
	}
}
