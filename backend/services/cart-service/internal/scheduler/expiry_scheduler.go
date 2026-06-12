package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/usecase"
)

type CleanupUsecase interface {
	Execute(ctx context.Context) (usecase.CleanupExpiredCartsReport, error)
}

type ExpirySchedulerMetrics interface {
	RecordLockSkipped(ctx context.Context)
}

type NoopExpirySchedulerMetrics struct{}

func (NoopExpirySchedulerMetrics) RecordLockSkipped(context.Context) {}

type ExpiryScheduler struct {
	cleanup    CleanupUsecase
	locker     Locker
	logger     *slog.Logger
	metrics    ExpirySchedulerMetrics
	interval   time.Duration
	runTimeout time.Duration
	lockTTL    time.Duration
	lockKey    string
	workerID   string
}

type ExpirySchedulerConfig struct {
	Interval   time.Duration
	RunTimeout time.Duration
	LockTTL    time.Duration
	LockKey    string
	WorkerID   string
}

type ExpirySchedulerDependencies struct {
	Cleanup CleanupUsecase
	Locker  Locker
	Logger  *slog.Logger
	Metrics ExpirySchedulerMetrics
	Config  ExpirySchedulerConfig
}

type RunOnceResult struct {
	Report      usecase.CleanupExpiredCartsReport
	LockSkipped bool
}

func NewExpiryScheduler(deps ExpirySchedulerDependencies) (*ExpiryScheduler, error) {
	if deps.Cleanup == nil {
		return nil, errors.New("cleanup usecase is required")
	}
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	if deps.Metrics == nil {
		deps.Metrics = NoopExpirySchedulerMetrics{}
	}
	if deps.Config.Interval <= 0 {
		return nil, errors.New("cleanup interval must be greater than zero")
	}
	if deps.Config.RunTimeout <= 0 {
		return nil, errors.New("cleanup run timeout must be greater than zero")
	}
	if deps.Config.LockTTL <= 0 {
		return nil, errors.New("cleanup lock ttl must be greater than zero")
	}
	deps.Config.LockKey = strings.TrimSpace(deps.Config.LockKey)
	if deps.Config.LockKey == "" {
		return nil, errors.New("cleanup lock key is required")
	}
	deps.Config.WorkerID = strings.TrimSpace(deps.Config.WorkerID)
	if deps.Config.WorkerID == "" {
		deps.Config.WorkerID = defaultWorkerID()
	}
	return &ExpiryScheduler{
		cleanup:    deps.Cleanup,
		locker:     deps.Locker,
		logger:     deps.Logger,
		metrics:    deps.Metrics,
		interval:   deps.Config.Interval,
		runTimeout: deps.Config.RunTimeout,
		lockTTL:    deps.Config.LockTTL,
		lockKey:    deps.Config.LockKey,
		workerID:   deps.Config.WorkerID,
	}, nil
}

func (s *ExpiryScheduler) RunOnce(ctx context.Context) (RunOnceResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var lock LockHandle
	if s.locker != nil {
		handle, acquired, err := s.locker.Acquire(ctx, s.lockKey, s.workerID, s.lockTTL)
		if err != nil {
			return RunOnceResult{}, err
		}
		if !acquired {
			s.metrics.RecordLockSkipped(ctx)
			s.logger.Info("cart.expiry.lock_not_acquired", slog.String("worker_id", s.workerID), slog.String("lock_key", s.lockKey))
			return RunOnceResult{LockSkipped: true}, nil
		}
		lock = handle
		defer func() {
			releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := lock.Release(releaseCtx); err != nil {
				s.logger.Warn("cart.expiry.lock_release_failed", slog.String("worker_id", s.workerID), slog.String("error", err.Error()))
			}
		}()
	}

	runCtx, cancel := context.WithTimeout(ctx, s.runTimeout)
	defer cancel()
	report, err := s.cleanup.Execute(runCtx)
	if err != nil {
		return RunOnceResult{Report: report}, err
	}
	return RunOnceResult{Report: report}, nil
}

func (s *ExpiryScheduler) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if _, err := s.RunOnce(ctx); err != nil {
				s.logger.Error("cart.expiry.cleanup_failed", slog.String("error", err.Error()))
			}
		}
	}
}

func defaultWorkerID() string {
	hostname, err := os.Hostname()
	if err != nil || strings.TrimSpace(hostname) == "" {
		return "cart-expiry-worker"
	}
	return "cart-expiry-worker:" + strings.TrimSpace(hostname)
}
