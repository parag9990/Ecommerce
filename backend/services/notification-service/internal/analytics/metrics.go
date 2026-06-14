package analytics

import (
	"fmt"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
)

type NoopObserver struct{}

func (NoopObserver) ObserveDeliveryEvent(domain.ProviderEvent)    {}
func (NoopObserver) ObserveWebhook(string, string)                {}
func (NoopObserver) ObserveWebhookDuration(string, time.Duration) {}

type PrometheusObserver struct {
	events   *prometheus.CounterVec
	webhooks *prometheus.CounterVec
	duration *prometheus.HistogramVec
	lag      *prometheus.HistogramVec
}

func NewPrometheusObserver(registerer prometheus.Registerer) (*PrometheusObserver, error) {
	if registerer == nil {
		return nil, fmt.Errorf("Prometheus registerer is required")
	}
	observer := &PrometheusObserver{
		events: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "notification_delivery_events_total",
			Help: "Count of deduplicated notification delivery milestone events.",
		}, []string{"event", "channel", "provider", "template_key"}),
		webhooks: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "notification_provider_webhooks_total",
			Help: "Count of authenticated provider webhook handling outcomes.",
		}, []string{"provider", "outcome"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "notification_provider_webhook_processing_seconds",
			Help:    "Provider webhook handling duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"provider"}),
		lag: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "notification_delivery_event_lag_seconds",
			Help:    "Seconds between provider event occurrence and service ingestion.",
			Buckets: []float64{1, 5, 15, 60, 300, 1800},
		}, []string{"event", "channel", "provider"}),
	}
	for _, collector := range []prometheus.Collector{observer.events, observer.webhooks, observer.duration, observer.lag} {
		if err := registerer.Register(collector); err != nil {
			return nil, fmt.Errorf("register notification metric: %w", err)
		}
	}
	return observer, nil
}

func (m *PrometheusObserver) ObserveDeliveryEvent(event domain.ProviderEvent) {
	m.events.WithLabelValues(string(event.Type), string(event.Channel), event.Provider, event.TemplateKey).Inc()
	lag := event.ReceivedAt.Sub(event.OccurredAt).Seconds()
	if lag < 0 {
		lag = 0
	}
	m.lag.WithLabelValues(string(event.Type), string(event.Channel), event.Provider).Observe(lag)
}

func (m *PrometheusObserver) ObserveWebhook(provider, outcome string) {
	m.webhooks.WithLabelValues(provider, outcome).Inc()
}

func (m *PrometheusObserver) ObserveWebhookDuration(provider string, duration time.Duration) {
	m.duration.WithLabelValues(provider).Observe(duration.Seconds())
}
