package retry

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
)

type workerStore struct {
	claim             domain.AttemptClaim
	accepted          bool
	retryScheduled    bool
	deadLettered      bool
	suppressed        bool
	code              string
	suppressionReason domain.ConsentReason
	nextRetryAt       time.Time
}

func (s *workerStore) BeginAttempt(context.Context, string, string, int, time.Time, time.Time) (domain.AttemptClaim, error) {
	return s.claim, nil
}

func (s *workerStore) MarkAccepted(context.Context, string, int, string, string, time.Time) error {
	s.accepted = true
	return nil
}

func (s *workerStore) MarkSuppressed(_ context.Context, _ string, _ int, reason domain.ConsentReason, _ time.Time) error {
	s.suppressed, s.suppressionReason = true, reason
	return nil
}

func (s *workerStore) MarkRetryScheduled(_ context.Context, _ string, _ int, next time.Time, code string, _ time.Time) error {
	s.retryScheduled, s.nextRetryAt, s.code = true, next, code
	return nil
}

func (s *workerStore) MarkDeadLettered(_ context.Context, _ string, _ int, code string, _ time.Time) error {
	s.deadLettered, s.code = true, code
	return nil
}

type workerSender struct {
	result provider.Result
	err    error
	calls  int
}

func (s *workerSender) SendDeliveryAttempt(context.Context, string) (provider.Result, error) {
	s.calls++
	return s.result, s.err
}

type workerPublisher struct {
	retryRoute string
	retryJob   domain.RetryJob
	deadRoute  string
	deadJob    domain.RetryJob
}

type workerAnalytics struct {
	sent     bool
	failed   bool
	provider string
	code     string
}

func (a *workerAnalytics) RecordSent(context.Context, string, string, string, time.Time) (domain.DeliveryEventApplyResult, error) {
	a.sent = true
	return domain.DeliveryEventApplyResult{MilestoneChanged: true}, nil
}

func (a *workerAnalytics) RecordFailed(_ context.Context, _ string, providerName string, code string, _ time.Time) (domain.DeliveryEventApplyResult, error) {
	a.failed, a.provider, a.code = true, providerName, code
	return domain.DeliveryEventApplyResult{MilestoneChanged: true}, nil
}

func (*workerPublisher) PublishAttempt(context.Context, domain.RetryJob) error { return nil }
func (p *workerPublisher) PublishRetry(_ context.Context, route string, job domain.RetryJob) error {
	p.retryRoute, p.retryJob = route, job
	return nil
}
func (p *workerPublisher) PublishDeadLetter(_ context.Context, route string, job domain.RetryJob) error {
	p.deadRoute, p.deadJob = route, job
	return nil
}

func TestWorkerAcceptsSuccessfulAttempt(t *testing.T) {
	t.Parallel()

	store := &workerStore{claim: domain.AttemptAcquired}
	sender := &workerSender{result: provider.Result{Status: provider.StatusAccepted, ProviderName: "smtp", ProviderMessageID: "msg"}}
	publisher := &workerPublisher{}
	worker := newTestWorker(t, store, sender, publisher)
	if err := worker.Process(context.Background(), retryJob(1)); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if !store.accepted || publisher.retryRoute != "" || publisher.deadRoute != "" {
		t.Fatalf("outcome accepted=%t retry=%q dead=%q", store.accepted, publisher.retryRoute, publisher.deadRoute)
	}
	if metrics := worker.analytics.(*workerAnalytics); !metrics.sent || metrics.failed {
		t.Fatalf("analytics sent=%t failed=%t", metrics.sent, metrics.failed)
	}
}

func TestWorkerSchedulesRetryThenDeadLettersExhaustedFailure(t *testing.T) {
	t.Parallel()

	store := &workerStore{claim: domain.AttemptAcquired}
	publisher := &workerPublisher{}
	worker := newTestWorker(t, store, &workerSender{err: provider.ErrProviderUnavailable}, publisher)
	if err := worker.Process(context.Background(), retryJob(1)); err != nil {
		t.Fatalf("Process(attempt 1) error = %v", err)
	}
	if !store.retryScheduled || publisher.retryRoute != "after.30s" ||
		publisher.retryJob.Attempt != 2 || store.code != "provider_unavailable" {
		t.Fatalf("scheduled retry = %+v, route = %q, job = %+v", store, publisher.retryRoute, publisher.retryJob)
	}
	if metrics := worker.analytics.(*workerAnalytics); metrics.failed {
		t.Fatal("transient retry must not record final failed analytics")
	}

	store = &workerStore{claim: domain.AttemptAcquired}
	publisher = &workerPublisher{}
	worker = newTestWorker(t, store, &workerSender{err: provider.ErrProviderUnavailable}, publisher)
	if err := worker.Process(context.Background(), retryJob(4)); err != nil {
		t.Fatalf("Process(attempt 4) error = %v", err)
	}
	if !store.deadLettered || publisher.deadRoute != "failed" ||
		store.code != "provider_unavailable_exhausted" {
		t.Fatalf("dead letter = %+v, route = %q", store, publisher.deadRoute)
	}
	if metrics := worker.analytics.(*workerAnalytics); !metrics.failed || metrics.code != "provider_unavailable_exhausted" {
		t.Fatalf("terminal analytics = %+v", metrics)
	}
}

func TestWorkerCompletesSuppressedAttemptWithoutRetryOrDeadLetter(t *testing.T) {
	t.Parallel()

	store := &workerStore{claim: domain.AttemptAcquired}
	publisher := &workerPublisher{}
	sender := &workerSender{err: &domain.DeliverySuppressedError{
		Reason: domain.ConsentReasonMarketingOptedOut, Purpose: domain.PurposeMarketing,
	}}
	worker := newTestWorker(t, store, sender, publisher)
	if err := worker.Process(context.Background(), retryJob(1)); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if !store.suppressed || store.suppressionReason != domain.ConsentReasonMarketingOptedOut ||
		publisher.retryRoute != "" || publisher.deadRoute != "" {
		t.Fatalf("suppressed outcome = %+v, retry = %q, dead = %q", store, publisher.retryRoute, publisher.deadRoute)
	}
	if metrics := worker.analytics.(*workerAnalytics); metrics.failed || metrics.sent {
		t.Fatalf("suppressed analytics = %+v", metrics)
	}
}

func TestWorkerDeadLettersTerminalAndSkipsDuplicateAttempts(t *testing.T) {
	t.Parallel()

	store := &workerStore{claim: domain.AttemptAcquired}
	sender := &workerSender{result: provider.Result{ProviderName: "smtp"}, err: provider.ErrProviderRejected}
	publisher := &workerPublisher{}
	worker := newTestWorker(t, store, sender, publisher)
	if err := worker.Process(context.Background(), retryJob(1)); err != nil {
		t.Fatalf("Process(rejected) error = %v", err)
	}
	if !store.deadLettered || publisher.retryRoute != "" || store.code != "provider_rejected_recipient" {
		t.Fatalf("terminal outcome = %+v, retry route = %q", store, publisher.retryRoute)
	}
	if metrics := worker.analytics.(*workerAnalytics); metrics.provider != "smtp" {
		t.Fatalf("terminal provider analytics = %+v", metrics)
	}

	store = &workerStore{claim: domain.AttemptCompleted}
	sender = &workerSender{}
	worker = newTestWorker(t, store, sender, &workerPublisher{})
	if err := worker.Process(context.Background(), retryJob(1)); err != nil || sender.calls != 0 {
		t.Fatalf("Process(duplicate) error = %v, sends = %d", err, sender.calls)
	}
}

func TestWorkerPublishesInvalidPolicyJobWithoutProviderCall(t *testing.T) {
	t.Parallel()

	sender := &workerSender{}
	publisher := &workerPublisher{}
	worker := newTestWorker(t, &workerStore{claim: domain.AttemptAcquired}, sender, publisher)
	job := retryJob(1)
	job.MaxAttempts = 7
	err := worker.Process(context.Background(), job)
	if err != nil || sender.calls != 0 || publisher.deadRoute != "invalid" {
		t.Fatalf("Process(invalid) error = %v, sends = %d, route = %q", err, sender.calls, publisher.deadRoute)
	}
}

func newTestWorker(t *testing.T, store DeliveryStore, sender AttemptSender, publisher Publisher) *Worker {
	t.Helper()
	worker, err := NewWorker(DefaultPolicy(), store, sender, publisher, &workerAnalytics{}, time.Minute,
		slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewWorker() error = %v", err)
	}
	worker.WithClock(func() time.Time {
		return time.Date(2026, time.May, 27, 14, 0, 0, 0, time.UTC)
	})
	return worker
}

func retryJob(attempt int) domain.RetryJob {
	return domain.RetryJob{
		DeliveryID: "delivery_1", IdempotencyKey: "evt:template:email",
		Attempt: attempt, MaxAttempts: 4, QueuedAt: time.Now().UTC(),
	}
}
