package domain

import (
	"fmt"
	"strings"
	"time"
)

// RetryJob deliberately contains no recipient or rendered notification content.
type RetryJob struct {
	DeliveryID      string    `json:"delivery_id"`
	IdempotencyKey  string    `json:"idempotency_key"`
	Attempt         int       `json:"attempt"`
	MaxAttempts     int       `json:"max_attempts"`
	LastFailureCode string    `json:"last_failure_code,omitempty"`
	TraceID         string    `json:"trace_id,omitempty"`
	QueuedAt        time.Time `json:"queued_at"`
}

func (j RetryJob) Validate() error {
	if strings.TrimSpace(j.DeliveryID) == "" || strings.TrimSpace(j.IdempotencyKey) == "" {
		return fmt.Errorf("%w: identity is required", ErrInvalidRetryJob)
	}
	if j.Attempt < 1 || j.MaxAttempts < 1 || j.Attempt > j.MaxAttempts {
		return fmt.Errorf("%w: invalid attempt bounds", ErrInvalidRetryJob)
	}
	if j.QueuedAt.IsZero() {
		return fmt.Errorf("%w: queued_at is required", ErrInvalidRetryJob)
	}
	return nil
}

type AttemptClaim string

const (
	AttemptAcquired  AttemptClaim = "acquired"
	AttemptCompleted AttemptClaim = "completed"
	AttemptBusy      AttemptClaim = "busy"
)
