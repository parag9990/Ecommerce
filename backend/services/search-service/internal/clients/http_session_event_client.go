package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

const (
	defaultSessionEventIngestPath = "/api/v1/sessions/events"
	defaultSessionClientTimeout   = 150 * time.Millisecond
	maxSessionEventResponseBytes  = 1 << 20
)

type HTTPSessionEventClientConfig struct {
	BaseURL    string
	IngestPath string
	Timeout    time.Duration
}

type HTTPSessionEventClient struct {
	baseURL    *url.URL
	ingestPath string
	httpClient *http.Client
}

func NewHTTPSessionEventClient(cfg HTTPSessionEventClientConfig, httpClient *http.Client) (*HTTPSessionEventClient, error) {
	baseURL, err := url.Parse(strings.TrimSpace(cfg.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("parse session service url: %w", err)
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, errors.New("SESSION_SERVICE_URL must use http or https")
	}
	if strings.TrimSpace(baseURL.Host) == "" {
		return nil, errors.New("SESSION_SERVICE_URL host is required")
	}
	if strings.TrimSpace(cfg.IngestPath) == "" {
		cfg.IngestPath = defaultSessionEventIngestPath
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultSessionClientTimeout
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.Timeout}
	}
	return &HTTPSessionEventClient{
		baseURL:    baseURL,
		ingestPath: cfg.IngestPath,
		httpClient: httpClient,
	}, nil
}

func (c *HTTPSessionEventClient) IngestSearchEvent(ctx context.Context, event domain.ZeroResultSearchEvent) error {
	event = event.Normalize()
	if !event.ShouldTrack() {
		return fmt.Errorf("%w: %s", domain.ErrInvalidZeroResultEvent, event.SkipReason())
	}

	body, err := json.Marshal(sessionEventRequestFromZeroResult(event))
	if err != nil {
		return fmt.Errorf("%w: encode request: %v", domain.ErrSessionEventUnavailable, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("%w: build request: %v", domain.ErrSessionEventUnavailable, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if event.RequestID != "" {
		req.Header.Set("X-Request-ID", event.RequestID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: call session service: %v", domain.ErrSessionEventUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxSessionEventResponseBytes))
		return fmt.Errorf("%w: session service returned status %d", domain.ErrSessionEventUnavailable, resp.StatusCode)
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxSessionEventResponseBytes))
	return nil
}

func (c *HTTPSessionEventClient) endpoint() string {
	endpoint := *c.baseURL
	basePath := strings.TrimRight(endpoint.Path, "/")
	ingestPath := "/" + strings.TrimLeft(c.ingestPath, "/")
	endpoint.Path = basePath + ingestPath
	return endpoint.String()
}

type sessionEventRequest struct {
	EventType   string         `json:"event_type"`
	AnonymousID string         `json:"anonymous_id"`
	SessionID   string         `json:"session_id"`
	UserID      string         `json:"user_id,omitempty"`
	OccurredAt  string         `json:"occurred_at"`
	Path        string         `json:"path,omitempty"`
	Properties  map[string]any `json:"properties"`
}

func sessionEventRequestFromZeroResult(event domain.ZeroResultSearchEvent) sessionEventRequest {
	properties := map[string]any{
		"query":            event.Query,
		"normalized_query": event.NormalizedQuery,
		"result_count":     0,
		"zero_result":      true,
		"filters":          event.Filters,
		"sort":             event.Sort,
		"page":             event.Page,
		"page_size":        event.PageSize,
		"source":           domain.ZeroResultEventSource,
	}
	if event.RequestID != "" {
		properties["request_id"] = event.RequestID
	}
	return sessionEventRequest{
		EventType:   domain.ZeroResultEventTypeSearch,
		AnonymousID: event.AnonymousID,
		SessionID:   event.SessionID,
		UserID:      event.UserID,
		OccurredAt:  event.OccurredAt.UTC().Format(time.RFC3339Nano),
		Path:        event.Path,
		Properties:  properties,
	}
}
