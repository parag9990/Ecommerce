package observability

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	serviceName string
	registry    *prometheus.Registry

	httpRequestsTotal         *prometheus.CounterVec
	httpRequestDuration       *prometheus.HistogramVec
	grpcClientRequestsTotal   *prometheus.CounterVec
	grpcClientDuration        *prometheus.HistogramVec
	gatewayRateLimitedTotal   *prometheus.CounterVec
	gatewayAuthFailuresTotal  *prometheus.CounterVec
	gatewayValidationFailures *prometheus.CounterVec
	gatewayDownstreamErrors   *prometheus.CounterVec
	gatewayInflightRequests   *prometheus.GaugeVec
}

func NewMetrics(cfg Config, registry *prometheus.Registry) (*Metrics, error) {
	cfg = cfg.Normalize(cfg.ServiceName, cfg.Environment)
	if !cfg.MetricsEnabled {
		return nil, nil
	}
	if registry == nil {
		registry = prometheus.NewRegistry()
	}
	metrics := &Metrics{
		serviceName: cfg.ServiceName,
		registry:    registry,
		httpRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests handled by the API Gateway.",
		}, []string{"service", "method", "route", "status_code", "error_code"}),
		httpRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"service", "method", "route"}),
		grpcClientRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "grpc_client_requests_total",
			Help: "Total outbound gRPC requests from the API Gateway.",
		}, []string{"service", "target_service", "grpc_method", "grpc_code"}),
		grpcClientDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "grpc_client_duration_seconds",
			Help:    "Outbound gRPC request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"service", "target_service", "grpc_method"}),
		gatewayRateLimitedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gateway_rate_limited_total",
			Help: "Total API Gateway requests blocked by rate limiting.",
		}, []string{"service", "route", "limit_type"}),
		gatewayAuthFailuresTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gateway_auth_failures_total",
			Help: "Total API Gateway authentication or authorization failures.",
		}, []string{"service", "route", "reason"}),
		gatewayValidationFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gateway_validation_failures_total",
			Help: "Total API Gateway request validation failures.",
		}, []string{"service", "route", "field_group"}),
		gatewayDownstreamErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gateway_downstream_errors_total",
			Help: "Total downstream gRPC errors observed by the API Gateway.",
		}, []string{"service", "target_service", "grpc_code"}),
		gatewayInflightRequests: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gateway_inflight_requests",
			Help: "Current in-flight API Gateway HTTP requests.",
		}, []string{"service", "route"}),
	}
	if err := registry.Register(metrics.httpRequestsTotal); err != nil {
		return nil, err
	}
	collectors := []prometheus.Collector{
		metrics.httpRequestDuration,
		metrics.grpcClientRequestsTotal,
		metrics.grpcClientDuration,
		metrics.gatewayRateLimitedTotal,
		metrics.gatewayAuthFailuresTotal,
		metrics.gatewayValidationFailures,
		metrics.gatewayDownstreamErrors,
		metrics.gatewayInflightRequests,
	}
	for _, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			return nil, err
		}
	}
	return metrics, nil
}

func (m *Metrics) Registry() *prometheus.Registry {
	if m == nil {
		return nil
	}
	return m.registry
}

func (m *Metrics) Handler() http.Handler {
	if m == nil || m.registry == nil {
		return http.NotFoundHandler()
	}
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *Metrics) ObserveHTTP(method, route string, status int, errorCode string, started time.Time) {
	if m == nil {
		return
	}
	route = emptyToUnknown(route)
	m.httpRequestsTotal.WithLabelValues(
		m.serviceName,
		method,
		route,
		strconv.Itoa(status),
		emptyToNone(errorCode),
	).Inc()
	m.httpRequestDuration.WithLabelValues(m.serviceName, method, route).Observe(time.Since(started).Seconds())
}

func (m *Metrics) ObserveGRPCClient(targetService, grpcMethod, grpcCode string, started time.Time) {
	if m == nil {
		return
	}
	targetService = emptyToUnknown(targetService)
	grpcMethod = emptyToUnknown(grpcMethod)
	grpcCode = emptyToNone(grpcCode)
	m.grpcClientRequestsTotal.WithLabelValues(m.serviceName, targetService, grpcMethod, grpcCode).Inc()
	m.grpcClientDuration.WithLabelValues(m.serviceName, targetService, grpcMethod).Observe(time.Since(started).Seconds())
}

func (m *Metrics) ObserveRateLimited(route, limitType string) {
	if m == nil {
		return
	}
	m.gatewayRateLimitedTotal.WithLabelValues(m.serviceName, emptyToUnknown(route), emptyToUnknown(limitType)).Inc()
}

func (m *Metrics) ObserveAuthFailure(route, reason string) {
	if m == nil {
		return
	}
	m.gatewayAuthFailuresTotal.WithLabelValues(m.serviceName, emptyToUnknown(route), emptyToUnknown(reason)).Inc()
}

func (m *Metrics) ObserveValidationFailure(route, fieldGroup string) {
	if m == nil {
		return
	}
	m.gatewayValidationFailures.WithLabelValues(m.serviceName, emptyToUnknown(route), emptyToUnknown(fieldGroup)).Inc()
}

func (m *Metrics) ObserveDownstreamError(targetService, grpcCode string) {
	if m == nil || targetService == "" {
		return
	}
	m.gatewayDownstreamErrors.WithLabelValues(m.serviceName, targetService, emptyToNone(grpcCode)).Inc()
}

func (m *Metrics) IncInflight(route string) {
	if m == nil {
		return
	}
	m.gatewayInflightRequests.WithLabelValues(m.serviceName, emptyToUnknown(route)).Inc()
}

func (m *Metrics) DecInflight(route string) {
	if m == nil {
		return
	}
	m.gatewayInflightRequests.WithLabelValues(m.serviceName, emptyToUnknown(route)).Dec()
}

func MetricsMiddleware(metrics *Metrics, labeler RouteLabeler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if metrics == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			recorder := NewStatusRecorder(w)
			route := RouteTemplateFromRequest(r, labeler)
			RecordRoute(recorder, route)
			metrics.IncInflight(route)
			defer metrics.DecInflight(route)

			next.ServeHTTP(recorder, r)

			route = fallbackRoute(route, recorder.Route(), RouteTemplateFromRequest(r, labeler))
			errorInfo := recorder.ErrorInfo()
			metrics.ObserveHTTP(r.Method, route, recorder.Status(), errorInfo.Code, started)
		})
	}
}

func emptyToNone(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "none"
	}
	return value
}

func emptyToUnknown(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return RouteUnknown
	}
	return value
}

func IsAlreadyRegistered(err error) bool {
	var already prometheus.AlreadyRegisteredError
	return errors.As(err, &already)
}
