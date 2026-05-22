package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const (
	DefaultSessionMetadataRetentionDays = 365
	DefaultJourneySummaryRetentionDays  = 365
	DefaultHeatmapRetentionDays         = 730
	DefaultAggregateRetentionYears      = 5
	DefaultRetentionWorkerBatchSize     = 1000
	DefaultRetentionWorkerInterval      = time.Hour
	DefaultSmallCohortThreshold         = 5
)

type RetentionClass string

const (
	RetentionClassRawEvent          RetentionClass = "raw_event"
	RetentionClassSessionMetadata   RetentionClass = "session_metadata"
	RetentionClassDerivedSummary    RetentionClass = "derived_summary"
	RetentionClassAggregateFine     RetentionClass = "aggregate_fine"
	RetentionClassAggregateBusiness RetentionClass = "aggregate_business"
	RetentionClassOperationalLog    RetentionClass = "operational_log"
	RetentionClassDeletionAudit     RetentionClass = "deletion_audit"
)

func (c RetentionClass) Valid() bool {
	switch c {
	case RetentionClassRawEvent,
		RetentionClassSessionMetadata,
		RetentionClassDerivedSummary,
		RetentionClassAggregateFine,
		RetentionClassAggregateBusiness,
		RetentionClassOperationalLog,
		RetentionClassDeletionAudit:
		return true
	default:
		return false
	}
}

type RetentionPolicy struct {
	RawEventTTLDays              int           `json:"raw_event_ttl_days"`
	SessionMetadataRetentionDays int           `json:"session_metadata_retention_days"`
	JourneySummaryRetentionDays  int           `json:"journey_summary_retention_days"`
	HeatmapRetentionDays         int           `json:"heatmap_retention_days"`
	AggregateRetentionYears      int           `json:"aggregate_retention_years"`
	WorkerBatchSize              int           `json:"worker_batch_size"`
	WorkerInterval               time.Duration `json:"worker_interval"`
	DryRun                       bool          `json:"dry_run"`
	SmallCohortThreshold         int64         `json:"small_cohort_threshold"`
}

func (p RetentionPolicy) WithDefaults() RetentionPolicy {
	if p.RawEventTTLDays <= 0 {
		p.RawEventTTLDays = 90
	}
	if p.SessionMetadataRetentionDays <= 0 {
		p.SessionMetadataRetentionDays = DefaultSessionMetadataRetentionDays
	}
	if p.JourneySummaryRetentionDays <= 0 {
		p.JourneySummaryRetentionDays = DefaultJourneySummaryRetentionDays
	}
	if p.HeatmapRetentionDays <= 0 {
		p.HeatmapRetentionDays = DefaultHeatmapRetentionDays
	}
	if p.AggregateRetentionYears <= 0 {
		p.AggregateRetentionYears = DefaultAggregateRetentionYears
	}
	if p.WorkerBatchSize <= 0 {
		p.WorkerBatchSize = DefaultRetentionWorkerBatchSize
	}
	if p.WorkerInterval <= 0 {
		p.WorkerInterval = DefaultRetentionWorkerInterval
	}
	if p.SmallCohortThreshold <= 0 {
		p.SmallCohortThreshold = DefaultSmallCohortThreshold
	}
	return p
}

func (p RetentionPolicy) Validate() error {
	p = p.WithDefaults()
	if p.RawEventTTLDays < 30 || p.RawEventTTLDays > 180 {
		return fmt.Errorf("%w: raw event TTL must be between 30 and 180 days", ErrInvalidRetentionPolicy)
	}
	if p.SessionMetadataRetentionDays < p.RawEventTTLDays {
		return fmt.Errorf("%w: session metadata retention cannot be shorter than raw event TTL", ErrInvalidRetentionPolicy)
	}
	if p.JourneySummaryRetentionDays < p.RawEventTTLDays {
		return fmt.Errorf("%w: journey summary retention cannot be shorter than raw event TTL", ErrInvalidRetentionPolicy)
	}
	if p.HeatmapRetentionDays < p.JourneySummaryRetentionDays {
		return fmt.Errorf("%w: heatmap retention cannot be shorter than journey summary retention", ErrInvalidRetentionPolicy)
	}
	if p.AggregateRetentionYears < 1 || p.AggregateRetentionYears > 10 {
		return fmt.Errorf("%w: aggregate retention must be between 1 and 10 years", ErrInvalidRetentionPolicy)
	}
	if p.WorkerBatchSize <= 0 || p.WorkerBatchSize > 10000 {
		return fmt.Errorf("%w: worker batch size must be between 1 and 10000", ErrInvalidRetentionPolicy)
	}
	if p.WorkerInterval <= 0 {
		return fmt.Errorf("%w: worker interval must be greater than zero", ErrInvalidRetentionPolicy)
	}
	if p.SmallCohortThreshold < 2 || p.SmallCohortThreshold > 100 {
		return fmt.Errorf("%w: small cohort threshold must be between 2 and 100", ErrInvalidRetentionPolicy)
	}
	return nil
}

type SessionDataIdentity struct {
	UserID       string   `json:"user_id,omitempty"`
	AnonymousIDs []string `json:"anonymous_ids,omitempty"`
}

func (i SessionDataIdentity) Normalize() SessionDataIdentity {
	out := i
	out.UserID = strings.TrimSpace(out.UserID)
	out.AnonymousIDs = normalizeUniqueStrings(out.AnonymousIDs)
	return out
}

func (i SessionDataIdentity) Validate() error {
	identity := i.Normalize()
	if identity.UserID == "" && len(identity.AnonymousIDs) == 0 {
		return fmt.Errorf("%w: user_id or anonymous_ids is required", ErrInvalidRetentionPolicy)
	}
	violations := make([]FieldViolation, 0)
	validateID(&violations, "user_id", identity.UserID, false, DefaultMaxIDLength)
	for idx, anonymousID := range identity.AnonymousIDs {
		validateID(&violations, fmt.Sprintf("anonymous_ids[%d]", idx), anonymousID, true, DefaultMaxIDLength)
	}
	if len(violations) > 0 {
		return RetentionValidationError{Violations: violations}
	}
	return nil
}

type DeleteUserSessionDataRequest struct {
	RequestID   string              `json:"request_id,omitempty"`
	Identity    SessionDataIdentity `json:"identity"`
	Reason      string              `json:"reason"`
	HardDelete  bool                `json:"hard_delete"`
	RequestedBy string              `json:"requested_by,omitempty"`
	RequestedAt time.Time           `json:"requested_at,omitempty"`
}

func (r DeleteUserSessionDataRequest) Normalize() DeleteUserSessionDataRequest {
	out := r
	out.RequestID = strings.TrimSpace(out.RequestID)
	out.Identity = out.Identity.Normalize()
	out.Reason = strings.TrimSpace(out.Reason)
	out.RequestedBy = strings.TrimSpace(out.RequestedBy)
	out.RequestedAt = normalizeTime(out.RequestedAt)
	return out
}

func (r DeleteUserSessionDataRequest) Validate() error {
	req := r.Normalize()
	if err := req.Identity.Validate(); err != nil {
		return err
	}
	violations := make([]FieldViolation, 0)
	validateID(&violations, "request_id", req.RequestID, false, DefaultMaxIDLength)
	validateID(&violations, "requested_by", req.RequestedBy, false, DefaultMaxIDLength)
	validateRequiredText(&violations, "reason", req.Reason, 512)
	if len(violations) > 0 {
		return RetentionValidationError{Violations: violations}
	}
	return nil
}

type DeleteUserSessionDataResult struct {
	RequestID          string    `json:"request_id"`
	RawEventsDeleted   int64     `json:"raw_events_deleted"`
	SessionsAnonymized int64     `json:"sessions_anonymized"`
	SessionsDeleted    int64     `json:"sessions_deleted"`
	JourneysAnonymized int64     `json:"journeys_anonymized"`
	JourneysDeleted    int64     `json:"journeys_deleted"`
	RedisKeysDeleted   int64     `json:"redis_keys_deleted"`
	RedisErrors        int64     `json:"redis_errors,omitempty"`
	HardDelete         bool      `json:"hard_delete"`
	CompletedAt        time.Time `json:"completed_at"`
}

type RetentionCleanupResult struct {
	RequestID                   string    `json:"request_id"`
	DryRun                      bool      `json:"dry_run"`
	SessionsAnonymized          int64     `json:"sessions_anonymized"`
	JourneysAnonymized          int64     `json:"journeys_anonymized"`
	HeatmapPointsPurged         int64     `json:"heatmap_points_purged"`
	HeatmapSessionMarkersPurged int64     `json:"heatmap_session_markers_purged"`
	AnalyticsAggregatesPurged   int64     `json:"analytics_aggregates_purged"`
	AnalyticsAggregatesWithPII  int64     `json:"analytics_aggregates_with_pii"`
	LegalHoldRecordsSkipped     int64     `json:"legal_hold_records_skipped"`
	StartedAt                   time.Time `json:"started_at"`
	CompletedAt                 time.Time `json:"completed_at"`
}

type SetLegalHoldRequest struct {
	SessionID string    `json:"session_id"`
	Hold      bool      `json:"hold"`
	Reason    string    `json:"reason,omitempty"`
	ActorID   string    `json:"actor_id,omitempty"`
	SetAt     time.Time `json:"set_at,omitempty"`
}

func (r SetLegalHoldRequest) Normalize() SetLegalHoldRequest {
	out := r
	out.SessionID = strings.TrimSpace(out.SessionID)
	out.Reason = strings.TrimSpace(out.Reason)
	out.ActorID = strings.TrimSpace(out.ActorID)
	out.SetAt = normalizeTime(out.SetAt)
	return out
}

func (r SetLegalHoldRequest) Validate() error {
	req := r.Normalize()
	violations := make([]FieldViolation, 0)
	validateID(&violations, "session_id", req.SessionID, true, DefaultMaxIDLength)
	validateID(&violations, "actor_id", req.ActorID, false, DefaultMaxIDLength)
	if req.Hold {
		validateRequiredText(&violations, "reason", req.Reason, 512)
	}
	if len(violations) > 0 {
		return RetentionValidationError{Violations: violations}
	}
	return nil
}

type SetLegalHoldResult struct {
	SessionID       string    `json:"session_id"`
	LegalHold       bool      `json:"legal_hold"`
	SessionsMatched int64     `json:"sessions_matched"`
	JourneysMatched int64     `json:"journeys_matched"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type RetentionValidationError struct {
	Violations []FieldViolation
}

func (e RetentionValidationError) Error() string {
	if len(e.Violations) == 0 {
		return ErrInvalidRetentionPolicy.Error()
	}
	parts := make([]string, 0, len(e.Violations))
	for _, violation := range e.Violations {
		parts = append(parts, fmt.Sprintf("%s: %s", violation.Field, violation.Message))
	}
	return ErrInvalidRetentionPolicy.Error() + ": " + strings.Join(parts, "; ")
}

func (e RetentionValidationError) Is(target error) bool {
	return target == ErrInvalidRetentionPolicy
}

func SuppressSmallCohort(count int64, minThreshold int64) (int64, bool) {
	if minThreshold <= 0 {
		minThreshold = DefaultSmallCohortThreshold
	}
	if count > 0 && count < minThreshold {
		return 0, true
	}
	return count, false
}

func HashRetentionIdentifier(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func normalizeUniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
