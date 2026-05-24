package httptransport

import (
	"context"
	"sync"
	"time"
)

type AdminRateLimiter interface {
	Allow(ctx context.Context, key string) bool
}

type AllowAllAdminRateLimiter struct{}

func (AllowAllAdminRateLimiter) Allow(context.Context, string) bool {
	return true
}

type FixedWindowAdminRateLimiter struct {
	limit  int
	window time.Duration
	now    func() time.Time

	mu      sync.Mutex
	buckets map[string]fixedWindowBucket
}

type fixedWindowBucket struct {
	start time.Time
	count int
}

func NewFixedWindowAdminRateLimiter(limit int, window time.Duration) *FixedWindowAdminRateLimiter {
	if limit <= 0 {
		return nil
	}
	if window <= 0 {
		window = time.Minute
	}
	return &FixedWindowAdminRateLimiter{
		limit:   limit,
		window:  window,
		now:     time.Now,
		buckets: map[string]fixedWindowBucket{},
	}
}

func (l *FixedWindowAdminRateLimiter) Allow(_ context.Context, key string) bool {
	if l == nil {
		return true
	}
	if key == "" {
		key = "anonymous"
	}

	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pruneLocked(now)

	bucket := l.buckets[key]
	if bucket.start.IsZero() || now.Sub(bucket.start) >= l.window {
		l.buckets[key] = fixedWindowBucket{start: now, count: 1}
		return true
	}
	if bucket.count >= l.limit {
		return false
	}
	bucket.count++
	l.buckets[key] = bucket
	return true
}

func (l *FixedWindowAdminRateLimiter) pruneLocked(now time.Time) {
	for key, bucket := range l.buckets {
		if now.Sub(bucket.start) >= 2*l.window {
			delete(l.buckets, key)
		}
	}
}
