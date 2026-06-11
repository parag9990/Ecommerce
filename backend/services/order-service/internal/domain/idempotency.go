package domain

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type IdempotencyStatus string

const (
	IdempotencyStatusProcessing IdempotencyStatus = "processing"
	IdempotencyStatusCompleted  IdempotencyStatus = "completed"
	IdempotencyStatusFailed     IdempotencyStatus = "failed"
)

type ClaimDecision string

const (
	ClaimAcquired   ClaimDecision = "acquired"
	ClaimReplay     ClaimDecision = "replay"
	ClaimInProgress ClaimDecision = "in_progress"
	ClaimResume     ClaimDecision = "resume"
	ClaimFailed     ClaimDecision = "failed"
)

type IdempotencyClaim struct {
	UserID      string
	Key         string
	RequestHash string
	OrderID     string
	Status      IdempotencyStatus
	Decision    ClaimDecision
	ExpiresAt   time.Time
}

func ParseIdempotencyStatus(value string) (IdempotencyStatus, error) {
	status := IdempotencyStatus(value)
	switch status {
	case IdempotencyStatusProcessing, IdempotencyStatusCompleted, IdempotencyStatusFailed:
		return status, nil
	default:
		return "", fmt.Errorf("%w: unknown idempotency status %q", ErrTemporarilyUnavailable, value)
	}
}

func ValidateIdempotencyKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrIdempotencyKeyRequired
	}
	if len(key) > 128 {
		return ErrIdempotencyKeyInvalid
	}
	return nil
}

func (c IdempotencyClaim) ValidateForOrderBinding(userID string, key string) error {
	userID = strings.TrimSpace(userID)
	key = strings.TrimSpace(key)
	if c.Decision != ClaimAcquired || c.Status != IdempotencyStatusProcessing ||
		strings.TrimSpace(c.OrderID) != "" || userID == "" || len(userID) > 64 ||
		c.UserID != userID || c.Key != key || c.ExpiresAt.IsZero() ||
		ValidateIdempotencyKey(c.Key) != nil {
		return ErrInvalidCheckoutCommand
	}
	decoded, err := hex.DecodeString(c.RequestHash)
	if err != nil || len(decoded) != 32 {
		return ErrInvalidCheckoutCommand
	}
	return nil
}
