package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoReportsRepository struct {
	schedules  *mongo.Collection
	aggregates *mongo.Collection
	sessions   *mongo.Collection
	journeys   *mongo.Collection
	heatmaps   *mongo.Collection
	audit      *mongo.Collection
}

func NewMongoReportsRepository(db *mongo.Database) (*MongoReportsRepository, error) {
	if db == nil {
		return nil, errors.New("mongo database is required")
	}
	return &MongoReportsRepository{
		schedules:  db.Collection("report_schedules"),
		aggregates: db.Collection(analyticsAggregatesCollectionName),
		sessions:   db.Collection(sessionsCollectionName),
		journeys:   db.Collection(journeySummariesCollectionName),
		heatmaps:   db.Collection(heatmapPointsCollectionName),
		audit:      db.Collection("admin_audit_events"),
	}, nil
}

func (r *MongoReportsRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.schedules.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "next_run_at", Value: 1}}, Options: options.Index().SetName("report_schedules_due")},
		{Keys: bson.D{{Key: "created_by", Value: 1}, {Key: "created_at", Value: -1}}, Options: options.Index().SetName("report_schedules_actor_created")},
		{Keys: bson.D{{Key: "updated_at", Value: -1}}, Options: options.Index().SetName("report_schedules_updated")},
	})
	if err != nil {
		return fmt.Errorf("create report schedule indexes: %w", err)
	}
	return nil
}

func (r *MongoReportsRepository) BuildReport(ctx context.Context, query domain.ReportQuery, maxRows int) (domain.TabularReport, error) {
	if maxRows <= 0 {
		maxRows = 10001
	}
	switch query.ReportType {
	case domain.ReportOverview:
		return r.aggregateReport(ctx, query, "", maxRows)
	case domain.ReportActiveSessions:
		return r.activeSessionsReport(ctx, query, maxRows)
	case domain.ReportJourneySummary:
		return r.journeyReport(ctx, query, maxRows)
	case domain.ReportFunnel:
		return r.aggregateReport(ctx, query, string(domain.AnalyticsMetricFunnelCheckout), maxRows)
	case domain.ReportHeatmap:
		return r.heatmapReport(ctx, query, maxRows)
	case domain.ReportRetention:
		return r.retentionReport(ctx, query, maxRows)
	default:
		return domain.TabularReport{}, domain.NewFieldError("reportType", "unsupported report type")
	}
}

func (r *MongoReportsRepository) aggregateReport(ctx context.Context, query domain.ReportQuery, metric string, maxRows int) (domain.TabularReport, error) {
	filter := bson.D{{Key: "bucket_start", Value: bson.D{{Key: "$gte", Value: query.From}, {Key: "$lt", Value: query.To}}}}
	if metric != "" {
		filter = append(filter, bson.E{Key: "metric", Value: metric})
	}
	filter = appendAggregateReportFilters(filter, query.Filters)
	cursor, err := r.aggregates.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "bucket_start", Value: 1}}).SetLimit(int64(maxRows)))
	if err != nil {
		return domain.TabularReport{}, err
	}
	defer cursor.Close(ctx)

	rows := make([][]string, 0)
	for cursor.Next(ctx) {
		var aggregate domain.AnalyticsAggregate
		if err := cursor.Decode(&aggregate); err != nil {
			return domain.TabularReport{}, err
		}
		for _, name := range sortedFloatKeys(aggregate.Values) {
			rows = append(rows, []string{string(aggregate.Metric), formatReportTime(aggregate.BucketStart), formatReportTime(aggregate.BucketEnd), name, strconv.FormatFloat(aggregate.Values[name], 'f', -1, 64)})
		}
		for _, step := range aggregate.Steps {
			rows = append(rows, []string{string(aggregate.Metric), formatReportTime(aggregate.BucketStart), formatReportTime(aggregate.BucketEnd), string(step.EventType), strconv.FormatInt(step.UniqueSessions, 10)})
		}
		if len(rows) >= maxRows {
			break
		}
	}
	if err := cursor.Err(); err != nil {
		return domain.TabularReport{}, err
	}
	return domain.TabularReport{Headers: []string{"metric", "from", "to", "measure", "value"}, Rows: rows}, nil
}

func (r *MongoReportsRepository) activeSessionsReport(ctx context.Context, query domain.ReportQuery, maxRows int) (domain.TabularReport, error) {
	filter := bson.D{{Key: "started_at", Value: bson.D{{Key: "$gte", Value: query.From}, {Key: "$lt", Value: query.To}}}}
	filter = appendSessionReportFilters(filter, query.Filters)
	cursor, err := r.sessions.Find(ctx, filter, options.Find().
		SetProjection(bson.D{{Key: "session_id", Value: 0}, {Key: "anonymous_id", Value: 0}, {Key: "user_id", Value: 0}, {Key: "ip_hash", Value: 0}, {Key: "user_agent", Value: 0}, {Key: "device_fingerprint_hash", Value: 0}}).
		SetSort(bson.D{{Key: "started_at", Value: -1}}).SetLimit(int64(maxRows)))
	if err != nil {
		return domain.TabularReport{}, err
	}
	defer cursor.Close(ctx)
	rows := make([][]string, 0)
	for cursor.Next(ctx) {
		var session domain.Session
		if err := cursor.Decode(&session); err != nil {
			return domain.TabularReport{}, err
		}
		duration := session.LastSeenAt.Sub(session.StartedAt).Seconds()
		country := ""
		if session.Geo.Country != nil {
			country = *session.Geo.Country
		}
		rows = append(rows, []string{formatReportTime(session.StartedAt), string(session.Status), string(session.Channel), session.EntryPage, string(session.Device.Type), country, strconv.FormatInt(int64(duration), 10)})
	}
	return domain.TabularReport{Headers: []string{"started_at", "status", "channel", "entry_page", "device_type", "country", "duration_seconds"}, Rows: rows}, cursor.Err()
}

func (r *MongoReportsRepository) journeyReport(ctx context.Context, query domain.ReportQuery, maxRows int) (domain.TabularReport, error) {
	filter := bson.D{{Key: "first_event_at", Value: bson.D{{Key: "$gte", Value: query.From}, {Key: "$lt", Value: query.To}}}}
	cursor, err := r.journeys.Find(ctx, filter, options.Find().
		SetProjection(bson.D{{Key: "session_id", Value: 0}, {Key: "anonymous_id", Value: 0}, {Key: "user_id", Value: 0}, {Key: "milestones", Value: 0}}).
		SetSort(bson.D{{Key: "first_event_at", Value: -1}}).SetLimit(int64(maxRows)))
	if err != nil {
		return domain.TabularReport{}, err
	}
	defer cursor.Close(ctx)
	rows := make([][]string, 0)
	for cursor.Next(ctx) {
		var summary domain.JourneySummary
		if err := cursor.Decode(&summary); err != nil {
			return domain.TabularReport{}, err
		}
		rows = append(rows, []string{formatReportTime(summary.FirstEventAt), summary.EntryPage, summary.ExitPage, strconv.FormatInt(summary.DurationSeconds, 10), strconv.Itoa(summary.TotalEvents), strconv.Itoa(summary.ProductsViewed), strconv.Itoa(summary.CartActions), strconv.FormatBool(summary.CheckoutStarted), strconv.FormatBool(summary.PaymentCompleted)})
	}
	return domain.TabularReport{Headers: []string{"started_at", "entry_page", "exit_page", "duration_seconds", "total_events", "products_viewed", "cart_actions", "checkout_started", "payment_completed"}, Rows: rows}, cursor.Err()
}

func (r *MongoReportsRepository) heatmapReport(ctx context.Context, query domain.ReportQuery, maxRows int) (domain.TabularReport, error) {
	filter := bson.D{{Key: "last_seen_at", Value: bson.D{{Key: "$gte", Value: query.From}, {Key: "$lt", Value: query.To}}}}
	if value := cleanAll(query.Filters.DeviceType); value != "" {
		filter = append(filter, bson.E{Key: "device_type", Value: value})
	}
	cursor, err := r.heatmaps.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "day", Value: 1}}).SetLimit(int64(maxRows)))
	if err != nil {
		return domain.TabularReport{}, err
	}
	defer cursor.Close(ctx)
	rows := make([][]string, 0)
	for cursor.Next(ctx) {
		var point domain.HeatmapPoint
		if err := cursor.Decode(&point); err != nil {
			return domain.TabularReport{}, err
		}
		rows = append(rows, []string{point.Day, string(point.HeatmapType), point.Path, string(point.DeviceType), strconv.Itoa(point.X), strconv.Itoa(point.Y), optionalInt(point.DepthBucket), strconv.Itoa(point.Weight), strconv.Itoa(point.UniqueSessions)})
	}
	return domain.TabularReport{Headers: []string{"day", "type", "path", "device_type", "x", "y", "depth", "weight", "unique_sessions"}, Rows: rows}, cursor.Err()
}

func (r *MongoReportsRepository) retentionReport(ctx context.Context, query domain.ReportQuery, maxRows int) (domain.TabularReport, error) {
	filter := bson.D{{Key: "metric", Value: bson.D{{Key: "$in", Value: bson.A{domain.AnalyticsMetricRetentionUsers, domain.AnalyticsMetricRetentionGuests}}}}, {Key: "cohort_start", Value: bson.D{{Key: "$gte", Value: query.From}, {Key: "$lte", Value: query.To}}}}
	filter = appendAggregateReportFilters(filter, query.Filters)
	cursor, err := r.aggregates.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "cohort_start", Value: 1}}).SetLimit(int64(maxRows)))
	if err != nil {
		return domain.TabularReport{}, err
	}
	defer cursor.Close(ctx)
	rows := make([][]string, 0)
	for cursor.Next(ctx) {
		var aggregate domain.RetentionAggregate
		if err := cursor.Decode(&aggregate); err != nil {
			return domain.TabularReport{}, err
		}
		keys := make([]string, 0, len(aggregate.Retention))
		for key := range aggregate.Retention {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			rows = append(rows, []string{formatReportTime(aggregate.CohortStart), formatReportTime(aggregate.CohortEnd), strconv.FormatInt(aggregate.CohortSize, 10), key, strconv.FormatInt(aggregate.Retention[key], 10), strconv.FormatFloat(aggregate.Rates[key], 'f', 2, 64)})
		}
		if len(rows) >= maxRows {
			break
		}
	}
	return domain.TabularReport{Headers: []string{"cohort_start", "cohort_end", "cohort_size", "offset", "retained_users", "retention_rate"}, Rows: rows}, cursor.Err()
}

func (r *MongoReportsRepository) ListSchedules(ctx context.Context, limit int) ([]domain.ReportSchedule, error) {
	cursor, err := r.schedules.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var items []domain.ReportSchedule
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MongoReportsRepository) CreateSchedule(ctx context.Context, schedule domain.ReportSchedule) error {
	_, err := r.schedules.InsertOne(ctx, schedule)
	return err
}

func (r *MongoReportsRepository) GetSchedule(ctx context.Context, id string) (domain.ReportSchedule, error) {
	var out domain.ReportSchedule
	err := r.schedules.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&out)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.ReportSchedule{}, domain.ErrNotFound
	}
	return out, err
}

func (r *MongoReportsRepository) UpdateScheduleStatus(ctx context.Context, id string, status domain.ReportScheduleStatus, nextRunAt *time.Time, updatedAt time.Time) (domain.ReportSchedule, error) {
	set := bson.D{{Key: "status", Value: status}, {Key: "updated_at", Value: updatedAt}}
	update := bson.D{{Key: "$set", Value: set}, {Key: "$unset", Value: bson.D{{Key: "last_error", Value: ""}}}}
	if nextRunAt != nil {
		set = append(set, bson.E{Key: "next_run_at", Value: *nextRunAt})
		update[0].Value = set
	} else {
		update[1].Value = bson.D{{Key: "last_error", Value: ""}, {Key: "next_run_at", Value: ""}}
	}
	var out domain.ReportSchedule
	err := r.schedules.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: id}}, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&out)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.ReportSchedule{}, domain.ErrNotFound
	}
	return out, err
}

func (r *MongoReportsRepository) DeleteSchedule(ctx context.Context, id string) error {
	result, err := r.schedules.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *MongoReportsRepository) CreateAuditEvent(ctx context.Context, event domain.AuditEvent) error {
	_, err := r.audit.InsertOne(ctx, auditEventDocumentFromDomain(event))
	return err
}

func appendAggregateReportFilters(filter bson.D, values domain.ReportFilters) bson.D {
	if value := cleanAll(values.Channel); value != "" {
		filter = append(filter, bson.E{Key: "segment.channel", Value: value})
	}
	if value := cleanAll(values.DeviceType); value != "" {
		filter = append(filter, bson.E{Key: "segment.device_type", Value: value})
	}
	if value := cleanAll(values.Country); value != "" {
		filter = append(filter, bson.E{Key: "segment.country", Value: strings.ToUpper(value)})
	}
	if value := cleanAll(values.Source); value != "" {
		filter = append(filter, bson.E{Key: "segment.source", Value: value})
	}
	return filter
}

func appendSessionReportFilters(filter bson.D, values domain.ReportFilters) bson.D {
	if value := cleanAll(values.Channel); value != "" {
		filter = append(filter, bson.E{Key: "channel", Value: value})
	}
	if value := cleanAll(values.DeviceType); value != "" {
		filter = append(filter, bson.E{Key: "device.type", Value: value})
	}
	if value := cleanAll(values.Country); value != "" {
		filter = append(filter, bson.E{Key: "geo.country", Value: strings.ToUpper(value)})
	}
	if value := cleanAll(values.Source); value != "" {
		filter = append(filter, bson.E{Key: "utm.source", Value: value})
	}
	switch cleanAll(values.UserType) {
	case "anonymous":
		filter = append(filter, bson.E{Key: "user_id", Value: bson.D{{Key: "$exists", Value: false}}})
	case "logged_in":
		filter = append(filter, bson.E{Key: "user_id", Value: bson.D{{Key: "$type", Value: "string"}}})
	}
	return filter
}

func cleanAll(value string) string {
	value = strings.TrimSpace(value)
	if strings.EqualFold(value, "all") {
		return ""
	}
	return value
}
func formatReportTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
func optionalInt(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}
func sortedFloatKeys(values map[string]float64) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
