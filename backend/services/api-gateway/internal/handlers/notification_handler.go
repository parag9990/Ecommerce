package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"ecommerce/api-gateway/internal/authctx"
	"ecommerce/api-gateway/internal/clients"
	"ecommerce/api-gateway/internal/middleware"
	notificationv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/notification/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type notificationPreferenceInput struct {
	EmailEnabled     *bool `json:"email_enabled"`
	SMSEnabled       *bool `json:"sms_enabled"`
	PushEnabled      *bool `json:"push_enabled"`
	MarketingEnabled *bool `json:"marketing_enabled"`
}

type notificationPreferenceResponse struct {
	EmailEnabled     bool `json:"email_enabled"`
	SMSEnabled       bool `json:"sms_enabled"`
	PushEnabled      bool `json:"push_enabled"`
	MarketingEnabled bool `json:"marketing_enabled"`
}

type NotificationHandler struct {
	client  clients.NotificationClient
	logger  *slog.Logger
	timeout time.Duration
}

func NewNotificationHandler(client clients.NotificationClient, logger *slog.Logger) *NotificationHandler {
	if client == nil {
		panic("notification client is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &NotificationHandler{client: client, logger: logger, timeout: defaultGRPCTimeout}
}

func (h *NotificationHandler) GetPreference(w http.ResponseWriter, r *http.Request) {
	ctx, cancel, ok := h.grpcContext(w, r)
	if !ok {
		return
	}
	defer cancel()

	preference, err := h.client.GetNotificationPreference(ctx, &notificationv1.GetNotificationPreferenceRequest{})
	if err != nil {
		h.writeGRPCError(w, r, err)
		return
	}
	writeData(w, r, http.StatusOK, mapNotificationPreference(preference))
}

func (h *NotificationHandler) UpdatePreference(w http.ResponseWriter, r *http.Request) {
	var input notificationPreferenceInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeAPIError(w, r, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if input.EmailEnabled == nil && input.SMSEnabled == nil && input.PushEnabled == nil && input.MarketingEnabled == nil {
		writeAPIError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "At least one preference field is required")
		return
	}

	ctx, cancel, ok := h.grpcContext(w, r)
	if !ok {
		return
	}
	defer cancel()
	preference, err := h.client.UpdateNotificationPreference(ctx, &notificationv1.UpdateNotificationPreferenceRequest{
		EmailEnabled: input.EmailEnabled, SmsEnabled: input.SMSEnabled,
		PushEnabled: input.PushEnabled, MarketingEnabled: input.MarketingEnabled,
	})
	if err != nil {
		h.writeGRPCError(w, r, err)
		return
	}
	writeData(w, r, http.StatusOK, mapNotificationPreference(preference))
}

func (h *NotificationHandler) grpcContext(
	w http.ResponseWriter,
	r *http.Request,
) (context.Context, context.CancelFunc, bool) {
	claims, ok := authctx.ClaimsFromContext(r.Context())
	if !ok || strings.TrimSpace(claims.UserID) == "" || len(claims.Roles) == 0 {
		writeAPIError(w, r, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication required")
		return nil, func() {}, false
	}
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	md := metadata.Pairs(
		"x-service-name", gatewayServiceName,
		"x-request-id", middleware.RequestIDFromRequest(r),
		"x-user-id", strings.TrimSpace(claims.UserID),
		"x-roles", strings.Join(claims.Roles, ","),
	)
	return metadata.NewOutgoingContext(ctx, md), cancel, true
}

func (h *NotificationHandler) writeGRPCError(w http.ResponseWriter, r *http.Request, err error) {
	grpcStatus, ok := status.FromError(err)
	if !ok {
		h.logger.ErrorContext(r.Context(), "gateway_notification_http_error", slog.String("error_type", "non_grpc"))
		writeAPIError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}
	switch grpcStatus.Code() {
	case codes.InvalidArgument:
		writeAPIError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", grpcStatus.Message())
	case codes.Unauthenticated:
		writeAPIError(w, r, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication required")
	case codes.PermissionDenied:
		writeAPIError(w, r, http.StatusForbidden, "PERMISSION_DENIED", "Permission denied")
	case codes.DeadlineExceeded:
		writeAPIError(w, r, http.StatusGatewayTimeout, "UPSTREAM_TIMEOUT", "Notification service timed out")
	case codes.Unavailable:
		writeAPIError(w, r, http.StatusServiceUnavailable, "UPSTREAM_UNAVAILABLE", "Notification service is temporarily unavailable")
	default:
		h.logger.ErrorContext(r.Context(), "gateway_notification_grpc_error", slog.String("code", grpcStatus.Code().String()))
		writeAPIError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

func mapNotificationPreference(preference *notificationv1.NotificationPreference) notificationPreferenceResponse {
	if preference == nil {
		return notificationPreferenceResponse{}
	}
	return notificationPreferenceResponse{
		EmailEnabled: preference.GetEmailEnabled(), SMSEnabled: preference.GetSmsEnabled(),
		PushEnabled: preference.GetPushEnabled(), MarketingEnabled: preference.GetMarketingEnabled(),
	}
}
