package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

var ErrInvalidSessionInput = errors.New("invalid session input")
var ErrIngestStorageUnavailable = errors.New("session ingest storage unavailable")
var ErrJourneyStorageUnavailable = errors.New("session journey storage unavailable")
var ErrHeatmapStorageUnavailable = errors.New("session heatmap storage unavailable")
var ErrAnalyticsStorageUnavailable = errors.New("session analytics storage unavailable")
var ErrAnalyticsAggregateNotReady = errors.New("session analytics aggregate not ready")
var ErrRetentionStorageUnavailable = errors.New("session retention storage unavailable")
var ErrRetentionPolicyInvalid = errors.New("session retention policy invalid")
var ErrRetentionAggregatePII = errors.New("session retention aggregate contains pii")

type Clock interface {
	Now() time.Time
}

type SessionRepository interface {
	UpsertSession(ctx context.Context, session domain.Session) error
	FindSessionByID(ctx context.Context, sessionID string) (domain.Session, error)
	ListUserSessions(ctx context.Context, userID string, limit int) ([]domain.Session, error)
	MarkEnded(ctx context.Context, sessionID string, endedAt time.Time, reason domain.SessionEndReason) error
}

type AnalyticsSessionRepository interface {
	ListSessions(ctx context.Context, filter domain.SessionListFilter) ([]domain.Session, int64, error)
}

type EventRepository interface {
	InsertEvent(ctx context.Context, event domain.SessionEvent) error
	ListEventsBySession(ctx context.Context, sessionID string, limit int) ([]domain.SessionEvent, error)
}

type JourneyEventRepository interface {
	ListEventsBySessionTimeline(ctx context.Context, sessionID string, limit int, cursorOccurredAt *time.Time) ([]domain.SessionEvent, error)
}

type JourneySummaryRepository interface {
	UpsertJourneySummary(ctx context.Context, summary domain.JourneySummary) error
	FindJourneySummaryBySessionID(ctx context.Context, sessionID string) (domain.JourneySummary, error)
}

type HeatmapEventRepository interface {
	ListHeatmapEvents(ctx context.Context, from time.Time, to time.Time, limit int, cursorAfter *domain.HeatmapEventCursor) ([]domain.SessionEvent, error)
}

type HeatmapRepository interface {
	UpsertHeatmapPoint(ctx context.Context, point domain.HeatmapPoint, sessionID string) error
	ListHeatmapPoints(ctx context.Context, filter domain.HeatmapFilter, limit int) ([]domain.HeatmapPoint, error)
	FindHeatmapCheckpoint(ctx context.Context, workerName string) (domain.HeatmapAggregationCheckpoint, error)
	SaveHeatmapCheckpoint(ctx context.Context, checkpoint domain.HeatmapAggregationCheckpoint) error
}

type LiveMetricsRepository interface {
	CountActiveUsers(ctx context.Context, window time.Duration) (int64, error)
	CountActiveSessions(ctx context.Context, window time.Duration) (int64, error)
	EventsPerMinute(ctx context.Context, window time.Duration) (float64, error)
}

type AnalyticsAggregateRepository interface {
	GetFunnelAggregate(ctx context.Context, filter domain.FunnelReportFilter) ([]domain.FunnelStep, error)
	BuildFunnelFromRawEvents(ctx context.Context, filter domain.FunnelReportFilter) ([]domain.FunnelStep, error)
}

type SessionTouchRepository interface {
	TouchSession(ctx context.Context, touch domain.SessionTouch) error
}

type ActiveSessionStore interface {
	Touch(ctx context.Context, snapshot domain.ActiveSessionSnapshot, ttl time.Duration) error
	Get(ctx context.Context, sessionID string) (domain.ActiveSessionSnapshot, error)
	Delete(ctx context.Context, sessionID string) error
	ListByUser(ctx context.Context, userID string) ([]string, error)
	ListByAnonymousID(ctx context.Context, anonymousID string) ([]string, error)
}

type SessionEventPublisher interface {
	PublishSessionEvent(ctx context.Context, event domain.SessionEvent) error
}

type RetentionRepository interface {
	AnonymizeExpiredSessions(ctx context.Context, now time.Time, retentionDays int, limit int, dryRun bool) (int64, error)
	AnonymizeExpiredJourneys(ctx context.Context, now time.Time, retentionDays int, limit int, dryRun bool) (int64, error)
	PurgeExpiredHeatmap(ctx context.Context, now time.Time, retentionDays int, limit int, dryRun bool) (int64, int64, error)
	PurgeExpiredAggregates(ctx context.Context, now time.Time, retentionYears int, limit int, dryRun bool) (int64, error)
	CountAggregatesWithPII(ctx context.Context) (int64, error)
	CountLegalHoldRetentionDue(ctx context.Context, now time.Time, sessionDays int, journeyDays int) (int64, error)
	DeleteRawEventsByIdentity(ctx context.Context, identity domain.SessionDataIdentity, dryRun bool) (int64, error)
	DeleteSessionsByIdentity(ctx context.Context, identity domain.SessionDataIdentity, dryRun bool) (int64, error)
	AnonymizeSessionsByIdentity(ctx context.Context, identity domain.SessionDataIdentity, at time.Time, dryRun bool) (int64, error)
	DeleteJourneysByIdentity(ctx context.Context, identity domain.SessionDataIdentity, dryRun bool) (int64, error)
	AnonymizeJourneysByIdentity(ctx context.Context, identity domain.SessionDataIdentity, at time.Time, dryRun bool) (int64, error)
	WriteDeletionAudit(ctx context.Context, request domain.DeleteUserSessionDataRequest, result domain.DeleteUserSessionDataResult) error
	SetLegalHold(ctx context.Context, request domain.SetLegalHoldRequest) (domain.SetLegalHoldResult, error)
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now().UTC()
}
