package events_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/events"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
)

type triggerSender struct {
	request domain.EventNotificationTrigger
	err     error
	calls   int
}

func (s *triggerSender) TriggerFromEvent(_ context.Context, request domain.EventNotificationTrigger) error {
	s.calls++
	s.request = request
	return s.err
}

func TestHandlerRoutesSupportedEventFamilies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		body        string
		templateKey domain.TemplateKey
		variable    string
		value       string
	}{
		{
			name:        "order paid",
			body:        `{"event_id":"evt_1","event_type":"OrderPaid","version":1,"occurred_at":"2026-05-27T14:00:00Z","producer":"order-service","trace_id":"trace_1","payload":{"order_id":"order_1","user_id":"user_1","email":"buyer@example.com","name":"Riya"}}`,
			templateKey: domain.TemplateOrderStatusUpdate, variable: "status", value: "Paid",
		},
		{
			name:        "payment failed",
			body:        `{"event_id":"evt_2","event_type":"PaymentFailed","version":1,"occurred_at":"2026-05-27T14:00:00Z","producer":"payment-service","payload":{"order_id":"order_1","user_id":"user_1","email":"buyer@example.com","name":"Riya","amount":"USD 20"}}`,
			templateKey: domain.TemplatePaymentStatusUpdate, variable: "payment_status", value: "Failed",
		},
		{
			name:        "user created",
			body:        `{"event_id":"evt_3","event_type":"UserCreated","version":1,"occurred_at":"2026-05-27T14:00:00Z","producer":"user-service","payload":{"user_id":"user_1","email":"buyer@example.com","name":"Riya"}}`,
			templateKey: domain.TemplateWelcomeUser, variable: "name", value: "Riya",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			sender := &triggerSender{}
			handler := newHandler(t, sender)
			outcome, err := handler.Handle(context.Background(), []byte(test.body))
			if err != nil || outcome != events.OutcomeProcessed {
				t.Fatalf("Handle() = %q, %v", outcome, err)
			}
			if sender.calls != 1 || sender.request.TemplateKey != test.templateKey ||
				sender.request.Variables[test.variable] != test.value {
				t.Fatalf("trigger request = %+v", sender.request)
			}
		})
	}
}

func TestHandlerAcksDuplicateAndUnsupportedEventsWithoutSend(t *testing.T) {
	t.Parallel()

	duplicate := &triggerSender{err: domain.ErrDuplicateEventDelivery}
	outcome, err := newHandler(t, duplicate).Handle(context.Background(), []byte(
		`{"event_id":"evt_1","event_type":"OrderPaid","version":1,"occurred_at":"2026-05-27T14:00:00Z","producer":"order-service","payload":{"order_id":"order_1","user_id":"user_1","email":"buyer@example.com","name":"Riya"}}`,
	))
	if err != nil || outcome != events.OutcomeDuplicate || duplicate.calls != 1 {
		t.Fatalf("duplicate Handle() = %q, %v; calls = %d", outcome, err, duplicate.calls)
	}

	unsupported := &triggerSender{}
	outcome, err = newHandler(t, unsupported).Handle(context.Background(), []byte(
		`{"event_id":"evt_2","event_type":"UserRenamed","version":1,"occurred_at":"2026-05-27T14:00:00Z","producer":"user-service","payload":{}}`,
	))
	if err != nil || outcome != events.OutcomeUnsupported || unsupported.calls != 0 {
		t.Fatalf("unsupported Handle() = %q, %v; calls = %d", outcome, err, unsupported.calls)
	}
}

func TestHandlerClassifiesInvalidPermanentAndTemporaryFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		body   string
		err    error
		want   error
		called bool
	}{
		{name: "invalid envelope", body: `{"event_type":"OrderPaid"}`, want: events.ErrInvalidEvent},
		{
			name: "wrong producer",
			body: `{"event_id":"evt_1","event_type":"OrderPaid","version":1,"occurred_at":"2026-05-27T14:00:00Z","producer":"user-service","payload":{}}`,
			want: events.ErrInvalidEvent,
		},
		{
			name: "provider rejection",
			body: validOrderBody(),
			err:  provider.ErrProviderRejected, want: events.ErrNonRetryableProcessing, called: true,
		},
		{
			name: "temporary provider outage",
			body: validOrderBody(),
			err:  provider.ErrProviderUnavailable, want: events.ErrTemporaryProcessing, called: true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			sender := &triggerSender{err: test.err}
			_, err := newHandler(t, sender).Handle(context.Background(), []byte(test.body))
			if !errors.Is(err, test.want) {
				t.Fatalf("Handle() error = %v, want %v", err, test.want)
			}
			if (sender.calls == 1) != test.called {
				t.Fatalf("sender calls = %d, called want %v", sender.calls, test.called)
			}
		})
	}
}

func TestQueueNamesValidate(t *testing.T) {
	t.Parallel()

	if err := (events.QueueNames{OrderEvents: "order", PaymentEvents: "payment", UserEvents: "user"}).Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if err := (events.QueueNames{OrderEvents: "same", PaymentEvents: "same", UserEvents: "user"}).Validate(); err == nil {
		t.Fatal("Validate() returned nil for duplicate queues")
	}
}

func newHandler(t *testing.T, sender events.TriggerSender) *events.Handler {
	t.Helper()
	handler, err := events.NewHandler(sender, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	return handler
}

func validOrderBody() string {
	return `{"event_id":"evt_1","event_type":"OrderPaid","version":1,"occurred_at":"2026-05-27T14:00:00Z","producer":"order-service","payload":{"order_id":"order_1","user_id":"user_1","email":"buyer@example.com","name":"Riya"}}`
}
