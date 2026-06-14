package retry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
)

type FailureKind string

const (
	FailureNone      FailureKind = ""
	FailureTransient FailureKind = "transient"
	FailureTerminal  FailureKind = "terminal"
)

type Policy struct {
	MaxAttempts int
	Delays      []time.Duration
}

func DefaultPolicy() Policy {
	return Policy{
		MaxAttempts: 4,
		Delays: []time.Duration{
			30 * time.Second,
			2 * time.Minute,
			8 * time.Minute,
		},
	}
}

func NewPolicy(maxAttempts int, delays []time.Duration) (Policy, error) {
	if maxAttempts < 1 || maxAttempts > 10 {
		return Policy{}, errors.New("notification retry max attempts must be between 1 and 10")
	}
	if len(delays) != maxAttempts-1 {
		return Policy{}, errors.New("notification retry delays must define one tier for each retry")
	}
	for index, delay := range delays {
		if delay <= 0 {
			return Policy{}, errors.New("notification retry delays must be positive")
		}
		if index > 0 && delay <= delays[index-1] {
			return Policy{}, errors.New("notification retry delays must be strictly increasing")
		}
	}
	return Policy{MaxAttempts: maxAttempts, Delays: append([]time.Duration(nil), delays...)}, nil
}

func (p Policy) NextDelay(failedAttempt int) (time.Duration, bool) {
	if failedAttempt < 1 || failedAttempt >= p.MaxAttempts {
		return 0, false
	}
	index := failedAttempt - 1
	if index >= len(p.Delays) {
		return 0, false
	}
	return p.Delays[index], true
}

func ClassifyError(err error) FailureKind {
	switch {
	case err == nil:
		return FailureNone
	case errors.Is(err, provider.ErrProviderUnavailable),
		errors.Is(err, context.DeadlineExceeded):
		return FailureTransient
	case errors.Is(err, provider.ErrProviderRejected),
		errors.Is(err, provider.ErrChannelDisabled),
		errors.Is(err, provider.ErrProviderNotConfigured),
		errors.Is(err, domain.ErrInvalidRetryInstruction),
		errors.Is(err, domain.ErrDurableOTPRetryForbidden),
		errors.Is(err, domain.ErrInvalidRenderRequest),
		errors.Is(err, domain.ErrUnsupportedTemplateKey),
		errors.Is(err, domain.ErrTemplateNotFound),
		errors.Is(err, domain.ErrTemplateRender),
		errors.Is(err, domain.ErrUnsupportedChannel):
		return FailureTerminal
	default:
		return FailureTransient
	}
}

func safeFailureCode(err error, exhausted bool) string {
	switch {
	case errors.Is(err, domain.ErrDurableOTPRetryForbidden):
		return "security_payload_blocked"
	case errors.Is(err, domain.ErrInvalidRetryInstruction):
		return "invalid_retry_instruction"
	case errors.Is(err, domain.ErrInvalidRenderRequest),
		errors.Is(err, domain.ErrUnsupportedTemplateKey),
		errors.Is(err, domain.ErrTemplateNotFound),
		errors.Is(err, domain.ErrTemplateRender):
		return "template_invalid"
	case errors.Is(err, provider.ErrProviderRejected):
		return "provider_rejected_recipient"
	case errors.Is(err, provider.ErrChannelDisabled):
		return "channel_disabled"
	case errors.Is(err, provider.ErrProviderNotConfigured):
		return "provider_not_configured"
	case errors.Is(err, provider.ErrProviderUnavailable),
		errors.Is(err, context.DeadlineExceeded):
		if exhausted {
			return "provider_unavailable_exhausted"
		}
		return "provider_unavailable"
	default:
		if exhausted {
			return "provider_unknown_failure_exhausted"
		}
		return "provider_unknown_failure"
	}
}

func routeForDelay(delay time.Duration) string {
	return "after." + strings.ReplaceAll(delay.String(), ".", "_")
}

func (p Policy) ValidateJob(job domain.RetryJob) error {
	if err := job.Validate(); err != nil {
		return err
	}
	if job.MaxAttempts != p.MaxAttempts {
		return fmt.Errorf("%w: max_attempts differs from active policy", domain.ErrInvalidRetryJob)
	}
	return nil
}
