package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

func TestCleanupExpiredCartsReturnsZeroWhenNoCandidates(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	repo := &fakeExpiryRepository{}
	cache := &fakeExpiryCache{}
	uc := newTestCleanupExpiredCartsUsecase(t, repo, cache, now)

	report, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if report.ScannedCount != 0 || report.ExpiredCount != 0 || repo.markCalls != 0 {
		t.Fatalf("report = %#v markCalls=%d, want no work", report, repo.markCalls)
	}
	if cache.activeDeletes != 0 || cache.summaryDeletes != 0 {
		t.Fatalf("cache deletes active=%d summary=%d, want 0", cache.activeDeletes, cache.summaryDeletes)
	}
}

func TestCleanupExpiredCartsMarksExpiredAndInvalidatesOwners(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	guestSessionID := "sess_123"
	repo := &fakeExpiryRepository{
		batches: [][]domain.ExpiredCartCandidate{
			{
				{CartID: "cart_user", UserID: &userID, Version: 3, ExpiresAt: now.Add(-time.Second)},
				{CartID: "cart_guest", GuestSessionID: &guestSessionID, Version: 2, ExpiresAt: now.Add(-time.Second)},
			},
		},
		markCount: 2,
	}
	cache := &fakeExpiryCache{}
	uc := newTestCleanupExpiredCartsUsecase(t, repo, cache, now)

	report, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if report.ScannedCount != 2 || report.ExpiredCount != 2 || report.BatchCount != 1 {
		t.Fatalf("report = %#v, want one expired batch", report)
	}
	if repo.markCalls != 1 || len(repo.markedIDs) != 2 {
		t.Fatalf("markCalls=%d markedIDs=%v, want two ids marked once", repo.markCalls, repo.markedIDs)
	}
	if cache.activeDeletes != 2 || cache.summaryDeletes != 2 {
		t.Fatalf("cache deletes active=%d summary=%d, want 2/2", cache.activeDeletes, cache.summaryDeletes)
	}
}

func TestCleanupExpiredCartsReportsCacheDeleteFailuresWithoutFailing(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	repo := &fakeExpiryRepository{
		batches: [][]domain.ExpiredCartCandidate{
			{{CartID: "cart_user", UserID: &userID, Version: 3, ExpiresAt: now.Add(-time.Second)}},
		},
		markCount: 1,
	}
	cache := &fakeExpiryCache{err: domain.ErrCacheUnavailable}
	uc := newTestCleanupExpiredCartsUsecase(t, repo, cache, now)

	report, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if report.CacheDeleteFailures != 2 {
		t.Fatalf("CacheDeleteFailures = %d, want 2", report.CacheDeleteFailures)
	}
}

func TestCleanupExpiredCartsReturnsRepositoryError(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	repo := &fakeExpiryRepository{
		batches: [][]domain.ExpiredCartCandidate{
			{{CartID: "cart_user", UserID: &userID, Version: 3, ExpiresAt: now.Add(-time.Second)}},
		},
		markErr: domain.ErrCartStoreUnavailable,
	}
	uc := newTestCleanupExpiredCartsUsecase(t, repo, &fakeExpiryCache{}, now)

	_, err := uc.Execute(context.Background())
	if !errors.Is(err, domain.ErrCartStoreUnavailable) {
		t.Fatalf("Execute() error = %v, want ErrCartStoreUnavailable", err)
	}
}

func newTestCleanupExpiredCartsUsecase(t *testing.T, repo *fakeExpiryRepository, cache *fakeExpiryCache, now time.Time) *CleanupExpiredCartsUsecase {
	t.Helper()
	uc, err := NewCleanupExpiredCartsUsecase(CleanupExpiredCartsDependencies{
		Repository: repo,
		Cache:      cache,
		Clock:      fakeClock{now: now},
		BatchSize:  10,
	})
	if err != nil {
		t.Fatalf("NewCleanupExpiredCartsUsecase() error = %v", err)
	}
	return uc
}

type fakeExpiryRepository struct {
	batches   [][]domain.ExpiredCartCandidate
	findCalls int
	markCalls int
	markedIDs []string
	markCount int64
	findErr   error
	markErr   error
}

func (f *fakeExpiryRepository) FindExpiredActiveCarts(ctx context.Context, now time.Time, limit int64) ([]domain.ExpiredCartCandidate, error) {
	f.findCalls++
	if f.findErr != nil {
		return nil, f.findErr
	}
	if len(f.batches) == 0 {
		return nil, nil
	}
	batch := f.batches[0]
	f.batches = f.batches[1:]
	return append([]domain.ExpiredCartCandidate(nil), batch...), nil
}

func (f *fakeExpiryRepository) MarkCartsExpired(ctx context.Context, cartIDs []string, now time.Time) (int64, error) {
	f.markCalls++
	if f.markErr != nil {
		return 0, f.markErr
	}
	f.markedIDs = append(f.markedIDs, cartIDs...)
	if f.markCount > 0 {
		return f.markCount, nil
	}
	return int64(len(cartIDs)), nil
}

type fakeExpiryCache struct {
	err            error
	activeDeletes  int
	summaryDeletes int
}

func (f *fakeExpiryCache) DeleteActiveCart(ctx context.Context, owner domain.CartOwner) error {
	f.activeDeletes++
	return f.err
}

func (f *fakeExpiryCache) DeleteCartSummary(ctx context.Context, owner domain.CartOwner) error {
	f.summaryDeletes++
	return f.err
}
