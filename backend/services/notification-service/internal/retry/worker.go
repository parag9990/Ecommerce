package retry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
)

var ErrAttemptBusy = errors.New("notification delivery attempt is already processing")

type AttemptSender interface {
	SendDeliveryAttempt(ctx context.Context, deliveryID string) (provider.Result, error)
}

type DeliveryStore interface {
	BeginAttempt(ctx context.Context, deliveryID string, idempotencyKey string, attempt int, leaseUntil time.Time, updatedAt time.Time) (domain.AttemptClaim, error)
	MarkAccepted(ctx context.Context, deliveryID string, attempt int, providerName, providerMessageID string, updatedAt time.Time) error
	MarkSuppressed(ctx context.Context, deliveryID string, attempt int, reason domain.ConsentReason, updatedAt time.Time) error
	MarkRetryScheduled(ctx context.Context, deliveryID string, attempt int, nextRetryAt time.Time, failureCode string, updatedAt time.Time) error
	MarkDeadLettered(ctx context.Context, deliveryID string, attempt int, failureCode string, deadLetteredAt time.Time) error
}

type AnalyticsRecorder interface {
	RecordSent(ctx context.Context, deliveryID, providerName, providerMessageID string, occurredAt time.Time) (domain.DeliveryEventApplyResult, error)
	RecordFailed(ctx context.Context, deliveryID, providerName, failureCode string, occurredAt time.Time) (domain.DeliveryEventApplyResult, error)
}

type Worker struct {
	policy       Policy
	store        DeliveryStore
	sender       AttemptSender
	publisher    Publisher
	analytics    AnalyticsRecorder
	logger       *slog.Logger
	attemptLease time.Duration
	now          func() time.Time
}

func NewWorker(
	policy Policy,
	store DeliveryStore,
	sender AttemptSender,
	publisher Publisher,
	analytics AnalyticsRecorder,
	attemptLease time.Duration,
	logger *slog.Logger,
) (*Worker, error) {
	if _, err := NewPolicy(policy.MaxAttempts, policy.Delays); err != nil {
		return nil, err
	}
	if absent(store) || absent(sender) || absent(publisher) || absent(analytics) {
		return nil, errors.New("notification retry worker dependencies are required")
	}
	if attemptLease <= 0 {
		return nil, errors.New("notification retry processing lease must be positive")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Worker{
		policy:       policy,
		store:        store,
		sender:       sender,
		publisher:    publisher,
		analytics:    analytics,
		logger:       logger,
		attemptLease: attemptLease,
		now:          time.Now,
	}, nil
}

func (w *Worker) WithClock(now func() time.Time) {
	if now != nil {
		w.now = now
	}
}

func (w *Worker) Process(ctx context.Context, job domain.RetryJob) error {
	if ctx == nil {
		return errors.New("notification retry worker context is required")
	}
	if err := w.policy.ValidateJob(job); err != nil {
		return w.PublishInvalid(ctx, job, "invalid")
	}
	now := w.now().UTC()
	claim, err := w.store.BeginAttempt(ctx, job.DeliveryID, job.IdempotencyKey, job.Attempt, now.Add(w.attemptLease), now)
	if err != nil {
		if errors.Is(err, domain.ErrRetryDeliveryReference) {
			return w.PublishInvalid(ctx, job, "invalid")
		}
		return fmt.Errorf("begin notification delivery attempt: %w", err)
	}
	switch claim {
	case domain.AttemptCompleted:
		w.logger.InfoContext(ctx, "notification.retry.duplicate_skipped",
			slog.String("delivery_id", job.DeliveryID),
			slog.Int("attempt", job.Attempt),
		)
		return nil
	case domain.AttemptBusy:
		return ErrAttemptBusy
	case domain.AttemptAcquired:
	default:
		return fmt.Errorf("unknown notification attempt claim %q", claim)
	}

	result, sendErr := w.sender.SendDeliveryAttempt(ctx, job.DeliveryID)
	var suppressed *domain.DeliverySuppressedError
	if errors.As(sendErr, &suppressed) {
		if err := w.store.MarkSuppressed(ctx, job.DeliveryID, job.Attempt, suppressed.Reason, w.now().UTC()); err != nil {
			return fmt.Errorf("mark notification delivery suppressed: %w", err)
		}
		w.logger.InfoContext(ctx, "notification.delivery.suppressed",
			slog.String("delivery_id", job.DeliveryID),
			slog.String("trace_id", job.TraceID),
			slog.String("reason", string(suppressed.Reason)),
			slog.String("purpose", string(suppressed.Purpose)),
		)
		return nil
	}
	if sendErr == nil && result.Status == provider.StatusAccepted {
		acceptedAt := w.now().UTC()
		if _, err := w.analytics.RecordSent(ctx, job.DeliveryID,
			strings.TrimSpace(result.ProviderName), strings.TrimSpace(result.ProviderMessageID), acceptedAt); err != nil {
			return fmt.Errorf("record sent notification milestone: %w", err)
		}
		if err := w.store.MarkAccepted(ctx, job.DeliveryID, job.Attempt,
			strings.TrimSpace(result.ProviderName), strings.TrimSpace(result.ProviderMessageID), w.now().UTC()); err != nil {
			return fmt.Errorf("mark notification delivery accepted: %w", err)
		}
		w.logger.InfoContext(ctx, "notification.delivery.accepted",
			slog.String("delivery_id", job.DeliveryID),
			slog.Int("attempt", job.Attempt),
			slog.String("provider", strings.TrimSpace(result.ProviderName)),
		)
		return nil
	}
	if errors.Is(sendErr, context.Canceled) && ctx.Err() != nil {
		return ctx.Err()
	}
	if sendErr == nil {
		sendErr = provider.ErrProviderUnavailable
	}

	if ClassifyError(sendErr) == FailureTransient {
		if delay, ok := w.policy.NextDelay(job.Attempt); ok {
			code := safeFailureCode(sendErr, false)
			job.Attempt++
			job.LastFailureCode = code
			job.QueuedAt = w.now().UTC()
			if err := w.publisher.PublishRetry(ctx, routeForDelay(delay), job); err != nil {
				return fmt.Errorf("publish notification retry: %w", err)
			}
			scheduledAt := w.now().UTC()
			if err := w.store.MarkRetryScheduled(ctx, job.DeliveryID, job.Attempt-1,
				scheduledAt.Add(delay), code, scheduledAt); err != nil {
				return fmt.Errorf("mark notification retry scheduled: %w", err)
			}
			w.logger.WarnContext(ctx, "notification.retry.scheduled",
				slog.String("delivery_id", job.DeliveryID),
				slog.String("trace_id", job.TraceID),
				slog.Int("attempt", job.Attempt-1),
				slog.Int("next_attempt", job.Attempt),
				slog.String("delay", delay.String()),
				slog.String("failure_code", code),
			)
			return nil
		}
	}

	code := safeFailureCode(sendErr, ClassifyError(sendErr) == FailureTransient)
	job.LastFailureCode = code
	failedAt := w.now().UTC()
	if _, err := w.analytics.RecordFailed(ctx, job.DeliveryID, strings.TrimSpace(result.ProviderName), code, failedAt); err != nil {
		return fmt.Errorf("record failed notification milestone: %w", err)
	}
	if err := w.publisher.PublishDeadLetter(ctx, "failed", job); err != nil {
		return fmt.Errorf("publish notification dead letter: %w", err)
	}
	if err := w.store.MarkDeadLettered(ctx, job.DeliveryID, job.Attempt, code, w.now().UTC()); err != nil {
		return fmt.Errorf("mark notification dead lettered: %w", err)
	}
	w.logger.ErrorContext(ctx, "notification.delivery.dead_lettered",
		slog.String("delivery_id", job.DeliveryID),
		slog.String("trace_id", job.TraceID),
		slog.Int("attempt", job.Attempt),
		slog.String("failure_code", code),
	)
	return nil
}

func (w *Worker) PublishInvalid(ctx context.Context, job domain.RetryJob, route string) error {
	job = sanitizedInvalidJob(job)
	if route == "security" {
		job.LastFailureCode = "security_payload_blocked"
	} else {
		job.LastFailureCode = "invalid_retry_job"
		route = "invalid"
	}
	job.QueuedAt = w.now().UTC()
	if err := w.publisher.PublishDeadLetter(ctx, route, job); err != nil {
		return fmt.Errorf("publish invalid notification dead letter: %w", err)
	}
	w.logger.WarnContext(ctx, "notification.retry.invalid_dead_lettered",
		slog.String("delivery_id", job.DeliveryID),
		slog.String("failure_code", job.LastFailureCode),
	)
	return nil
}

var safeMetadataID = regexp.MustCompile(`^[A-Za-z0-9:_-]{1,160}$`)

func sanitizedInvalidJob(job domain.RetryJob) domain.RetryJob {
	sanitized := domain.RetryJob{}
	if safeMetadataID.MatchString(job.DeliveryID) {
		sanitized.DeliveryID = job.DeliveryID
	}
	if safeMetadataID.MatchString(job.IdempotencyKey) {
		sanitized.IdempotencyKey = job.IdempotencyKey
	}
	if safeMetadataID.MatchString(job.TraceID) {
		sanitized.TraceID = job.TraceID
	}
	if job.Attempt >= 1 && job.Attempt <= 10 {
		sanitized.Attempt = job.Attempt
	}
	if job.MaxAttempts >= 1 && job.MaxAttempts <= 10 {
		sanitized.MaxAttempts = job.MaxAttempts
	}
	return sanitized
}

func absent(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
