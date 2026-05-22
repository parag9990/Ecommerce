package worker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/usecase"
)

type RetentionCleanupUsecase interface {
	RunRetentionCleanup(ctx context.Context, input usecase.RunRetentionCleanupInput) (domain.RetentionCleanupResult, error)
}

type RetentionWorker struct {
	cleanup  RetentionCleanupUsecase
	interval time.Duration
	logger   *slog.Logger
}

func NewRetentionWorker(cleanup RetentionCleanupUsecase, interval time.Duration, logger *slog.Logger) *RetentionWorker {
	if interval <= 0 {
		interval = time.Hour
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &RetentionWorker{
		cleanup:  cleanup,
		interval: interval,
		logger:   logger,
	}
}

func (w *RetentionWorker) RunOnce(ctx context.Context) error {
	if w == nil || w.cleanup == nil {
		return nil
	}
	_, err := w.cleanup.RunRetentionCleanup(ctx, usecase.RunRetentionCleanupInput{})
	if err != nil && !errors.Is(err, context.Canceled) {
		w.logger.WarnContext(ctx, "session.retention.worker_run_failed", slog.String("error", err.Error()))
	}
	return err
}

func (w *RetentionWorker) Run(ctx context.Context) error {
	if w == nil || w.cleanup == nil {
		return nil
	}
	if err := w.RunOnce(ctx); err != nil && errors.Is(err, context.Canceled) {
		return err
	}
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.RunOnce(ctx); err != nil && errors.Is(err, context.Canceled) {
				return err
			}
		}
	}
}
