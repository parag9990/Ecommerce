package domain

import (
	"errors"
	"testing"
	"time"
)

func TestCanTransitionPayment(t *testing.T) {
	tests := []struct {
		name string
		from PaymentStatus
		to   PaymentStatus
		want bool
	}{
		{
			name: "new payment starts initiated",
			from: PaymentStatusUnspecified,
			to:   PaymentStatusInitiated,
			want: true,
		},
		{
			name: "initiated to authorized is allowed",
			from: PaymentStatusInitiated,
			to:   PaymentStatusAuthorized,
			want: true,
		},
		{
			name: "initiated to captured is allowed for immediate capture providers",
			from: PaymentStatusInitiated,
			to:   PaymentStatusCaptured,
			want: true,
		},
		{
			name: "captured to refunded is allowed",
			from: PaymentStatusCaptured,
			to:   PaymentStatusRefunded,
			want: true,
		},
		{
			name: "partial refund can receive another partial refund",
			from: PaymentStatusPartiallyRefunded,
			to:   PaymentStatusPartiallyRefunded,
			want: true,
		},
		{
			name: "failed to retry allowed is allowed",
			from: PaymentStatusFailed,
			to:   PaymentStatusRetryAllowed,
			want: true,
		},
		{
			name: "failed to captured is blocked",
			from: PaymentStatusFailed,
			to:   PaymentStatusCaptured,
			want: false,
		},
		{
			name: "refunded to captured is blocked",
			from: PaymentStatusRefunded,
			to:   PaymentStatusCaptured,
			want: false,
		},
		{
			name: "duplicate captured webhook is no-op allowed",
			from: PaymentStatusCaptured,
			to:   PaymentStatusCaptured,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanTransitionPayment(tt.from, tt.to)
			if got != tt.want {
				t.Fatalf("CanTransitionPayment(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestNextStatusForEvent(t *testing.T) {
	tests := []struct {
		name     string
		from     PaymentStatus
		event    PaymentEvent
		want     PaymentStatus
		wantNoop bool
		wantErr  error
	}{
		{
			name:  "requires action completes into authorized",
			from:  PaymentStatusRequiresAction,
			event: PaymentEventActionCompleted,
			want:  PaymentStatusAuthorized,
		},
		{
			name:  "authorized captures",
			from:  PaymentStatusAuthorized,
			event: PaymentEventProviderCaptured,
			want:  PaymentStatusCaptured,
		},
		{
			name:     "duplicate captured event is no-op",
			from:     PaymentStatusCaptured,
			event:    PaymentEventProviderCaptured,
			want:     PaymentStatusCaptured,
			wantNoop: true,
		},
		{
			name:    "failed attempt cannot become captured",
			from:    PaymentStatusFailed,
			event:   PaymentEventProviderCaptured,
			want:    PaymentStatusCaptured,
			wantErr: ErrInvalidPaymentTransition,
		},
		{
			name:    "unknown event is rejected",
			from:    PaymentStatusInitiated,
			event:   PaymentEvent("provider_settled"),
			wantErr: ErrInvalidPaymentEvent,
		},
	}

	machine := NewStateMachine()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, noop, err := machine.NextStatusForEvent(tt.from, tt.event)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NextStatusForEvent error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("NextStatusForEvent status = %q, want %q", got, tt.want)
			}
			if noop != tt.wantNoop {
				t.Fatalf("NextStatusForEvent noop = %v, want %v", noop, tt.wantNoop)
			}
		})
	}
}

func TestProviderEventNormalization(t *testing.T) {
	mapping, err := NormalizeProviderEvent(" PAYMENT_INTENT.SUCCEEDED ")
	if err != nil {
		t.Fatalf("NormalizeProviderEvent returned error: %v", err)
	}
	if mapping.InternalEvent != PaymentEventProviderCaptured {
		t.Fatalf("internal event = %q, want %q", mapping.InternalEvent, PaymentEventProviderCaptured)
	}
	if mapping.NextStatus != PaymentStatusCaptured {
		t.Fatalf("next status = %q, want %q", mapping.NextStatus, PaymentStatusCaptured)
	}

	if _, err := NormalizeProviderEvent("invoice.paid"); !errors.Is(err, ErrUnknownProviderEvent) {
		t.Fatalf("NormalizeProviderEvent unknown error = %v, want %v", err, ErrUnknownProviderEvent)
	}
}

func TestSuggestedOrderStatus(t *testing.T) {
	tests := []struct {
		paymentStatus PaymentStatus
		want          OrderPaymentStatus
	}{
		{paymentStatus: PaymentStatusInitiated, want: OrderPaymentStatusPendingPayment},
		{paymentStatus: PaymentStatusAuthorized, want: OrderPaymentStatusPaymentAuthorized},
		{paymentStatus: PaymentStatusCaptured, want: OrderPaymentStatusPaid},
		{paymentStatus: PaymentStatusFailed, want: OrderPaymentStatusPaymentFailed},
		{paymentStatus: PaymentStatusPartiallyRefunded, want: OrderPaymentStatusPartiallyRefunded},
		{paymentStatus: PaymentStatusRefunded, want: OrderPaymentStatusRefunded},
	}

	for _, tt := range tests {
		t.Run(string(tt.paymentStatus), func(t *testing.T) {
			got, ok := SuggestedOrderStatus(tt.paymentStatus)
			if !ok {
				t.Fatalf("SuggestedOrderStatus(%q) not found", tt.paymentStatus)
			}
			if got != tt.want {
				t.Fatalf("SuggestedOrderStatus(%q) = %q, want %q", tt.paymentStatus, got, tt.want)
			}
		})
	}
}

func TestPaymentApplyEventUpdatesStatus(t *testing.T) {
	payment, err := NewPayment(NewPaymentInput{
		PaymentID: "pay_123",
		OrderID:   "ord_123",
		UserID:    "usr_123",
		Provider:  "stripe_like",
		Amount:    Money{Amount: 129900, Currency: "inr"},
		Now:       time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewPayment returned error: %v", err)
	}

	decision, err := payment.ApplyEvent(PaymentEventProviderRequiresAction, time.Date(2026, 5, 26, 10, 1, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ApplyEvent requires_action returned error: %v", err)
	}
	if !decision.Allowed || payment.Status != PaymentStatusRequiresAction {
		t.Fatalf("status = %q decision=%+v, want requires_action allowed", payment.Status, decision)
	}

	decision, err = payment.ApplyEvent(PaymentEventActionCompleted, time.Date(2026, 5, 26, 10, 2, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ApplyEvent action_completed returned error: %v", err)
	}
	if payment.Status != PaymentStatusAuthorized {
		t.Fatalf("status = %q, want authorized", payment.Status)
	}

	decision, err = payment.ApplyEvent(PaymentEventActionCompleted, time.Date(2026, 5, 26, 10, 3, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ApplyEvent duplicate action_completed returned error: %v", err)
	}
	if !decision.Allowed || !decision.Noop {
		t.Fatalf("duplicate action_completed from authorized should be an idempotent no-op, got %+v", decision)
	}
}

func TestPersistedStatusesExcludeRetryAllowed(t *testing.T) {
	if IsPersistedPaymentStatus(PaymentStatusRetryAllowed) {
		t.Fatal("retry_allowed should be a helper state and not persisted")
	}
	if !IsPersistedPaymentStatus(PaymentStatusCaptured) {
		t.Fatal("captured should be persisted")
	}
}
