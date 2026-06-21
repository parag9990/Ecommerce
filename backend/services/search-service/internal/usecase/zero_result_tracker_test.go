package usecase

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

func TestZeroResultTrackerTrackNowSendsFirstSeenEvent(t *testing.T) {
	dedupe := &fakeZeroResultDedupe{firstSeen: true}
	sink := &fakeSessionSink{}
	metrics := &capturingZeroResultMetrics{}
	tracker := newZeroResultTrackerForTest(t, dedupe, sink, metrics)

	tracker.trackNow(context.Background(), trackableZeroResultEvent())

	if dedupe.key == "" {
		t.Fatal("dedupe was not called")
	}
	if len(sink.events) != 1 {
		t.Fatalf("sink events = %d", len(sink.events))
	}
	if metrics.last.Outcome != ZeroResultOutcomeTracked {
		t.Fatalf("metrics = %#v", metrics.last)
	}
}

func TestZeroResultTrackerSkipsDuplicate(t *testing.T) {
	dedupe := &fakeZeroResultDedupe{firstSeen: false}
	sink := &fakeSessionSink{}
	metrics := &capturingZeroResultMetrics{}
	tracker := newZeroResultTrackerForTest(t, dedupe, sink, metrics)

	tracker.trackNow(context.Background(), trackableZeroResultEvent())

	if len(sink.events) != 0 {
		t.Fatalf("sink events = %d", len(sink.events))
	}
	if metrics.last.Outcome != ZeroResultOutcomeDuplicate {
		t.Fatalf("metrics = %#v", metrics.last)
	}
}

func TestZeroResultTrackerHandlesDedupeAndSinkFailure(t *testing.T) {
	t.Run("dedupe", func(t *testing.T) {
		metrics := &capturingZeroResultMetrics{}
		tracker := newZeroResultTrackerForTest(t, &fakeZeroResultDedupe{err: errors.New("redis down")}, &fakeSessionSink{}, metrics)

		tracker.trackNow(context.Background(), trackableZeroResultEvent())

		if metrics.last.Outcome != ZeroResultOutcomeFailed || metrics.last.Reason != "dedupe_error" {
			t.Fatalf("metrics = %#v", metrics.last)
		}
	})

	t.Run("sink", func(t *testing.T) {
		metrics := &capturingZeroResultMetrics{}
		dedupe := &fakeZeroResultDedupe{firstSeen: true}
		tracker := newZeroResultTrackerForTest(t, dedupe, &fakeSessionSink{err: errors.New("session down")}, metrics)

		tracker.trackNow(context.Background(), trackableZeroResultEvent())

		if metrics.last.Outcome != ZeroResultOutcomeFailed || metrics.last.Reason != "session_ingest_error" {
			t.Fatalf("metrics = %#v", metrics.last)
		}
		if dedupe.releasedKey == "" || dedupe.releasedKey != dedupe.key {
			t.Fatalf("dedupe release = %q, marked = %q", dedupe.releasedKey, dedupe.key)
		}
	})
}

func TestZeroResultTrackerSkipsUntrackableEvent(t *testing.T) {
	metrics := &capturingZeroResultMetrics{}
	tracker := newZeroResultTrackerForTest(t, &fakeZeroResultDedupe{firstSeen: true}, &fakeSessionSink{}, metrics)

	tracker.trackNow(context.Background(), domain.ZeroResultSearchEvent{Query: domain.DefaultSearchQuery, AnonymousID: "anon_1", SessionID: "sess_1"})

	if metrics.last.Outcome != ZeroResultOutcomeSkipped || metrics.last.Reason != domain.ZeroResultSkipBrowseQuery {
		t.Fatalf("metrics = %#v", metrics.last)
	}
}

func newZeroResultTrackerForTest(t *testing.T, dedupe ZeroResultDedupeStore, sink SessionEventSink, metrics ZeroResultMetricsRecorder) *ZeroResultTracker {
	t.Helper()
	tracker, err := NewZeroResultTracker(dedupe, sink, ZeroResultTrackerOptions{
		Enabled:     true,
		DedupeTTL:   time.Minute,
		SendTimeout: time.Second,
		QueueSize:   1,
		WorkerCount: 1,
		Metrics:     metrics,
	}, slog.Default())
	if err != nil {
		t.Fatalf("tracker: %v", err)
	}
	t.Cleanup(func() {
		_ = tracker.Close(context.Background())
	})
	return tracker
}

func trackableZeroResultEvent() domain.ZeroResultSearchEvent {
	return domain.ZeroResultSearchEvent{
		Query:       "waterproof laptop bag",
		Filters:     map[string]string{domain.SearchFilterBrand: "Acme"},
		Sort:        domain.SortRelevance,
		Page:        1,
		PageSize:    20,
		RequestID:   "req_1",
		AnonymousID: "anon_1",
		SessionID:   "sess_1",
		Path:        "/search",
		OccurredAt:  time.Date(2026, 5, 24, 10, 30, 0, 0, time.UTC),
	}
}

type fakeZeroResultDedupe struct {
	key         string
	ttl         time.Duration
	firstSeen   bool
	err         error
	releasedKey string
	releaseErr  error
}

func (d *fakeZeroResultDedupe) Release(_ context.Context, key string) error {
	d.releasedKey = key
	return d.releaseErr
}

func (d *fakeZeroResultDedupe) MarkFirstSeen(_ context.Context, key string, ttl time.Duration) (bool, error) {
	d.key = key
	d.ttl = ttl
	return d.firstSeen, d.err
}

type fakeSessionSink struct {
	events []domain.ZeroResultSearchEvent
	err    error
}

func (s *fakeSessionSink) IngestSearchEvent(_ context.Context, event domain.ZeroResultSearchEvent) error {
	if s.err != nil {
		return s.err
	}
	s.events = append(s.events, event)
	return nil
}

type capturingZeroResultMetrics struct {
	last ZeroResultMetrics
}

func (m *capturingZeroResultMetrics) RecordZeroResult(_ context.Context, metrics ZeroResultMetrics) {
	m.last = metrics
}
