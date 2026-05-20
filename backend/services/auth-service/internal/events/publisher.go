package events

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, event domain.OutboxEvent) error
}

type HTTPPublisher struct {
	endpoint string
	client   *http.Client
	logger   *slog.Logger
}

func NewHTTPPublisher(endpoint string, timeout time.Duration, logger *slog.Logger) (*HTTPPublisher, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return nil, errors.New("event publisher endpoint is required")
	}
	if timeout <= 0 {
		return nil, errors.New("event publisher timeout must be greater than zero")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &HTTPPublisher{
		endpoint: endpoint,
		client:   &http.Client{Timeout: timeout},
		logger:   logger,
	}, nil
}

func (p *HTTPPublisher) Publish(ctx context.Context, topic string, event domain.OutboxEvent) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(event.Payload))
	if err != nil {
		return fmt.Errorf("create event publish request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Event-Topic", topic)
	req.Header.Set("X-Event-ID", event.EventID)
	req.Header.Set("X-Event-Type", event.EventType)
	req.Header.Set("X-Event-Routing-Key", event.RoutingKey)
	if event.TraceID != "" {
		req.Header.Set("X-Trace-ID", event.TraceID)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("publish event: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("publish event status %d", resp.StatusCode)
	}

	p.logger.DebugContext(ctx, "auth.outbox.publish_succeeded",
		slog.String("event_id", event.EventID),
		slog.String("event_type", event.EventType),
		slog.String("topic", topic),
	)
	return nil
}
