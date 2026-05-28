package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type ConsumerMetrics struct {
	consumed                  *prometheus.CounterVec
	processingDuration        *prometheus.HistogramVec
	dlq                       *prometheus.CounterVec
	duplicates                *prometheus.CounterVec
	mongoInsertErrors         *prometheus.CounterVec
	cacheErrors               *prometheus.CounterVec
	consumerLag               *prometheus.GaugeVec
	featureEvents             prometheus.Counter
	featureErrors             *prometheus.CounterVec
	featureDocuments          *prometheus.CounterVec
	featureDuration           *prometheus.HistogramVec
	featureLag                prometheus.Gauge
	rankingRuns               *prometheus.CounterVec
	rankingDuration           prometheus.Observer
	rankingCandidates         *prometheus.GaugeVec
	rankingItems              *prometheus.CounterVec
	rankingEmptySets          *prometheus.CounterVec
	rankingUpsertError        *prometheus.CounterVec
	rankingCacheError         *prometheus.CounterVec
	rankingFallbacks          *prometheus.CounterVec
	personalizedBuilds        *prometheus.CounterVec
	personalizedDuration      prometheus.Observer
	personalizedProfileMisses *prometheus.CounterVec
	personalizedCandidates    prometheus.Observer
	personalizedItems         prometheus.Observer
	personalizedFallbacks     *prometheus.CounterVec
	personalizedBackfills     *prometheus.CounterVec
	personalizedCacheHits     *prometheus.CounterVec
	personalizedCacheErrors   *prometheus.CounterVec
	personalizedEmptyResults  prometheus.Counter
	personalizedUpsertErrors  prometheus.Counter
	grpcRequests              *prometheus.CounterVec
	grpcDuration              *prometheus.HistogramVec
	grpcInflight              *prometheus.GaugeVec
	serveSources              *prometheus.CounterVec
	serveFallbacks            *prometheus.CounterVec
	serveEmptyResults         *prometheus.CounterVec
	strategyServed            *prometheus.CounterVec
	abAssignments             *prometheus.CounterVec
	abAssignmentErrors        *prometheus.CounterVec
}

func NewConsumerMetrics(registerer prometheus.Registerer) (*ConsumerMetrics, error) {
	if registerer == nil {
		registerer = prometheus.DefaultRegisterer
	}

	metrics := &ConsumerMetrics{}
	var err error
	metrics.consumed, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_events_consumed_total",
		Help: "Recommendation events consumed by result.",
	}, []string{"event_type", "producer", "result"}))
	if err != nil {
		return nil, err
	}
	metrics.processingDuration, err = registerHistogramVec(registerer, prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "recommendation_event_processing_duration_ms",
		Help:    "Recommendation event processing duration in milliseconds.",
		Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
	}, []string{"event_type"}))
	if err != nil {
		return nil, err
	}
	metrics.dlq, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_event_dlq_total",
		Help: "Recommendation events sent to the dead-letter queue.",
	}, []string{"reason", "event_type"}))
	if err != nil {
		return nil, err
	}
	metrics.duplicates, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_event_duplicate_total",
		Help: "Duplicate recommendation interaction events ignored by idempotency.",
	}, []string{"event_type"}))
	if err != nil {
		return nil, err
	}
	metrics.mongoInsertErrors, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_mongo_insert_errors_total",
		Help: "MongoDB insert errors while ingesting recommendation interactions.",
	}, []string{"error_type"}))
	if err != nil {
		return nil, err
	}
	metrics.cacheErrors, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_cache_invalidation_errors_total",
		Help: "Redis cache invalidation errors after recommendation interaction ingestion.",
	}, []string{"cache_action"}))
	if err != nil {
		return nil, err
	}
	metrics.consumerLag, err = registerGaugeVec(registerer, prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "queue_consumer_lag",
		Help: "Queue consumer lag by topic and consumer group when reported by the provider.",
	}, []string{"topic", "consumer_group"}))
	if err != nil {
		return nil, err
	}
	metrics.featureEvents, err = registerCounter(registerer, prometheus.NewCounter(prometheus.CounterOpts{
		Name: "recommendation_feature_events_processed_total",
		Help: "Raw interactions applied to derived recommendation features.",
	}))
	if err != nil {
		return nil, err
	}
	metrics.featureErrors, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_feature_update_errors_total",
		Help: "Feature builder update failures.",
	}, []string{"error_type"}))
	if err != nil {
		return nil, err
	}
	metrics.featureDocuments, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_feature_docs_updated_total",
		Help: "Derived feature documents updated by collection.",
	}, []string{"collection"}))
	if err != nil {
		return nil, err
	}
	metrics.featureDuration, err = registerHistogramVec(registerer, prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "recommendation_feature_job_duration_ms",
		Help:    "Feature builder job duration in milliseconds.",
		Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
	}, []string{"job_type"}))
	if err != nil {
		return nil, err
	}
	metrics.featureLag, err = registerGauge(registerer, prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "recommendation_feature_job_lag_seconds",
		Help: "Lag between current time and the latest processed interaction watermark.",
	}))
	if err != nil {
		return nil, err
	}
	metrics.rankingRuns, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_ranking_job_runs_total",
		Help: "Rule-based ranking rebuild runs by result.",
	}, []string{"result"}))
	if err != nil {
		return nil, err
	}
	rankingDuration, err := registerHistogram(registerer, prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "recommendation_ranking_job_duration_ms",
		Help:    "Rule-based ranking rebuild duration in milliseconds.",
		Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
	}))
	if err != nil {
		return nil, err
	}
	metrics.rankingDuration = rankingDuration
	metrics.rankingCandidates, err = registerGaugeVec(registerer, prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "recommendation_ranking_candidates_total",
		Help: "Eligible product candidates observed by generated ranking scope.",
	}, []string{"scope"}))
	if err != nil {
		return nil, err
	}
	metrics.rankingItems, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_ranking_items_generated_total",
		Help: "Items written into generated rule-based sets.",
	}, []string{"scope"}))
	if err != nil {
		return nil, err
	}
	metrics.rankingEmptySets, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_empty_set_total",
		Help: "Generated rule-based sets containing no product items.",
	}, []string{"scope"}))
	if err != nil {
		return nil, err
	}
	metrics.rankingUpsertError, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_set_upsert_errors_total",
		Help: "Durable rule-based recommendation set write failures.",
	}, []string{"scope"}))
	if err != nil {
		return nil, err
	}
	metrics.rankingCacheError, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_cache_set_errors_total",
		Help: "Rule-based recommendation Redis publish failures.",
	}, []string{"scope"}))
	if err != nil {
		return nil, err
	}
	metrics.rankingFallbacks, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_fallback_total",
		Help: "Scoped rule-based list resolution attempts falling back to global trending.",
	}, []string{"scope"}))
	if err != nil {
		return nil, err
	}
	metrics.personalizedBuilds, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_personalized_build_total",
		Help: "Personalized ranking builds by outcome.",
	}, []string{"outcome"}))
	if err != nil {
		return nil, err
	}
	personalizedDuration, err := registerHistogram(registerer, prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "recommendation_personalized_build_duration_ms",
		Help:    "Personalized ranking build duration in milliseconds.",
		Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
	}))
	if err != nil {
		return nil, err
	}
	metrics.personalizedDuration = personalizedDuration
	metrics.personalizedProfileMisses, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_personalized_profile_miss_total",
		Help: "Personalized ranking profile misses or weak profile decisions.",
	}, []string{"reason"}))
	if err != nil {
		return nil, err
	}
	personalizedCandidates, err := registerHistogram(registerer, prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "recommendation_personalized_candidates_total",
		Help:    "Candidate pool size evaluated by personalized ranking.",
		Buckets: []float64{0, 1, 5, 10, 25, 50, 100, 250, 500, 1000},
	}))
	if err != nil {
		return nil, err
	}
	metrics.personalizedCandidates = personalizedCandidates
	personalizedItems, err := registerHistogram(registerer, prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "recommendation_personalized_result_items_total",
		Help:    "Final item count returned by personalized ranking.",
		Buckets: []float64{0, 1, 2, 4, 8, 12, 24, 48, 100},
	}))
	if err != nil {
		return nil, err
	}
	metrics.personalizedItems = personalizedItems
	metrics.personalizedFallbacks, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_personalized_fallback_total",
		Help: "Cold-start fallback usage by source.",
	}, []string{"source"}))
	if err != nil {
		return nil, err
	}
	metrics.personalizedBackfills, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_personalized_backfill_total",
		Help: "Personalized ranking backfilled items by fallback source.",
	}, []string{"source"}))
	if err != nil {
		return nil, err
	}
	metrics.personalizedCacheHits, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_personalized_cache_hit_total",
		Help: "Personalized cache hits by identity type.",
	}, []string{"identity_type"}))
	if err != nil {
		return nil, err
	}
	metrics.personalizedCacheErrors, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_personalized_cache_error_total",
		Help: "Personalized cache errors by action.",
	}, []string{"action"}))
	if err != nil {
		return nil, err
	}
	metrics.personalizedEmptyResults, err = registerCounter(registerer, prometheus.NewCounter(prometheus.CounterOpts{
		Name: "recommendation_personalized_empty_result_total",
		Help: "Personalized builds returning no recommendation items.",
	}))
	if err != nil {
		return nil, err
	}
	metrics.personalizedUpsertErrors, err = registerCounter(registerer, prometheus.NewCounter(prometheus.CounterOpts{
		Name: "recommendation_personalized_set_upsert_error_total",
		Help: "Personalized recommendation set durable write failures.",
	}))
	if err != nil {
		return nil, err
	}
	metrics.grpcRequests, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_grpc_requests_total",
		Help: "Recommendation gRPC requests by method and status code.",
	}, []string{"method", "code"}))
	if err != nil {
		return nil, err
	}
	metrics.grpcDuration, err = registerHistogramVec(registerer, prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "recommendation_grpc_duration_seconds",
		Help:    "Recommendation gRPC request duration in seconds.",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"method"}))
	if err != nil {
		return nil, err
	}
	metrics.grpcInflight, err = registerGaugeVec(registerer, prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "recommendation_grpc_inflight_requests",
		Help: "Recommendation gRPC requests currently in flight.",
	}, []string{"method"}))
	if err != nil {
		return nil, err
	}
	metrics.serveSources, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_serve_source_total",
		Help: "Recommendation serving results by source.",
	}, []string{"source"}))
	if err != nil {
		return nil, err
	}
	metrics.serveFallbacks, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_serve_fallback_total",
		Help: "Recommendation serving fallback usage by result type.",
	}, []string{"type"}))
	if err != nil {
		return nil, err
	}
	metrics.serveEmptyResults, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_serve_empty_total",
		Help: "Recommendation serving empty responses by request context.",
	}, []string{"context"}))
	if err != nil {
		return nil, err
	}
	metrics.strategyServed, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_strategy_served_total",
		Help: "Recommendation responses served by stable strategy id.",
	}, []string{"strategy_id", "context", "source"}))
	if err != nil {
		return nil, err
	}
	metrics.abAssignments, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_ab_assignment_total",
		Help: "Recommendation A/B assignments by experiment, variant, and strategy.",
	}, []string{"experiment_id", "variant_id", "strategy_id"}))
	if err != nil {
		return nil, err
	}
	metrics.abAssignmentErrors, err = registerCounterVec(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recommendation_ab_assignment_error_total",
		Help: "Recommendation A/B assignment hook failures by low-cardinality reason.",
	}, []string{"reason"}))
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (m *ConsumerMetrics) RecordConsumed(eventType string, producer string, result string) {
	if m != nil && m.consumed != nil {
		m.consumed.WithLabelValues(eventType, producer, result).Inc()
	}
}

func (m *ConsumerMetrics) ObserveProcessingDuration(eventType string, duration time.Duration) {
	if m != nil && m.processingDuration != nil {
		m.processingDuration.WithLabelValues(eventType).Observe(float64(duration.Milliseconds()))
	}
}

func (m *ConsumerMetrics) RecordDLQ(reason string, eventType string) {
	if m != nil && m.dlq != nil {
		m.dlq.WithLabelValues(reason, eventType).Inc()
	}
}

func (m *ConsumerMetrics) RecordDuplicate(eventType string) {
	if m != nil && m.duplicates != nil {
		m.duplicates.WithLabelValues(eventType).Inc()
	}
}

func (m *ConsumerMetrics) RecordMongoInsertError(errorType string) {
	if m != nil && m.mongoInsertErrors != nil {
		m.mongoInsertErrors.WithLabelValues(errorType).Inc()
	}
}

func (m *ConsumerMetrics) RecordCacheInvalidationError(action string) {
	if m != nil && m.cacheErrors != nil {
		m.cacheErrors.WithLabelValues(action).Inc()
	}
}

func (m *ConsumerMetrics) SetConsumerLag(topic string, consumerGroup string, lag float64) {
	if m != nil && m.consumerLag != nil {
		m.consumerLag.WithLabelValues(topic, consumerGroup).Set(lag)
	}
}

func (m *ConsumerMetrics) RecordFeatureEventsProcessed(count int) {
	if m != nil && m.featureEvents != nil && count > 0 {
		m.featureEvents.Add(float64(count))
	}
}

func (m *ConsumerMetrics) RecordFeatureUpdateError(errorType string) {
	if m != nil && m.featureErrors != nil {
		m.featureErrors.WithLabelValues(errorType).Inc()
	}
}

func (m *ConsumerMetrics) RecordFeatureDocumentsUpdated(collection string, count int) {
	if m != nil && m.featureDocuments != nil && count > 0 {
		m.featureDocuments.WithLabelValues(collection).Add(float64(count))
	}
}

func (m *ConsumerMetrics) ObserveFeatureJobDuration(jobType string, duration time.Duration) {
	if m != nil && m.featureDuration != nil {
		m.featureDuration.WithLabelValues(jobType).Observe(float64(duration.Milliseconds()))
	}
}

func (m *ConsumerMetrics) SetFeatureJobLag(lag time.Duration) {
	if m != nil && m.featureLag != nil {
		m.featureLag.Set(lag.Seconds())
	}
}

func (m *ConsumerMetrics) RecordRankingRun(result string) {
	if m != nil && m.rankingRuns != nil {
		m.rankingRuns.WithLabelValues(result).Inc()
	}
}

func (m *ConsumerMetrics) ObserveRankingJobDuration(duration time.Duration) {
	if m != nil && m.rankingDuration != nil {
		m.rankingDuration.Observe(float64(duration.Milliseconds()))
	}
}

func (m *ConsumerMetrics) SetRankingCandidates(scope string, count int) {
	if m != nil && m.rankingCandidates != nil {
		m.rankingCandidates.WithLabelValues(scope).Set(float64(count))
	}
}

func (m *ConsumerMetrics) RecordRankingItemsGenerated(scope string, count int) {
	if m != nil && m.rankingItems != nil && count > 0 {
		m.rankingItems.WithLabelValues(scope).Add(float64(count))
	}
}

func (m *ConsumerMetrics) RecordRankingEmptySet(scope string) {
	if m != nil && m.rankingEmptySets != nil {
		m.rankingEmptySets.WithLabelValues(scope).Inc()
	}
}

func (m *ConsumerMetrics) RecordRankingSetUpsertError(scope string) {
	if m != nil && m.rankingUpsertError != nil {
		m.rankingUpsertError.WithLabelValues(scope).Inc()
	}
}

func (m *ConsumerMetrics) RecordRankingCacheSetError(scope string) {
	if m != nil && m.rankingCacheError != nil {
		m.rankingCacheError.WithLabelValues(scope).Inc()
	}
}

func (m *ConsumerMetrics) RecordRankingFallback(scope string) {
	if m != nil && m.rankingFallbacks != nil {
		m.rankingFallbacks.WithLabelValues(scope).Inc()
	}
}

func (m *ConsumerMetrics) RecordPersonalizedBuild(outcome string) {
	if m != nil && m.personalizedBuilds != nil {
		m.personalizedBuilds.WithLabelValues(outcome).Inc()
	}
}

func (m *ConsumerMetrics) ObservePersonalizedBuildDuration(duration time.Duration) {
	if m != nil && m.personalizedDuration != nil {
		m.personalizedDuration.Observe(float64(duration.Milliseconds()))
	}
}

func (m *ConsumerMetrics) RecordPersonalizedProfileMiss(reason string) {
	if m != nil && m.personalizedProfileMisses != nil {
		m.personalizedProfileMisses.WithLabelValues(reason).Inc()
	}
}

func (m *ConsumerMetrics) ObservePersonalizedCandidates(count int) {
	if m != nil && m.personalizedCandidates != nil {
		m.personalizedCandidates.Observe(float64(count))
	}
}

func (m *ConsumerMetrics) ObservePersonalizedResultItems(count int) {
	if m != nil && m.personalizedItems != nil {
		m.personalizedItems.Observe(float64(count))
	}
}

func (m *ConsumerMetrics) RecordPersonalizedFallback(source string) {
	if m != nil && m.personalizedFallbacks != nil {
		m.personalizedFallbacks.WithLabelValues(source).Inc()
	}
}

func (m *ConsumerMetrics) RecordPersonalizedBackfill(source string, count int) {
	if m != nil && m.personalizedBackfills != nil && count > 0 {
		m.personalizedBackfills.WithLabelValues(source).Add(float64(count))
	}
}

func (m *ConsumerMetrics) RecordPersonalizedCacheHit(identityType string) {
	if m != nil && m.personalizedCacheHits != nil {
		m.personalizedCacheHits.WithLabelValues(identityType).Inc()
	}
}

func (m *ConsumerMetrics) RecordPersonalizedCacheError(action string) {
	if m != nil && m.personalizedCacheErrors != nil {
		m.personalizedCacheErrors.WithLabelValues(action).Inc()
	}
}

func (m *ConsumerMetrics) RecordPersonalizedEmptyResult() {
	if m != nil && m.personalizedEmptyResults != nil {
		m.personalizedEmptyResults.Inc()
	}
}

func (m *ConsumerMetrics) RecordPersonalizedSetUpsertError() {
	if m != nil && m.personalizedUpsertErrors != nil {
		m.personalizedUpsertErrors.Inc()
	}
}

func (m *ConsumerMetrics) RecordGRPCRequest(method string, code string) {
	if m != nil && m.grpcRequests != nil {
		m.grpcRequests.WithLabelValues(method, code).Inc()
	}
}

func (m *ConsumerMetrics) ObserveGRPCDuration(method string, duration time.Duration) {
	if m != nil && m.grpcDuration != nil {
		m.grpcDuration.WithLabelValues(method).Observe(duration.Seconds())
	}
}

func (m *ConsumerMetrics) IncGRPCInflight(method string) {
	if m != nil && m.grpcInflight != nil {
		m.grpcInflight.WithLabelValues(method).Inc()
	}
}

func (m *ConsumerMetrics) DecGRPCInflight(method string) {
	if m != nil && m.grpcInflight != nil {
		m.grpcInflight.WithLabelValues(method).Dec()
	}
}

func (m *ConsumerMetrics) RecordRecommendationServeSource(source string) {
	if m != nil && m.serveSources != nil {
		m.serveSources.WithLabelValues(source).Inc()
	}
}

func (m *ConsumerMetrics) RecordRecommendationServeFallback(kind string) {
	if m != nil && m.serveFallbacks != nil {
		m.serveFallbacks.WithLabelValues(kind).Inc()
	}
}

func (m *ConsumerMetrics) RecordRecommendationServeEmpty(context string) {
	if m != nil && m.serveEmptyResults != nil {
		m.serveEmptyResults.WithLabelValues(context).Inc()
	}
}

func (m *ConsumerMetrics) RecordRecommendationStrategyServed(strategyID string, context string, source string) {
	if m != nil && m.strategyServed != nil {
		m.strategyServed.WithLabelValues(strategyID, context, source).Inc()
	}
}

func (m *ConsumerMetrics) RecordABAssignment(experimentID string, variantID string, strategyID string) {
	if m != nil && m.abAssignments != nil {
		m.abAssignments.WithLabelValues(experimentID, variantID, strategyID).Inc()
	}
}

func (m *ConsumerMetrics) RecordABAssignmentError(reason string) {
	if m != nil && m.abAssignmentErrors != nil {
		m.abAssignmentErrors.WithLabelValues(reason).Inc()
	}
}

func registerCounterVec(registerer prometheus.Registerer, collector *prometheus.CounterVec) (*prometheus.CounterVec, error) {
	if err := registerer.Register(collector); err != nil {
		if already, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if existing, ok := already.ExistingCollector.(*prometheus.CounterVec); ok {
				return existing, nil
			}
		}
		return nil, err
	}
	return collector, nil
}

func registerHistogramVec(registerer prometheus.Registerer, collector *prometheus.HistogramVec) (*prometheus.HistogramVec, error) {
	if err := registerer.Register(collector); err != nil {
		if already, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if existing, ok := already.ExistingCollector.(*prometheus.HistogramVec); ok {
				return existing, nil
			}
		}
		return nil, err
	}
	return collector, nil
}

func registerGaugeVec(registerer prometheus.Registerer, collector *prometheus.GaugeVec) (*prometheus.GaugeVec, error) {
	if err := registerer.Register(collector); err != nil {
		if already, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if existing, ok := already.ExistingCollector.(*prometheus.GaugeVec); ok {
				return existing, nil
			}
		}
		return nil, err
	}
	return collector, nil
}

func registerCounter(registerer prometheus.Registerer, collector prometheus.Counter) (prometheus.Counter, error) {
	if err := registerer.Register(collector); err != nil {
		if already, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if existing, ok := already.ExistingCollector.(prometheus.Counter); ok {
				return existing, nil
			}
		}
		return nil, err
	}
	return collector, nil
}

func registerGauge(registerer prometheus.Registerer, collector prometheus.Gauge) (prometheus.Gauge, error) {
	if err := registerer.Register(collector); err != nil {
		if already, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if existing, ok := already.ExistingCollector.(prometheus.Gauge); ok {
				return existing, nil
			}
		}
		return nil, err
	}
	return collector, nil
}

func registerHistogram(registerer prometheus.Registerer, collector prometheus.Histogram) (prometheus.Histogram, error) {
	if err := registerer.Register(collector); err != nil {
		if already, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if existing, ok := already.ExistingCollector.(prometheus.Histogram); ok {
				return existing, nil
			}
		}
		return nil, err
	}
	return collector, nil
}
