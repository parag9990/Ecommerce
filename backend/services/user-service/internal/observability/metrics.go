package observability

import (
	"context"
	"log/slog"
	"time"

	"github.com/parag/ecommerce/backend/services/user-service/internal/events"
	platformmetrics "github.com/parag/ecommerce/backend/shared/platform/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const serviceName = "user-service"

type Metrics struct {
	Registry       *prometheus.Registry
	GRPC           *platformmetrics.GRPC
	outboxCreated  *prometheus.CounterVec
	publishTotal   *prometheus.CounterVec
	publishLatency *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)

	metrics := &Metrics{
		Registry: registry,
		GRPC:     platformmetrics.NewGRPC(registry, "ecommerce", serviceName),
		outboxCreated: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "ecommerce",
			Subsystem: "user_outbox",
			Name:      "events_created_total",
			Help:      "Total number of User Service outbox events created.",
		}, []string{"event_type"}),
		publishTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "ecommerce",
			Subsystem: "user_events",
			Name:      "publish_total",
			Help:      "Total number of User Service broker publish attempts.",
		}, []string{"event_type", "result"}),
		publishLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "ecommerce",
			Subsystem: "user_events",
			Name:      "publish_duration_seconds",
			Help:      "Duration of User Service broker publish attempts.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"event_type"}),
	}
	registry.MustRegister(metrics.outboxCreated, metrics.publishTotal, metrics.publishLatency)
	return metrics
}

func (m *Metrics) RecordEnqueued(eventType string) {
	m.outboxCreated.WithLabelValues(eventType).Inc()
}

func (m *Metrics) ObservePublish(eventType string, result string, duration time.Duration) {
	m.publishTotal.WithLabelValues(eventType, result).Inc()
	m.publishLatency.WithLabelValues(eventType).Observe(duration.Seconds())
}

func (m *Metrics) RegisterOutboxStats(repo events.OutboxStatsRepository, logger *slog.Logger) {
	if repo == nil {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}
	m.Registry.MustRegister(newOutboxCollector(repo, logger))
}

type outboxCollector struct {
	repo      events.OutboxStatsRepository
	logger    *slog.Logger
	pending   *prometheus.Desc
	oldestAge *prometheus.Desc
}

func newOutboxCollector(repo events.OutboxStatsRepository, logger *slog.Logger) *outboxCollector {
	return &outboxCollector{
		repo:   repo,
		logger: logger,
		pending: prometheus.NewDesc(
			"ecommerce_user_outbox_pending",
			"Current number of pending or failed User Service outbox events.",
			nil,
			nil,
		),
		oldestAge: prometheus.NewDesc(
			"ecommerce_user_outbox_oldest_age_seconds",
			"Age in seconds of the oldest pending or failed User Service outbox event.",
			nil,
			nil,
		),
	}
}

func (c *outboxCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.pending
	ch <- c.oldestAge
}

func (c *outboxCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stats, err := c.repo.Stats(ctx)
	if err != nil {
		c.logger.Warn("user_outbox_metrics_collection_failed", slog.String("error_type", "outbox_stats"))
		return
	}
	ch <- prometheus.MustNewConstMetric(c.pending, prometheus.GaugeValue, float64(stats.Pending))
	ch <- prometheus.MustNewConstMetric(c.oldestAge, prometheus.GaugeValue, stats.OldestPendingAge.Seconds())
}
