package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPSMSClientConfig struct {
	Endpoint    string
	BearerToken string
	Timeout     time.Duration
}

type HTTPSMSClient struct {
	endpoint    string
	bearerToken string
	client      *http.Client
}

func NewHTTPSMSClient(config HTTPSMSClientConfig) (*HTTPSMSClient, error) {
	endpoint := strings.TrimSpace(config.Endpoint)
	parsed, err := url.ParseRequestURI(endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("SMS HTTP endpoint must be an absolute URL")
	}
	if parsed.Scheme != "https" && !smsLoopbackHost(parsed.Hostname()) {
		return nil, errors.New("SMS HTTP endpoint must use HTTPS outside local development")
	}
	if strings.TrimSpace(config.BearerToken) == "" && !smsLoopbackHost(parsed.Hostname()) {
		return nil, errors.New("SMS HTTP bearer token is required outside local development")
	}
	if config.Timeout <= 0 {
		return nil, errors.New("SMS HTTP timeout must be greater than zero")
	}
	return &HTTPSMSClient{
		endpoint:    endpoint,
		bearerToken: strings.TrimSpace(config.BearerToken),
		client:      &http.Client{Timeout: config.Timeout},
	}, nil
}

func (c *HTTPSMSClient) Deliver(ctx context.Context, phoneE164, textBody string) (string, error) {
	if c == nil || c.client == nil {
		return "", ErrProviderUnavailable
	}
	if ctx == nil {
		return "", ErrProviderRejected
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	body, err := json.Marshal(struct {
		To   string `json:"to"`
		Text string `json:"text"`
	}{
		To:   strings.TrimSpace(phoneE164),
		Text: textBody,
	})
	if err != nil {
		return "", ErrProviderUnavailable
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", ErrProviderUnavailable
	}
	request.Header.Set("Content-Type", "application/json")
	if c.bearerToken != "" {
		request.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}

	response, err := c.client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", ErrProviderUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusBadRequest && response.StatusCode < http.StatusInternalServerError {
		return "", ErrProviderRejected
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", ErrProviderUnavailable
	}
	var result struct {
		MessageID string `json:"message_id"`
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 64<<10))
	if err := decoder.Decode(&result); err != nil || strings.TrimSpace(result.MessageID) == "" {
		return "", ErrProviderUnavailable
	}
	return strings.TrimSpace(result.MessageID), nil
}

func smsLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
