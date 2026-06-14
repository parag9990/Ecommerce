package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var otpDigitsPattern = regexp.MustCompile(`^[0-9]{6,}$`)

type OTPPurpose string

const (
	OTPPurposeSignup        OTPPurpose = "signup"
	OTPPurposeLogin         OTPPurpose = "login"
	OTPPurposePasswordReset OTPPurpose = "password_reset"
	OTPPurposePhoneVerify   OTPPurpose = "phone_verify"
	OTPPurposeEmailVerify   OTPPurpose = "email_verify"
)

// SendOTPRequest contains a plaintext OTP only while it is being rendered and
// dispatched. It must never be persisted or included in application logs.
type SendOTPRequest struct {
	ChallengeID      string
	UserID           string
	Channel          Channel
	Target           string
	OTP              string
	ExpiresInMinutes int
	Purpose          string
	CorrelationID    string
}

type SendOTPResult struct {
	DeliveryID string
	Status     string
}

func (r SendOTPRequest) Validate() error {
	if strings.TrimSpace(r.ChallengeID) == "" {
		return fmt.Errorf("%w: challenge id is required", ErrInvalidOTPRequest)
	}
	if r.Channel != ChannelEmail && r.Channel != ChannelSMS {
		return fmt.Errorf("%w: OTP supports email or sms only", ErrInvalidOTPRequest)
	}
	if strings.TrimSpace(r.Target) == "" {
		return fmt.Errorf("%w: target is required", ErrInvalidOTPRequest)
	}
	if !otpDigitsPattern.MatchString(strings.TrimSpace(r.OTP)) {
		return fmt.Errorf("%w: OTP must contain at least 6 digits", ErrInvalidOTPRequest)
	}
	if r.ExpiresInMinutes <= 0 {
		return fmt.Errorf("%w: expiry is required", ErrInvalidOTPRequest)
	}
	if !validOTPPurpose(strings.TrimSpace(r.Purpose)) {
		return fmt.Errorf("%w: unsupported purpose", ErrInvalidOTPRequest)
	}
	return nil
}

func validOTPPurpose(value string) bool {
	switch OTPPurpose(value) {
	case OTPPurposeSignup, OTPPurposeLogin, OTPPurposePasswordReset, OTPPurposePhoneVerify, OTPPurposeEmailVerify:
		return true
	default:
		return false
	}
}
