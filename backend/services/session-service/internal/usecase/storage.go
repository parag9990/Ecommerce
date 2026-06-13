package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

const (
	defaultActiveSessionTTL = 35 * time.Minute
	defaultListLimit        = 100
	defaultMaxListLimit     = 500
)

type StorageConfig struct {
	ActiveSessionTTL time.Duration
	DefaultListLimit int
	MaxListLimit     int
	SessionEvent     domain.SessionEventValidationConfig
}

func DefaultStorageConfig() StorageConfig {
	return StorageConfig{
		ActiveSessionTTL: defaultActiveSessionTTL,
		DefaultListLimit: defaultListLimit,
		MaxListLimit:     defaultMaxListLimit,
		SessionEvent:     domain.DefaultSessionEventValidationConfig(),
	}
}

type StorageUsecase struct {
	sessions SessionRepository
	events   EventRepository
	active   ActiveSessionStore
	cfg      StorageConfig
	logger   *slog.Logger
	clock    Clock
}

func NewStorageUsecase(
	sessions SessionRepository,
	events EventRepository,
	active ActiveSessionStore,
	cfg StorageConfig,
	logger *slog.Logger,
) (*StorageUsecase, error) {
	if sessions == nil {
		return nil, errors.New("session repository is required")
	}
	if events == nil {
		return nil, errors.New("event repository is required")
	}
	if active == nil {
		return nil, errors.New("active session store is required")
	}
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &StorageUsecase{
		sessions: sessions,
		events:   events,
		active:   active,
		cfg:      cfg,
		logger:   logger,
		clock:    realClock{},
	}, nil
}

func (u *StorageUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func (u *StorageUsecase) UpsertSession(ctx context.Context, session domain.Session) (domain.Session, error) {
	if err := ctx.Err(); err != nil {
		return domain.Session{}, err
	}
	normalized := session.Normalize()
	if err := normalized.Validate(); err != nil {
		return domain.Session{}, fmt.Errorf("%w: %w", ErrInvalidSessionInput, err)
	}
	if err := u.sessions.UpsertSession(ctx, normalized); err != nil {
		return domain.Session{}, err
	}
	u.logger.DebugContext(ctx, "session.storage.session_upserted",
		slog.String("session_id", normalized.SessionID),
		slog.String("anonymous_id", normalized.AnonymousID),
	)
	return normalized, nil
}

func (u *StorageUsecase) FindSessionByID(ctx context.Context, sessionID string) (domain.Session, error) {
	if err := ctx.Err(); err != nil {
		return domain.Session{}, err
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return domain.Session{}, fmt.Errorf("%w: session_id is required", ErrInvalidSessionInput)
	}
	return u.sessions.FindSessionByID(ctx, sessionID)
}

func (u *StorageUsecase) ListUserSessions(ctx context.Context, userID string, limit int) ([]domain.Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidSessionInput)
	}
	return u.sessions.ListUserSessions(ctx, userID, u.normalizeLimit(limit))
}

func (u *StorageUsecase) MarkSessionEnded(ctx context.Context, sessionID string, endedAt time.Time, reason domain.SessionEndReason) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return fmt.Errorf("%w: session_id is required", ErrInvalidSessionInput)
	}
	if reason == "" {
		reason = domain.EndReasonUnknown
	}
	if !reason.Valid() {
		return fmt.Errorf("%w: unsupported end reason", ErrInvalidSessionInput)
	}
	if endedAt.IsZero() {
		endedAt = u.clock.Now()
	}
	if err := u.sessions.MarkEnded(ctx, sessionID, endedAt.UTC(), reason); err != nil {
		return err
	}
	if err := u.active.Delete(ctx, sessionID); err != nil {
		u.logger.WarnContext(ctx, "session.storage.active_delete_failed",
			slog.String("session_id", sessionID),
			slog.String("error", err.Error()),
		)
	}
	return nil
}

func (u *StorageUsecase) InsertEvent(ctx context.Context, event domain.SessionEvent) (domain.SessionEvent, error) {
	if err := ctx.Err(); err != nil {
		return domain.SessionEvent{}, err
	}
	normalized := event.Normalize()
	if normalized.SchemaVersion == 0 {
		normalized.SchemaVersion = domain.CurrentSessionEventSchemaVersion
	}
	if normalized.ReceivedAt.IsZero() {
		normalized.ReceivedAt = u.clock.Now().UTC()
	}
	if err := normalized.ValidateWithConfig(u.cfg.SessionEvent); err != nil {
		return domain.SessionEvent{}, fmt.Errorf("%w: %w", ErrInvalidSessionInput, err)
	}
	if err := u.events.InsertEvent(ctx, normalized); err != nil {
		return domain.SessionEvent{}, err
	}
	u.logger.DebugContext(ctx, "session.storage.event_inserted",
		slog.String("event_id", normalized.EventID),
		slog.String("session_id", normalized.SessionID),
		slog.String("event_type", string(normalized.EventType)),
	)
	return normalized, nil
}

func (u *StorageUsecase) ListSessionEvents(ctx context.Context, sessionID string, limit int) ([]domain.SessionEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("%w: session_id is required", ErrInvalidSessionInput)
	}
	return u.events.ListEventsBySession(ctx, sessionID, u.normalizeLimit(limit))
}

func (u *StorageUsecase) TouchActiveSession(ctx context.Context, snapshot domain.ActiveSessionSnapshot) (domain.ActiveSessionSnapshot, error) {
	if err := ctx.Err(); err != nil {
		return domain.ActiveSessionSnapshot{}, err
	}
	normalized := snapshot.Normalize()
	if normalized.LastSeenAt.IsZero() {
		normalized.LastSeenAt = u.clock.Now().UTC()
	}
	if err := normalized.Validate(); err != nil {
		return domain.ActiveSessionSnapshot{}, fmt.Errorf("%w: %w", ErrInvalidSessionInput, err)
	}
	if err := u.active.Touch(ctx, normalized, u.cfg.ActiveSessionTTL); err != nil {
		return domain.ActiveSessionSnapshot{}, err
	}
	return normalized, nil
}

func (u *StorageUsecase) GetActiveSession(ctx context.Context, sessionID string) (domain.ActiveSessionSnapshot, error) {
	if err := ctx.Err(); err != nil {
		return domain.ActiveSessionSnapshot{}, err
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return domain.ActiveSessionSnapshot{}, fmt.Errorf("%w: session_id is required", ErrInvalidSessionInput)
	}
	return u.active.Get(ctx, sessionID)
}

func (u *StorageUsecase) DeleteActiveSession(ctx context.Context, sessionID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return fmt.Errorf("%w: session_id is required", ErrInvalidSessionInput)
	}
	return u.active.Delete(ctx, sessionID)
}

func (u *StorageUsecase) ListActiveSessionsByUser(ctx context.Context, userID string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidSessionInput)
	}
	return u.active.ListByUser(ctx, userID)
}

func (u *StorageUsecase) normalizeLimit(limit int) int {
	if limit <= 0 {
		return u.cfg.DefaultListLimit
	}
	if limit > u.cfg.MaxListLimit {
		return u.cfg.MaxListLimit
	}
	return limit
}

func (c StorageConfig) Validate() error {
	if c.ActiveSessionTTL <= 0 {
		return errors.New("SESSION_ACTIVE_TTL must be greater than zero")
	}
	if c.DefaultListLimit <= 0 {
		return errors.New("SESSION_DEFAULT_LIST_LIMIT must be greater than zero")
	}
	if c.MaxListLimit < c.DefaultListLimit {
		return errors.New("SESSION_MAX_LIST_LIMIT must be greater than or equal to default list limit")
	}
	return nil
}

func (c StorageConfig) withDefaults() StorageConfig {
	defaults := DefaultStorageConfig()
	if c.ActiveSessionTTL == 0 {
		c.ActiveSessionTTL = defaults.ActiveSessionTTL
	}
	if c.DefaultListLimit == 0 {
		c.DefaultListLimit = defaults.DefaultListLimit
	}
	if c.MaxListLimit == 0 {
		c.MaxListLimit = defaults.MaxListLimit
	}
	c.SessionEvent = c.SessionEvent.WithDefaults()
	return c
}
