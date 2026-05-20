package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SendOTPRequest struct {
	Target           string
	Channel          string
	Purpose          string
	OTP              string
	ChallengeID      string
	ExpiresInSeconds int
}

type NotificationClient interface {
	SendOTP(ctx context.Context, req SendOTPRequest) error
}

type HTTPNotificationClient struct {
	endpoint string
	client   *http.Client
}

func NewHTTPNotificationClient(endpoint string, timeout time.Duration) (*HTTPNotificationClient, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return nil, errors.New("notification otp endpoint is required")
	}
	parsed, err := url.ParseRequestURI(endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("notification otp endpoint must be an absolute URL")
	}
	if timeout <= 0 {
		return nil, errors.New("notification timeout must be greater than zero")
	}

	return &HTTPNotificationClient{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *HTTPNotificationClient) SendOTP(ctx context.Context, req SendOTPRequest) error {
	if c == nil || c.client == nil {
		return errors.New("notification client is not initialized")
	}
	if strings.TrimSpace(req.Target) == "" ||
		strings.TrimSpace(req.Channel) == "" ||
		strings.TrimSpace(req.Purpose) == "" ||
		strings.TrimSpace(req.OTP) == "" ||
		strings.TrimSpace(req.ChallengeID) == "" ||
		req.ExpiresInSeconds <= 0 {
		return errors.New("notification otp request is incomplete")
	}

	body, err := json.Marshal(struct {
		Target           string `json:"target"`
		Channel          string `json:"channel"`
		Purpose          string `json:"purpose"`
		OTP              string `json:"otp"`
		ChallengeID      string `json:"challenge_id"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}{
		Target:           req.Target,
		Channel:          req.Channel,
		Purpose:          req.Purpose,
		OTP:              req.OTP,
		ChallengeID:      req.ChallengeID,
		ExpiresInSeconds: req.ExpiresInSeconds,
	})
	if err != nil {
		return fmt.Errorf("marshal notification otp request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build notification otp request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send notification otp request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("notification otp request failed with status %d", resp.StatusCode)
	}
	return nil
}
