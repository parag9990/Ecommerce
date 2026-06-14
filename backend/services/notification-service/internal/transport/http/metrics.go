package http

import (
	nethttp "net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func MetricsHandler(gatherer prometheus.Gatherer) nethttp.Handler {
	return promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})
}
