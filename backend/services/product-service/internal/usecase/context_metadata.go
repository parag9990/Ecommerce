package usecase

import (
	"context"
	"strings"
)

type metadataContextKey string

const (
	requestIDContextKey metadataContextKey = "request_id"
	traceIDContextKey   metadataContextKey = "trace_id"
)

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, strings.TrimSpace(requestID))
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDContextKey, strings.TrimSpace(traceID))
}

func RequestIDFromContext(ctx context.Context) string {
	return firstContextString(ctx, requestIDContextKey, "request_id", "requestID", "x-request-id")
}

func TraceIDFromContext(ctx context.Context) string {
	return firstContextString(ctx, traceIDContextKey, "trace_id", "traceID", "x-trace-id")
}

func firstContextString(ctx context.Context, keys ...any) string {
	if ctx == nil {
		return ""
	}
	for _, key := range keys {
		if value, ok := ctx.Value(key).(string); ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}
