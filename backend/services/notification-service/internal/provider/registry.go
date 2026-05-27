package provider

import (
	"fmt"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

type ChannelConfiguration interface {
	ChannelEnabled(channel domain.Channel) bool
	ProviderName(channel domain.Channel) string
}

type Registry struct {
	byChannel map[domain.Channel]Provider
	disabled  map[domain.Channel]struct{}
}

func NewRegistry(providers ...Provider) (*Registry, error) {
	registry := &Registry{
		byChannel: make(map[domain.Channel]Provider),
		disabled:  make(map[domain.Channel]struct{}),
	}
	for _, item := range providers {
		if err := registry.register(item); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func NewConfiguredRegistry(configuration ChannelConfiguration, providers ...Provider) (*Registry, error) {
	if missingDependency(configuration) {
		return nil, fmt.Errorf("%w: channel configuration is required", ErrInvalidProvider)
	}

	candidates := make(map[domain.Channel]map[string]Provider)
	for _, item := range providers {
		if err := validateRegistration(item); err != nil {
			return nil, err
		}
		if candidates[item.Channel()] == nil {
			candidates[item.Channel()] = make(map[string]Provider)
		}
		if _, exists := candidates[item.Channel()][item.Name()]; exists {
			return nil, fmt.Errorf("%w: provider %q registered twice for channel %q", ErrInvalidProvider, item.Name(), item.Channel())
		}
		candidates[item.Channel()][item.Name()] = item
	}

	registry := &Registry{
		byChannel: make(map[domain.Channel]Provider),
		disabled:  make(map[domain.Channel]struct{}),
	}
	for _, channel := range domain.SupportedChannels() {
		if !configuration.ChannelEnabled(channel) {
			registry.disabled[channel] = struct{}{}
			continue
		}

		name := strings.TrimSpace(configuration.ProviderName(channel))
		if name == "" {
			return nil, fmt.Errorf("%w: enabled channel %q has no selected provider", ErrProviderNotConfigured, channel)
		}
		item, exists := candidates[channel][name]
		if !exists {
			return nil, fmt.Errorf("%w: provider %q for channel %q", ErrProviderNotConfigured, name, channel)
		}
		registry.byChannel[channel] = item
	}
	return registry, nil
}

func (r *Registry) Resolve(channel domain.Channel) (Provider, error) {
	if r == nil {
		return nil, ErrProviderNotConfigured
	}
	if !channel.IsSupported() {
		return nil, fmt.Errorf("%w: %q", domain.ErrUnsupportedChannel, channel)
	}
	if _, disabled := r.disabled[channel]; disabled {
		return nil, fmt.Errorf("%w: %q", ErrChannelDisabled, channel)
	}
	item, ok := r.byChannel[channel]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrProviderNotConfigured, channel)
	}
	return item, nil
}

func (r *Registry) register(item Provider) error {
	if err := validateRegistration(item); err != nil {
		return err
	}
	if _, exists := r.byChannel[item.Channel()]; exists {
		return fmt.Errorf("%w: provider already registered for channel %q", ErrInvalidProvider, item.Channel())
	}
	r.byChannel[item.Channel()] = item
	return nil
}

func validateRegistration(item Provider) error {
	if missingDependency(item) {
		return fmt.Errorf("%w: provider is required", ErrInvalidProvider)
	}
	if strings.TrimSpace(item.Name()) == "" {
		return fmt.Errorf("%w: provider name is required", ErrInvalidProvider)
	}
	if !item.Channel().IsSupported() {
		return fmt.Errorf("%w: provider %q: %w", ErrInvalidProvider, item.Name(), domain.ErrUnsupportedChannel)
	}
	return nil
}
