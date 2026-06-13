package grpctransport

import (
	"context"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	metadataUserID      = "x-user-id"
	metadataSessionID   = "x-session-id"
	metadataRoles       = "x-roles"
	metadataSellerID    = "x-seller-id"
	metadataTenantID    = "x-tenant-id"
	metadataStaffStatus = "x-staff-status"
	metadataRequestID   = "x-request-id"
	metadataTraceID     = "x-trace-id"
	metadataServiceName = "x-service-name"
)

type requestMetadata struct {
	Actor       domain.ActorContext
	ServiceName string
	RequestID   string
	TraceID     string
}

func metadataFromContext(ctx context.Context) requestMetadata {
	md, _ := metadata.FromIncomingContext(ctx)
	requestID := metadataValue(md, metadataRequestID)
	traceID := metadataValue(md, metadataTraceID)
	if requestID == "" {
		requestID = traceID
	}
	return requestMetadata{
		Actor: domain.ActorContext{
			UserID:      metadataValue(md, metadataUserID),
			SessionID:   metadataValue(md, metadataSessionID),
			SellerID:    metadataValue(md, metadataSellerID),
			TenantID:    metadataValue(md, metadataTenantID),
			RequestID:   requestID,
			Roles:       domain.NormalizeRoles(splitCommaValues(md.Get(metadataRoles))),
			StaffStatus: domain.StaffStatus(metadataValue(md, metadataStaffStatus)),
		}.Normalized(),
		ServiceName: strings.TrimSpace(metadataValue(md, metadataServiceName)),
		RequestID:   requestID,
		TraceID:     traceID,
	}
}

func (s *Server) requireInternalCaller(ctx context.Context, methodAllowed map[string]struct{}) (requestMetadata, error) {
	md := metadataFromContext(ctx)
	if md.ServiceName == "" {
		return md, status.Error(codes.Unauthenticated, "internal service metadata is required")
	}
	if len(s.allowedCaller) > 0 {
		if _, ok := s.allowedCaller[md.ServiceName]; !ok {
			return md, status.Error(codes.PermissionDenied, "internal service is not allowed")
		}
	}
	if len(methodAllowed) > 0 {
		if _, ok := methodAllowed[md.ServiceName]; !ok {
			return md, status.Error(codes.PermissionDenied, "internal service is not allowed for this method")
		}
	}
	if s.config.InternalAuthToken != "" {
		incoming, _ := metadata.FromIncomingContext(ctx)
		if metadataValue(incoming, strings.ToLower(s.config.InternalAuthHeader)) != s.config.InternalAuthToken {
			return md, status.Error(codes.Unauthenticated, "internal authorization is required")
		}
	}
	return md, nil
}

func (s *Server) internalCallerIfAuthenticated(ctx context.Context, methodAllowed map[string]struct{}) (requestMetadata, bool, error) {
	md := metadataFromContext(ctx)
	if md.Actor.Authenticated() {
		return md, false, nil
	}
	authorized, err := s.requireInternalCaller(ctx, methodAllowed)
	if err != nil {
		return md, false, err
	}
	return authorized, true, nil
}

func metadataValue(md metadata.MD, key string) string {
	values := md.Get(strings.ToLower(key))
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

func splitCommaValues(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

func allowedCallers(values ...string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out[value] = struct{}{}
		}
	}
	return out
}
