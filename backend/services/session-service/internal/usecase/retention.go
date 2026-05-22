package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

type RetentionConfig struct {
	RawEventTTLDays              int
	SessionMetadataRetentionDays int
	JourneySummaryRetentionDays  int
	HeatmapRetentionDays         int
	AggregateRetentionYears      int
	WorkerBatchSize              int
	WorkerInterval               time.Duration
	DryRun                       bool
	SmallCohortThreshold         int64
}

type RunRetentionCleanupInput struct {
	RequestID string
	DryRun    *bool
	Now       time.Time
}

type DeleteUserSessionDataInput struct {
	RequestID    string
	UserID       string
	AnonymousIDs []string
	Reason       string
	HardDelete   bool
	RequestedBy  string
	RequestedAt  time.Time
}

type SetLegalHoldInput struct {
	SessionID string
	Hold      bool
	Reason    string
	ActorID   string
	SetAt     time.Time
}

type RetentionUsecase struct {
	repository RetentionRepository
	active     ActiveSessionStore
	cfg        RetentionConfig
	logger     *slog.Logger
	clock      Clock
}

func NewRetentionUsecase(repository RetentionRepository, active ActiveSessionStore, cfg RetentionConfig, logger *slog.Logger) (*RetentionUsecase, error) {
	if repository == nil {
		return nil, errors.New("retention repository is required")
	}
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &RetentionUsecase{
		repository: repository,
		active:     active,
		cfg:        cfg,
		logger:     logger,
		clock:      realClock{},
	}, nil
}

func (u *RetentionUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func (u *RetentionUsecase) RunRetentionCleanup(ctx context.Context, input RunRetentionCleanupInput) (domain.RetentionCleanupResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.RetentionCleanupResult{}, err
	}
	now := input.Now.UTC()
	if now.IsZero() {
		now = u.clock.Now().UTC()
	}
	dryRun := u.cfg.DryRun
	if input.DryRun != nil {
		dryRun = *input.DryRun
	}
	requestID := input.RequestID
	if requestID == "" {
		requestID = newRequestID("ret")
	}
	startedAt := u.clock.Now().UTC()
	if startedAt.IsZero() {
		startedAt = now
	}

	result := domain.RetentionCleanupResult{
		RequestID: requestID,
		DryRun:    dryRun,
		StartedAt: startedAt,
	}
	var err error
	result.LegalHoldRecordsSkipped, err = u.repository.CountLegalHoldRetentionDue(ctx, now, u.cfg.SessionMetadataRetentionDays, u.cfg.JourneySummaryRetentionDays)
	if err != nil {
		return result, fmt.Errorf("%w: count legal hold records: %w", ErrRetentionStorageUnavailable, err)
	}
	result.SessionsAnonymized, err = u.repository.AnonymizeExpiredSessions(ctx, now, u.cfg.SessionMetadataRetentionDays, u.cfg.WorkerBatchSize, dryRun)
	if err != nil {
		return result, fmt.Errorf("%w: anonymize expired sessions: %w", ErrRetentionStorageUnavailable, err)
	}
	result.JourneysAnonymized, err = u.repository.AnonymizeExpiredJourneys(ctx, now, u.cfg.JourneySummaryRetentionDays, u.cfg.WorkerBatchSize, dryRun)
	if err != nil {
		return result, fmt.Errorf("%w: anonymize expired journeys: %w", ErrRetentionStorageUnavailable, err)
	}
	result.HeatmapPointsPurged, result.HeatmapSessionMarkersPurged, err = u.repository.PurgeExpiredHeatmap(ctx, now, u.cfg.HeatmapRetentionDays, u.cfg.WorkerBatchSize, dryRun)
	if err != nil {
		return result, fmt.Errorf("%w: purge expired heatmap: %w", ErrRetentionStorageUnavailable, err)
	}
	result.AnalyticsAggregatesPurged, err = u.repository.PurgeExpiredAggregates(ctx, now, u.cfg.AggregateRetentionYears, u.cfg.WorkerBatchSize, dryRun)
	if err != nil {
		return result, fmt.Errorf("%w: purge expired aggregates: %w", ErrRetentionStorageUnavailable, err)
	}
	result.AnalyticsAggregatesWithPII, err = u.repository.CountAggregatesWithPII(ctx)
	if err != nil {
		return result, fmt.Errorf("%w: verify aggregate pii policy: %w", ErrRetentionStorageUnavailable, err)
	}
	result.CompletedAt = u.clock.Now().UTC()
	if result.CompletedAt.IsZero() {
		result.CompletedAt = time.Now().UTC()
	}

	u.logger.InfoContext(ctx, "session.retention.cleanup.completed",
		slog.String("request_id", result.RequestID),
		slog.Bool("dry_run", result.DryRun),
		slog.Int64("sessions_anonymized", result.SessionsAnonymized),
		slog.Int64("journeys_anonymized", result.JourneysAnonymized),
		slog.Int64("heatmap_points_purged", result.HeatmapPointsPurged),
		slog.Int64("analytics_aggregates_purged", result.AnalyticsAggregatesPurged),
		slog.Int64("analytics_aggregates_with_pii", result.AnalyticsAggregatesWithPII),
	)
	if result.AnalyticsAggregatesWithPII > 0 {
		return result, fmt.Errorf("%w: %d aggregate documents violate pii policy", ErrRetentionAggregatePII, result.AnalyticsAggregatesWithPII)
	}
	return result, nil
}

func (u *RetentionUsecase) DeleteUserSessionData(ctx context.Context, input DeleteUserSessionDataInput) (domain.DeleteUserSessionDataResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.DeleteUserSessionDataResult{}, err
	}
	now := u.clock.Now().UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	req := domain.DeleteUserSessionDataRequest{
		RequestID: input.RequestID,
		Identity: domain.SessionDataIdentity{
			UserID:       input.UserID,
			AnonymousIDs: input.AnonymousIDs,
		},
		Reason:      input.Reason,
		HardDelete:  input.HardDelete,
		RequestedBy: input.RequestedBy,
		RequestedAt: input.RequestedAt,
	}.Normalize()
	if req.RequestID == "" {
		req.RequestID = newRequestID("del")
	}
	if req.RequestedAt.IsZero() {
		req.RequestedAt = now
	}
	if err := req.Validate(); err != nil {
		return domain.DeleteUserSessionDataResult{}, fmt.Errorf("%w: %w", ErrRetentionPolicyInvalid, err)
	}

	result := domain.DeleteUserSessionDataResult{
		RequestID:  req.RequestID,
		HardDelete: req.HardDelete,
	}
	var err error
	result.RawEventsDeleted, err = u.repository.DeleteRawEventsByIdentity(ctx, req.Identity, false)
	if err != nil {
		return result, fmt.Errorf("%w: delete raw events: %w", ErrRetentionStorageUnavailable, err)
	}
	if req.HardDelete {
		result.SessionsDeleted, err = u.repository.DeleteSessionsByIdentity(ctx, req.Identity, false)
		if err != nil {
			return result, fmt.Errorf("%w: delete sessions: %w", ErrRetentionStorageUnavailable, err)
		}
		result.JourneysDeleted, err = u.repository.DeleteJourneysByIdentity(ctx, req.Identity, false)
		if err != nil {
			return result, fmt.Errorf("%w: delete journeys: %w", ErrRetentionStorageUnavailable, err)
		}
	} else {
		result.SessionsAnonymized, err = u.repository.AnonymizeSessionsByIdentity(ctx, req.Identity, now, false)
		if err != nil {
			return result, fmt.Errorf("%w: anonymize sessions: %w", ErrRetentionStorageUnavailable, err)
		}
		result.JourneysAnonymized, err = u.repository.AnonymizeJourneysByIdentity(ctx, req.Identity, now, false)
		if err != nil {
			return result, fmt.Errorf("%w: anonymize journeys: %w", ErrRetentionStorageUnavailable, err)
		}
	}
	result.RedisKeysDeleted, result.RedisErrors = u.deleteActiveRedisKeys(ctx, req.Identity)
	result.CompletedAt = u.clock.Now().UTC()
	if result.CompletedAt.IsZero() {
		result.CompletedAt = time.Now().UTC()
	}
	if err := u.repository.WriteDeletionAudit(ctx, req, result); err != nil {
		return result, fmt.Errorf("%w: write deletion audit: %w", ErrRetentionStorageUnavailable, err)
	}
	u.logger.InfoContext(ctx, "session.retention.user_data_deleted",
		slog.String("request_id", result.RequestID),
		slog.Bool("hard_delete", result.HardDelete),
		slog.Int64("raw_events_deleted", result.RawEventsDeleted),
		slog.Int64("sessions_anonymized", result.SessionsAnonymized),
		slog.Int64("sessions_deleted", result.SessionsDeleted),
		slog.Int64("journeys_anonymized", result.JourneysAnonymized),
		slog.Int64("journeys_deleted", result.JourneysDeleted),
		slog.Int64("redis_keys_deleted", result.RedisKeysDeleted),
		slog.Int64("redis_errors", result.RedisErrors),
	)
	return result, nil
}

func (u *RetentionUsecase) SetLegalHold(ctx context.Context, input SetLegalHoldInput) (domain.SetLegalHoldResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.SetLegalHoldResult{}, err
	}
	now := input.SetAt.UTC()
	if now.IsZero() {
		now = u.clock.Now().UTC()
	}
	req := domain.SetLegalHoldRequest{
		SessionID: input.SessionID,
		Hold:      input.Hold,
		Reason:    input.Reason,
		ActorID:   input.ActorID,
		SetAt:     now,
	}.Normalize()
	if err := req.Validate(); err != nil {
		return domain.SetLegalHoldResult{}, fmt.Errorf("%w: %w", ErrRetentionPolicyInvalid, err)
	}
	result, err := u.repository.SetLegalHold(ctx, req)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return domain.SetLegalHoldResult{}, err
		}
		return domain.SetLegalHoldResult{}, fmt.Errorf("%w: set legal hold: %w", ErrRetentionStorageUnavailable, err)
	}
	u.logger.InfoContext(ctx, "session.retention.legal_hold_updated",
		slog.String("session_id", result.SessionID),
		slog.Bool("legal_hold", result.LegalHold),
		slog.String("actor_id", req.ActorID),
	)
	return result, nil
}

func (u *RetentionUsecase) deleteActiveRedisKeys(ctx context.Context, identity domain.SessionDataIdentity) (int64, int64) {
	if u.active == nil {
		return 0, 0
	}
	ids := make(map[string]struct{})
	var redisErrors int64
	if identity.UserID != "" {
		userIDs, err := u.active.ListByUser(ctx, identity.UserID)
		if err != nil {
			redisErrors++
			u.logger.WarnContext(ctx, "session.retention.redis_list_user_failed", slog.String("error", err.Error()))
		}
		for _, sessionID := range userIDs {
			ids[sessionID] = struct{}{}
		}
	}
	for _, anonymousID := range identity.AnonymousIDs {
		sessionIDs, err := u.active.ListByAnonymousID(ctx, anonymousID)
		if err != nil {
			redisErrors++
			u.logger.WarnContext(ctx, "session.retention.redis_list_anonymous_failed", slog.String("error", err.Error()))
			continue
		}
		for _, sessionID := range sessionIDs {
			ids[sessionID] = struct{}{}
		}
	}
	var deleted int64
	for sessionID := range ids {
		if err := u.active.Delete(ctx, sessionID); err != nil {
			redisErrors++
			u.logger.WarnContext(ctx, "session.retention.redis_delete_session_failed",
				slog.String("session_id", sessionID),
				slog.String("error", err.Error()),
			)
			continue
		}
		deleted++
	}
	return deleted, redisErrors
}

func (c RetentionConfig) Validate() error {
	policy := domain.RetentionPolicy{
		RawEventTTLDays:              c.RawEventTTLDays,
		SessionMetadataRetentionDays: c.SessionMetadataRetentionDays,
		JourneySummaryRetentionDays:  c.JourneySummaryRetentionDays,
		HeatmapRetentionDays:         c.HeatmapRetentionDays,
		AggregateRetentionYears:      c.AggregateRetentionYears,
		WorkerBatchSize:              c.WorkerBatchSize,
		WorkerInterval:               c.WorkerInterval,
		DryRun:                       c.DryRun,
		SmallCohortThreshold:         c.SmallCohortThreshold,
	}
	if err := policy.Validate(); err != nil {
		return err
	}
	return nil
}

func (c RetentionConfig) withDefaults() RetentionConfig {
	if c.SessionMetadataRetentionDays <= 0 {
		c.SessionMetadataRetentionDays = domain.DefaultSessionMetadataRetentionDays
	}
	if c.RawEventTTLDays <= 0 {
		c.RawEventTTLDays = 90
	}
	if c.JourneySummaryRetentionDays <= 0 {
		c.JourneySummaryRetentionDays = domain.DefaultJourneySummaryRetentionDays
	}
	if c.HeatmapRetentionDays <= 0 {
		c.HeatmapRetentionDays = domain.DefaultHeatmapRetentionDays
	}
	if c.AggregateRetentionYears <= 0 {
		c.AggregateRetentionYears = domain.DefaultAggregateRetentionYears
	}
	if c.WorkerBatchSize <= 0 {
		c.WorkerBatchSize = domain.DefaultRetentionWorkerBatchSize
	}
	if c.WorkerInterval <= 0 {
		c.WorkerInterval = domain.DefaultRetentionWorkerInterval
	}
	if c.SmallCohortThreshold <= 0 {
		c.SmallCohortThreshold = domain.DefaultSmallCohortThreshold
	}
	return c
}

func newRequestID(prefix string) string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err == nil {
		return prefix + "_" + hex.EncodeToString(bytes[:])
	}
	return prefix + "_" + time.Now().UTC().Format("20060102150405.000000000")
}
