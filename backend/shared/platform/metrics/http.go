package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTP struct {
	Requests *prometheus.CounterVec
	Duration *prometheus.HistogramVec
}

func NewHTTP(registry prometheus.Registerer, namespace, service string) *HTTP {
	metrics := &HTTP{
		Requests: prometheus.NewCounterVec(prometheus.CounterOpts{Namespace: namespace, Subsystem: "http", Name: "requests_total", ConstLabels: prometheus.Labels{"service": service}}, []string{"method", "route", "status_code"}),
		Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Namespace: namespace, Subsystem: "http", Name: "request_duration_seconds", ConstLabels: prometheus.Labels{"service": service}}, []string{"method", "route"}),
	}
	registry.MustRegister(metrics.Requests, metrics.Duration)
	return metrics
}

func (metrics *HTTP) Middleware(route func(*http.Request) string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		routeName := r.URL.Path
		if route != nil {
			routeName = route(r)
		}
		metrics.Requests.WithLabelValues(r.Method, routeName, strconv.Itoa(recorder.status)).Inc()
		metrics.Duration.WithLabelValues(r.Method, routeName).Observe(time.Since(started).Seconds())
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (recorder *statusRecorder) WriteHeader(status int) {
	recorder.status = status
	recorder.ResponseWriter.WriteHeader(status)
}

func Handler(gatherer prometheus.Gatherer) http.Handler {
	return promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})
}
