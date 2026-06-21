package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestLoggerRedactsSensitiveFieldsAndAddsCorrelation(t *testing.T) {
	var output bytes.Buffer
	log := NewWithWriter(&output, "test-service", "test", slog.LevelInfo)
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1},
		SpanID:  trace.SpanID{2},
	})
	ctx := trace.ContextWithSpanContext(WithRequestID(context.Background(), "req-1"), spanContext)

	FromContext(ctx, log).InfoContext(ctx, "handled", "db_password", "do-not-log")
	line := output.String()
	for _, expected := range []string{"test-service", "req-1", spanContext.TraceID().String(), "[REDACTED]"} {
		if !strings.Contains(line, expected) {
			t.Fatalf("log output %q does not contain %q", line, expected)
		}
	}
	if strings.Contains(line, "do-not-log") {
		t.Fatalf("sensitive value leaked in %q", line)
	}
}
