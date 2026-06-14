package grpc

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	notificationv1 "github.com/example/ecommerce-platform/backend/services/notification-service/api/notification/v1"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type sendOTPStub struct {
	request domain.SendOTPRequest
	result  domain.SendOTPResult
	err     error
	calls   int
}

type preferenceStub struct {
	userID string
	patch  usecase.PreferencePatch
	result domain.Preference
	err    error
	calls  int
}

func (s *preferenceStub) Get(_ context.Context, userID string) (domain.Preference, error) {
	s.calls++
	s.userID = userID
	return s.result, s.err
}

func (s *preferenceStub) Update(_ context.Context, userID string, patch usecase.PreferencePatch) (domain.Preference, error) {
	s.calls++
	s.userID, s.patch = userID, patch
	return s.result, s.err
}

func (s *sendOTPStub) Send(_ context.Context, request domain.SendOTPRequest) (domain.SendOTPResult, error) {
	s.calls++
	s.request = request
	return s.result, s.err
}

func TestNotificationHandlerSendOTPMapsRequest(t *testing.T) {
	t.Parallel()

	service := &sendOTPStub{result: domain.SendOTPResult{DeliveryID: "delivery_1", Status: "accepted"}}
	handler := newHandler(t, service)
	got, err := handler.SendOTP(context.Background(), &notificationv1.SendOTPRequest{
		ChallengeId:      "challenge_1",
		UserId:           "user_1",
		Channel:          notificationv1.OTPChannel_OTP_CHANNEL_EMAIL,
		Target:           "person@example.com",
		Otp:              "482991",
		ExpiresInMinutes: 5,
		Purpose:          "login",
		CorrelationId:    "request_1",
	})
	if err != nil {
		t.Fatalf("SendOTP() error = %v", err)
	}
	if got.GetDeliveryId() != "delivery_1" || got.GetStatus() != "accepted" {
		t.Fatalf("SendOTP() response = %+v", got)
	}
	if service.request.Channel != domain.ChannelEmail || service.request.OTP != "482991" ||
		service.request.Target != "person@example.com" {
		t.Fatalf("mapped request = %+v", service.request)
	}
}

func TestNotificationHandlerRejectsUnsupportedChannelBeforeUsecase(t *testing.T) {
	t.Parallel()

	service := &sendOTPStub{}
	handler := newHandler(t, service)
	_, err := handler.SendOTP(context.Background(), &notificationv1.SendOTPRequest{
		Channel: notificationv1.OTPChannel_OTP_CHANNEL_UNSPECIFIED,
	})
	if status.Code(err) != codes.InvalidArgument || service.calls != 0 {
		t.Fatalf("SendOTP() error = %v, calls = %d", err, service.calls)
	}
}

func TestNotificationHandlerSanitizesFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "validation", err: domain.ErrInvalidOTPRequest, code: codes.InvalidArgument},
		{name: "template", err: domain.ErrTemplateNotFound, code: codes.FailedPrecondition},
		{name: "provider", err: provider.ErrProviderUnavailable, code: codes.Unavailable},
		{name: "deadline", err: context.DeadlineExceeded, code: codes.DeadlineExceeded},
		{name: "unknown", err: errors.New("OTP 482991 for person@example.com"), code: codes.Internal},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			handler := newHandler(t, &sendOTPStub{err: test.err})
			_, err := handler.SendOTP(context.Background(), &notificationv1.SendOTPRequest{
				Channel: notificationv1.OTPChannel_OTP_CHANNEL_SMS,
			})
			if status.Code(err) != test.code {
				t.Fatalf("SendOTP() code = %v, want %v", status.Code(err), test.code)
			}
			if strings.Contains(err.Error(), "482991") || strings.Contains(err.Error(), "person@example.com") {
				t.Fatalf("SendOTP() exposed sensitive failure: %v", err)
			}
		})
	}
}

func TestNotificationHandlerGetsAuthenticatedUsersPreference(t *testing.T) {
	t.Parallel()

	preferences := &preferenceStub{result: domain.Preference{EmailEnabled: true}}
	handler := newHandlerWithPreferences(t, &sendOTPStub{}, preferences)
	got, err := handler.GetNotificationPreference(authenticatedContext(), &notificationv1.GetNotificationPreferenceRequest{})
	if err != nil || preferences.userID != "user_1" || !got.GetEmailEnabled() {
		t.Fatalf("GetNotificationPreference() = %+v, %v; user = %q", got, err, preferences.userID)
	}
}

func TestNotificationHandlerPatchPreservesExplicitFalse(t *testing.T) {
	t.Parallel()

	disabled := false
	preferences := &preferenceStub{}
	handler := newHandlerWithPreferences(t, &sendOTPStub{}, preferences)
	_, err := handler.UpdateNotificationPreference(authenticatedContext(), &notificationv1.UpdateNotificationPreferenceRequest{
		MarketingEnabled: &disabled,
	})
	if err != nil || preferences.patch.MarketingEnabled == nil || *preferences.patch.MarketingEnabled {
		t.Fatalf("UpdateNotificationPreference() error = %v; patch = %+v", err, preferences.patch)
	}
}

func TestNotificationHandlerRequiresGatewayAuthMetadataForPreferences(t *testing.T) {
	t.Parallel()

	preferences := &preferenceStub{}
	handler := newHandlerWithPreferences(t, &sendOTPStub{}, preferences)
	_, err := handler.GetNotificationPreference(context.Background(), &notificationv1.GetNotificationPreferenceRequest{})
	if status.Code(err) != codes.Unauthenticated || preferences.calls != 0 {
		t.Fatalf("GetNotificationPreference() error = %v; calls = %d", err, preferences.calls)
	}
}

func newHandler(t *testing.T, service OTPService) *NotificationHandler {
	t.Helper()
	return newHandlerWithPreferences(t, service, &preferenceStub{})
}

func newHandlerWithPreferences(t *testing.T, service OTPService, preferences PreferenceService) *NotificationHandler {
	t.Helper()
	handler, err := NewNotificationHandler(service, preferences, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewNotificationHandler() error = %v", err)
	}
	return handler
}

func authenticatedContext() context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-user-id", "user_1",
		"x-roles", "buyer",
	))
}
