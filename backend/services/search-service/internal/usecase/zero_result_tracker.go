package usecase

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/requestctx"
)

const (
	defaultZeroResultDedupeTTL   = 30 * time.Minute
	defaultZeroResultSendTimeout = 150 * time.Millisecond
	defaultZeroResultQueueSize   = 1024
	defaultZeroResultWorkerCount = 2
)

const (
	ZeroResultOutcomeTracked   = "tracked"
	ZeroResultOutcomeSkipped   = "skipped"
	ZeroResultOutcomeDuplicate = "duplicate"
	ZeroResultOutcomeFailed    = "failed"
)

type ZeroResultTrackerOptions struct {
	Enabled     bool
	DedupeTTL   time.Duration
	SendTimeout time.Duration
	QueueSize   int
	WorkerCount int
	Metrics     ZeroResultMetricsRecorder
}

type ZeroResultTracker struct {
	dedupe      ZeroResultDedupeStore
	sink        SessionEventSink
	logger      *slog.Logger
	metrics     ZeroResultMetricsRecorder
	enabled     bool
	dedupeTTL   time.Duration
	sendTimeout time.Duration
	jobs        chan domain.ZeroResultSearchEvent
	stop        chan struct{}
	wg          sync.WaitGroup
	stopped     atomic.Bool
}

func NewZeroResultTracker(dedupe ZeroResultDedupeStore, sink SessionEventSink, options ZeroResultTrackerOptions, logger *slog.Logger) (*ZeroResultTracker, error) {
	if logger == nil {
		logger = slog.Default()
	}
	if options.Metrics == nil {
		options.Metrics = NopZeroResultMetricsRecorder{}
	}
	if !options.Enabled {
		return &ZeroResultTracker{
			logger:  logger,
			metrics: options.Metrics,
			enabled: false,
		}, nil
	}
	if dedupe == nil {
		return nil, errors.New("zero-result dedupe store is required")
	}
	if sink == nil {
		return nil, errors.New("session event sink is required")
	}
	if options.DedupeTTL <= 0 {
		options.DedupeTTL = defaultZeroResultDedupeTTL
	}
	if options.SendTimeout <= 0 {
		options.SendTimeout = defaultZeroResultSendTimeout
	}
	if options.QueueSize <= 0 {
		options.QueueSize = defaultZeroResultQueueSize
	}
	if options.WorkerCount <= 0 {
		options.WorkerCount = defaultZeroResultWorkerCount
	}

	tracker := &ZeroResultTracker{
		dedupe:      dedupe,
		sink:        sink,
		logger:      logger,
		metrics:     options.Metrics,
		enabled:     true,
		dedupeTTL:   options.DedupeTTL,
		sendTimeout: options.SendTimeout,
		jobs:        make(chan domain.ZeroResultSearchEvent, options.QueueSize),
		stop:        make(chan struct{}),
	}
	for i := 0; i < options.WorkerCount; i++ {
		tracker.wg.Add(1)
		go tracker.worker()
	}
	return tracker, nil
}

func (t *ZeroResultTracker) Track(ctx context.Context, event domain.ZeroResultSearchEvent) {
	if t == nil || !t.enabled {
		return
	}
	event = t.prepareEvent(ctx, event)
	if reason := event.SkipReason(); reason != "" {
		t.logSkip(ctx, event, reason)
		t.record(ctx, ZeroResultOutcomeSkipped, reason, 0)
		return
	}
	if t.stopped.Load() {
		t.logSkip(ctx, event, "tracker_stopped")
		t.record(ctx, ZeroResultOutcomeSkipped, "tracker_stopped", 0)
		return
	}

	select {
	case t.jobs <- event:
	default:
		t.logger.WarnContext(ctx, "search.zero_result.queue_full",
			slog.String("request_id", event.RequestID),
			slog.String("query_hash", event.QueryHash()),
			slog.String("session_id_hash", event.SessionIDHash()),
		)
		t.record(ctx, ZeroResultOutcomeSkipped, "queue_full", 0)
	}
}

func (t *ZeroResultTracker) Close(ctx context.Context) error {
	if t == nil || !t.enabled {
		return nil
	}
	if !t.stopped.CompareAndSwap(false, true) {
		return nil
	}
	close(t.stop)

	done := make(chan struct{})
	go func() {
		t.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *ZeroResultTracker) worker() {
	defer t.wg.Done()
	for {
		select {
		case event := <-t.jobs:
			t.trackNow(context.Background(), event)
		case <-t.stop:
			return
		}
	}
}

func (t *ZeroResultTracker) trackNow(ctx context.Context, event domain.ZeroResultSearchEvent) {
	if t == nil || !t.enabled {
		return
	}
	event = event.Normalize()
	if reason := event.SkipReason(); reason != "" {
		t.logSkip(ctx, event, reason)
		t.record(ctx, ZeroResultOutcomeSkipped, reason, 0)
		return
	}

	started := time.Now()
	sendCtx, cancel := context.WithTimeout(ctx, t.sendTimeout)
	defer cancel()

	dedupeKey := event.DedupeKey()
	firstSeen, err := t.dedupe.MarkFirstSeen(sendCtx, dedupeKey, t.dedupeTTL)
	if err != nil {
		t.logger.WarnContext(ctx, "search.zero_result.dedupe_failed",
			slog.String("request_id", event.RequestID),
			slog.String("query_hash", event.QueryHash()),
			slog.String("session_id_hash", event.SessionIDHash()),
			slog.String("error", err.Error()),
		)
		t.record(ctx, ZeroResultOutcomeFailed, "dedupe_error", elapsedMS(started))
		return
	}
	if !firstSeen {
		t.logger.DebugContext(ctx, "search.zero_result.duplicate_skipped",
			slog.String("request_id", event.RequestID),
			slog.String("query_hash", event.QueryHash()),
			slog.String("session_id_hash", event.SessionIDHash()),
		)
		t.record(ctx, ZeroResultOutcomeDuplicate, "duplicate", elapsedMS(started))
		return
	}

	if err := t.sink.IngestSearchEvent(sendCtx, event); err != nil {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), t.sendTimeout)
		cleanupErr := t.dedupe.Release(cleanupCtx, dedupeKey)
		cleanupCancel()
		t.logger.WarnContext(ctx, "search.zero_result.ingest_failed",
			slog.String("request_id", event.RequestID),
			slog.String("query_hash", event.QueryHash()),
			slog.String("session_id_hash", event.SessionIDHash()),
			slog.String("error", err.Error()),
			slog.Any("dedupe_release_error", cleanupErr),
		)
		t.record(ctx, ZeroResultOutcomeFailed, "session_ingest_error", elapsedMS(started))
		return
	}

	t.logger.InfoContext(ctx, "search.zero_result.tracked",
		slog.String("request_id", event.RequestID),
		slog.String("query_hash", event.QueryHash()),
		slog.String("session_id_hash", event.SessionIDHash()),
		slog.Int("query_length", event.QueryLength()),
		slog.Int("latency_ms", elapsedMS(started)),
	)
	t.record(ctx, ZeroResultOutcomeTracked, "", elapsedMS(started))
}

func (t *ZeroResultTracker) prepareEvent(ctx context.Context, event domain.ZeroResultSearchEvent) domain.ZeroResultSearchEvent {
	if event.RequestID == "" {
		if analytics := requestctx.Analytics(ctx); analytics.RequestID != "" {
			event.RequestID = analytics.RequestID
		} else {
			event.RequestID = requestctx.RequestID(ctx)
		}
	}
	return event.Normalize()
}

func (t *ZeroResultTracker) logSkip(ctx context.Context, event domain.ZeroResultSearchEvent, reason string) {
	t.logger.DebugContext(ctx, "search.zero_result.skipped",
		slog.String("request_id", event.RequestID),
		slog.String("reason", reason),
		slog.String("query_hash", event.QueryHash()),
		slog.String("session_id_hash", event.SessionIDHash()),
	)
}

func (t *ZeroResultTracker) record(ctx context.Context, outcome string, reason string, durationMS int) {
	if t == nil || t.metrics == nil {
		return
	}
	t.metrics.RecordZeroResult(ctx, ZeroResultMetrics{
		Outcome:    outcome,
		Reason:     reason,
		DurationMS: durationMS,
	})
}

type NopZeroResultTracker struct{}

func (NopZeroResultTracker) Track(context.Context, domain.ZeroResultSearchEvent) {}
