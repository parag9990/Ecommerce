package ratelimit

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidRequest = errors.New("invalid rate limit request")

type Request struct {
	Key    string
	Limit  int64
	Window time.Duration
	Cost   int64
}

type Result struct {
	Allowed    bool
	Limit      int64
	Remaining  int64
	RetryAfter time.Duration
	ResetAfter time.Duration
}

type Limiter interface {
	Allow(ctx context.Context, req Request) (Result, error)
}
