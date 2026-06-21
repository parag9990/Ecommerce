package tracing

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type Config struct {
	ServiceName, Environment, OTLPEndpoint string
	SampleRatio                            float64
}

func Setup(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	if strings.TrimSpace(cfg.ServiceName) == "" {
		return nil, errors.New("service name is required")
	}
	ratio := cfg.SampleRatio
	if ratio <= 0 || ratio > 1 {
		ratio = 0.1
	}
	// Use an empty schema URL when merging with resource.Default so SDK upgrades do
	// not create an otherwise harmless schema-version conflict at startup.
	res, err := resource.Merge(resource.Default(), resource.NewWithAttributes("", semconv.ServiceName(cfg.ServiceName), semconv.DeploymentEnvironmentName(cfg.Environment)))
	if err != nil {
		return nil, err
	}
	providerOptions := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
		sdktrace.WithResource(res),
	}
	if endpoint := strings.TrimSpace(cfg.OTLPEndpoint); endpoint != "" {
		exporter, exporterErr := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithInsecure())
		if exporterErr != nil {
			return nil, exporterErr
		}
		providerOptions = append(providerOptions, sdktrace.WithBatcher(exporter))
	}
	provider := sdktrace.NewTracerProvider(providerOptions...)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return provider.Shutdown, nil
}

func ExtractHTTP(ctx context.Context, header http.Header) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(header))
}

func InjectHTTP(ctx context.Context, header http.Header) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))
}
