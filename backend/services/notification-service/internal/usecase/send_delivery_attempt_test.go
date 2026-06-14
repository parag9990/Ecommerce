package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/usecase"
)

type retryDeliveryLookup struct {
	delivery domain.Delivery
}

func (l *retryDeliveryLookup) FindDeliveryByID(context.Context, string) (domain.Delivery, error) {
	return l.delivery, nil
}

func TestDeliveryAttemptRendersStoredInstructionAndSends(t *testing.T) {
	t.Parallel()

	sender := &otpSender{channel: domain.ChannelEmail, result: provider.Result{
		Status: provider.StatusAccepted, ProviderName: "smtp", ProviderMessageID: "message_1",
	}}
	service, err := usecase.NewDeliveryAttemptService(
		&otpRenderer{result: domain.RenderedMessage{Subject: "Paid", Body: "Your order is paid."}},
		&otpRegistry{sender: sender},
		&retryDeliveryLookup{delivery: retryableDelivery()},
		&recipientProtector{revealed: "buyer@example.com"},
		allowedConsentGate(t),
	)
	if err != nil {
		t.Fatalf("NewDeliveryAttemptService() error = %v", err)
	}
	result, err := service.SendDeliveryAttempt(context.Background(), "delivery_1")
	if err != nil || result.ProviderMessageID != "message_1" || sender.calls != 1 ||
		sender.message.Recipient.Email != "buyer@example.com" {
		t.Fatalf("SendDeliveryAttempt() = %+v, %v; message = %+v", result, err, sender.message)
	}
}

func TestDeliveryAttemptNeverReplaysOTP(t *testing.T) {
	t.Parallel()

	delivery := retryableDelivery()
	delivery.TemplateKey = string(domain.TemplateOTPVerification)
	sender := &otpSender{channel: domain.ChannelEmail}
	service, err := usecase.NewDeliveryAttemptService(
		&otpRenderer{}, &otpRegistry{sender: sender},
		&retryDeliveryLookup{delivery: delivery},
		&recipientProtector{revealed: "buyer@example.com"},
		allowedConsentGate(t),
	)
	if err != nil {
		t.Fatalf("NewDeliveryAttemptService() error = %v", err)
	}
	_, err = service.SendDeliveryAttempt(context.Background(), delivery.ID)
	if !errors.Is(err, domain.ErrDurableOTPRetryForbidden) || sender.calls != 0 {
		t.Fatalf("SendDeliveryAttempt() error = %v, sends = %d", err, sender.calls)
	}
}

func TestDeliveryAttemptSuppressesOptedOutDeliveryBeforeRendering(t *testing.T) {
	t.Parallel()

	renderer := &otpRenderer{}
	sender := &otpSender{channel: domain.ChannelEmail}
	gate := newConsentGate(t, &preferenceReader{preference: domain.Preference{}})
	service, err := usecase.NewDeliveryAttemptService(
		renderer, &otpRegistry{sender: sender}, &retryDeliveryLookup{delivery: retryableDelivery()},
		&recipientProtector{revealed: "buyer@example.com"}, gate,
	)
	if err != nil {
		t.Fatalf("NewDeliveryAttemptService() error = %v", err)
	}
	_, err = service.SendDeliveryAttempt(context.Background(), "delivery_1")
	if !errors.Is(err, domain.ErrDeliverySuppressed) || renderer.calls != 0 || sender.calls != 0 {
		t.Fatalf("SendDeliveryAttempt() error = %v, renders = %d, sends = %d", err, renderer.calls, sender.calls)
	}
}

func TestDeliveryAttemptPreservesConfiguredProviderOnUnavailableOutcome(t *testing.T) {
	t.Parallel()

	sender := &otpSender{channel: domain.ChannelEmail, err: provider.ErrProviderUnavailable}
	service, err := usecase.NewDeliveryAttemptService(
		&otpRenderer{result: domain.RenderedMessage{Subject: "Paid", Body: "Your order is paid."}},
		&otpRegistry{sender: sender},
		&retryDeliveryLookup{delivery: retryableDelivery()},
		&recipientProtector{revealed: "buyer@example.com"},
		allowedConsentGate(t),
	)
	if err != nil {
		t.Fatalf("NewDeliveryAttemptService() error = %v", err)
	}
	result, sendErr := service.SendDeliveryAttempt(context.Background(), "delivery_1")
	if !errors.Is(sendErr, provider.ErrProviderUnavailable) || result.ProviderName != sender.Name() {
		t.Fatalf("SendDeliveryAttempt() = %+v, %v", result, sendErr)
	}
}

func retryableDelivery() domain.Delivery {
	now := time.Date(2026, time.May, 27, 14, 0, 0, 0, time.UTC)
	return domain.Delivery{
		ID: "delivery_1", UserID: "user_1", Channel: domain.ChannelEmail,
		TemplateKey: string(domain.TemplateOrderStatusUpdate), Status: domain.DeliveryStatusProcessing,
		IdempotencyKey: "evt_1:order_status_update:email", SourceEventID: "evt_1",
		SourceEventType: "OrderPaid", RecipientCiphertext: "encrypted-recipient", MaxAttempts: 4,
		ProcessingAttempt: 1, ProcessingLeaseUntil: &now, CreatedAt: now, UpdatedAt: now,
		Payload: map[string]any{"name": "Riya", "order_id": "order_1", "status": "Paid"},
	}
}
