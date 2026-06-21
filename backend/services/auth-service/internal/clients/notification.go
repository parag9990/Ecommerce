package clients

import (
	"context"
)

type SendOTPRequest struct {
	Target           string
	Channel          string
	Purpose          string
	OTP              string
	ChallengeID      string
	ExpiresInSeconds int
}

type NotificationClient interface {
	SendOTP(ctx context.Context, req SendOTPRequest) error
}
