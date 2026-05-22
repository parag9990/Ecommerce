package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

const (
	defaultJourneyLimit           = 1000
	defaultJourneyMaxLimit        = 1000
	defaultJourneySummaryTopPaths = 10
)

type JourneyConfig struct {
	DefaultLimit          int
	MaxLimit              int
	SummaryTopPathsLimit  int
	SummaryUpsertDisabled bool
}

type GetJourneyInput struct {
	SessionID        string
	Limit            int
	CursorOccurredAt *time.Time
}

type JourneyOutput struct {
	Session          domain.Session
	Events           []domain.SessionEvent
	Summary          domain.JourneySummary
	SummaryPersisted bool
}

type JourneyUsecase struct {
	sessions  SessionRepository
	events    JourneyEventRepository
	summaries JourneySummaryRepository
	cfg       JourneyConfig
	logger    *slog.Logger
	clock     Clock
}

func NewJourneyUsecase(
	sessions SessionRepository,
	events JourneyEventRepository,
	summaries JourneySummaryRepository,
	cfg JourneyConfig,
	logger *slog.Logger,
) (*JourneyUsecase, error) {
	if sessions == nil {
		return nil, errors.New("session repository is required")
	}
	if events == nil {
		return nil, errors.New("journey event repository is required")
	}
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.SummaryUpsertEnabled() && summaries == nil {
		return nil, errors.New("journey summary repository is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &JourneyUsecase{
		sessions:  sessions,
		events:    events,
		summaries: summaries,
		cfg:       cfg,
		logger:    logger,
		clock:     realClock{},
	}, nil
}

func (u *JourneyUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func (u *JourneyUsecase) GetJourney(ctx context.Context, input GetJourneyInput) (JourneyOutput, error) {
	if err := ctx.Err(); err != nil {
		return JourneyOutput{}, err
	}

	started := time.Now()
	normalized, limit, err := u.normalizeInput(input)
	if err != nil {
		return JourneyOutput{}, err
	}

	session, err := u.sessions.FindSessionByID(ctx, normalized.SessionID)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return JourneyOutput{}, err
		}
		return JourneyOutput{}, fmt.Errorf("%w: find session: %w", ErrJourneyStorageUnavailable, err)
	}

	events, err := u.events.ListEventsBySessionTimeline(ctx, normalized.SessionID, limit, normalized.CursorOccurredAt)
	if err != nil {
		return JourneyOutput{}, fmt.Errorf("%w: list timeline events: %w", ErrJourneyStorageUnavailable, err)
	}
	events = deduplicateJourneyEvents(events)
	safeEvents := sanitizeJourneyEvents(events)

	summary := domain.BuildJourneySummary(session, events, u.clock.Now().UTC(), u.cfg.SummaryTopPathsLimit)
	persisted := false
	if u.cfg.SummaryUpsertEnabled() && u.summaries != nil && normalized.CursorOccurredAt == nil {
		if err := u.summaries.UpsertJourneySummary(ctx, summary); err != nil {
			u.logger.WarnContext(ctx, "session.journey.summary_upsert_failed",
				slog.String("session_id", normalized.SessionID),
				slog.Int("events_count", len(events)),
				slog.String("error", err.Error()),
			)
		} else {
			persisted = true
		}
	}

	u.logger.InfoContext(ctx, "session.journey.fetched",
		slog.String("session_id", normalized.SessionID),
		slog.Int("events_count", len(safeEvents)),
		slog.Int64("duration_ms", time.Since(started).Milliseconds()),
	)
	return JourneyOutput{
		Session:          session.Normalize(),
		Events:           safeEvents,
		Summary:          summary,
		SummaryPersisted: persisted,
	}, nil
}

func (u *JourneyUsecase) normalizeInput(input GetJourneyInput) (GetJourneyInput, int, error) {
	out := input
	out.SessionID = strings.TrimSpace(out.SessionID)
	if out.SessionID == "" {
		return GetJourneyInput{}, 0, fmt.Errorf("%w: session_id is required", ErrInvalidSessionInput)
	}
	if len(out.SessionID) > domain.DefaultMaxIDLength {
		return GetJourneyInput{}, 0, fmt.Errorf("%w: session_id is too long", ErrInvalidSessionInput)
	}
	if !strings.HasPrefix(out.SessionID, "sess_") {
		return GetJourneyInput{}, 0, fmt.Errorf("%w: session_id must start with sess_", ErrInvalidSessionInput)
	}

	limit := out.Limit
	if limit == 0 {
		limit = u.cfg.DefaultLimit
	}
	if limit < 0 {
		return GetJourneyInput{}, 0, fmt.Errorf("%w: limit must be greater than zero", ErrInvalidSessionInput)
	}
	if limit > u.cfg.MaxLimit {
		return GetJourneyInput{}, 0, fmt.Errorf("%w: limit cannot exceed %d", ErrInvalidSessionInput, u.cfg.MaxLimit)
	}
	if out.CursorOccurredAt != nil {
		cursor := out.CursorOccurredAt.UTC()
		if cursor.IsZero() {
			return GetJourneyInput{}, 0, fmt.Errorf("%w: cursor is invalid", ErrInvalidSessionInput)
		}
		out.CursorOccurredAt = &cursor
	}
	return out, limit, nil
}

func deduplicateJourneyEvents(events []domain.SessionEvent) []domain.SessionEvent {
	if len(events) == 0 {
		return nil
	}
	normalized := make([]domain.SessionEvent, 0, len(events))
	for _, event := range events {
		normalized = append(normalized, event.Normalize())
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		left := normalized[i]
		right := normalized[j]
		if !left.OccurredAt.Equal(right.OccurredAt) {
			return left.OccurredAt.Before(right.OccurredAt)
		}
		if !left.ReceivedAt.Equal(right.ReceivedAt) {
			return left.ReceivedAt.Before(right.ReceivedAt)
		}
		return left.EventID < right.EventID
	})

	seenEventIDs := make(map[string]struct{}, len(normalized))
	seenRetryKeys := make(map[string]struct{}, len(normalized))
	out := make([]domain.SessionEvent, 0, len(normalized))
	for _, event := range normalized {
		if event.EventID != "" {
			if _, exists := seenEventIDs[event.EventID]; exists {
				continue
			}
			seenEventIDs[event.EventID] = struct{}{}
		}
		if key := journeyRetryDedupKey(event); key != "" {
			if _, exists := seenRetryKeys[key]; exists {
				continue
			}
			seenRetryKeys[key] = struct{}{}
		}
		out = append(out, event)
	}
	return out
}

func sanitizeJourneyEvents(events []domain.SessionEvent) []domain.SessionEvent {
	if len(events) == 0 {
		return nil
	}
	out := make([]domain.SessionEvent, 0, len(events))
	for _, event := range events {
		safe := event.Normalize()
		safe.Properties = sanitizeProperties(safe.Properties)
		out = append(out, safe)
	}
	return out
}

func journeyRetryDedupKey(event domain.SessionEvent) string {
	if event.RequestID == nil || strings.TrimSpace(*event.RequestID) == "" {
		return ""
	}
	path := ""
	if event.Path != nil {
		path = strings.TrimSpace(*event.Path)
	}
	return strings.Join([]string{
		event.SessionID,
		string(event.EventType),
		event.OccurredAt.Format(time.RFC3339Nano),
		path,
		strings.TrimSpace(*event.RequestID),
	}, "|")
}

func (c JourneyConfig) Validate() error {
	if c.DefaultLimit <= 0 {
		return errors.New("SESSION_JOURNEY_DEFAULT_LIMIT must be greater than zero")
	}
	if c.MaxLimit < c.DefaultLimit {
		return errors.New("SESSION_JOURNEY_MAX_LIMIT must be greater than or equal to default journey limit")
	}
	if c.SummaryTopPathsLimit <= 0 {
		return errors.New("SESSION_JOURNEY_SUMMARY_TOP_PATHS_LIMIT must be greater than zero")
	}
	return nil
}

func (c JourneyConfig) withDefaults() JourneyConfig {
	if c.DefaultLimit == 0 {
		c.DefaultLimit = defaultJourneyLimit
	}
	if c.MaxLimit == 0 {
		c.MaxLimit = defaultJourneyMaxLimit
	}
	if c.SummaryTopPathsLimit == 0 {
		c.SummaryTopPathsLimit = defaultJourneySummaryTopPaths
	}
	return c
}

func (c JourneyConfig) SummaryUpsertEnabled() bool {
	return !c.SummaryUpsertDisabled
}
