package events

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

const (
	metadataRequestID   = "x-request-id"
	metadataTraceID     = "x-trace-id"
	metadataTraceparent = "traceparent"
)

func requestIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md, ok = metadata.FromOutgoingContext(ctx)
	}
	if !ok {
		return ""
	}
	return firstMetadataValue(md, metadataRequestID)
}

func traceIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md, ok = metadata.FromOutgoingContext(ctx)
	}
	if !ok {
		return ""
	}
	if traceID := firstMetadataValue(md, metadataTraceID); traceID != "" {
		return traceID
	}
	return traceIDFromTraceparent(firstMetadataValue(md, metadataTraceparent))
}

func firstMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

func traceIDFromTraceparent(traceparent string) string {
	parts := strings.Split(strings.TrimSpace(traceparent), "-")
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}
