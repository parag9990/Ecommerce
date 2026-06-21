package clients

import (
	"context"
	"errors"
	"testing"
	"time"

	notificationv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/notification/v1"
	"google.golang.org/grpc"
)

type notificationRPCStub struct {
	request *notificationv1.SendOTPRequest
	result  *notificationv1.SendOTPResponse
	err     error
}

func (s *notificationRPCStub) SendOTP(_ context.Context, request *notificationv1.SendOTPRequest, _ ...grpc.CallOption) (*notificationv1.SendOTPResponse, error) {
	s.request = request
	return s.result, s.err
}

func TestGRPCNotificationClientMapsOTPContract(t *testing.T) {
	rpc := &notificationRPCStub{result: &notificationv1.SendOTPResponse{DeliveryId: "delivery_123", Status: "accepted"}}
	client := NewGRPCNotificationClient(rpc, time.Second)

	err := client.SendOTP(context.Background(), SendOTPRequest{
		Target:           "+919876543210",
		Channel:          "phone",
		Purpose:          "password_reset",
		OTP:              "482991",
		ChallengeID:      "otp_123",
		ExpiresInSeconds: 301,
	})
	if err != nil {
		t.Fatalf("SendOTP() error = %v", err)
	}
	if rpc.request.GetChannel() != notificationv1.OTPChannel_OTP_CHANNEL_SMS {
		t.Fatalf("channel = %v", rpc.request.GetChannel())
	}
	if rpc.request.GetExpiresInMinutes() != 6 {
		t.Fatalf("expires_in_minutes = %d", rpc.request.GetExpiresInMinutes())
	}
	if rpc.request.GetOtp() != "482991" || rpc.request.GetChallengeId() != "otp_123" {
		t.Fatalf("request mapping = %+v", rpc.request)
	}
}

func TestGRPCNotificationClientRejectsInvalidOrIncompleteResponse(t *testing.T) {
	tests := []struct {
		name    string
		request SendOTPRequest
		result  *notificationv1.SendOTPResponse
		err     error
	}{
		{
			name:    "invalid channel",
			request: validNotificationOTPRequest("push"),
			result:  &notificationv1.SendOTPResponse{DeliveryId: "delivery_123"},
		},
		{
			name:    "missing delivery id",
			request: validNotificationOTPRequest("email"),
			result:  &notificationv1.SendOTPResponse{Status: "accepted"},
		},
		{
			name:    "rpc failure",
			request: validNotificationOTPRequest("email"),
			err:     errors.New("provider unavailable"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rpc := &notificationRPCStub{result: test.result, err: test.err}
			client := NewGRPCNotificationClient(rpc, time.Second)
			if err := client.SendOTP(context.Background(), test.request); err == nil {
				t.Fatal("SendOTP() error = nil")
			}
		})
	}
}

func validNotificationOTPRequest(channel string) SendOTPRequest {
	return SendOTPRequest{
		Target:           "buyer@example.com",
		Channel:          channel,
		Purpose:          "login",
		OTP:              "482991",
		ChallengeID:      "otp_123",
		ExpiresInSeconds: 300,
	}
}
