package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *MongoAnalyticsRepository) GetRetentionAggregates(ctx context.Context, filter domain.RetentionReportFilter) ([]domain.RetentionAggregate, error) {
	metric := domain.AnalyticsMetricRetentionUsers
	if strings.EqualFold(strings.TrimSpace(filter.UserType), "anonymous") {
		metric = domain.AnalyticsMetricRetentionGuests
	}
	query := bson.D{
		{Key: "metric", Value: metric},
		{Key: "cohort_start", Value: bson.D{{Key: "$gte", Value: filter.From.UTC()}, {Key: "$lte", Value: filter.To.UTC()}}},
	}
	if filter.DeviceType != "" && filter.DeviceType != "all" {
		query = append(query, bson.E{Key: "segment.device_type", Value: string(filter.DeviceType)})
	}
	if filter.Channel != "" && filter.Channel != "all" {
		query = append(query, bson.E{Key: "segment.channel", Value: string(filter.Channel)})
	}
	if value := strings.TrimSpace(filter.Source); value != "" && !strings.EqualFold(value, "all") {
		query = append(query, bson.E{Key: "segment.source", Value: value})
	}

	cursor, err := r.aggregates.Find(ctx, query, options.Find().
		SetSort(bson.D{{Key: "cohort_start", Value: 1}}).
		SetLimit(512))
	if err != nil {
		return nil, fmt.Errorf("find retention aggregates: %w", err)
	}
	defer cursor.Close(ctx)

	var aggregates []domain.RetentionAggregate
	if err := cursor.All(ctx, &aggregates); err != nil {
		return nil, fmt.Errorf("decode retention aggregates: %w", err)
	}
	return aggregates, nil
}

const analyticsAggregatesCollectionName = "analytics_aggregates"

type MongoAnalyticsRepository struct {
	aggregates *mongo.Collection
	events     *mongo.Collection
	logger     *slog.Logger
}

func NewMongoAnalyticsRepository(database *mongo.Database, logger *slog.Logger) (*MongoAnalyticsRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &MongoAnalyticsRepository{
		aggregates: database.Collection(analyticsAggregatesCollectionName),
		events:     database.Collection(sessionEventsCollectionName),
		logger:     logger,
	}, nil
}

func (r *MongoAnalyticsRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "metric", Value: 1},
				{Key: "bucket", Value: 1},
				{Key: "bucket_start", Value: 1},
				{Key: "bucket_end", Value: 1},
				{Key: "segment.channel", Value: 1},
				{Key: "segment.device_type", Value: 1},
				{Key: "segment.country", Value: 1},
				{Key: "segment.campaign", Value: 1},
			},
			Options: options.Index().
				SetName("uniq_metric_bucket_time_segment").
				SetUnique(true).
				SetPartialFilterExpression(bson.D{
					{Key: "bucket_start", Value: bson.D{{Key: "$type", Value: "date"}}},
					{Key: "bucket_end", Value: bson.D{{Key: "$type", Value: "date"}}},
				}),
		},
		{
			Keys:    bson.D{{Key: "metric", Value: 1}, {Key: "bucket_start", Value: -1}},
			Options: options.Index().SetName("idx_metric_time"),
		},
		{
			Keys:    bson.D{{Key: "segment.channel", Value: 1}, {Key: "segment.device_type", Value: 1}, {Key: "bucket_start", Value: -1}},
			Options: options.Index().SetName("idx_segment_time"),
		},
		{
			Keys:    bson.D{{Key: "metric", Value: 1}, {Key: "cohort_start", Value: -1}},
			Options: options.Index().SetName("idx_retention_cohort"),
		},
		{
			Keys:    bson.D{{Key: "contains_pii", Value: 1}, {Key: "updated_at", Value: -1}},
			Options: options.Index().SetName("idx_aggregates_pii_policy"),
		},
		{
			Keys:    bson.D{{Key: "bucket_end", Value: 1}, {Key: "calculated_at", Value: 1}},
			Options: options.Index().SetName("idx_aggregates_retention_due"),
		},
	}
	if _, err := r.aggregates.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("create analytics aggregate indexes: %w", err)
	}
	r.logger.InfoContext(ctx, "session.mongo.analytics_indexes_ready")
	return nil
}

func (r *MongoAnalyticsRepository) GetFunnelAggregate(ctx context.Context, filter domain.FunnelReportFilter) ([]domain.FunnelStep, error) {
	normalized := filter.Normalize()
	query := bson.D{
		{Key: "metric", Value: normalized.Metric},
		{Key: "bucket_start", Value: bson.D{{Key: "$gte", Value: normalized.From}, {Key: "$lt", Value: normalized.To}}},
	}
	query = appendSegmentFilters(query, normalized)

	cursor, err := r.aggregates.Find(
		ctx,
		query,
		options.Find().SetSort(bson.D{{Key: "bucket_start", Value: 1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("find funnel aggregate: %w", err)
	}
	defer cursor.Close(ctx)

	merged := orderedFunnelSteps(normalized.Steps)
	found := false
	for cursor.Next(ctx) {
		var aggregate domain.AnalyticsAggregate
		if err := cursor.Decode(&aggregate); err != nil {
			return nil, fmt.Errorf("decode funnel aggregate: %w", err)
		}
		for _, step := range aggregate.Normalize().Steps {
			key := funnelStepKey(step)
			if current, ok := merged[key]; ok {
				current.Count += step.Count
				current.UniqueSessions += step.UniqueSessions
				current.UniqueUsers += step.UniqueUsers
				merged[key] = current
				found = true
			}
		}
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate funnel aggregates: %w", err)
	}
	if !found {
		return nil, nil
	}
	return orderedMergedSteps(normalized.Steps, merged), nil
}

func (r *MongoAnalyticsRepository) BuildFunnelFromRawEvents(ctx context.Context, filter domain.FunnelReportFilter) ([]domain.FunnelStep, error) {
	normalized := filter.Normalize()
	eventTypes := make(bson.A, 0, len(normalized.Steps))
	for _, step := range normalized.Steps {
		eventTypes = append(eventTypes, string(step))
	}
	match := bson.D{
		{Key: "event_type", Value: bson.D{{Key: "$in", Value: eventTypes}}},
		{Key: "occurred_at", Value: bson.D{{Key: "$gte", Value: normalized.From}, {Key: "$lt", Value: normalized.To}}},
	}
	if normalized.DeviceType != "" {
		match = append(match, bson.E{Key: "device.type", Value: normalized.DeviceType})
	}
	if normalized.Channel != "" {
		match = append(match, bson.E{Key: "client.channel", Value: normalized.Channel})
	}
	if normalized.Country != "" {
		match = append(match, bson.E{Key: "geo.country", Value: normalized.Country})
	}
	andConditions := bson.A{}
	if normalized.Source != "" && !strings.EqualFold(normalized.Source, "all") {
		andConditions = append(andConditions, bson.D{{Key: "$or", Value: bson.A{
			bson.D{{Key: "properties.source", Value: normalized.Source}},
			bson.D{{Key: "properties.utm_source", Value: normalized.Source}},
			bson.D{{Key: "properties.utm.source", Value: normalized.Source}},
		}}})
	}
	switch normalized.UserType {
	case "logged_in":
		andConditions = append(andConditions, bson.D{{Key: "user_id", Value: bson.D{{Key: "$type", Value: "string"}, {Key: "$ne", Value: ""}}}})
	case "anonymous":
		andConditions = append(andConditions, bson.D{{Key: "$or", Value: bson.A{
			bson.D{{Key: "user_id", Value: bson.D{{Key: "$exists", Value: false}}}},
			bson.D{{Key: "user_id", Value: nil}},
			bson.D{{Key: "user_id", Value: ""}},
		}}})
	}
	if len(andConditions) > 0 {
		match = append(match, bson.E{Key: "$and", Value: andConditions})
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: match}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$event_type"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "unique_sessions", Value: bson.D{{Key: "$addToSet", Value: "$session_id"}}},
			{Key: "unique_users", Value: bson.D{{Key: "$addToSet", Value: "$user_id"}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "event_type", Value: "$_id"},
			{Key: "count", Value: 1},
			{Key: "unique_sessions", Value: bson.D{{Key: "$size", Value: "$unique_sessions"}}},
			{Key: "unique_users", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$setDifference", Value: bson.A{"$unique_users", bson.A{nil, ""}}}}}}},
		}}},
	}

	cursor, err := r.events.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate raw funnel events: %w", err)
	}
	defer cursor.Close(ctx)

	merged := orderedFunnelSteps(normalized.Steps)
	for cursor.Next(ctx) {
		var row rawFunnelStepDocument
		if err := cursor.Decode(&row); err != nil {
			return nil, fmt.Errorf("decode raw funnel step: %w", err)
		}
		key := domain.EventType(row.EventType)
		if step, ok := merged[key]; ok {
			step.Count = row.Count
			step.UniqueSessions = row.UniqueSessions
			step.UniqueUsers = row.UniqueUsers
			merged[key] = step
		}
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate raw funnel steps: %w", err)
	}
	return orderedMergedSteps(normalized.Steps, merged), nil
}

type rawFunnelStepDocument struct {
	EventType      string `bson:"event_type"`
	Count          int64  `bson:"count"`
	UniqueSessions int64  `bson:"unique_sessions"`
	UniqueUsers    int64  `bson:"unique_users"`
}

func appendSegmentFilters(query bson.D, filter domain.FunnelReportFilter) bson.D {
	if filter.DeviceType != "" {
		query = append(query, bson.E{Key: "segment.device_type", Value: string(filter.DeviceType)})
	}
	if filter.Channel != "" {
		query = append(query, bson.E{Key: "segment.channel", Value: string(filter.Channel)})
	}
	if filter.Source != "" && !strings.EqualFold(filter.Source, "all") {
		query = append(query, bson.E{Key: "segment.source", Value: filter.Source})
	}
	if filter.UserType != "" && !strings.EqualFold(filter.UserType, "all") {
		query = append(query, bson.E{Key: "segment.user_type", Value: filter.UserType})
	}
	if filter.Country != "" {
		query = append(query, bson.E{Key: "segment.country", Value: filter.Country})
	}
	if filter.Campaign != "" {
		query = append(query, bson.E{Key: "segment.campaign", Value: filter.Campaign})
	}
	return query
}

func orderedFunnelSteps(steps []domain.EventType) map[domain.EventType]domain.FunnelStep {
	out := make(map[domain.EventType]domain.FunnelStep, len(steps))
	for _, step := range steps {
		out[step] = domain.FunnelStep{Name: string(step), EventType: step}
	}
	return out
}

func orderedMergedSteps(order []domain.EventType, merged map[domain.EventType]domain.FunnelStep) []domain.FunnelStep {
	out := make([]domain.FunnelStep, 0, len(order))
	for _, step := range order {
		out = append(out, merged[step].Normalize())
	}
	return out
}

func funnelStepKey(step domain.FunnelStep) domain.EventType {
	normalized := step.Normalize()
	if normalized.EventType != "" {
		return normalized.EventType
	}
	return domain.EventType(normalized.Name)
}
