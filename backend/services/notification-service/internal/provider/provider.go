package provider

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

var (
	ErrInvalidProvider       = errors.New("invalid notification provider")
	ErrProviderNotConfigured = errors.New("notification provider not configured")
	ErrChannelDisabled       = errors.New("notification channel disabled")
	ErrProviderUnavailable   = errors.New("notification provider unavailable")
	ErrProviderRejected      = errors.New("notification provider rejected message")
)

type DeliveryStatus string

const (
	StatusAccepted DeliveryStatus = "accepted"
	StatusRejected DeliveryStatus = "rejected"
)

type Result struct {
	Channel           domain.Channel
	ProviderName      string
	ProviderMessageID string
	Status            DeliveryStatus
}

type Provider interface {
	Name() string
	Channel() domain.Channel
	Send(ctx context.Context, message domain.Message) (Result, error)
}

func validatedName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("%w: provider name is required", ErrInvalidProvider)
	}
	return name, nil
}

func missingDependency(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func validateSend(ctx context.Context, channel domain.Channel, message domain.Message) error {
	if ctx == nil {
		return fmt.Errorf("%w: context is required", ErrProviderRejected)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if message.Channel != channel {
		return fmt.Errorf("%w: provider handles %q but received %q", ErrProviderRejected, channel, message.Channel)
	}
	if err := message.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrProviderRejected, err)
	}
	return nil
}

func result(channel domain.Channel, providerName, messageID string) Result {
	return Result{
		Channel:           channel,
		ProviderName:      providerName,
		ProviderMessageID: messageID,
		Status:            StatusAccepted,
	}
}

func deliveryResult(channel domain.Channel, providerName, messageID string, err error) (Result, error) {
	if err == nil {
		return result(channel, providerName, messageID), nil
	}
	switch {
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return Result{}, err
	case errors.Is(err, ErrProviderRejected):
		return Result{Channel: channel, ProviderName: providerName, Status: StatusRejected}, ErrProviderRejected
	case errors.Is(err, ErrProviderUnavailable):
		return Result{}, ErrProviderUnavailable
	default:
		// Provider errors may contain recipient or message data; do not leak them upward.
		return Result{}, ErrProviderUnavailable
	}
}
