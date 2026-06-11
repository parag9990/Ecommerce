package provider

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"
)

type Registry struct {
	providers map[string]Provider
}

func NewRegistry(providers ...Provider) (*Registry, error) {
	registry := &Registry{
		providers: make(map[string]Provider, len(providers)),
	}
	for _, p := range providers {
		if p == nil {
			return nil, fmt.Errorf("%w: provider is nil", ErrInvalidProviderRequest)
		}
		name := NormalizeProviderName(p.Name())
		if name == "" {
			return nil, fmt.Errorf("%w: payment provider name is required", ErrInvalidProviderRequest)
		}
		if _, exists := registry.providers[name]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateProvider, name)
		}
		registry.providers[name] = p
	}
	return registry, nil
}

func NewRegistryFromConfig(cfg Config, logger *slog.Logger) (*Registry, error) {
	cfg = cfg.Normalized()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	providers := make([]Provider, 0, len(cfg.AllowedProviders))
	for _, providerName := range cfg.AllowedProviders {
		providerConfig, ok := cfg.ProviderConfig(providerName)
		if !ok {
			return nil, fmt.Errorf("%w: provider config missing for %s", ErrInvalidProviderRequest, providerName)
		}
		switch providerName {
		case ProviderNameStripeLike:
			p, err := NewStripeLikeProvider(providerConfig, logger)
			if err != nil {
				return nil, err
			}
			providers = append(providers, p)
		case ProviderNameRazorpayLike:
			p, err := NewRazorpayLikeProvider(providerConfig, logger)
			if err != nil {
				return nil, err
			}
			providers = append(providers, p)
		default:
			return nil, fmt.Errorf("%w: unsupported provider %q", ErrInvalidProviderRequest, providerName)
		}
	}
	return NewRegistry(providers...)
}

func (r *Registry) Get(name string) (Provider, error) {
	if r == nil {
		return nil, NewProviderNotFoundError(name)
	}
	name = NormalizeProviderName(name)
	provider, ok := r.providers[name]
	if !ok {
		return nil, NewProviderNotFoundError(name)
	}
	return provider, nil
}

func (r *Registry) Contains(name string) bool {
	if r == nil {
		return false
	}
	_, ok := r.providers[NormalizeProviderName(name)]
	return ok
}

func (r *Registry) Names() []string {
	if r == nil {
		return nil
	}
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *Registry) ValidateDefaultProvider(name string) error {
	if NormalizeProviderName(name) == "" {
		return fmt.Errorf("%w: default provider is required", ErrInvalidProviderRequest)
	}
	if _, err := r.Get(name); err != nil {
		if errors.Is(err, ErrProviderNotFound) {
			return err
		}
		return err
	}
	return nil
}
