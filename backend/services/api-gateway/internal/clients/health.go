package clients

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"ecommerce/api-gateway/internal/config"

	"google.golang.org/grpc"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

const healthCheckTimeout = 300 * time.Millisecond

type HealthStatus string

const (
	HealthStatusServing     HealthStatus = "serving"
	HealthStatusNotServing  HealthStatus = "not_serving"
	HealthStatusUnknown     HealthStatus = "unknown"
	HealthStatusUnavailable HealthStatus = "unavailable"
)

type DependencyHealth struct {
	Service string       `json:"service"`
	Status  HealthStatus `json:"status"`
	Error   string       `json:"error,omitempty"`
}

type HealthReport map[Downstream]DependencyHealth

func (c *Clients) Check(ctx context.Context) HealthReport {
	return c.checkServices(ctx, serviceOrder)
}

func (c *Clients) checkServices(ctx context.Context, services []Downstream) HealthReport {
	report := make(HealthReport, len(services))
	if c == nil {
		for _, service := range services {
			report[service] = DependencyHealth{
				Service: string(service),
				Status:  HealthStatusUnavailable,
				Error:   "grpc client registry is not initialized",
			}
		}
		return report
	}

	var wg sync.WaitGroup
	results := make(chan dependencyHealthResult, len(services))
	for _, service := range services {
		descriptor, descriptorOK := c.descriptors[service]
		conn, connOK := c.conns[service]
		if !descriptorOK || !connOK || conn == nil {
			report[service] = DependencyHealth{
				Service: string(service),
				Status:  HealthStatusUnavailable,
				Error:   "grpc connection is not initialized",
			}
			continue
		}

		wg.Add(1)
		go func(descriptor ServiceDescriptor, conn *grpc.ClientConn) {
			defer wg.Done()
			results <- dependencyHealthResult{
				service: descriptor.Name,
				health:  checkHealth(ctx, descriptor, conn),
			}
		}(descriptor, conn)
	}

	wg.Wait()
	close(results)
	for result := range results {
		report[result.service] = result.health
	}
	return report
}

type HTTPHealthTarget struct {
	URL  string
	Path string
}

type ReadinessChecker struct {
	grpc        *Clients
	httpClient  *http.Client
	httpTargets map[Downstream]HTTPHealthTarget
}

func NewReadinessCheckerFromConfig(cfg config.Config, grpcClients *Clients) *ReadinessChecker {
	return &ReadinessChecker{
		grpc:       grpcClients,
		httpClient: &http.Client{},
		httpTargets: map[Downstream]HTTPHealthTarget{
			DownstreamAuth:         {URL: cfg.AuthHTTPURL, Path: "/readyz"},
			DownstreamCart:         {URL: cfg.CartHTTPURL, Path: "/readyz"},
			DownstreamWishlist:     {URL: cfg.WishlistHTTPURL, Path: "/readyz"},
			DownstreamPayment:      {URL: cfg.PaymentHTTPURL, Path: "/healthz"},
			DownstreamCMS:          {URL: cfg.CMSHTTPURL, Path: "/healthz"},
			DownstreamSession:      {URL: cfg.SessionHTTPURL, Path: "/readyz"},
			DownstreamNotification: {URL: cfg.NotificationHTTPURL, Path: "/readyz"},
			DownstreamSuperadmin:   {URL: cfg.SuperadminHTTPURL, Path: "/readyz"},
		},
	}
}

func (c *ReadinessChecker) Check(ctx context.Context) HealthReport {
	report := make(HealthReport, len(serviceOrder))
	if c == nil {
		return (*Clients)(nil).Check(ctx)
	}

	grpcReport := c.grpc.checkServices(ctx, grpcReadinessServices)
	for _, service := range serviceOrder {
		if target, ok := c.httpTargets[service]; ok {
			report[service] = checkHTTPHealth(ctx, service, target, c.httpClient)
			continue
		}
		if health, ok := grpcReport[service]; ok {
			report[service] = health
			continue
		}
		report[service] = DependencyHealth{
			Service: string(service),
			Status:  HealthStatusUnavailable,
			Error:   "readiness check is not configured",
		}
	}
	return report
}

func (r HealthReport) Ready() bool {
	if len(r) == 0 {
		return false
	}
	for _, service := range serviceOrder {
		health, ok := r[service]
		if !ok || health.Status != HealthStatusServing {
			return false
		}
	}
	return true
}

func (r HealthReport) Statuses() map[string]string {
	statuses := make(map[string]string, len(r))
	for _, service := range serviceOrder {
		if health, ok := r[service]; ok {
			statuses[string(service)] = string(health.Status)
		}
	}
	return statuses
}

func (r HealthReport) Errors() []string {
	errs := make([]string, 0)
	for _, service := range serviceOrder {
		health, ok := r[service]
		if !ok || health.Error == "" {
			continue
		}
		errs = append(errs, fmt.Sprintf("%s: %s", service, health.Error))
	}
	sort.Strings(errs)
	return errs
}

type dependencyHealthResult struct {
	service Downstream
	health  DependencyHealth
}

var grpcReadinessServices = []Downstream{
	DownstreamUser,
	DownstreamProduct,
	DownstreamOrder,
	DownstreamSearch,
	DownstreamRecommendation,
}

func checkHealth(ctx context.Context, descriptor ServiceDescriptor, conn *grpc.ClientConn) DependencyHealth {
	healthCtx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	resp, err := healthv1.NewHealthClient(conn).Check(healthCtx, &healthv1.HealthCheckRequest{
		Service: descriptor.HealthService,
	})
	if err != nil {
		return DependencyHealth{
			Service: string(descriptor.Name),
			Status:  HealthStatusUnavailable,
			Error:   err.Error(),
		}
	}
	status := healthStatusFromGRPC(resp.GetStatus())
	if status != HealthStatusServing {
		return DependencyHealth{
			Service: string(descriptor.Name),
			Status:  status,
			Error:   resp.GetStatus().String(),
		}
	}
	return DependencyHealth{
		Service: string(descriptor.Name),
		Status:  HealthStatusServing,
	}
}

func checkHTTPHealth(ctx context.Context, service Downstream, target HTTPHealthTarget, client *http.Client) DependencyHealth {
	rawURL := strings.TrimSpace(target.URL)
	if rawURL == "" {
		return DependencyHealth{
			Service: string(service),
			Status:  HealthStatusUnavailable,
			Error:   "http health URL is not configured",
		}
	}
	if client == nil {
		client = http.DefaultClient
	}

	healthCtx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(healthCtx, http.MethodGet, joinHealthURL(rawURL, target.Path), nil)
	if err != nil {
		return DependencyHealth{
			Service: string(service),
			Status:  HealthStatusUnavailable,
			Error:   err.Error(),
		}
	}
	response, err := client.Do(request)
	if err != nil {
		return DependencyHealth{
			Service: string(service),
			Status:  HealthStatusUnavailable,
			Error:   err.Error(),
		}
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return DependencyHealth{
			Service: string(service),
			Status:  HealthStatusUnavailable,
			Error:   fmt.Sprintf("http health returned %s", response.Status),
		}
	}
	return DependencyHealth{
		Service: string(service),
		Status:  HealthStatusServing,
	}
}

func joinHealthURL(rawURL string, path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "/healthz"
	}
	return strings.TrimRight(rawURL, "/") + "/" + strings.TrimLeft(path, "/")
}

func healthStatusFromGRPC(status healthv1.HealthCheckResponse_ServingStatus) HealthStatus {
	switch status {
	case healthv1.HealthCheckResponse_SERVING:
		return HealthStatusServing
	case healthv1.HealthCheckResponse_NOT_SERVING:
		return HealthStatusNotServing
	default:
		return HealthStatusUnknown
	}
}
