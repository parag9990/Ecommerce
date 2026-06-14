package grpc

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	notificationv1 "github.com/example/ecommerce-platform/backend/services/notification-service/api/notification/v1"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type OTPService interface {
	Send(ctx context.Context, req domain.SendOTPRequest) (domain.SendOTPResult, error)
}

type PreferenceService interface {
	Get(ctx context.Context, userID string) (domain.Preference, error)
	Update(ctx context.Context, userID string, patch usecase.PreferencePatch) (domain.Preference, error)
}

type NotificationHandler struct {
	notificationv1.UnimplementedNotificationServiceServer
	sendOTP     OTPService
	preferences PreferenceService
	logger      *slog.Logger
}

func NewNotificationHandler(sendOTP OTPService, preferences PreferenceService, logger *slog.Logger) (*NotificationHandler, error) {
	if sendOTP == nil || preferences == nil {
		return nil, errors.New("notification handler services are required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &NotificationHandler{sendOTP: sendOTP, preferences: preferences, logger: logger}, nil
}

func (h *NotificationHandler) SendOTP(
	ctx context.Context,
	in *notificationv1.SendOTPRequest,
) (*notificationv1.SendOTPResponse, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "OTP delivery request is required")
	}
	channel, err := otpChannelFromProto(in.GetChannel())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "unsupported OTP channel")
	}

	result, err := h.sendOTP.Send(ctx, domain.SendOTPRequest{
		ChallengeID:      in.GetChallengeId(),
		UserID:           in.GetUserId(),
		Channel:          channel,
		Target:           in.GetTarget(),
		OTP:              in.GetOtp(),
		ExpiresInMinutes: int(in.GetExpiresInMinutes()),
		Purpose:          in.GetPurpose(),
		CorrelationID:    in.GetCorrelationId(),
	})
	if err != nil {
		mapped := mapSendOTPError(err)
		h.logger.WarnContext(ctx, "notification.otp.rpc_failed",
			slog.String("channel", string(channel)),
			slog.String("grpc_code", status.Code(mapped).String()),
		)
		return nil, mapped
	}
	return &notificationv1.SendOTPResponse{
		DeliveryId: result.DeliveryID,
		Status:     result.Status,
	}, nil
}

func (h *NotificationHandler) GetNotificationPreference(
	ctx context.Context,
	in *notificationv1.GetNotificationPreferenceRequest,
) (*notificationv1.NotificationPreference, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "preference request is required")
	}
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}
	preference, err := h.preferences.Get(ctx, userID)
	if err != nil {
		return nil, mapPreferenceError(err)
	}
	return preferenceResponse(preference), nil
}

func (h *NotificationHandler) UpdateNotificationPreference(
	ctx context.Context,
	in *notificationv1.UpdateNotificationPreferenceRequest,
) (*notificationv1.NotificationPreference, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "preference update is required")
	}
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}
	patch := usecase.PreferencePatch{
		EmailEnabled: in.EmailEnabled, SMSEnabled: in.SmsEnabled,
		PushEnabled: in.PushEnabled, MarketingEnabled: in.MarketingEnabled,
	}
	preference, err := h.preferences.Update(ctx, userID, patch)
	if err != nil {
		return nil, mapPreferenceError(err)
	}
	h.logger.InfoContext(ctx, "notification.preference.updated",
		slog.String("user_id", userID),
		slog.Int("fields_changed", preferenceFieldCount(patch)),
	)
	return preferenceResponse(preference), nil
}

func preferenceResponse(preference domain.Preference) *notificationv1.NotificationPreference {
	return &notificationv1.NotificationPreference{
		EmailEnabled:     preference.EmailEnabled,
		SmsEnabled:       preference.SMSEnabled,
		PushEnabled:      preference.PushEnabled,
		MarketingEnabled: preference.MarketingEnabled,
	}
}

func authenticatedUserID(ctx context.Context) (string, error) {
	values, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "authenticated user context is required")
	}
	userIDs := values.Get("x-user-id")
	if len(userIDs) != 1 || strings.TrimSpace(userIDs[0]) == "" || !permittedUserRole(values.Get("x-roles")) {
		return "", status.Error(codes.Unauthenticated, "authenticated user context is required")
	}
	return strings.TrimSpace(userIDs[0]), nil
}

func permittedUserRole(values []string) bool {
	allowed := map[string]struct{}{"buyer": {}, "seller": {}, "admin": {}, "superadmin": {}}
	for _, value := range values {
		for _, role := range strings.Split(value, ",") {
			if _, exists := allowed[strings.TrimSpace(role)]; exists {
				return true
			}
		}
	}
	return false
}

func preferenceFieldCount(patch usecase.PreferencePatch) int {
	count := 0
	for _, value := range []*bool{patch.EmailEnabled, patch.SMSEnabled, patch.PushEnabled, patch.MarketingEnabled} {
		if value != nil {
			count++
		}
	}
	return count
}

func mapPreferenceError(err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "preference operation canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "preference operation deadline exceeded")
	case errors.Is(err, domain.ErrInvalidPreference):
		return status.Error(codes.InvalidArgument, "invalid notification preference update")
	default:
		return status.Error(codes.Internal, "notification preference operation failed")
	}
}

func otpChannelFromProto(channel notificationv1.OTPChannel) (domain.Channel, error) {
	switch channel {
	case notificationv1.OTPChannel_OTP_CHANNEL_EMAIL:
		return domain.ChannelEmail, nil
	case notificationv1.OTPChannel_OTP_CHANNEL_SMS:
		return domain.ChannelSMS, nil
	default:
		return "", domain.ErrUnsupportedChannel
	}
}

func mapSendOTPError(err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "OTP delivery canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "OTP delivery deadline exceeded")
	case errors.Is(err, domain.ErrInvalidOTPRequest),
		errors.Is(err, domain.ErrInvalidRecipient),
		errors.Is(err, domain.ErrInvalidContent),
		errors.Is(err, domain.ErrUnsupportedChannel),
		errors.Is(err, provider.ErrProviderRejected):
		return status.Error(codes.InvalidArgument, "invalid OTP delivery request")
	case errors.Is(err, domain.ErrInvalidRenderRequest),
		errors.Is(err, domain.ErrTemplateNotFound),
		errors.Is(err, domain.ErrTemplateRender),
		errors.Is(err, provider.ErrProviderNotConfigured),
		errors.Is(err, provider.ErrChannelDisabled):
		return status.Error(codes.FailedPrecondition, "OTP delivery is not configured")
	case errors.Is(err, provider.ErrProviderUnavailable):
		return status.Error(codes.Unavailable, "OTP delivery provider unavailable")
	default:
		return status.Error(codes.Internal, "OTP delivery failed")
	}
}
