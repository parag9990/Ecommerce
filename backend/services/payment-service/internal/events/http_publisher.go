package events

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
)

type HTTPPublisherConfig struct {
	Endpoint  string
	AuthToken string
	Timeout   time.Duration
}

func (c HTTPPublisherConfig) Normalized() HTTPPublisherConfig {
	c.Endpoint = strings.TrimSpace(c.Endpoint)
	c.AuthToken = strings.TrimSpace(c.AuthToken)
	if c.Timeout == 0 {
		c.Timeout = 5 * time.Second
	}
	return c
}

func (c HTTPPublisherConfig) Validate() error {
	c = c.Normalized()
	if c.Endpoint == "" {
		return errors.New("PAYMENT_EVENTS_ENDPOINT is required when payment event publishing is enabled")
	}
	endpoint, err := url.Parse(c.Endpoint)
	if err != nil || endpoint.Scheme == "" || endpoint.Host == "" {
		return errors.New("PAYMENT_EVENTS_ENDPOINT must be a valid URL")
	}
	if endpoint.Scheme != "https" && !(endpoint.Scheme == "http" && isLoopbackHost(endpoint.Hostname())) {
		return errors.New("PAYMENT_EVENTS_ENDPOINT must use HTTPS outside local testing")
	}
	if c.Timeout <= 0 {
		return errors.New("PAYMENT_EVENTS_TIMEOUT must be greater than zero")
	}
	if len(c.AuthToken) < 32 {
		return errors.New("PAYMENT_EVENTS_AUTH_TOKEN must be at least 32 characters")
	}
	return nil
}

type HTTPPublisher struct {
	config HTTPPublisherConfig
	client *http.Client
}

func NewHTTPPublisher(cfg HTTPPublisherConfig) (*HTTPPublisher, error) {
	cfg = cfg.Normalized()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &HTTPPublisher{
		config: cfg,
		client: &http.Client{Timeout: cfg.Timeout},
	}, nil
}

type publishedEnvelope struct {
	Topic string                    `json:"topic"`
	Event domain.PaymentDomainEvent `json:"event"`
}

func (p *HTTPPublisher) Publish(ctx context.Context, topic string, event domain.PaymentDomainEvent) error {
	if p == nil {
		return errors.New("payment event publisher is not initialized")
	}
	topic = strings.TrimSpace(topic)
	if topic == "" || strings.TrimSpace(event.EventID) == "" {
		return errors.New("payment event topic and event id are required")
	}
	return p.publishJSON(ctx, event.EventID, publishedEnvelope{Topic: topic, Event: event})
}

func (p *HTTPPublisher) publishJSON(ctx context.Context, idempotencyKey string, value any) error {
	body, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal payment event: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.Endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create payment event request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.AuthToken)
	req.Header.Set("Idempotency-Key", idempotencyKey)
	res, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("publish payment event: %w", err)
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("publish payment event: endpoint returned status %d", res.StatusCode)
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
