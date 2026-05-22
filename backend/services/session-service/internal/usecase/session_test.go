package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

func TestPrepareSessionAppliesModelDefaults(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	usecase, err := NewSessionUsecase(Config{}, discardLogger())
	if err != nil {
		t.Fatalf("new usecase: %v", err)
	}
	usecase.WithClock(fixedClock{now: now})

	session, err := usecase.PrepareSession(context.Background(), domain.Session{
		SessionID:   "sess_123",
		AnonymousID: "anon_456",
		EntryPage:   "/products/prod_123",
		UserAgent:   "Mozilla/5.0",
		IPHash:      "b94d27b9934d3e08a52e52d7da7dabfade64b6d5",
	})
	if err != nil {
		t.Fatalf("prepare session: %v", err)
	}

	if session.SchemaVersion != domain.CurrentSessionSchemaVersion {
		t.Fatalf("unexpected schema version %d", session.SchemaVersion)
	}
	if session.Status != domain.SessionStatusActive || session.Channel != domain.ChannelUnknown {
		t.Fatalf("unexpected default status/channel: %s/%s", session.Status, session.Channel)
	}
	if session.Device.Type != domain.DeviceTypeUnknown || session.IPVersion != domain.IPVersionUnknown || session.RiskLevel != domain.RiskLevelUnknown {
		t.Fatalf("unexpected device/ip/risk defaults: %s/%s/%s", session.Device.Type, session.IPVersion, session.RiskLevel)
	}
	if !session.StartedAt.Equal(now) || !session.LastSeenAt.Equal(now) || !session.CreatedAt.Equal(now) || !session.UpdatedAt.Equal(now) {
		t.Fatalf("expected timestamps to default to fixed clock, got %+v", session)
	}
}

func TestPrepareSessionPropagatesContextCancellation(t *testing.T) {
	usecase, err := NewSessionUsecase(Config{}, discardLogger())
	if err != nil {
		t.Fatalf("new usecase: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = usecase.PrepareSession(ctx, domain.Session{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestPrepareSessionWrapsValidationErrors(t *testing.T) {
	usecase, err := NewSessionUsecase(Config{}, discardLogger())
	if err != nil {
		t.Fatalf("new usecase: %v", err)
	}

	_, err = usecase.PrepareSession(context.Background(), domain.Session{
		AnonymousID: "anon_456",
		EntryPage:   "/",
		UserAgent:   "Mozilla/5.0",
		IPHash:      "b94d27b9934d3e08a52e52d7da7dabfade64b6d5",
	})
	if !errors.Is(err, ErrInvalidSessionInput) {
		t.Fatalf("expected usecase validation wrapper, got %v", err)
	}
	if !errors.Is(err, domain.ErrInvalidSession) {
		t.Fatalf("expected domain validation error, got %v", err)
	}
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
