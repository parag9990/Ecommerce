package analytics

import (
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
)

func TestPrometheusObserverExportsDeliveryAndWebhookMetrics(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	observer, err := NewPrometheusObserver(registry)
	if err != nil {
		t.Fatalf("NewPrometheusObserver() error = %v", err)
	}
	event := analyticsEvent()
	observer.ObserveDeliveryEvent(event)
	observer.ObserveWebhook(event.Provider, "accepted")
	observer.ObserveWebhookDuration(event.Provider, 10*time.Millisecond)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v", err)
	}
	found := map[string]bool{}
	for _, metric := range metrics {
		found[metric.GetName()] = true
	}
	for _, expected := range []string{
		"notification_delivery_events_total",
		"notification_delivery_event_lag_seconds",
		"notification_provider_webhooks_total",
		"notification_provider_webhook_processing_seconds",
	} {
		if !found[expected] {
			t.Fatalf("Gather() missing metric %q in %+v", expected, found)
		}
	}
}

func analyticsEvent() domain.ProviderEvent {
	now := time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC)
	return domain.ProviderEvent{
		ID: "event_1", Provider: "smtp", ProviderEventID: "evt_1", ProviderMessageID: "msg_1",
		DeliveryID: "delivery_1", Type: domain.DeliveryEventDelivered, Channel: domain.ChannelEmail,
		TemplateKey: "order_status_update", OccurredAt: now, ReceivedAt: now.Add(time.Second),
	}
}
