package clients

import (
	"context"
	"time"
)

func TimeoutFor(service Downstream) time.Duration {
	switch service {
	case DownstreamSession:
		return 200 * time.Millisecond
	case DownstreamAuth, DownstreamSearch:
		return 300 * time.Millisecond
	case DownstreamProduct:
		return 500 * time.Millisecond
	case DownstreamOrder, DownstreamPayment:
		return 1500 * time.Millisecond
	case DownstreamCMS, DownstreamSuperadmin:
		return time.Second
	default:
		return 700 * time.Millisecond
	}
}

func WithDeadline(ctx context.Context, service Downstream) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, TimeoutFor(service))
}
