package domain

import (
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestDeliveryValidate(t *testing.T) {
	t.Parallel()

	if err := validDelivery().Validate(); err != nil {
		t.Fatalf("Validate() returned error for valid delivery: %v", err)
	}
}

func TestDeliveryValidateAllowsPreSignupOTPWithoutUser(t *testing.T) {
	t.Parallel()

	delivery := validDelivery()
	delivery.UserID = ""
	delivery.TemplateKey = string(TemplateOTPVerification)
	if err := delivery.Validate(); err != nil {
		t.Fatalf("Validate() returned error for pre-signup OTP delivery: %v", err)
	}
}

func TestDeliveryBSONOmitsEmptyPreSignupUser(t *testing.T) {
	t.Parallel()

	delivery := validDelivery()
	delivery.UserID = ""
	delivery.TemplateKey = string(TemplateOTPVerification)
	raw, err := bson.Marshal(delivery)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var document bson.M
	if err := bson.Unmarshal(raw, &document); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if _, exists := document["user_id"]; exists {
		t.Fatalf("pre-signup OTP document contains empty user_id: %#v", document["user_id"])
	}
}

func TestDeliveryValidateAllowsClaimedEventDelivery(t *testing.T) {
	t.Parallel()

	delivery := validDelivery()
	delivery.Status = DeliveryStatusProcessing
	delivery.Provider = ""
	delivery.ProviderMessageID = ""
	delivery.Attempts = 0
	delivery.IdempotencyKey = "evt_1:order_status_update:email"
	delivery.SourceEventID = "evt_1"
	delivery.SourceEventType = "OrderPaid"
	delivery.TraceID = "trace_1"
	if err := delivery.Validate(); err != nil {
		t.Fatalf("Validate() returned error for event claim: %v", err)
	}
	delivery.SourceEventID = ""
	if err := delivery.Validate(); !errors.Is(err, ErrInvalidDelivery) {
		t.Fatalf("Validate() error = %v, want ErrInvalidDelivery", err)
	}
}

func TestDeliveryValidateRejectsInvalidDocuments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		change func(*Delivery)
	}{
		{name: "missing user", change: func(delivery *Delivery) { delivery.UserID = "" }},
		{name: "unknown channel", change: func(delivery *Delivery) { delivery.Channel = "fax" }},
		{name: "missing status", change: func(delivery *Delivery) { delivery.Status = "" }},
		{name: "negative attempts", change: func(delivery *Delivery) { delivery.Attempts = -1 }},
		{name: "provider reference without provider", change: func(delivery *Delivery) { delivery.Provider = "" }},
		{name: "timestamps reversed", change: func(delivery *Delivery) { delivery.UpdatedAt = delivery.CreatedAt.Add(-time.Second) }},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			delivery := validDelivery()
			test.change(&delivery)
			if err := delivery.Validate(); !errors.Is(err, ErrInvalidDelivery) {
				t.Fatalf("Validate() error = %v, want %v", err, ErrInvalidDelivery)
			}
		})
	}
}

func TestDeliveryValidateRejectsSensitivePayloadFields(t *testing.T) {
	t.Parallel()

	tests := []map[string]any{
		{"otp": "123456"},
		{"access_token": "token"},
		{"provider": map[string]any{"api_key": "key"}},
		{"items": []any{map[string]any{"password": "secret"}}},
	}
	for _, payload := range tests {
		delivery := validDelivery()
		delivery.Payload = payload
		if err := delivery.Validate(); !errors.Is(err, ErrInvalidDelivery) {
			t.Fatalf("Validate() error = %v, want %v for payload %+v", err, ErrInvalidDelivery, payload)
		}
	}
}

func TestDeliveryValidateAllowsSuppressedConsentAudit(t *testing.T) {
	t.Parallel()

	delivery := validDelivery()
	delivery.Status = DeliveryStatusSuppressed
	delivery.Provider = ""
	delivery.ProviderMessageID = ""
	delivery.Attempts = 0
	delivery.SuppressionReason = ConsentReasonMarketingOptedOut
	if err := delivery.Validate(); err != nil {
		t.Fatalf("Validate() returned error for suppression audit: %v", err)
	}
	delivery.SuppressionReason = ""
	if err := delivery.Validate(); !errors.Is(err, ErrInvalidDelivery) {
		t.Fatalf("Validate() error = %v, want ErrInvalidDelivery", err)
	}
}

func TestDeliveryValidateAnalyticsMilestones(t *testing.T) {
	t.Parallel()

	delivery := validDelivery()
	failedAt := delivery.UpdatedAt.Add(time.Minute)
	delivery.FailedAt = &failedAt
	delivery.FailureCode = "hard_bounce"
	if err := delivery.Validate(); err != nil {
		t.Fatalf("Validate() analytics delivery error = %v", err)
	}
	delivery.FailureCode = ""
	if err := delivery.Validate(); !errors.Is(err, ErrInvalidDelivery) {
		t.Fatalf("Validate() missing failure code error = %v, want ErrInvalidDelivery", err)
	}
}

func validDelivery() Delivery {
	now := time.Date(2026, time.May, 18, 0, 0, 0, 0, time.UTC)
	return Delivery{
		ID:                "delivery_123",
		UserID:            "user_123",
		Channel:           ChannelEmail,
		TemplateKey:       "order_paid",
		Status:            DeliveryStatusAccepted,
		Provider:          "ses",
		ProviderMessageID: "msg_123",
		Attempts:          1,
		Payload:           map[string]any{"order_id": "order_123", "trace_id": "trace_123"},
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}
