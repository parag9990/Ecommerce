package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

type Logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

type SlogLogger struct {
	logger *slog.Logger
}

func New(serviceName string, level string) SlogLogger {
	var slogLevel slog.Level
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn", "warning":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel})
	return SlogLogger{logger: slog.New(handler).With("service", serviceName)}
}

func NewNop() SlogLogger {
	return SlogLogger{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
}

func (l SlogLogger) Debug(ctx context.Context, msg string, args ...any) {
	l.logger.DebugContext(ctx, msg, args...)
}

func (l SlogLogger) Info(ctx context.Context, msg string, args ...any) {
	l.logger.InfoContext(ctx, msg, args...)
}

func (l SlogLogger) Warn(ctx context.Context, msg string, args ...any) {
	l.logger.WarnContext(ctx, msg, args...)
}

func (l SlogLogger) Error(ctx context.Context, msg string, args ...any) {
	l.logger.ErrorContext(ctx, msg, args...)
}
