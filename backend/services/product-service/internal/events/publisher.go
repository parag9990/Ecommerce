package events

import (
	"context"
	"time"

	"product-service/internal/domain"
)

type ProductEventPublisher interface {
	Publish(ctx context.Context, topic string, envelope domain.EventEnvelope) error
}

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}
