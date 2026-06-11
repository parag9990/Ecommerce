package domain

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestPaymentSchemaDefinition(t *testing.T) {
	schema := PaymentSchema()
	if schema.DatabaseName != PaymentDatabaseName {
		t.Fatalf("database name = %q, want %q", schema.DatabaseName, PaymentDatabaseName)
	}
	if len(schema.Tables) != 5 {
		t.Fatalf("tables len = %d, want 5", len(schema.Tables))
	}

	payments, ok := findSchemaTable(schema, "payments")
	if !ok {
		t.Fatal("payments table definition not found")
	}
	if !contains(payments.UniqueKeys, "uk_payments_idempotency") {
		t.Fatalf("payments unique keys = %+v, want uk_payments_idempotency", payments.UniqueKeys)
	}
	if !contains(payments.Indexes, "idx_payments_provider_payment") {
		t.Fatalf("payments indexes = %+v, want idx_payments_provider_payment", payments.Indexes)
	}
	if !contains(payments.UniqueKeys, "uk_payments_retry_request") || !contains(payments.Indexes, "idx_payments_root_status") ||
		!contains(payments.ForeignKeys, "fk_payments_retry_of") {
		t.Fatalf("payments retry schema = %+v, want retry uniqueness/index/foreign key", payments)
	}

	refunds, ok := findSchemaTable(schema, "refunds")
	if !ok {
		t.Fatal("refunds table definition not found")
	}
	if !contains(refunds.ForeignKeys, "fk_refunds_payment") {
		t.Fatalf("refund foreign keys = %+v, want fk_refunds_payment", refunds.ForeignKeys)
	}
	if !contains(refunds.Columns, "review_reason") || !contains(refunds.Indexes, "idx_refunds_provider_refund") {
		t.Fatalf("refund schema = %+v, want review_reason and provider refund lookup index", refunds)
	}
}

func TestPaymentValidateForPersistence(t *testing.T) {
	payment := Payment{
		PaymentID:      " pay_123 ",
		OrderID:        "ord_123",
		UserID:         "usr_123",
		Provider:       "Stripe_Like",
		Amount:         Money{Amount: 129900, Currency: "inr"},
		Status:         PaymentStatusInitiated,
		IdempotencyKey: " payment_intent:ord_123:attempt_1 ",
		AttemptNo:      1,
	}
	if err := payment.ValidateForPersistence(); err != nil {
		t.Fatalf("ValidateForPersistence returned error: %v", err)
	}

	payment.IdempotencyKey = ""
	if err := payment.ValidateForPersistence(); !errors.Is(err, ErrInvalidPayment) {
		t.Fatalf("missing idempotency error = %v, want invalid payment", err)
	}

	payment.IdempotencyKey = "payment_intent:ord_123:attempt_1"
	payment.Status = PaymentStatusRetryAllowed
	if err := payment.ValidateForPersistence(); !errors.Is(err, ErrInvalidPayment) {
		t.Fatalf("retry_allowed persistence error = %v, want invalid payment", err)
	}
}

func TestSchemaJSONPayloadValidationRejectsSensitiveKeys(t *testing.T) {
	attempt := PaymentAttempt{
		AttemptID:           "pat_123",
		PaymentID:           "pay_123",
		Status:              PaymentAttemptStatusInitiated,
		RawProviderResponse: json.RawMessage(`{"provider_status":"failed","cvv":"123"}`),
	}
	if err := attempt.Validate(); !errors.Is(err, ErrInvalidPayment) {
		t.Fatalf("sensitive raw provider response error = %v, want invalid payment", err)
	}

	processedAt := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	event := PaymentWebhookEvent{
		WebhookEventID:  "wh_123",
		Provider:        "stripe_like",
		ProviderEventID: "evt_123",
		EventType:       "payment.captured",
		Processed:       true,
		ProcessedAt:     &processedAt,
		Payload:         json.RawMessage(`{"provider_payment_id":"pi_123"}`),
	}
	if err := event.Validate(); err != nil {
		t.Fatalf("PaymentWebhookEvent.Validate returned error: %v", err)
	}
}

func findSchemaTable(schema PaymentSchemaDefinition, name string) (PaymentTableDefinition, bool) {
	for _, table := range schema.Tables {
		if table.Name == name {
			return table, true
		}
	}
	return PaymentTableDefinition{}, false
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
