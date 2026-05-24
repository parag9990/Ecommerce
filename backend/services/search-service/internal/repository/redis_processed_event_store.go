package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

type RedisProcessedEventStore struct {
	redis *RedisClient
	ttl   time.Duration
}

func NewRedisProcessedEventStore(redis *RedisClient, ttl time.Duration) (*RedisProcessedEventStore, error) {
	if redis == nil {
		return nil, fmt.Errorf("redis client is required")
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("processed event ttl must be greater than zero")
	}
	return &RedisProcessedEventStore{redis: redis, ttl: ttl}, nil
}

func (s *RedisProcessedEventStore) WasProcessed(ctx context.Context, eventID string) (bool, error) {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return false, fmt.Errorf("%w: event_id is required", domain.ErrInvalidProductEvent)
	}
	processed, err := s.redis.Exists(ctx, processedEventKey(eventID))
	if err != nil {
		return false, fmt.Errorf("%w: check event %q: %v", domain.ErrProcessedEventUnavailable, eventID, err)
	}
	return processed, nil
}

func (s *RedisProcessedEventStore) MarkProcessed(ctx context.Context, eventID string) error {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return fmt.Errorf("%w: event_id is required", domain.ErrInvalidProductEvent)
	}
	if _, err := s.redis.SetNX(ctx, processedEventKey(eventID), "1", s.ttl); err != nil {
		return fmt.Errorf("%w: mark event %q: %v", domain.ErrProcessedEventUnavailable, eventID, err)
	}
	return nil
}

func processedEventKey(eventID string) string {
	return "search:indexer:processed:" + eventID
}
