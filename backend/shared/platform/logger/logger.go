package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

var sensitiveFragments = []string{
	"authorization", "card_number", "cookie", "credential", "cvv", "otp", "password", "refresh_token", "secret", "signature", "token",
}

func New(service, environment string, level slog.Level) *slog.Logger {
	return NewWithWriter(os.Stdout, service, environment, level)
}

func NewWithWriter(writer io.Writer, service, environment string, level slog.Level) *slog.Logger {
	if writer == nil {
		writer = io.Discard
	}
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if IsSensitiveKey(attr.Key) {
				return slog.String(attr.Key, "[REDACTED]")
			}
			return attr
		},
	})
	return slog.New(handler).With("service", service, "environment", environment)
}

func FromContext(ctx context.Context, fallback *slog.Logger) *slog.Logger {
	if fallback == nil {
		fallback = slog.Default()
	}
	attributes := make([]any, 0, 6)
	if requestID := RequestIDFromContext(ctx); requestID != "" {
		attributes = append(attributes, "request_id", requestID)
	}
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		attributes = append(attributes, "trace_id", spanContext.TraceID().String(), "span_id", spanContext.SpanID().String())
	}
	return fallback.With(attributes...)
}

type requestIDKey struct{}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, strings.TrimSpace(requestID))
}

func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	requestID, _ := ctx.Value(requestIDKey{}).(string)
	return strings.TrimSpace(requestID)
}

func IsSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	for _, fragment := range sensitiveFragments {
		if normalized == fragment || strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}
