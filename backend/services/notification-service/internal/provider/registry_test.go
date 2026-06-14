package provider_test

import (
	"context"
	"errors"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/config"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
)

type stubProvider struct {
	name    string
	channel domain.Channel
}

func (p stubProvider) Name() string {
	return p.name
}

func (p stubProvider) Channel() domain.Channel {
	return p.channel
}

func (p stubProvider) Send(context.Context, domain.Message) (provider.Result, error) {
	return provider.Result{Channel: p.channel, ProviderName: p.name, Status: provider.StatusAccepted}, nil
}

func TestRegistryResolvesProviderForEachChannel(t *testing.T) {
	t.Parallel()

	items := []provider.Provider{
		stubProvider{name: "email", channel: domain.ChannelEmail},
		stubProvider{name: "sms", channel: domain.ChannelSMS},
		stubProvider{name: "push", channel: domain.ChannelPush},
		stubProvider{name: "messaging", channel: domain.ChannelWhatsAppLike},
	}
	registry, err := provider.NewRegistry(items...)
	if err != nil {
		t.Fatalf("NewRegistry() returned error: %v", err)
	}
	for _, expected := range items {
		got, err := registry.Resolve(expected.Channel())
		if err != nil {
			t.Fatalf("Resolve(%q) returned error: %v", expected.Channel(), err)
		}
		if got.Name() != expected.Name() {
			t.Fatalf("Resolve(%q).Name() = %q, want %q", expected.Channel(), got.Name(), expected.Name())
		}
	}
}

func TestRegistryRejectsDuplicateChannel(t *testing.T) {
	t.Parallel()

	_, err := provider.NewRegistry(
		stubProvider{name: "one", channel: domain.ChannelEmail},
		stubProvider{name: "two", channel: domain.ChannelEmail},
	)
	if !errors.Is(err, provider.ErrInvalidProvider) {
		t.Fatalf("NewRegistry() error = %v, want ErrInvalidProvider", err)
	}
}

func TestRegistryReportsUnconfiguredProvider(t *testing.T) {
	t.Parallel()

	registry, err := provider.NewRegistry(stubProvider{name: "email", channel: domain.ChannelEmail})
	if err != nil {
		t.Fatalf("NewRegistry() returned error: %v", err)
	}
	if _, err := registry.Resolve(domain.ChannelSMS); !errors.Is(err, provider.ErrProviderNotConfigured) {
		t.Fatalf("Resolve() error = %v, want ErrProviderNotConfigured", err)
	}
}

func TestConfiguredRegistrySelectsEnabledProviderAndRejectsDisabledChannel(t *testing.T) {
	t.Parallel()

	settings := config.Config{
		Email: config.ChannelConfig{Enabled: true, ProviderName: "selected"},
		SMS:   config.ChannelConfig{Enabled: false, ProviderName: "sms"},
	}
	registry, err := provider.NewConfiguredRegistry(
		settings,
		stubProvider{name: "alternate", channel: domain.ChannelEmail},
		stubProvider{name: "selected", channel: domain.ChannelEmail},
		stubProvider{name: "sms", channel: domain.ChannelSMS},
	)
	if err != nil {
		t.Fatalf("NewConfiguredRegistry() returned error: %v", err)
	}
	got, err := registry.Resolve(domain.ChannelEmail)
	if err != nil || got.Name() != "selected" {
		t.Fatalf("Resolve(email) = %v, %v; want selected provider", got, err)
	}
	if _, err := registry.Resolve(domain.ChannelSMS); !errors.Is(err, provider.ErrChannelDisabled) {
		t.Fatalf("Resolve(SMS) error = %v, want ErrChannelDisabled", err)
	}
}
