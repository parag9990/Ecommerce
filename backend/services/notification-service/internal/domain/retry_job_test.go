package domain

import (
	"errors"
	"testing"
	"time"
)

func TestRetryJobValidate(t *testing.T) {
	t.Parallel()

	job := RetryJob{
		DeliveryID: "delivery_1", IdempotencyKey: "evt_1:welcome_user:email",
		Attempt: 1, MaxAttempts: 4, QueuedAt: time.Now().UTC(),
	}
	if err := job.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	job.Attempt = 5
	if err := job.Validate(); !errors.Is(err, ErrInvalidRetryJob) {
		t.Fatalf("Validate() error = %v, want ErrInvalidRetryJob", err)
	}
}

func TestDeliveryRejectsDurableOTPRetryMetadata(t *testing.T) {
	t.Parallel()

	delivery := validDelivery()
	delivery.TemplateKey = string(TemplateOTPVerification)
	delivery.UserID = ""
	delivery.IdempotencyKey = "otp-is-not-queueable"
	delivery.SourceEventID = "event"
	delivery.SourceEventType = "OTP"
	delivery.RecipientCiphertext = "ciphertext"
	delivery.MaxAttempts = 4
	if err := delivery.Validate(); !errors.Is(err, ErrDurableOTPRetryForbidden) {
		t.Fatalf("Validate() error = %v, want ErrDurableOTPRetryForbidden", err)
	}
}

func TestDeliveryAllowsEncryptedQueuedEventInstruction(t *testing.T) {
	t.Parallel()

	delivery := validDelivery()
	delivery.Status = DeliveryStatusPending
	delivery.Provider = ""
	delivery.ProviderMessageID = ""
	delivery.Attempts = 0
	delivery.IdempotencyKey = "evt_1:order_status_update:email"
	delivery.SourceEventID = "evt_1"
	delivery.SourceEventType = "OrderPaid"
	delivery.RecipientCiphertext = "encrypted-recipient"
	delivery.MaxAttempts = 4
	if err := delivery.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
