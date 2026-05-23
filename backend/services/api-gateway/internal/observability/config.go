package observability

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	defaultMetricsAddr           = ":9090"
	defaultOTLPEndpoint          = "otel-collector:4317"
	defaultTraceShutdownTimeout  = 5 * time.Second
	defaultTraceSampleRatio      = 1.0
	defaultRequestIDHeader       = "X-Request-Id"
	defaultDeploymentEnvironment = "local"
)

type Config struct {
	ServiceName          string
	Environment          string
	RequestIDHeader      string
	MetricsEnabled       bool
	MetricsAddress       string
	TracingEnabled       bool
	TraceOTLPEndpoint    string
	TraceOTLPInsecure    bool
	TraceSampleRatio     float64
	TraceShutdownTimeout time.Duration
	UserHashSalt         string
}

func DefaultConfig(serviceName, environment string) Config {
	return Config{
		ServiceName:          defaultString(serviceName, "api-gateway"),
		Environment:          defaultString(environment, defaultDeploymentEnvironment),
		RequestIDHeader:      defaultRequestIDHeader,
		MetricsEnabled:       true,
		MetricsAddress:       defaultMetricsAddr,
		TracingEnabled:       true,
		TraceOTLPEndpoint:    defaultOTLPEndpoint,
		TraceOTLPInsecure:    true,
		TraceSampleRatio:     defaultTraceSampleRatio,
		TraceShutdownTimeout: defaultTraceShutdownTimeout,
	}
}

func (c Config) Normalize(serviceName, environment string) Config {
	if strings.TrimSpace(c.ServiceName) == "" {
		c.ServiceName = strings.TrimSpace(serviceName)
	}
	if strings.TrimSpace(c.ServiceName) == "" {
		c.ServiceName = "api-gateway"
	}
	if strings.TrimSpace(c.Environment) == "" {
		c.Environment = strings.TrimSpace(environment)
	}
	if strings.TrimSpace(c.Environment) == "" {
		c.Environment = defaultDeploymentEnvironment
	}
	if strings.TrimSpace(c.RequestIDHeader) == "" {
		c.RequestIDHeader = defaultRequestIDHeader
	}
	if c.MetricsEnabled && strings.TrimSpace(c.MetricsAddress) == "" {
		c.MetricsAddress = defaultMetricsAddr
	}
	if c.TracingEnabled {
		if strings.TrimSpace(c.TraceOTLPEndpoint) == "" {
			c.TraceOTLPEndpoint = defaultOTLPEndpoint
		}
		if c.TraceSampleRatio == 0 {
			c.TraceSampleRatio = defaultTraceSampleRatio
		}
		if c.TraceShutdownTimeout <= 0 {
			c.TraceShutdownTimeout = defaultTraceShutdownTimeout
		}
	}
	return c
}

func (c Config) Validate() error {
	var errs []error
	if strings.TrimSpace(c.ServiceName) == "" {
		errs = append(errs, errors.New("observability service name is required"))
	}
	if strings.TrimSpace(c.Environment) == "" {
		errs = append(errs, errors.New("observability environment is required"))
	}
	if strings.TrimSpace(c.RequestIDHeader) == "" {
		errs = append(errs, errors.New("request id header is required"))
	}
	if strings.ContainsAny(c.RequestIDHeader, " \t\r\n") {
		errs = append(errs, errors.New("request id header must not contain whitespace"))
	}
	if c.MetricsEnabled && strings.TrimSpace(c.MetricsAddress) == "" {
		errs = append(errs, errors.New("METRICS_ADDR is required when metrics are enabled"))
	}
	if c.TracingEnabled {
		if strings.TrimSpace(c.TraceOTLPEndpoint) == "" {
			errs = append(errs, errors.New("TRACE_EXPORTER_OTLP_ENDPOINT is required when tracing is enabled"))
		}
		if c.TraceSampleRatio < 0 || c.TraceSampleRatio > 1 {
			errs = append(errs, fmt.Errorf("TRACE_SAMPLE_RATIO must be between 0 and 1, got %v", c.TraceSampleRatio))
		}
		if c.TraceShutdownTimeout <= 0 {
			errs = append(errs, errors.New("TRACE_SHUTDOWN_TIMEOUT must be positive"))
		}
	}
	return errors.Join(errs...)
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
