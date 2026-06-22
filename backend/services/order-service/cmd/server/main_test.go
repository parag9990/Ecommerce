package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/usecase"
)

type fakePaymentResultExecutor struct {
	command usecase.ApplyPaymentResultCommand
}

func (f *fakePaymentResultExecutor) Execute(_ context.Context, command usecase.ApplyPaymentResultCommand) error {
	f.command = command
	return nil
}

func TestPaymentEventIngressAuthenticatesAndAppliesCapture(t *testing.T) {
	executor := &fakePaymentResultExecutor{}
	handler := serviceMux(nil, executor, "events-token", slog.New(slog.NewTextHandler(io.Discard, nil)))
	body := `{"topic":"payment.events","event":{"event_id":"evt_1","event_type":"PaymentCaptured","payment_id":"pay_1","order_id":"ord_1","provider":"stripe_like","amount":1200,"currency":"INR","occurred_at":"2026-06-21T08:00:00Z"}}`
	request := httptest.NewRequest(http.MethodPost, "/internal/v1/payment-events", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer events-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if executor.command.Result != "captured" || executor.command.OrderID != "ord_1" || executor.command.ProviderEventID != "evt_1" {
		t.Fatalf("unexpected payment result command: %+v", executor.command)
	}
}

func TestPaymentEventIngressRejectsMissingAuthorization(t *testing.T) {
	handler := serviceMux(nil, &fakePaymentResultExecutor{}, "events-token", slog.Default())
	request := httptest.NewRequest(http.MethodPost, "/internal/v1/payment-events", strings.NewReader(`{}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestPaymentEventIngressRejectsTrailingJSON(t *testing.T) {
	handler := serviceMux(nil, &fakePaymentResultExecutor{}, "events-token", slog.Default())
	body := `{"topic":"payment.events","event":{"event_id":"evt_1","event_type":"PaymentCaptured","payment_id":"pay_1","order_id":"ord_1","amount":1200,"currency":"INR","occurred_at":"2026-06-21T08:00:00Z"}} {}`
	request := httptest.NewRequest(http.MethodPost, "/internal/v1/payment-events", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer events-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want %d", response.Code, http.StatusBadRequest)
	}
}
