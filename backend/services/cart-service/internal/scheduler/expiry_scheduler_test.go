package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/usecase"
)

func TestExpirySchedulerRunOnceSkipsWhenLockNotAcquired(t *testing.T) {
	cleanup := &fakeCleanupUsecase{}
	locker := &fakeLocker{acquired: false}
	metrics := &fakeSchedulerMetrics{}
	scheduler := newTestExpiryScheduler(t, cleanup, locker, metrics)

	result, err := scheduler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
	if !result.LockSkipped || cleanup.calls != 0 || metrics.lockSkipped != 1 {
		t.Fatalf("result=%#v cleanupCalls=%d lockSkipped=%d, want skipped without cleanup", result, cleanup.calls, metrics.lockSkipped)
	}
}

func TestExpirySchedulerRunOnceRunsCleanupAndReleasesLock(t *testing.T) {
	cleanup := &fakeCleanupUsecase{report: usecase.CleanupExpiredCartsReport{ScannedCount: 2, ExpiredCount: 2}}
	locker := &fakeLocker{acquired: true}
	scheduler := newTestExpiryScheduler(t, cleanup, locker, &fakeSchedulerMetrics{})

	result, err := scheduler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
	if result.Report.ExpiredCount != 2 || cleanup.calls != 1 || !locker.handle.released {
		t.Fatalf("result=%#v cleanupCalls=%d released=%v, want cleanup and release", result, cleanup.calls, locker.handle.released)
	}
}

func TestExpirySchedulerRunOnceReturnsCleanupError(t *testing.T) {
	cleanupErr := errors.New("cleanup failed")
	cleanup := &fakeCleanupUsecase{err: cleanupErr}
	locker := &fakeLocker{acquired: true}
	scheduler := newTestExpiryScheduler(t, cleanup, locker, &fakeSchedulerMetrics{})

	_, err := scheduler.RunOnce(context.Background())
	if !errors.Is(err, cleanupErr) {
		t.Fatalf("RunOnce() error = %v, want cleanup error", err)
	}
	if !locker.handle.released {
		t.Fatal("lock was not released after cleanup error")
	}
}

func newTestExpiryScheduler(t *testing.T, cleanup *fakeCleanupUsecase, locker *fakeLocker, metrics *fakeSchedulerMetrics) *ExpiryScheduler {
	t.Helper()
	s, err := NewExpiryScheduler(ExpirySchedulerDependencies{
		Cleanup: cleanup,
		Locker:  locker,
		Metrics: metrics,
		Config: ExpirySchedulerConfig{
			Interval:   time.Minute,
			RunTimeout: time.Second,
			LockTTL:    time.Minute,
			LockKey:    "cart:lock:expiry-cleanup",
			WorkerID:   "worker_1",
		},
	})
	if err != nil {
		t.Fatalf("NewExpiryScheduler() error = %v", err)
	}
	return s
}

type fakeCleanupUsecase struct {
	calls  int
	report usecase.CleanupExpiredCartsReport
	err    error
}

func (f *fakeCleanupUsecase) Execute(ctx context.Context) (usecase.CleanupExpiredCartsReport, error) {
	f.calls++
	return f.report, f.err
}

type fakeLocker struct {
	acquired bool
	handle   *fakeLockHandle
}

func (f *fakeLocker) Acquire(ctx context.Context, key string, owner string, ttl time.Duration) (LockHandle, bool, error) {
	if !f.acquired {
		return nil, false, nil
	}
	f.handle = &fakeLockHandle{}
	return f.handle, true, nil
}

type fakeLockHandle struct {
	released bool
}

func (f *fakeLockHandle) Release(ctx context.Context) error {
	f.released = true
	return nil
}

type fakeSchedulerMetrics struct {
	lockSkipped int
}

func (f *fakeSchedulerMetrics) RecordLockSkipped(ctx context.Context) {
	f.lockSkipped++
}
