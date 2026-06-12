package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

const defaultCleanupExpiredCartsBatchSize int64 = 500

type CartExpiryRepository interface {
	FindExpiredActiveCarts(ctx context.Context, now time.Time, limit int64) ([]domain.ExpiredCartCandidate, error)
	MarkCartsExpired(ctx context.Context, cartIDs []string, now time.Time) (int64, error)
}

type CartExpiryCache interface {
	DeleteActiveCart(ctx context.Context, owner domain.CartOwner) error
	DeleteCartSummary(ctx context.Context, owner domain.CartOwner) error
}

type CartExpiryMetrics interface {
	RecordCleanupRun(ctx context.Context, report CleanupExpiredCartsReport, duration time.Duration, err error)
}

type NoopCartExpiryMetrics struct{}

func (NoopCartExpiryMetrics) RecordCleanupRun(context.Context, CleanupExpiredCartsReport, time.Duration, error) {
}

type CleanupExpiredCartsUsecase struct {
	repository CartExpiryRepository
	cache      CartExpiryCache
	clock      Clock
	logger     *slog.Logger
	metrics    CartExpiryMetrics
	batchSize  int64
}

type CleanupExpiredCartsDependencies struct {
	Repository CartExpiryRepository
	Cache      CartExpiryCache
	Clock      Clock
	Logger     *slog.Logger
	Metrics    CartExpiryMetrics
	BatchSize  int64
}

type CleanupExpiredCartsReport struct {
	ScannedCount        int   `json:"scanned_count"`
	ExpiredCount        int64 `json:"expired_count"`
	BatchCount          int   `json:"batch_count"`
	CacheDeleteFailures int   `json:"cache_delete_failures"`
	InvalidOwnerCount   int   `json:"invalid_owner_count"`
}

func NewCleanupExpiredCartsUsecase(deps CleanupExpiredCartsDependencies) (*CleanupExpiredCartsUsecase, error) {
	if deps.Repository == nil {
		return nil, errors.New("cart expiry repository is required")
	}
	if deps.Cache == nil {
		return nil, errors.New("cart expiry cache is required")
	}
	if deps.Clock == nil {
		deps.Clock = SystemClock{}
	}
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	if deps.Metrics == nil {
		deps.Metrics = NoopCartExpiryMetrics{}
	}
	if deps.BatchSize <= 0 {
		deps.BatchSize = defaultCleanupExpiredCartsBatchSize
	}
	return &CleanupExpiredCartsUsecase{
		repository: deps.Repository,
		cache:      deps.Cache,
		clock:      deps.Clock,
		logger:     deps.Logger,
		metrics:    deps.Metrics,
		batchSize:  deps.BatchSize,
	}, nil
}

func (u *CleanupExpiredCartsUsecase) Execute(ctx context.Context) (CleanupExpiredCartsReport, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	startedAt := u.clock.Now().UTC()
	report := CleanupExpiredCartsReport{}
	u.logger.Info(
		"cart.expiry.cleanup_started",
		slog.Time("now", startedAt),
		slog.Int64("batch_size", u.batchSize),
	)

	var runErr error
	defer func() {
		u.metrics.RecordCleanupRun(ctx, report, u.clock.Now().UTC().Sub(startedAt), runErr)
	}()

	for {
		select {
		case <-ctx.Done():
			runErr = ctx.Err()
			u.logger.Error(
				"cart.expiry.cleanup_failed",
				slog.String("error", runErr.Error()),
				slog.Int("total_scanned", report.ScannedCount),
				slog.Int64("total_expired", report.ExpiredCount),
			)
			return report, runErr
		default:
		}

		now := u.clock.Now().UTC()
		candidates, err := u.repository.FindExpiredActiveCarts(ctx, now, u.batchSize)
		if err != nil {
			runErr = err
			u.logger.Error("cart.expiry.cleanup_failed", slog.String("error", err.Error()))
			return report, err
		}
		if len(candidates) == 0 {
			break
		}

		cartIDs := make([]string, 0, len(candidates))
		for _, candidate := range candidates {
			cartIDs = append(cartIDs, candidate.CartID)
		}

		expiredCount, err := u.repository.MarkCartsExpired(ctx, cartIDs, now)
		if err != nil {
			runErr = err
			u.logger.Error("cart.expiry.cleanup_failed", slog.String("error", err.Error()))
			return report, err
		}

		report.BatchCount++
		report.ScannedCount += len(candidates)
		report.ExpiredCount += expiredCount
		cacheFailures, invalidOwners := u.invalidateCandidateCaches(ctx, candidates)
		report.CacheDeleteFailures += cacheFailures
		report.InvalidOwnerCount += invalidOwners

		u.logger.Info(
			"cart.expiry.batch_expired",
			slog.Int("scanned_count", len(candidates)),
			slog.Int64("expired_count", expiredCount),
			slog.Int("cache_delete_failures", report.CacheDeleteFailures),
		)

		if int64(len(candidates)) < u.batchSize {
			break
		}
	}

	u.logger.Info(
		"cart.expiry.cleanup_finished",
		slog.Int("total_scanned", report.ScannedCount),
		slog.Int64("total_expired", report.ExpiredCount),
		slog.Int("batch_count", report.BatchCount),
		slog.Int("cache_delete_failures", report.CacheDeleteFailures),
		slog.Int("invalid_owner_count", report.InvalidOwnerCount),
	)
	return report, nil
}

func (u *CleanupExpiredCartsUsecase) invalidateCandidateCaches(ctx context.Context, candidates []domain.ExpiredCartCandidate) (int, int) {
	failures := 0
	invalidOwners := 0
	for _, candidate := range candidates {
		owner, ok := candidate.Owner()
		if !ok {
			invalidOwners++
			u.logger.Warn("cart.expiry.candidate_owner_missing", slog.String("cart_id", candidate.CartID))
			continue
		}
		failures += invalidateExpiredCartCache(ctx, u.cache, u.logger, owner, candidate.CartID, "cleanup_job")
	}
	return failures, invalidOwners
}
