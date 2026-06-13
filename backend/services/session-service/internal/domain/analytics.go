package domain

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

const CurrentAnalyticsAggregateSchemaVersion = 1

type AnalyticsMetric string

const (
	AnalyticsMetricFunnelCheckout  AnalyticsMetric = "funnel.checkout"
	AnalyticsMetricConversion      AnalyticsMetric = "conversion.checkout"
	AnalyticsMetricRetentionUsers  AnalyticsMetric = "retention.users"
	AnalyticsMetricRetentionGuests AnalyticsMetric = "retention.guests"
)

type AnalyticsBucket string

const (
	AnalyticsBucketHourly AnalyticsBucket = "hourly"
	AnalyticsBucketDaily  AnalyticsBucket = "daily"
	AnalyticsBucketCustom AnalyticsBucket = "custom"
)

type LiveMetrics struct {
	ActiveUsers     int64     `json:"active_users" bson:"active_users"`
	ActiveSessions  int64     `json:"active_sessions" bson:"active_sessions"`
	EventsPerMinute float64   `json:"events_per_minute" bson:"events_per_minute"`
	WindowSeconds   int64     `json:"window_seconds,omitempty" bson:"window_seconds,omitempty"`
	MeasuredAt      time.Time `json:"measured_at,omitempty" bson:"measured_at,omitempty"`
}

type SessionListFilter struct {
	UserID      string
	AnonymousID string
	From        time.Time
	To          time.Time
	Page        int
	PageSize    int
	DeviceType  DeviceType
	Channel     Channel
	Status      SessionStatus
}

func (f SessionListFilter) Normalize() SessionListFilter {
	out := f
	out.UserID = strings.TrimSpace(out.UserID)
	out.AnonymousID = strings.TrimSpace(out.AnonymousID)
	out.From = normalizeTime(out.From)
	out.To = normalizeTime(out.To)
	out.DeviceType = DeviceType(strings.TrimSpace(string(out.DeviceType)))
	out.Channel = Channel(strings.TrimSpace(string(out.Channel)))
	out.Status = SessionStatus(strings.TrimSpace(string(out.Status)))
	return out
}

type FunnelReportFilter struct {
	Metric     AnalyticsMetric
	From       time.Time
	To         time.Time
	Steps      []EventType
	DeviceType DeviceType
	Channel    Channel
	Country    string
	Campaign   string
}

func (f FunnelReportFilter) Normalize() FunnelReportFilter {
	out := f
	out.Metric = AnalyticsMetric(strings.TrimSpace(string(out.Metric)))
	out.From = normalizeTime(out.From)
	out.To = normalizeTime(out.To)
	out.Steps = normalizeFunnelStepTypes(out.Steps)
	out.DeviceType = DeviceType(strings.TrimSpace(string(out.DeviceType)))
	out.Channel = Channel(strings.TrimSpace(string(out.Channel)))
	out.Country = strings.ToUpper(strings.TrimSpace(out.Country))
	out.Campaign = strings.TrimSpace(out.Campaign)
	return out
}

type FunnelStep struct {
	Name                   string    `json:"name" bson:"name"`
	EventType              EventType `json:"event_type,omitempty" bson:"event_type,omitempty"`
	Count                  int64     `json:"count" bson:"count"`
	UniqueSessions         int64     `json:"unique_sessions" bson:"unique_sessions"`
	UniqueUsers            int64     `json:"unique_users,omitempty" bson:"unique_users,omitempty"`
	ConversionFromPrevious float64   `json:"conversion_from_previous" bson:"conversion_from_previous"`
	DropoffFromPrevious    float64   `json:"dropoff_from_previous" bson:"dropoff_from_previous"`
}

func (s FunnelStep) Normalize() FunnelStep {
	out := s
	out.Name = strings.TrimSpace(out.Name)
	out.EventType = EventType(strings.TrimSpace(string(out.EventType)))
	if out.Name == "" && out.EventType != "" {
		out.Name = string(out.EventType)
	}
	if out.EventType == "" && out.Name != "" {
		out.EventType = EventType(out.Name)
	}
	out.ConversionFromPrevious = roundTwo(out.ConversionFromPrevious)
	out.DropoffFromPrevious = roundTwo(out.DropoffFromPrevious)
	return out
}

type AnalyticsAggregate struct {
	ID             string             `json:"id,omitempty" bson:"_id,omitempty"`
	Metric         AnalyticsMetric    `json:"metric" bson:"metric"`
	Bucket         AnalyticsBucket    `json:"bucket" bson:"bucket"`
	BucketStart    time.Time          `json:"bucket_start" bson:"bucket_start"`
	BucketEnd      time.Time          `json:"bucket_end" bson:"bucket_end"`
	Segment        map[string]string  `json:"segment,omitempty" bson:"segment,omitempty"`
	Steps          []FunnelStep       `json:"steps,omitempty" bson:"steps,omitempty"`
	Values         map[string]float64 `json:"values,omitempty" bson:"values,omitempty"`
	SchemaVersion  int                `json:"schema_version" bson:"schema_version"`
	CalculatedAt   time.Time          `json:"calculated_at" bson:"calculated_at"`
	CreatedAt      time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
	RetentionClass RetentionClass     `json:"retention_class,omitempty" bson:"retention_class,omitempty"`
	ContainsPII    bool               `json:"contains_pii,omitempty" bson:"contains_pii,omitempty"`
}

func (a AnalyticsAggregate) Normalize() AnalyticsAggregate {
	out := a
	out.ID = strings.TrimSpace(out.ID)
	out.Metric = AnalyticsMetric(strings.TrimSpace(string(out.Metric)))
	out.Bucket = AnalyticsBucket(strings.TrimSpace(string(out.Bucket)))
	out.BucketStart = normalizeTime(out.BucketStart)
	out.BucketEnd = normalizeTime(out.BucketEnd)
	out.Segment = normalizeAnalyticsSegment(out.Segment)
	out.Steps = NormalizeFunnelSteps(out.Steps)
	out.CalculatedAt = normalizeTime(out.CalculatedAt)
	out.CreatedAt = normalizeTime(out.CreatedAt)
	out.UpdatedAt = normalizeTime(out.UpdatedAt)
	out.RetentionClass = RetentionClass(strings.TrimSpace(string(out.RetentionClass)))
	return out
}

type RetentionAggregate struct {
	ID            string             `json:"id,omitempty" bson:"_id,omitempty"`
	Metric        AnalyticsMetric    `json:"metric" bson:"metric"`
	Bucket        AnalyticsBucket    `json:"bucket" bson:"bucket"`
	CohortStart   time.Time          `json:"cohort_start" bson:"cohort_start"`
	CohortEnd     time.Time          `json:"cohort_end" bson:"cohort_end"`
	Segment       map[string]string  `json:"segment,omitempty" bson:"segment,omitempty"`
	CohortSize    int64              `json:"cohort_size" bson:"cohort_size"`
	Retention     map[string]int64   `json:"retention" bson:"retention"`
	Rates         map[string]float64 `json:"rates" bson:"rates"`
	SchemaVersion int                `json:"schema_version" bson:"schema_version"`
	CalculatedAt  time.Time          `json:"calculated_at" bson:"calculated_at"`
	UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`
}

func NormalizeFunnelSteps(steps []FunnelStep) []FunnelStep {
	if len(steps) == 0 {
		return nil
	}
	out := make([]FunnelStep, 0, len(steps))
	for _, step := range steps {
		out = append(out, step.Normalize())
	}
	return out
}

func CalculateFunnelRates(steps []FunnelStep) []FunnelStep {
	out := NormalizeFunnelSteps(steps)
	for i := range out {
		if i == 0 {
			out[i].ConversionFromPrevious = 100
			out[i].DropoffFromPrevious = 0
			continue
		}
		previous := out[i-1].UniqueSessions
		if previous <= 0 {
			out[i].ConversionFromPrevious = 0
			out[i].DropoffFromPrevious = 100
			continue
		}
		conversion := float64(out[i].UniqueSessions) / float64(previous) * 100
		out[i].ConversionFromPrevious = roundTwo(conversion)
		out[i].DropoffFromPrevious = roundTwo(100 - conversion)
	}
	return out
}

func OverallFunnelConversion(steps []FunnelStep) float64 {
	if len(steps) < 2 || steps[0].UniqueSessions <= 0 {
		return 0
	}
	return roundTwo(float64(steps[len(steps)-1].UniqueSessions) / float64(steps[0].UniqueSessions) * 100)
}

func ValidateFunnelStepTypes(steps []EventType, minSteps int, maxSteps int) error {
	if minSteps <= 0 {
		minSteps = 2
	}
	if maxSteps < minSteps {
		maxSteps = minSteps
	}
	normalized := normalizeFunnelStepTypes(steps)
	if len(normalized) < minSteps {
		return fmt.Errorf("%w: at least %d funnel steps are required", ErrInvalidAnalytics, minSteps)
	}
	if len(normalized) > maxSteps {
		return fmt.Errorf("%w: funnel steps cannot exceed %d", ErrInvalidAnalytics, maxSteps)
	}
	for _, step := range normalized {
		if !step.SupportedForIngestion() {
			return fmt.Errorf("%w: unsupported funnel step %q", ErrInvalidAnalytics, step)
		}
	}
	return nil
}

func FunnelCacheStepKey(steps []EventType) string {
	normalized := normalizeFunnelStepTypes(steps)
	values := make([]string, 0, len(normalized))
	for _, step := range normalized {
		values = append(values, string(step))
	}
	return strings.Join(values, ",")
}

func sortedSegmentKey(segment map[string]string) string {
	if len(segment) == 0 {
		return ""
	}
	keys := make([]string, 0, len(segment))
	for key := range segment {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+segment[key])
	}
	return strings.Join(parts, "|")
}

func normalizeFunnelStepTypes(steps []EventType) []EventType {
	if len(steps) == 0 {
		return nil
	}
	out := make([]EventType, 0, len(steps))
	for _, step := range steps {
		step = EventType(strings.TrimSpace(string(step)))
		if step != "" {
			out = append(out, step)
		}
	}
	return out
}

func normalizeAnalyticsSegment(segment map[string]string) map[string]string {
	if len(segment) == 0 {
		return nil
	}
	out := make(map[string]string, len(segment))
	for key, value := range segment {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "" && value != "" {
			out[key] = value
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func roundTwo(value float64) float64 {
	return math.Round(value*100) / 100
}
