package clients

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	notificationv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/notification/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NotificationRPC interface {
	SendOTP(ctx context.Context, in *notificationv1.SendOTPRequest, opts ...grpc.CallOption) (*notificationv1.SendOTPResponse, error)
}

type GRPCNotificationClient struct {
	client  NotificationRPC
	timeout time.Duration
}

func DialGRPCNotificationClient(address string, timeout time.Duration) (*GRPCNotificationClient, *grpc.ClientConn, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return nil, nil, errors.New("notification gRPC address is required")
	}
	if timeout <= 0 {
		return nil, nil, errors.New("notification timeout must be greater than zero")
	}

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("create notification gRPC client: %w", err)
	}
	return NewGRPCNotificationClient(notificationv1.NewNotificationServiceClient(conn), timeout), conn, nil
}

func NewGRPCNotificationClient(client NotificationRPC, timeout time.Duration) *GRPCNotificationClient {
	return &GRPCNotificationClient{client: client, timeout: timeout}
}

func (c *GRPCNotificationClient) SendOTP(ctx context.Context, req SendOTPRequest) error {
	if c == nil || c.client == nil {
		return errors.New("notification client is not initialized")
	}
	if c.timeout <= 0 {
		return errors.New("notification timeout must be greater than zero")
	}
	if strings.TrimSpace(req.Target) == "" ||
		strings.TrimSpace(req.Channel) == "" ||
		strings.TrimSpace(req.Purpose) == "" ||
		strings.TrimSpace(req.OTP) == "" ||
		strings.TrimSpace(req.ChallengeID) == "" ||
		req.ExpiresInSeconds <= 0 {
		return errors.New("notification otp request is incomplete")
	}

	channel, err := notificationChannel(req.Channel)
	if err != nil {
		return err
	}

	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	response, err := c.client.SendOTP(callCtx, &notificationv1.SendOTPRequest{
		ChallengeId:      strings.TrimSpace(req.ChallengeID),
		Channel:          channel,
		Target:           strings.TrimSpace(req.Target),
		Otp:              strings.TrimSpace(req.OTP),
		ExpiresInMinutes: int32(minutesCeil(req.ExpiresInSeconds)),
		Purpose:          strings.TrimSpace(req.Purpose),
		CorrelationId:    strings.TrimSpace(req.ChallengeID),
	})
	if err != nil {
		return fmt.Errorf("send notification OTP: %w", err)
	}
	if response == nil || strings.TrimSpace(response.GetDeliveryId()) == "" {
		return errors.New("notification OTP response is incomplete")
	}
	return nil
}

func notificationChannel(value string) (notificationv1.OTPChannel, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "email":
		return notificationv1.OTPChannel_OTP_CHANNEL_EMAIL, nil
	case "phone", "sms":
		return notificationv1.OTPChannel_OTP_CHANNEL_SMS, nil
	default:
		return notificationv1.OTPChannel_OTP_CHANNEL_UNSPECIFIED, errors.New("notification OTP channel is invalid")
	}
}

func minutesCeil(seconds int) int {
	minutes := seconds / 60
	if seconds%60 != 0 {
		minutes++
	}
	if minutes < 1 {
		return 1
	}
	return minutes
}
