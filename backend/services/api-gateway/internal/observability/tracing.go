package observability

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
)

type TracingProvider struct {
	provider *sdktrace.TracerProvider
}

func InitTracing(ctx context.Context, cfg Config) (*TracingProvider, error) {
	cfg = cfg.Normalize(cfg.ServiceName, cfg.Environment)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	if !cfg.TracingEnabled {
		return &TracingProvider{}, nil
	}

	exporterOptions := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.TraceOTLPEndpoint),
	}
	if cfg.TraceOTLPInsecure {
		exporterOptions = append(exporterOptions, otlptracegrpc.WithInsecure())
	} else {
		exporterOptions = append(exporterOptions, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}
	exporter, err := otlptracegrpc.New(ctx, exporterOptions...)
	if err != nil {
		return nil, err
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", cfg.ServiceName),
			attribute.String("deployment.environment.name", cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.TraceSampleRatio))),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(provider)
	return &TracingProvider{provider: provider}, nil
}

func (p *TracingProvider) Shutdown(ctx context.Context) error {
	if p == nil || p.provider == nil {
		return nil
	}
	return p.provider.Shutdown(ctx)
}

func TracingMiddleware(cfg Config, labeler RouteLabeler) func(http.Handler) http.Handler {
	cfg = cfg.Normalize(cfg.ServiceName, cfg.Environment)
	tracer := otel.Tracer(cfg.ServiceName)
	return func(next http.Handler) http.Handler {
		if !cfg.TracingEnabled {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := RouteTemplateFromRequest(r, labeler)
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			ctx, span := tracer.Start(ctx, r.Method+" "+route, trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()

			recorder := NewStatusRecorder(w)
			recordSpan(recorder, span.SpanContext())
			RecordRoute(recorder, route)
			span.SetAttributes(
				attribute.String("service.name", cfg.ServiceName),
				attribute.String("deployment.environment.name", cfg.Environment),
				attribute.String("http.request.method", r.Method),
				attribute.String("http.route", route),
				attribute.String("url.path", r.URL.Path),
				attribute.String("request_id", RequestIDFromContext(r.Context())),
			)

			next.ServeHTTP(recorder, r.WithContext(ctx))

			route = fallbackRoute(route, recorder.Route(), RouteTemplateFromRequest(r, labeler))
			span.SetName(r.Method + " " + route)
			span.SetAttributes(
				attribute.String("http.route", route),
				attribute.Int("http.response.status_code", recorder.Status()),
			)
			if recorder.Status() >= http.StatusInternalServerError && !IsClientClosed(recorder.Status()) {
				span.SetStatus(codes.Error, http.StatusText(recorder.Status()))
				if info := recorder.ErrorInfo(); info.Code != "" {
					span.SetAttributes(attribute.String("error.code", info.Code))
				}
			}
		})
	}
}

func recordSpan(w http.ResponseWriter, spanContext trace.SpanContext) {
	if !spanContext.IsValid() {
		return
	}
	RecordSpan(w, SpanInfo{
		TraceID: spanContext.TraceID().String(),
		SpanID:  spanContext.SpanID().String(),
	})
}
