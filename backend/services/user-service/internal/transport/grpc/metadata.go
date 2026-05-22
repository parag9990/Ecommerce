package grpc

import (
	"context"
	"strings"

	"github.com/parag/ecommerce/backend/services/user-service/internal/usecase"
	"google.golang.org/grpc/metadata"
)

const (
	metadataUserID      = "x-user-id"
	metadataServiceName = "x-service-name"
	metadataRoles       = "x-roles"
	metadataRequestID   = "x-request-id"
)

func callerFromContext(ctx context.Context) usecase.Caller {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return usecase.Caller{}
	}

	return usecase.Caller{
		UserID:      firstMetadataValue(md, metadataUserID),
		ServiceName: firstMetadataValue(md, metadataServiceName),
		Roles:       splitMetadataValues(md.Get(metadataRoles)),
	}
}

func requestIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	return firstMetadataValue(md, metadataRequestID)
}

func firstMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

func splitMetadataValues(values []string) []string {
	roles := make([]string, 0)
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			role := strings.TrimSpace(part)
			if role != "" {
				roles = append(roles, role)
			}
		}
	}
	return roles
}
