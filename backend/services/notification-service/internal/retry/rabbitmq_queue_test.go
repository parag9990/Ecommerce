package retry

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

func TestDecodeRetryJobRejectsContentAndAcceptsMetadataOnlyJob(t *testing.T) {
	t.Parallel()

	valid := domain.RetryJob{
		DeliveryID: "delivery_1", IdempotencyKey: "event:template:email",
		Attempt: 1, MaxAttempts: 4, QueuedAt: time.Now().UTC(),
	}
	body, _ := json.Marshal(valid)
	if _, route, err := decodeRetryJob(body); err != nil || route != "" {
		t.Fatalf("decodeRetryJob(valid) route = %q, error = %v", route, err)
	}
	body = []byte(`{"delivery_id":"delivery_1","idempotency_key":"key","attempt":1,"max_attempts":4,"queued_at":"2026-05-27T00:00:00Z","body":"forbidden"}`)
	_, route, err := decodeRetryJob(body)
	if !errors.Is(err, domain.ErrInvalidRetryJob) || route != "security" {
		t.Fatalf("decodeRetryJob(content) route = %q, error = %v", route, err)
	}
}
