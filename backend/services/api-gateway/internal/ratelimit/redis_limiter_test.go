package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRedisLimiterValidatesRequest(t *testing.T) {
	limiter := NewRedisLimiter(nil)

	_, err := limiter.Allow(context.Background(), Request{
		Key:    "",
		Limit:  1,
		Window: time.Minute,
		Cost:   1,
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid request error, got %v", err)
	}
}
