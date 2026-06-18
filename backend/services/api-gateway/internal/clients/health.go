package clients

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

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
	report := make(HealthReport, len(serviceOrder))
	if c == nil {
		for _, service := range serviceOrder {
			report[service] = DependencyHealth{
				Service: string(service),
				Status:  HealthStatusUnavailable,
				Error:   "grpc client registry is not initialized",
			}
		}
		return report
	}

	var wg sync.WaitGroup
	results := make(chan dependencyHealthResult, len(serviceOrder))
	for _, service := range serviceOrder {
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
