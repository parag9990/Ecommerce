package domain

import "time"

type OTPChannel string

const (
	OTPChannelEmail OTPChannel = "email"
	OTPChannelPhone OTPChannel = "phone"
)

func (c OTPChannel) Valid() bool {
	switch c {
	case OTPChannelEmail, OTPChannelPhone:
		return true
	default:
		return false
	}
}

type OTPPurpose string

const (
	OTPPurposeSignup        OTPPurpose = "signup"
	OTPPurposeLogin         OTPPurpose = "login"
	OTPPurposePasswordReset OTPPurpose = "password_reset"
	OTPPurposePhoneVerify   OTPPurpose = "phone_verify"
	OTPPurposeEmailVerify   OTPPurpose = "email_verify"
)

func (p OTPPurpose) Valid() bool {
	switch p {
	case OTPPurposeSignup,
		OTPPurposeLogin,
		OTPPurposePasswordReset,
		OTPPurposePhoneVerify,
		OTPPurposeEmailVerify:
		return true
	default:
		return false
	}
}

type OTPChallenge struct {
	ChallengeID string
	AccountID   *string
	Target      string
	Channel     OTPChannel
	Purpose     OTPPurpose
	OTPHash     string
	Attempts    int
	MaxAttempts int
	VerifiedAt  *time.Time
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

func (c OTPChallenge) IsExpired(now time.Time) bool {
	return !now.UTC().Before(c.ExpiresAt.UTC())
}

func (c OTPChallenge) IsVerified() bool {
	return c.VerifiedAt != nil
}

func (c OTPChallenge) AttemptsExhausted() bool {
	return c.MaxAttempts > 0 && c.Attempts >= c.MaxAttempts
}

type OTPChallengeMutation struct {
	IncrementAttempts        bool
	MarkVerifiedAt           *time.Time
	MarkEmailVerifiedAccount string
	MarkPhoneVerifiedAccount string
}

func (m OTPChallengeMutation) IsZero() bool {
	return !m.IncrementAttempts &&
		m.MarkVerifiedAt == nil &&
		m.MarkEmailVerifiedAccount == "" &&
		m.MarkPhoneVerifiedAccount == ""
}

type OTPRateLimitError struct {
	Reason     string
	RetryAfter time.Duration
}

func (e *OTPRateLimitError) Error() string {
	if e == nil || e.Reason == "" {
		return ErrOTPRateLimited.Error()
	}
	return ErrOTPRateLimited.Error() + ": " + e.Reason
}

func (e *OTPRateLimitError) Is(target error) bool {
	return target == ErrOTPRateLimited
}

func (e *OTPRateLimitError) RetryAfterSeconds() int {
	if e == nil || e.RetryAfter <= 0 {
		return 0
	}
	seconds := int(e.RetryAfter.Seconds())
	if e.RetryAfter%time.Second != 0 {
		seconds++
	}
	if seconds < 1 {
		return 1
	}
	return seconds
}
