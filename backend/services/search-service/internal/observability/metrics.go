package observability

import (
	"context"
	"strconv"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/events"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/usecase"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	searchRequests       *prometheus.CounterVec
	searchDuration       *prometheus.HistogramVec
	autocompleteRequests *prometheus.CounterVec
	autocompleteDuration prometheus.Histogram
	zeroResults          *prometheus.CounterVec
	reindexRuns          *prometheus.CounterVec
	reindexProducts      *prometheus.CounterVec
	productEvents        *prometheus.CounterVec
	productEventLag      *prometheus.HistogramVec
	productIndexDuration *prometheus.HistogramVec
}

func NewMetrics(registerer prometheus.Registerer) *Metrics {
	if registerer == nil {
		registerer = prometheus.DefaultRegisterer
	}
	m := &Metrics{
		searchRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "search_service_requests_total", Help: "Search requests by outcome and zero-result status.",
		}, []string{"outcome", "sort", "zero_results"}),
		searchDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "search_service_request_duration_seconds", Help: "Search request latency by stage.", Buckets: prometheus.DefBuckets,
		}, []string{"stage"}),
		autocompleteRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "search_service_autocomplete_requests_total", Help: "Autocomplete requests by outcome and cache result.",
		}, []string{"outcome", "cache_hit"}),
		autocompleteDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "search_service_autocomplete_duration_seconds", Help: "Autocomplete request latency.", Buckets: prometheus.DefBuckets,
		}),
		zeroResults: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "search_service_zero_result_events_total", Help: "Zero-result analytics outcomes.",
		}, []string{"outcome", "reason"}),
		reindexRuns: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "search_service_reindex_runs_total", Help: "Catalog reindex runs by mode, status, and stage.",
		}, []string{"mode", "status", "stage", "error_code"}),
		reindexProducts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "search_service_reindex_products_total", Help: "Products processed by reindex outcome.",
		}, []string{"outcome"}),
		productEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "search_service_product_events_total", Help: "Product indexing events by type and outcome.",
		}, []string{"event_type", "outcome"}),
		productEventLag: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "search_service_product_event_lag_seconds", Help: "Lag from product event occurrence to Search consumption.", Buckets: prometheus.ExponentialBuckets(0.1, 2, 14),
		}, []string{"event_type"}),
		productIndexDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "search_service_product_index_duration_seconds", Help: "Typesense product indexing duration.", Buckets: prometheus.DefBuckets,
		}, []string{"event_type", "outcome"}),
	}
	registerer.MustRegister(m.searchRequests, m.searchDuration, m.autocompleteRequests, m.autocompleteDuration, m.zeroResults, m.reindexRuns, m.reindexProducts, m.productEvents, m.productEventLag, m.productIndexDuration)
	return m
}

var _ events.ProductConsumerMetricsRecorder = (*Metrics)(nil)

func (m *Metrics) RecordProductEvent(_ context.Context, eventType string, outcome string) {
	if m != nil {
		m.productEvents.WithLabelValues(eventType, outcome).Inc()
	}
}

func (m *Metrics) ObserveProductEventLag(_ context.Context, eventType string, lag time.Duration) {
	if m != nil {
		m.productEventLag.WithLabelValues(eventType).Observe(lag.Seconds())
	}
}

func (m *Metrics) ObserveProductIndexDuration(_ context.Context, eventType string, outcome string, duration time.Duration) {
	if m != nil {
		m.productIndexDuration.WithLabelValues(eventType, outcome).Observe(duration.Seconds())
	}
}

func (m *Metrics) RecordSearch(_ context.Context, value usecase.SearchMetrics) {
	if m == nil {
		return
	}
	outcome := "success"
	if value.ErrorCode != "" {
		outcome = value.ErrorCode
	}
	m.searchRequests.WithLabelValues(outcome, value.Sort, strconv.FormatBool(value.ZeroResults)).Inc()
	m.searchDuration.WithLabelValues("total").Observe(milliseconds(value.DurationMS))
	m.searchDuration.WithLabelValues("typesense").Observe(milliseconds(value.TypesenseMS))
	m.searchDuration.WithLabelValues("hydration").Observe(milliseconds(value.HydrationMS))
}

func (m *Metrics) RecordAutocomplete(_ context.Context, value usecase.AutocompleteMetrics) {
	if m == nil {
		return
	}
	outcome := "success"
	if value.ErrorCode != "" {
		outcome = value.ErrorCode
	}
	m.autocompleteRequests.WithLabelValues(outcome, strconv.FormatBool(value.CacheHit)).Inc()
	m.autocompleteDuration.Observe(milliseconds(value.DurationMS))
}

func (m *Metrics) RecordZeroResult(_ context.Context, value usecase.ZeroResultMetrics) {
	if m != nil {
		m.zeroResults.WithLabelValues(value.Outcome, value.Reason).Inc()
	}
}

func (m *Metrics) RecordReindexRun(_ context.Context, value usecase.ReindexMetrics) {
	if m == nil {
		return
	}
	m.reindexRuns.WithLabelValues(value.Mode, value.Status, value.Stage, value.ErrorCode).Inc()
	m.reindexProducts.WithLabelValues("read").Add(float64(value.ProductsRead))
	m.reindexProducts.WithLabelValues("indexed").Add(float64(value.ProductsIndexed))
	m.reindexProducts.WithLabelValues("skipped").Add(float64(value.ProductsSkipped))
}

func milliseconds(value int) float64 {
	if value <= 0 {
		return 0
	}
	return float64(value) / 1000
}
