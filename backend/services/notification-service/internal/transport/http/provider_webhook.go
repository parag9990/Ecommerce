package http

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	nethttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/analytics"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

const (
	SignatureHeader = "X-Notification-Signature"
	TimestampHeader = "X-Notification-Timestamp"
)

type CallbackRecorder interface {
	RecordCallback(ctx context.Context, callback domain.ProviderCallback) (domain.DeliveryEventApplyResult, error)
}

type WebhookEndpoint struct {
	Channel           domain.Channel
	Provider          string
	SigningSecret     string
	AllowOpenTracking bool
}

type ProviderWebhookRequest struct {
	ProviderEventID   string                   `json:"provider_event_id"`
	ProviderMessageID string                   `json:"provider_message_id"`
	Type              domain.DeliveryEventType `json:"type"`
	OccurredAt        time.Time                `json:"occurred_at"`
	FailureCode       string                   `json:"failure_code,omitempty"`
}

type ProviderWebhookHandler struct {
	endpoints    map[domain.Channel]WebhookEndpoint
	recorder     CallbackRecorder
	metrics      analytics.Observer
	logger       *slog.Logger
	maxBodyBytes int64
	replayWindow time.Duration
	now          func() time.Time
}

func NewProviderWebhookHandler(
	endpoints []WebhookEndpoint,
	recorder CallbackRecorder,
	metrics analytics.Observer,
	maxBodyBytes int64,
	replayWindow time.Duration,
	logger *slog.Logger,
) (*ProviderWebhookHandler, error) {
	if recorder == nil || metrics == nil || maxBodyBytes < 1 || replayWindow <= 0 {
		return nil, errors.New("provider webhook dependencies and limits are required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	configured := make(map[domain.Channel]WebhookEndpoint, len(endpoints))
	for _, endpoint := range endpoints {
		endpoint.Provider = strings.TrimSpace(endpoint.Provider)
		endpoint.SigningSecret = strings.TrimSpace(endpoint.SigningSecret)
		if !endpoint.Channel.IsSupported() || endpoint.Provider == "" || endpoint.SigningSecret == "" {
			return nil, errors.New("provider webhook endpoint channel, provider, and signing secret are required")
		}
		if endpoint.AllowOpenTracking && endpoint.Channel != domain.ChannelEmail {
			return nil, fmt.Errorf("observed-open events are unsupported for channel %q", endpoint.Channel)
		}
		if _, exists := configured[endpoint.Channel]; exists {
			return nil, fmt.Errorf("provider webhook endpoint already configured for channel %q", endpoint.Channel)
		}
		configured[endpoint.Channel] = endpoint
	}
	if len(configured) == 0 {
		return nil, errors.New("at least one provider webhook endpoint is required")
	}
	return &ProviderWebhookHandler{
		endpoints: configured, recorder: recorder, metrics: metrics, logger: logger,
		maxBodyBytes: maxBodyBytes, replayWindow: replayWindow, now: time.Now,
	}, nil
}

func (h *ProviderWebhookHandler) WithClock(now func() time.Time) {
	if now != nil {
		h.now = now
	}
}

func (h *ProviderWebhookHandler) Register(mux *nethttp.ServeMux, pathPrefix string) error {
	if mux == nil {
		return errors.New("provider webhook mux is required")
	}
	pathPrefix = strings.TrimRight(strings.TrimSpace(pathPrefix), "/")
	if !strings.HasPrefix(pathPrefix, "/") {
		return errors.New("provider webhook path prefix must be absolute")
	}
	mux.Handle("POST "+pathPrefix+"/{channel}", h)
	return nil
}

func (h *ProviderWebhookHandler) ServeHTTP(w nethttp.ResponseWriter, req *nethttp.Request) {
	startedAt := h.now()
	channel, err := domain.ParseChannel(req.PathValue("channel"))
	if err != nil {
		nethttp.NotFound(w, req)
		return
	}
	endpoint, exists := h.endpoints[channel]
	if !exists {
		nethttp.NotFound(w, req)
		return
	}
	defer func() {
		h.metrics.ObserveWebhookDuration(endpoint.Provider, h.now().Sub(startedAt))
	}()

	body, err := io.ReadAll(nethttp.MaxBytesReader(w, req.Body, h.maxBodyBytes))
	if err != nil {
		h.metrics.ObserveWebhook(endpoint.Provider, "invalid_payload")
		nethttp.Error(w, "invalid webhook payload", nethttp.StatusRequestEntityTooLarge)
		return
	}
	if !h.signatureValid(endpoint.SigningSecret, req.Header, body) {
		h.metrics.ObserveWebhook(endpoint.Provider, "invalid_signature")
		h.logger.WarnContext(req.Context(), "notification.webhook.authentication_failed",
			slog.String("provider", endpoint.Provider), slog.String("channel", string(channel)))
		nethttp.Error(w, "webhook authentication failed", nethttp.StatusUnauthorized)
		return
	}
	var payload ProviderWebhookRequest
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		h.metrics.ObserveWebhook(endpoint.Provider, "invalid_payload")
		nethttp.Error(w, "invalid webhook payload", nethttp.StatusBadRequest)
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		h.metrics.ObserveWebhook(endpoint.Provider, "invalid_payload")
		nethttp.Error(w, "invalid webhook payload", nethttp.StatusBadRequest)
		return
	}
	if payload.Type == domain.DeliveryEventOpened && !endpoint.AllowOpenTracking {
		h.metrics.ObserveWebhook(endpoint.Provider, "open_tracking_disabled")
		nethttp.Error(w, "open tracking is disabled", nethttp.StatusUnprocessableEntity)
		return
	}
	if payload.Type != domain.DeliveryEventFailed && strings.TrimSpace(payload.FailureCode) != "" {
		h.metrics.ObserveWebhook(endpoint.Provider, "invalid_payload")
		nethttp.Error(w, "failure code only applies to failed events", nethttp.StatusBadRequest)
		return
	}
	result, err := h.recorder.RecordCallback(req.Context(), domain.ProviderCallback{
		Provider: endpoint.Provider, ProviderEventID: strings.TrimSpace(payload.ProviderEventID),
		ProviderMessageID: strings.TrimSpace(payload.ProviderMessageID), Type: payload.Type,
		Channel: channel, OccurredAt: payload.OccurredAt.UTC(), FailureCode: payload.FailureCode,
	})
	switch {
	case err == nil && result.Duplicate:
		h.metrics.ObserveWebhook(endpoint.Provider, "duplicate")
		w.WriteHeader(nethttp.StatusNoContent)
	case err == nil:
		h.metrics.ObserveWebhook(endpoint.Provider, "accepted")
		w.WriteHeader(nethttp.StatusNoContent)
	case errors.Is(err, domain.ErrProviderEventNotFound):
		h.metrics.ObserveWebhook(endpoint.Provider, "unmatched")
		h.logger.WarnContext(req.Context(), "notification.webhook.unmatched",
			slog.String("provider", endpoint.Provider), slog.String("channel", string(channel)),
			slog.String("event_type", string(payload.Type)))
		w.WriteHeader(nethttp.StatusAccepted)
	case errors.Is(err, domain.ErrInvalidProviderEvent), errors.Is(err, domain.ErrUnsupportedChannel):
		h.metrics.ObserveWebhook(endpoint.Provider, "invalid_payload")
		nethttp.Error(w, "invalid webhook payload", nethttp.StatusBadRequest)
	default:
		h.metrics.ObserveWebhook(endpoint.Provider, "failed")
		h.logger.ErrorContext(req.Context(), "notification.webhook.processing_failed",
			slog.String("provider", endpoint.Provider), slog.String("channel", string(channel)),
			slog.String("error", err.Error()))
		nethttp.Error(w, "webhook processing unavailable", nethttp.StatusServiceUnavailable)
	}
}

func (h *ProviderWebhookHandler) signatureValid(secret string, headers nethttp.Header, body []byte) bool {
	timestampValue := strings.TrimSpace(headers.Get(TimestampHeader))
	seconds, err := strconv.ParseInt(timestampValue, 10, 64)
	if err != nil {
		return false
	}
	signedAt := time.Unix(seconds, 0)
	age := h.now().Sub(signedAt)
	if age < -h.replayWindow || age > h.replayWindow {
		return false
	}
	signature := strings.TrimSpace(strings.TrimPrefix(headers.Get(SignatureHeader), "sha256="))
	provided, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestampValue))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(body)
	return hmac.Equal(mac.Sum(nil), provided)
}
