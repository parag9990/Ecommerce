package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

type webhookRecorder struct {
	callback domain.ProviderCallback
	calls    int
	result   domain.DeliveryEventApplyResult
	err      error
}

func (r *webhookRecorder) RecordCallback(_ context.Context, callback domain.ProviderCallback) (domain.DeliveryEventApplyResult, error) {
	r.calls++
	r.callback = callback
	return r.result, r.err
}

type webhookObserver struct {
	outcomes []string
}

func (*webhookObserver) ObserveDeliveryEvent(domain.ProviderEvent) {}
func (o *webhookObserver) ObserveWebhook(_ string, outcome string) {
	o.outcomes = append(o.outcomes, outcome)
}
func (*webhookObserver) ObserveWebhookDuration(string, time.Duration) {}

func TestProviderWebhookHandlerAcceptsSignedDeliveryAndDuplicate(t *testing.T) {
	t.Parallel()

	recorder := &webhookRecorder{result: domain.DeliveryEventApplyResult{MilestoneChanged: true}}
	observer := &webhookObserver{}
	mux := webhookMux(t, recorder, observer, true)
	body := `{"provider_event_id":"evt_1","provider_message_id":"msg_1","type":"delivered","occurred_at":"2026-05-27T12:00:00Z"}`
	response := performWebhook(mux, body, "valid-secret", "1779883200")
	if response.Code != nethttp.StatusNoContent || recorder.callback.Provider != "smtp" ||
		recorder.callback.Channel != domain.ChannelEmail || observer.outcomes[0] != "accepted" {
		t.Fatalf("webhook response = %d callback = %+v outcomes = %+v", response.Code, recorder.callback, observer.outcomes)
	}

	recorder.result = domain.DeliveryEventApplyResult{Duplicate: true}
	response = performWebhook(mux, body, "valid-secret", "1779883200")
	if response.Code != nethttp.StatusNoContent || observer.outcomes[1] != "duplicate" {
		t.Fatalf("duplicate response = %d outcomes = %+v", response.Code, observer.outcomes)
	}
}

func TestProviderWebhookHandlerRejectsInvalidSignatureAndDisabledOpen(t *testing.T) {
	t.Parallel()

	recorder := &webhookRecorder{}
	observer := &webhookObserver{}
	mux := webhookMux(t, recorder, observer, false)
	delivered := `{"provider_event_id":"evt_1","provider_message_id":"msg_1","type":"delivered","occurred_at":"2026-05-27T12:00:00Z"}`
	response := performWebhook(mux, delivered, "wrong-secret", "1779883200")
	if response.Code != nethttp.StatusUnauthorized || recorder.calls != 0 {
		t.Fatalf("invalid signature response = %d recorder calls = %d", response.Code, recorder.calls)
	}
	response = performWebhook(mux, delivered, "valid-secret", "1779882000")
	if response.Code != nethttp.StatusUnauthorized || recorder.calls != 0 {
		t.Fatalf("stale signature response = %d recorder calls = %d", response.Code, recorder.calls)
	}
	opened := `{"provider_event_id":"evt_2","provider_message_id":"msg_1","type":"opened","occurred_at":"2026-05-27T12:00:00Z"}`
	response = performWebhook(mux, opened, "valid-secret", "1779883200")
	if response.Code != nethttp.StatusUnprocessableEntity || recorder.calls != 0 {
		t.Fatalf("disabled open response = %d recorder calls = %d", response.Code, recorder.calls)
	}
}

func webhookMux(t *testing.T, recorder *webhookRecorder, observer *webhookObserver, allowOpen bool) *nethttp.ServeMux {
	t.Helper()
	handler, err := NewProviderWebhookHandler([]WebhookEndpoint{{
		Channel: domain.ChannelEmail, Provider: "smtp", SigningSecret: "valid-secret", AllowOpenTracking: allowOpen,
	}}, recorder, observer, 4096, 5*time.Minute, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewProviderWebhookHandler() error = %v", err)
	}
	handler.WithClock(func() time.Time {
		return time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC)
	})
	mux := nethttp.NewServeMux()
	if err := handler.Register(mux, "/internal/provider-webhooks"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	return mux
}

func performWebhook(handler nethttp.Handler, body, secret, timestamp string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(nethttp.MethodPost, "/internal/provider-webhooks/email", strings.NewReader(body))
	request.Header.Set(TimestampHeader, timestamp)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "." + body))
	request.Header.Set(SignatureHeader, "sha256="+hex.EncodeToString(mac.Sum(nil)))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
