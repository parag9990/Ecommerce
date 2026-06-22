package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	defaultActiveSessionKeyPrefix = "session"
	defaultActiveCounterTTL       = 48 * time.Hour
)

type ActiveSessionStoreConfig struct {
	KeyPrefix  string
	CounterTTL time.Duration
}

type RedisActiveSessionStore struct {
	client     *redis.Client
	keyPrefix  string
	counterTTL time.Duration
	logger     *slog.Logger
}

func NewRedisActiveSessionStore(client *redis.Client, cfg ActiveSessionStoreConfig, logger *slog.Logger) (*RedisActiveSessionStore, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &RedisActiveSessionStore{
		client:     client,
		keyPrefix:  cfg.KeyPrefix,
		counterTTL: cfg.CounterTTL,
		logger:     logger,
	}, nil
}

func (s *RedisActiveSessionStore) Touch(ctx context.Context, snapshot domain.ActiveSessionSnapshot, ttl time.Duration) error {
	if ttl <= 0 {
		return errors.New("active session ttl must be greater than zero")
	}
	normalized := snapshot.Normalize()
	if err := normalized.Validate(); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInvalidActiveSession, err)
	}

	fields := s.snapshotFields(normalized)
	pipe := s.client.Pipeline()
	activeKey := s.activeKey(normalized.SessionID)
	lastSeenKey := s.lastSeenKey(normalized.SessionID)
	pipe.HSet(ctx, activeKey, fields)
	pipe.Expire(ctx, activeKey, ttl)
	pipe.Set(ctx, lastSeenKey, normalized.LastSeenAt.Format(time.RFC3339Nano), ttl)

	anonymousKey := s.anonymousKey(normalized.AnonymousID)
	pipe.SAdd(ctx, anonymousKey, normalized.SessionID)
	pipe.Expire(ctx, anonymousKey, ttl)
	if normalized.UserID != nil {
		userKey := s.userKey(*normalized.UserID)
		pipe.SAdd(ctx, userKey, normalized.SessionID)
		pipe.Expire(ctx, userKey, ttl)
	}

	counterKey := s.activeCounterKey(normalized.LastSeenAt)
	eventCounterKey := s.eventCounterKey(normalized.LastSeenAt)
	activeSessionsKey := s.activeSessionsKey()
	activeUsersKey := s.activeUsersKey()
	activeScore := float64(normalized.LastSeenAt.Unix())
	pipe.Incr(ctx, counterKey)
	pipe.Expire(ctx, counterKey, s.counterTTL)
	pipe.Incr(ctx, eventCounterKey)
	pipe.Expire(ctx, eventCounterKey, s.counterTTL)
	pipe.ZAdd(ctx, activeSessionsKey, redis.Z{Score: activeScore, Member: normalized.SessionID})
	pipe.Expire(ctx, activeSessionsKey, s.counterTTL)
	if normalized.UserID != nil {
		pipe.ZAdd(ctx, activeUsersKey, redis.Z{Score: activeScore, Member: *normalized.UserID})
		pipe.Expire(ctx, activeUsersKey, s.counterTTL)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("touch active session: %w", err)
	}
	s.logger.DebugContext(ctx, "session.redis.active_touched",
		slog.String("session_id", normalized.SessionID),
		slog.Duration("ttl", ttl),
	)
	return nil
}

func (s *RedisActiveSessionStore) Get(ctx context.Context, sessionID string) (domain.ActiveSessionSnapshot, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return domain.ActiveSessionSnapshot{}, fmt.Errorf("%w: session_id is required", domain.ErrInvalidActiveSession)
	}
	values, err := s.client.HGetAll(ctx, s.activeKey(sessionID)).Result()
	if err != nil {
		return domain.ActiveSessionSnapshot{}, fmt.Errorf("get active session: %w", err)
	}
	if len(values) == 0 {
		return domain.ActiveSessionSnapshot{}, domain.ErrSessionNotFound
	}
	snapshot, err := activeSnapshotFromRedis(values)
	if err != nil {
		return domain.ActiveSessionSnapshot{}, err
	}
	return snapshot, nil
}

func (s *RedisActiveSessionStore) Delete(ctx context.Context, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return fmt.Errorf("%w: session_id is required", domain.ErrInvalidActiveSession)
	}

	snapshot, err := s.Get(ctx, sessionID)
	if err != nil && !errors.Is(err, domain.ErrSessionNotFound) {
		return err
	}

	pipe := s.client.Pipeline()
	pipe.Del(ctx, s.activeKey(sessionID), s.lastSeenKey(sessionID))
	pipe.ZRem(ctx, s.activeSessionsKey(), sessionID)
	if err == nil {
		pipe.SRem(ctx, s.anonymousKey(snapshot.AnonymousID), sessionID)
		if snapshot.UserID != nil {
			pipe.SRem(ctx, s.userKey(*snapshot.UserID), sessionID)
			pipe.ZRem(ctx, s.activeUsersKey(), *snapshot.UserID)
		}
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("delete active session: %w", err)
	}
	return nil
}

func (s *RedisActiveSessionStore) PreviewDeletion(ctx context.Context, target domain.DeletionTarget) (int64, error) {
	ids, err := s.deletionSessionIDs(ctx, target)
	if err != nil {
		return 0, err
	}
	return int64(len(ids)), nil
}

func (s *RedisActiveSessionStore) DeleteMatching(ctx context.Context, target domain.DeletionTarget) (int64, error) {
	ids, err := s.deletionSessionIDs(ctx, target)
	if err != nil {
		return 0, err
	}
	var deleted int64
	for _, sessionID := range ids {
		if err := s.Delete(ctx, sessionID); err != nil {
			return deleted, err
		}
		deleted++
	}
	switch target.Type {
	case domain.DeletionTargetUserID:
		if err := s.client.Del(ctx, s.userKey(target.Value)).Err(); err != nil {
			return deleted, fmt.Errorf("delete active user index: %w", err)
		}
	case domain.DeletionTargetAnonymousID:
		if err := s.client.Del(ctx, s.anonymousKey(target.Value)).Err(); err != nil {
			return deleted, fmt.Errorf("delete active anonymous index: %w", err)
		}
	}
	return deleted, nil
}

func (s *RedisActiveSessionStore) ApplyRetentionTTL(ctx context.Context, ttl time.Duration) error {
	if ttl <= 0 {
		return errors.New("active session retention ttl must be greater than zero")
	}
	patterns := []string{s.keyPrefix + ":active:*", s.keyPrefix + ":last_seen:*", s.keyPrefix + ":user:*", s.keyPrefix + ":anon:*"}
	for _, pattern := range patterns {
		var cursor uint64
		for {
			keys, next, err := s.client.Scan(ctx, cursor, pattern, 250).Result()
			if err != nil {
				return fmt.Errorf("scan active session retention keys: %w", err)
			}
			if len(keys) > 0 {
				pipe := s.client.Pipeline()
				for _, key := range keys {
					pipe.Expire(ctx, key, ttl)
				}
				if _, err := pipe.Exec(ctx); err != nil {
					return fmt.Errorf("apply active session retention ttl: %w", err)
				}
			}
			cursor = next
			if cursor == 0 {
				break
			}
		}
	}
	return nil
}

func (s *RedisActiveSessionStore) deletionSessionIDs(ctx context.Context, target domain.DeletionTarget) ([]string, error) {
	switch target.Type {
	case domain.DeletionTargetUserID:
		return s.ListByUser(ctx, target.Value)
	case domain.DeletionTargetAnonymousID:
		return s.ListByAnonymousID(ctx, target.Value)
	case domain.DeletionTargetSessionID:
		if _, err := s.Get(ctx, target.Value); errors.Is(err, domain.ErrSessionNotFound) {
			return []string{}, nil
		} else if err != nil {
			return nil, err
		}
		return []string{target.Value}, nil
	default:
		return nil, fmt.Errorf("%w: unsupported deletion target type", domain.ErrInvalidInput)
	}
}

func (s *RedisActiveSessionStore) ListByUser(ctx context.Context, userID string) ([]string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", domain.ErrInvalidActiveSession)
	}
	ids, err := s.client.SMembers(ctx, s.userKey(userID)).Result()
	if err != nil {
		return nil, fmt.Errorf("list active sessions by user: %w", err)
	}
	return ids, nil
}

func (s *RedisActiveSessionStore) ListByAnonymousID(ctx context.Context, anonymousID string) ([]string, error) {
	anonymousID = strings.TrimSpace(anonymousID)
	if anonymousID == "" {
		return nil, fmt.Errorf("%w: anonymous_id is required", domain.ErrInvalidActiveSession)
	}
	ids, err := s.client.SMembers(ctx, s.anonymousKey(anonymousID)).Result()
	if err != nil {
		return nil, fmt.Errorf("list active sessions by anonymous id: %w", err)
	}
	return ids, nil
}

func (s *RedisActiveSessionStore) snapshotFields(snapshot domain.ActiveSessionSnapshot) map[string]any {
	fields := map[string]any{
		"session_id":   snapshot.SessionID,
		"anonymous_id": snapshot.AnonymousID,
		"status":       string(snapshot.Status),
		"entry_page":   snapshot.EntryPage,
		"device_type":  string(snapshot.DeviceType),
		"channel":      string(snapshot.Channel),
		"ip_hash":      snapshot.IPHash,
		"started_at":   snapshot.StartedAt.Format(time.RFC3339Nano),
		"last_seen_at": snapshot.LastSeenAt.Format(time.RFC3339Nano),
	}
	if snapshot.UserID != nil {
		fields["user_id"] = *snapshot.UserID
	}
	if snapshot.CurrentPage != nil {
		fields["current_page"] = *snapshot.CurrentPage
	}
	if snapshot.Browser != nil {
		fields["browser"] = *snapshot.Browser
	}
	if snapshot.OS != nil {
		fields["os"] = *snapshot.OS
	}
	if snapshot.Country != nil {
		fields["country"] = *snapshot.Country
	}
	if snapshot.City != nil {
		fields["city"] = *snapshot.City
	}
	if snapshot.LastEventType != nil {
		fields["last_event_type"] = string(*snapshot.LastEventType)
	}
	return fields
}

func activeSnapshotFromRedis(values map[string]string) (domain.ActiveSessionSnapshot, error) {
	startedAt, err := parseRedisTime(values["started_at"])
	if err != nil {
		return domain.ActiveSessionSnapshot{}, fmt.Errorf("%w: invalid started_at", domain.ErrInvalidActiveSession)
	}
	lastSeenAt, err := parseRedisTime(values["last_seen_at"])
	if err != nil {
		return domain.ActiveSessionSnapshot{}, fmt.Errorf("%w: invalid last_seen_at", domain.ErrInvalidActiveSession)
	}
	snapshot := domain.ActiveSessionSnapshot{
		SessionID:   values["session_id"],
		AnonymousID: values["anonymous_id"],
		UserID:      redisStringPtr(values["user_id"]),
		Status:      domain.SessionStatus(values["status"]),
		EntryPage:   values["entry_page"],
		CurrentPage: redisStringPtr(values["current_page"]),
		DeviceType:  domain.DeviceType(values["device_type"]),
		Browser:     redisStringPtr(values["browser"]),
		OS:          redisStringPtr(values["os"]),
		Country:     redisStringPtr(values["country"]),
		City:        redisStringPtr(values["city"]),
		Channel:     domain.Channel(values["channel"]),
		IPHash:      values["ip_hash"],
		StartedAt:   startedAt,
		LastSeenAt:  lastSeenAt,
	}
	if eventType := redisStringPtr(values["last_event_type"]); eventType != nil {
		typed := domain.EventType(*eventType)
		snapshot.LastEventType = &typed
	}
	if err := snapshot.Validate(); err != nil {
		return domain.ActiveSessionSnapshot{}, err
	}
	return snapshot.Normalize(), nil
}

func parseRedisTime(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, errors.New("empty time")
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func redisStringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func (s *RedisActiveSessionStore) activeKey(sessionID string) string {
	return s.keyPrefix + ":active:" + sessionID
}

func (s *RedisActiveSessionStore) userKey(userID string) string {
	return s.keyPrefix + ":user:" + userID
}

func (s *RedisActiveSessionStore) anonymousKey(anonymousID string) string {
	return s.keyPrefix + ":anon:" + anonymousID
}

func (s *RedisActiveSessionStore) lastSeenKey(sessionID string) string {
	return s.keyPrefix + ":last_seen:" + sessionID
}

func (s *RedisActiveSessionStore) activeCounterKey(at time.Time) string {
	return s.keyPrefix + ":counter:active:" + at.UTC().Format("2006010215")
}

func (s *RedisActiveSessionStore) eventCounterKey(at time.Time) string {
	return s.keyPrefix + ":counter:events:" + at.UTC().Format("200601021504")
}

func (s *RedisActiveSessionStore) activeSessionsKey() string {
	return s.keyPrefix + ":active_sessions"
}

func (s *RedisActiveSessionStore) activeUsersKey() string {
	return s.keyPrefix + ":active_users"
}

func (c ActiveSessionStoreConfig) Validate() error {
	if strings.TrimSpace(c.KeyPrefix) == "" {
		return errors.New("SESSION_REDIS_KEY_PREFIX cannot be empty")
	}
	if strings.ContainsAny(c.KeyPrefix, " \t\r\n") {
		return errors.New("SESSION_REDIS_KEY_PREFIX cannot contain whitespace")
	}
	if c.CounterTTL <= 0 {
		return errors.New("SESSION_ACTIVE_COUNTER_TTL must be greater than zero")
	}
	return nil
}

func (c ActiveSessionStoreConfig) withDefaults() ActiveSessionStoreConfig {
	c.KeyPrefix = strings.TrimSpace(c.KeyPrefix)
	if c.KeyPrefix == "" {
		c.KeyPrefix = defaultActiveSessionKeyPrefix
	}
	if c.CounterTTL == 0 {
		c.CounterTTL = defaultActiveCounterTTL
	}
	return c
}
