package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

func TestGetJourneyReturnsSortedDedupedSafeEventsAndUpsertsSummary(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	session := journeyTestSession(now)
	reqID := "req_retry"
	home := "/"
	product := "/products/prod_123"
	events := &fakeJourneyEventRepository{events: []domain.SessionEvent{
		journeyUsecaseEvent("evt_003", domain.EventAddToCart, product, now.Add(3*time.Minute), nil, map[string]any{"product_id": "prod_123"}),
		journeyUsecaseEvent("evt_001", domain.EventPageView, home, now, &reqID, map[string]any{"title": "Home", "password": "secret"}),
		journeyUsecaseEvent("evt_002", domain.EventPageView, home, now, &reqID, map[string]any{"title": "Home retry"}),
		journeyUsecaseEvent("evt_004", domain.EventPaymentResult, "/checkout/success", now.Add(5*time.Minute), nil, map[string]any{"status": "paid"}),
	}}
	summaries := &fakeJourneySummaryRepository{}
	uc, err := NewJourneyUsecase(&fakeJourneySessionRepository{session: session}, events, summaries, JourneyConfig{DefaultLimit: 10, MaxLimit: 10}, discardLogger())
	if err != nil {
		t.Fatalf("new journey usecase: %v", err)
	}
	uc.WithClock(fixedClock{now: now.Add(6 * time.Minute)})

	out, err := uc.GetJourney(context.Background(), GetJourneyInput{SessionID: session.SessionID, Limit: 10})
	if err != nil {
		t.Fatalf("get journey: %v", err)
	}
	if len(out.Events) != 3 {
		t.Fatalf("expected duplicate retry event to be removed, got %d", len(out.Events))
	}
	if out.Events[0].EventID != "evt_001" || out.Events[1].EventID != "evt_003" || out.Events[2].EventID != "evt_004" {
		t.Fatalf("events not sorted/deduped: %+v", out.Events)
	}
	if _, ok := out.Events[0].Properties["password"]; ok {
		t.Fatal("expected sensitive property to be removed from journey output")
	}
	if !out.Summary.PaymentCompleted || out.Summary.CartActions != 1 || out.Summary.TotalEvents != 3 {
		t.Fatalf("unexpected summary: %+v", out.Summary)
	}
	if !out.SummaryPersisted || summaries.upserted.SessionID != session.SessionID {
		t.Fatalf("expected summary upsert, persisted=%v summary=%+v", out.SummaryPersisted, summaries.upserted)
	}
}

func TestGetJourneyToleratesSummaryUpsertFailure(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	session := journeyTestSession(now)
	uc, err := NewJourneyUsecase(
		&fakeJourneySessionRepository{session: session},
		&fakeJourneyEventRepository{events: []domain.SessionEvent{journeyUsecaseEvent("evt_001", domain.EventPageView, "/", now, nil, nil)}},
		&fakeJourneySummaryRepository{upsertErr: errors.New("mongo unavailable")},
		JourneyConfig{},
		discardLogger(),
	)
	if err != nil {
		t.Fatalf("new journey usecase: %v", err)
	}
	uc.WithClock(fixedClock{now: now})

	out, err := uc.GetJourney(context.Background(), GetJourneyInput{SessionID: session.SessionID})
	if err != nil {
		t.Fatalf("summary upsert failure should not fail journey read, got %v", err)
	}
	if out.SummaryPersisted {
		t.Fatal("expected summary persisted flag to be false")
	}
	if len(out.Events) != 1 {
		t.Fatalf("expected timeline to be returned, got %d", len(out.Events))
	}
}

func TestGetJourneyRejectsExcessiveLimit(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	uc, err := NewJourneyUsecase(
		&fakeJourneySessionRepository{session: journeyTestSession(now)},
		&fakeJourneyEventRepository{},
		&fakeJourneySummaryRepository{},
		JourneyConfig{DefaultLimit: 5, MaxLimit: 5},
		discardLogger(),
	)
	if err != nil {
		t.Fatalf("new journey usecase: %v", err)
	}

	_, err = uc.GetJourney(context.Background(), GetJourneyInput{SessionID: "sess_123", Limit: 6})
	if !errors.Is(err, ErrInvalidSessionInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestGetJourneyPropagatesSessionNotFound(t *testing.T) {
	uc, err := NewJourneyUsecase(
		&fakeJourneySessionRepository{err: domain.ErrSessionNotFound},
		&fakeJourneyEventRepository{},
		&fakeJourneySummaryRepository{},
		JourneyConfig{},
		discardLogger(),
	)
	if err != nil {
		t.Fatalf("new journey usecase: %v", err)
	}

	_, err = uc.GetJourney(context.Background(), GetJourneyInput{SessionID: "sess_123"})
	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("expected session not found, got %v", err)
	}
}

type fakeJourneySessionRepository struct {
	session domain.Session
	err     error
}

func (r *fakeJourneySessionRepository) UpsertSession(ctx context.Context, session domain.Session) error {
	return nil
}

func (r *fakeJourneySessionRepository) FindSessionByID(ctx context.Context, sessionID string) (domain.Session, error) {
	if r.err != nil {
		return domain.Session{}, r.err
	}
	session := r.session
	session.SessionID = sessionID
	return session, nil
}

func (r *fakeJourneySessionRepository) ListUserSessions(ctx context.Context, userID string, limit int) ([]domain.Session, error) {
	return nil, nil
}

func (r *fakeJourneySessionRepository) MarkEnded(ctx context.Context, sessionID string, endedAt time.Time, reason domain.SessionEndReason) error {
	return nil
}

type fakeJourneyEventRepository struct {
	events []domain.SessionEvent
	err    error
	limit  int
	cursor *time.Time
}

func (r *fakeJourneyEventRepository) ListEventsBySessionTimeline(ctx context.Context, sessionID string, limit int, cursorOccurredAt *time.Time) ([]domain.SessionEvent, error) {
	r.limit = limit
	r.cursor = cursorOccurredAt
	if r.err != nil {
		return nil, r.err
	}
	return append([]domain.SessionEvent(nil), r.events...), nil
}

type fakeJourneySummaryRepository struct {
	upserted  domain.JourneySummary
	upsertErr error
}

func (r *fakeJourneySummaryRepository) UpsertJourneySummary(ctx context.Context, summary domain.JourneySummary) error {
	if r.upsertErr != nil {
		return r.upsertErr
	}
	r.upserted = summary
	return nil
}

func (r *fakeJourneySummaryRepository) FindJourneySummaryBySessionID(ctx context.Context, sessionID string) (domain.JourneySummary, error) {
	return r.upserted, nil
}

func journeyTestSession(now time.Time) domain.Session {
	return domain.Session{
		SessionID:     "sess_123",
		AnonymousID:   "anon_123",
		SchemaVersion: domain.CurrentSessionSchemaVersion,
		Status:        domain.SessionStatusActive,
		Channel:       domain.ChannelUserAppWeb,
		EntryPage:     "/",
		UserAgent:     "Mozilla/5.0",
		Device:        domain.Device{Type: domain.DeviceTypeDesktop},
		IPHash:        "7c9e6679f7425d42a01e0cfbdac7c7b39f4f5f9c31f5b22d9efb9b9f7436e75a",
		IPVersion:     domain.IPVersionIPv4,
		RiskLevel:     domain.RiskLevelLow,
		RiskReasons:   []string{},
		StartedAt:     now,
		LastSeenAt:    now.Add(time.Minute),
		CreatedAt:     now,
		UpdatedAt:     now.Add(time.Minute),
	}
}

func journeyUsecaseEvent(eventID string, eventType domain.EventType, path string, occurredAt time.Time, requestID *string, properties map[string]any) domain.SessionEvent {
	return domain.SessionEvent{
		EventID:       eventID,
		SessionID:     "sess_123",
		AnonymousID:   "anon_123",
		EventType:     eventType,
		Path:          &path,
		Properties:    properties,
		OccurredAt:    occurredAt,
		ReceivedAt:    occurredAt.Add(time.Second),
		SchemaVersion: domain.CurrentSessionEventSchemaVersion,
		RequestID:     requestID,
	}
}
