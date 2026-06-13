package requestctx

import (
	"context"
	"strings"
	"unicode/utf8"
)

const (
	DefaultClientPath = "/search"
	maxClientPathLen  = 256
)

type AnalyticsContext struct {
	RequestID   string
	AnonymousID string
	SessionID   string
	UserID      string
	ClientPath  string
}

type analyticsContextKey struct{}

func WithAnalyticsContext(ctx context.Context, value AnalyticsContext) context.Context {
	value = value.Normalize()
	return context.WithValue(ctx, analyticsContextKey{}, value)
}

func Analytics(ctx context.Context) AnalyticsContext {
	value, _ := ctx.Value(analyticsContextKey{}).(AnalyticsContext)
	return value.Normalize()
}

func (a AnalyticsContext) Normalize() AnalyticsContext {
	a.RequestID = strings.TrimSpace(a.RequestID)
	a.AnonymousID = strings.TrimSpace(a.AnonymousID)
	a.SessionID = strings.TrimSpace(a.SessionID)
	a.UserID = strings.TrimSpace(a.UserID)
	a.ClientPath = NormalizeClientPath(a.ClientPath)
	return a
}

func (a AnalyticsContext) CanIngestSessionEvent() bool {
	a = a.Normalize()
	return a.AnonymousID != "" && a.SessionID != ""
}

func NormalizeClientPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return DefaultClientPath
	}
	if index := strings.IndexAny(path, "?#"); index >= 0 {
		path = path[:index]
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return DefaultClientPath
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if utf8.RuneCountInString(path) > maxClientPathLen {
		path = string([]rune(path)[:maxClientPathLen])
	}
	return path
}
