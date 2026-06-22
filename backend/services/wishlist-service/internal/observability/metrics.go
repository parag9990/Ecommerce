package observability

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry                 *prometheus.Registry
	productEvents            *prometheus.CounterVec
	productEventFailures     *prometheus.CounterVec
	availabilityUpdates      *prometheus.CounterVec
	priceDropCandidates      prometheus.Counter
	priceDropNotifications   *prometheus.CounterVec
	priceSnapshotUpdates     *prometheus.CounterVec
	productProcessingSeconds *prometheus.HistogramVec
	productEventLagSeconds   *prometheus.HistogramVec
	outboxBatchSize          prometheus.Histogram
	outboxPublished          *prometheus.CounterVec
	outboxPublishFailures    *prometheus.CounterVec
	outboxPublishSeconds     *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	registry := prometheus.NewRegistry()
	m := &Metrics{
		registry:                 registry,
		productEvents:            prometheus.NewCounterVec(prometheus.CounterOpts{Name: "wishlist_product_events_total", Help: "Product events handled by Wishlist Service."}, []string{"event_type", "result"}),
		productEventFailures:     prometheus.NewCounterVec(prometheus.CounterOpts{Name: "wishlist_product_event_failures_total", Help: "Product event failures by reason."}, []string{"reason"}),
		availabilityUpdates:      prometheus.NewCounterVec(prometheus.CounterOpts{Name: "wishlist_availability_items_updated_total", Help: "Wishlist items updated from product availability events."}, []string{"availability"}),
		priceDropCandidates:      prometheus.NewCounter(prometheus.CounterOpts{Name: "wishlist_price_drop_candidates_total", Help: "Wishlist price-drop candidates evaluated."}),
		priceDropNotifications:   prometheus.NewCounterVec(prometheus.CounterOpts{Name: "wishlist_price_drop_notifications_total", Help: "Price-drop notification commands by result."}, []string{"result"}),
		priceSnapshotUpdates:     prometheus.NewCounterVec(prometheus.CounterOpts{Name: "wishlist_price_snapshot_updates_total", Help: "Wishlist price snapshots updated by result."}, []string{"result"}),
		productProcessingSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "wishlist_product_event_processing_seconds", Help: "Product event processing latency."}, []string{"event_type"}),
		productEventLagSeconds:   prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "wishlist_product_event_lag_seconds", Help: "Delay between product event occurrence and processing."}, []string{"topic"}),
		outboxBatchSize:          prometheus.NewHistogram(prometheus.HistogramOpts{Name: "wishlist_outbox_batch_size", Help: "Number of analytics events claimed per poll.", Buckets: []float64{0, 1, 5, 10, 25, 50, 100}}),
		outboxPublished:          prometheus.NewCounterVec(prometheus.CounterOpts{Name: "wishlist_outbox_published_total", Help: "Analytics events published from the outbox."}, []string{"event_type"}),
		outboxPublishFailures:    prometheus.NewCounterVec(prometheus.CounterOpts{Name: "wishlist_outbox_publish_failures_total", Help: "Analytics event publish failures."}, []string{"event_type"}),
		outboxPublishSeconds:     prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "wishlist_outbox_publish_seconds", Help: "Analytics event publish latency."}, []string{"event_type"}),
	}
	registry.MustRegister(
		prometheus.NewGoCollector(), prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		m.productEvents, m.productEventFailures, m.availabilityUpdates, m.priceDropCandidates,
		m.priceDropNotifications, m.priceSnapshotUpdates, m.productProcessingSeconds,
		m.productEventLagSeconds, m.outboxBatchSize, m.outboxPublished,
		m.outboxPublishFailures, m.outboxPublishSeconds,
	)
	return m
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *Metrics) RecordConsumed(_ context.Context, eventType, result string) {
	m.productEvents.WithLabelValues(eventType, result).Inc()
}

func (m *Metrics) RecordFailure(_ context.Context, reason string) {
	m.productEventFailures.WithLabelValues(reason).Inc()
}
func (m *Metrics) RecordAvailabilityUpdate(_ context.Context, availability string, count int64) {
	m.availabilityUpdates.WithLabelValues(availability).Add(nonNegative(count))
}
func (m *Metrics) RecordPriceDropCandidates(_ context.Context, count int64) {
	m.priceDropCandidates.Add(nonNegative(count))
}
func (m *Metrics) RecordPriceDropNotifications(_ context.Context, result string, count int64) {
	m.priceDropNotifications.WithLabelValues(result).Add(nonNegative(count))
}
func (m *Metrics) RecordPriceSnapshotUpdate(_ context.Context, result string, count int64) {
	m.priceSnapshotUpdates.WithLabelValues(result).Add(nonNegative(count))
}
func (m *Metrics) RecordProcessingDuration(_ context.Context, eventType string, duration time.Duration) {
	m.productProcessingSeconds.WithLabelValues(eventType).Observe(nonNegativeDuration(duration))
}
func (m *Metrics) RecordEventLag(_ context.Context, topic string, lag time.Duration) {
	m.productEventLagSeconds.WithLabelValues(topic).Observe(nonNegativeDuration(lag))
}
func (m *Metrics) RecordBatchSize(_ context.Context, size int) {
	m.outboxBatchSize.Observe(nonNegative(int64(size)))
}
func (m *Metrics) RecordPublished(_ context.Context, eventType string) {
	m.outboxPublished.WithLabelValues(eventType).Inc()
}
func (m *Metrics) RecordPublishFailure(_ context.Context, eventType string) {
	m.outboxPublishFailures.WithLabelValues(eventType).Inc()
}
func (m *Metrics) RecordPublishDuration(_ context.Context, eventType string, duration time.Duration) {
	m.outboxPublishSeconds.WithLabelValues(eventType).Observe(nonNegativeDuration(duration))
}

func nonNegative(value int64) float64 {
	if value < 0 {
		return 0
	}
	return float64(value)
}

func nonNegativeDuration(value time.Duration) float64 {
	if value < 0 {
		return 0
	}
	return value.Seconds()
}
