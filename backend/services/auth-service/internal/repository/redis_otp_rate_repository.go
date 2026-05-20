package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
	otpsec "github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/otp"
)

type RedisOTPRateRepository struct {
	redis  *RedisClient
	policy otpsec.Policy
	pepper string
}

func NewRedisOTPRateRepository(redis *RedisClient, policy otpsec.Policy, pepper string) (*RedisOTPRateRepository, error) {
	if redis == nil {
		return nil, fmt.Errorf("redis client is required")
	}
	if err := policy.Validate(); err != nil {
		return nil, fmt.Errorf("invalid otp policy: %w", err)
	}
	if strings.TrimSpace(pepper) == "" {
		return nil, fmt.Errorf("otp rate limit pepper is required")
	}
	return &RedisOTPRateRepository{
		redis:  redis,
		policy: policy,
		pepper: strings.TrimSpace(pepper),
	}, nil
}

func (r *RedisOTPRateRepository) ReserveSend(ctx context.Context, purpose domain.OTPPurpose, target string, now time.Time) (time.Duration, error) {
	targetHash, err := otpsec.HashLookupKey(target, r.pepper)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", domain.ErrOTPRateLimitUnavailable, err)
	}

	cooldownKey := fmt.Sprintf("otp:cooldown:%s:%s", purpose, targetHash)
	ttl, err := r.redis.TTL(ctx, cooldownKey)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", domain.ErrOTPRateLimitUnavailable, err)
	}
	if ttl > 0 {
		return 0, &domain.OTPRateLimitError{
			Reason:     "cooldown_active",
			RetryAfter: ttl,
		}
	}

	windowKey := fmt.Sprintf("otp:send:window:%s:%s", purpose, targetHash)
	windowCount, err := r.redis.Incr(ctx, windowKey)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", domain.ErrOTPRateLimitUnavailable, err)
	}
	if windowCount == 1 {
		if err := r.redis.Expire(ctx, windowKey, r.policy.SendLimitWindow); err != nil {
			return 0, fmt.Errorf("%w: %v", domain.ErrOTPRateLimitUnavailable, err)
		}
	}
	if windowCount > int64(r.policy.SendLimitPerWindow) {
		return 0, &domain.OTPRateLimitError{
			Reason:     "send_window_exceeded",
			RetryAfter: r.policy.SendLimitWindow,
		}
	}

	dailyKey := fmt.Sprintf("otp:send:daily:%s:%s:%s", purpose, targetHash, now.UTC().Format("20060102"))
	dailyCount, err := r.redis.Incr(ctx, dailyKey)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", domain.ErrOTPRateLimitUnavailable, err)
	}
	if dailyCount == 1 {
		if err := r.redis.Expire(ctx, dailyKey, durationUntilNextUTCDay(now)); err != nil {
			return 0, fmt.Errorf("%w: %v", domain.ErrOTPRateLimitUnavailable, err)
		}
	}
	if dailyCount > int64(r.policy.DailyLimit) {
		return 0, &domain.OTPRateLimitError{
			Reason:     "daily_limit_exceeded",
			RetryAfter: durationUntilNextUTCDay(now),
		}
	}

	ok, err := r.redis.SetNX(ctx, cooldownKey, "1", r.policy.ResendCooldown)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", domain.ErrOTPRateLimitUnavailable, err)
	}
	if !ok {
		ttl, _ := r.redis.TTL(ctx, cooldownKey)
		return 0, &domain.OTPRateLimitError{
			Reason:     "cooldown_active",
			RetryAfter: ttl,
		}
	}

	return r.policy.ResendCooldown, nil
}

func (r *RedisOTPRateRepository) MarkVerifyAttempt(ctx context.Context, challengeID string, verificationSource string, now time.Time) error {
	source := strings.TrimSpace(verificationSource)
	if source == "" {
		source = "unknown"
	}
	sourceHash, err := otpsec.HashLookupKey(source, r.pepper)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrOTPRateLimitUnavailable, err)
	}

	key := fmt.Sprintf("otp:verify:ip:%s:%s", sourceHash, challengeID)
	count, err := r.redis.Incr(ctx, key)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrOTPRateLimitUnavailable, err)
	}
	if count == 1 {
		if err := r.redis.Expire(ctx, key, r.policy.VerifyIPWindow); err != nil {
			return fmt.Errorf("%w: %v", domain.ErrOTPRateLimitUnavailable, err)
		}
	}
	if count > int64(r.policy.VerifyIPLimit) {
		return &domain.OTPRateLimitError{
			Reason:     "verify_source_limit_exceeded",
			RetryAfter: r.policy.VerifyIPWindow,
		}
	}
	return nil
}

func durationUntilNextUTCDay(now time.Time) time.Duration {
	utc := now.UTC()
	nextDay := time.Date(utc.Year(), utc.Month(), utc.Day()+1, 0, 0, 0, 0, time.UTC)
	return nextDay.Sub(utc)
}
