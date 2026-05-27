package usecase

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
)

func TestEvaluateTransitionWithProviderEvent(t *testing.T) {
	uc := newTestUsecase(t)

	out, err := uc.EvaluateTransition(context.Background(), EvaluateTransitionInput{
		From:          string(domain.PaymentStatusAuthorized),
		ProviderEvent: "payment.captured",
	})
	if err != nil {
		t.Fatalf("EvaluateTransition returned error: %v", err)
	}
	if !out.Allowed {
		t.Fatal("provider captured event should be allowed from authorized")
	}
	if out.To != domain.PaymentStatusCaptured {
		t.Fatalf("to = %q, want captured", out.To)
	}
	if out.Event != domain.PaymentEventProviderCaptured {
		t.Fatalf("event = %q, want provider_captured", out.Event)
	}
	if out.SuggestedOrderStatus != domain.OrderPaymentStatusPaid {
		t.Fatalf("order status = %q, want paid", out.SuggestedOrderStatus)
	}
}

func TestEvaluateTransitionBlocksInvalidTransition(t *testing.T) {
	logs := &bytes.Buffer{}
	uc, err := NewPaymentStateUsecase(domain.NewStateMachine(), slog.New(slog.NewTextHandler(logs, nil)))
	if err != nil {
		t.Fatalf("NewPaymentStateUsecase returned error: %v", err)
	}

	out, err := uc.EvaluateTransition(context.Background(), EvaluateTransitionInput{
		From: string(domain.PaymentStatusFailed),
		To:   string(domain.PaymentStatusCaptured),
	})
	if !errors.Is(err, domain.ErrInvalidPaymentTransition) {
		t.Fatalf("EvaluateTransition error = %v, want invalid transition", err)
	}
	if out.Allowed {
		t.Fatal("failed -> captured should be blocked")
	}
	if !strings.Contains(logs.String(), "payment.state_machine.validation_failed") {
		t.Fatalf("expected validation failure log, got %q", logs.String())
	}
}

func TestEvaluateTransitionRequiresTarget(t *testing.T) {
	uc := newTestUsecase(t)

	_, err := uc.EvaluateTransition(context.Background(), EvaluateTransitionInput{
		From: string(domain.PaymentStatusInitiated),
	})
	if !errors.Is(err, ErrTransitionTargetRequired) {
		t.Fatalf("EvaluateTransition error = %v, want target required", err)
	}
}

func TestEvaluateTransitionDetectsEventTargetConflict(t *testing.T) {
	uc := newTestUsecase(t)

	_, err := uc.EvaluateTransition(context.Background(), EvaluateTransitionInput{
		From:  string(domain.PaymentStatusAuthorized),
		Event: string(domain.PaymentEventProviderCaptured),
		To:    string(domain.PaymentStatusFailed),
	})
	if !errors.Is(err, ErrTransitionTargetConflict) {
		t.Fatalf("EvaluateTransition error = %v, want target conflict", err)
	}
}

func TestNormalizeProviderEvent(t *testing.T) {
	uc := newTestUsecase(t)

	out, err := uc.NormalizeProviderEvent(context.Background(), "refund.succeeded")
	if err != nil {
		t.Fatalf("NormalizeProviderEvent returned error: %v", err)
	}
	if out.InternalEvent != domain.PaymentEventFullRefundSucceeded {
		t.Fatalf("internal event = %q, want full_refund_succeeded", out.InternalEvent)
	}
	if out.NextStatus != domain.PaymentStatusRefunded {
		t.Fatalf("next status = %q, want refunded", out.NextStatus)
	}
}

func newTestUsecase(t *testing.T) *PaymentStateUsecase {
	t.Helper()

	uc, err := NewPaymentStateUsecase(domain.NewStateMachine(), slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	if err != nil {
		t.Fatalf("NewPaymentStateUsecase returned error: %v", err)
	}
	return uc
}
