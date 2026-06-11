package provider

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

const defaultProviderTimeout = 5 * time.Second
const defaultWebhookTimestampTolerance = 5 * time.Minute

type Config struct {
	DefaultProvider   string
	AllowedProviders  []string
	AllowedCurrencies []string
	CaptureMode       CaptureMode
	Providers         map[string]ProviderConfig
}

func (c Config) Normalized() Config {
	c.DefaultProvider = NormalizeProviderName(c.DefaultProvider)
	c.CaptureMode = c.CaptureMode.Normalized()
	if c.CaptureMode == "" {
		c.CaptureMode = CaptureModeAutomatic
	}

	c.AllowedProviders = normalizeStringList(c.AllowedProviders, NormalizeProviderName)
	if c.DefaultProvider != "" && len(c.AllowedProviders) == 0 {
		c.AllowedProviders = []string{c.DefaultProvider}
	}
	c.AllowedCurrencies = normalizeStringList(c.AllowedCurrencies, func(value string) string {
		return strings.ToUpper(strings.TrimSpace(value))
	})

	providers := make(map[string]ProviderConfig, len(c.Providers))
	for name, cfg := range c.Providers {
		cfg = cfg.Normalized()
		providers[NormalizeProviderName(name)] = cfg
	}
	c.Providers = providers
	return c
}

func (c Config) Validate() error {
	c = c.Normalized()
	if err := ValidateCaptureMode(c.CaptureMode); err != nil {
		return err
	}
	if len(c.AllowedProviders) > 0 && c.DefaultProvider == "" {
		return fmt.Errorf("%w: default provider is required when providers are enabled", ErrInvalidProviderRequest)
	}
	if c.DefaultProvider != "" && !contains(c.AllowedProviders, c.DefaultProvider) {
		return fmt.Errorf("%w: default provider %q is not allowed", ErrInvalidProviderRequest, c.DefaultProvider)
	}
	for _, currency := range c.AllowedCurrencies {
		if !isCurrencyCode(currency) {
			return fmt.Errorf("%w: allowed currency %q must be a 3-letter ISO code", ErrInvalidProviderRequest, currency)
		}
	}
	seen := make(map[string]struct{}, len(c.AllowedProviders))
	for _, providerName := range c.AllowedProviders {
		if providerName == "" {
			return fmt.Errorf("%w: allowed provider cannot be empty", ErrInvalidProviderRequest)
		}
		if _, ok := seen[providerName]; ok {
			return fmt.Errorf("%w: duplicate allowed provider %q", ErrInvalidProviderRequest, providerName)
		}
		seen[providerName] = struct{}{}
		cfg, ok := c.Providers[providerName]
		if !ok {
			return fmt.Errorf("%w: config missing for provider %q", ErrInvalidProviderRequest, providerName)
		}
		if err := cfg.Validate(providerName); err != nil {
			return err
		}
	}
	return nil
}

func (c Config) ProviderConfig(name string) (ProviderConfig, bool) {
	c = c.Normalized()
	cfg, ok := c.Providers[NormalizeProviderName(name)]
	return cfg, ok
}

func (c Config) AllowsCurrency(currency string) bool {
	c = c.Normalized()
	currency = strings.ToUpper(strings.TrimSpace(currency))
	for _, allowed := range c.AllowedCurrencies {
		if allowed == currency {
			return true
		}
	}
	return false
}

func (c Config) ValidateCurrency(currency string) error {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if !isCurrencyCode(currency) {
		return fmt.Errorf("%w: currency must be a 3-letter ISO code", ErrInvalidProviderRequest)
	}
	if !c.AllowsCurrency(currency) {
		return fmt.Errorf("%w: currency %q is not allowed", ErrInvalidProviderRequest, currency)
	}
	return nil
}

type ProviderConfig struct {
	PublicKey                 string
	SecretKey                 string
	WebhookSecret             string
	BaseURL                   string
	Timeout                   time.Duration
	WebhookTimestampTolerance time.Duration
}

func (c ProviderConfig) Normalized() ProviderConfig {
	c.PublicKey = strings.TrimSpace(c.PublicKey)
	c.SecretKey = strings.TrimSpace(c.SecretKey)
	c.WebhookSecret = strings.TrimSpace(c.WebhookSecret)
	c.BaseURL = strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	if c.Timeout == 0 {
		c.Timeout = defaultProviderTimeout
	}
	if c.WebhookTimestampTolerance == 0 {
		c.WebhookTimestampTolerance = defaultWebhookTimestampTolerance
	}
	return c
}

func (c ProviderConfig) Validate(providerName string) error {
	c = c.Normalized()
	providerName = NormalizeProviderName(providerName)
	if c.Timeout <= 0 {
		return fmt.Errorf("%w: provider %q timeout must be greater than zero", ErrInvalidProviderRequest, providerName)
	}
	if c.WebhookTimestampTolerance <= 0 {
		return fmt.Errorf("%w: provider %q webhook timestamp tolerance must be greater than zero", ErrInvalidProviderRequest, providerName)
	}
	if c.BaseURL != "" {
		baseURL, err := url.Parse(c.BaseURL)
		if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
			return fmt.Errorf("%w: provider %q base URL is invalid", ErrInvalidProviderRequest, providerName)
		}
		if baseURL.Scheme != "https" && baseURL.Scheme != "http" {
			return fmt.Errorf("%w: provider %q base URL must use HTTP or HTTPS", ErrInvalidProviderRequest, providerName)
		}
		if baseURL.Scheme == "http" && !isLoopbackHost(baseURL.Hostname()) {
			return fmt.Errorf("%w: provider %q base URL must use HTTPS outside local testing", ErrInvalidProviderRequest, providerName)
		}
	}
	switch providerName {
	case ProviderNameStripeLike, ProviderNameRazorpayLike:
		if c.PublicKey == "" {
			return fmt.Errorf("%w: provider %q public key is required", ErrInvalidProviderRequest, providerName)
		}
		if c.SecretKey == "" {
			return fmt.Errorf("%w: provider %q secret key is required", ErrInvalidProviderRequest, providerName)
		}
		if c.WebhookSecret == "" {
			return fmt.Errorf("%w: provider %q webhook secret is required", ErrInvalidProviderRequest, providerName)
		}
	default:
		return fmt.Errorf("%w: unsupported provider %q", ErrInvalidProviderRequest, providerName)
	}
	return nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(strings.TrimSpace(host), "localhost") {
		return true
	}
	ip := net.ParseIP(strings.TrimSpace(host))
	return ip != nil && ip.IsLoopback()
}

func normalizeStringList(values []string, normalize func(string) string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = normalize(value)
		if value == "" {
			continue
		}
		out = append(out, value)
	}
	return out
}

func contains(values []string, value string) bool {
	for _, existing := range values {
		if existing == value {
			return true
		}
	}
	return false
}
