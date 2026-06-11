package domain

import "errors"

var (
	ErrCredentialNotFound  = errors.New("credential not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrAccountLocked       = errors.New("account temporarily locked")
	ErrAccountInactive     = errors.New("account cannot authenticate")
	ErrDuplicateCredential = errors.New("credential already exists")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenReuse   = errors.New("refresh token reuse detected")
	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrTokenSubjectMissing = errors.New("token subject is incomplete")
	ErrRoleRequired        = errors.New("at least one role is required")
	ErrAccountNotFound     = errors.New("account not found")
	ErrUnauthenticated     = errors.New("authentication required")
	ErrForbidden           = errors.New("permission denied")

	ErrUnknownRole            = errors.New("unknown role")
	ErrInvalidRoleAssignment  = errors.New("invalid role assignment")
	ErrInvalidRoleScope       = errors.New("invalid role scope")
	ErrAuditReasonRequired    = errors.New("role change reason is required")
	ErrRoleAlreadyAssigned    = errors.New("role is already assigned")
	ErrRoleAssignmentNotFound = errors.New("role assignment not found")

	ErrInvalidOTPRequest       = errors.New("invalid otp request")
	ErrOTPChallengeNotFound    = errors.New("otp challenge not found")
	ErrInvalidOTP              = errors.New("invalid otp")
	ErrOTPExpired              = errors.New("otp expired")
	ErrOTPAlreadyUsed          = errors.New("otp already used")
	ErrOTPAttemptsExceeded     = errors.New("otp attempts exceeded")
	ErrOTPRateLimited          = errors.New("otp rate limited")
	ErrOTPRateLimitUnavailable = errors.New("otp rate limiter unavailable")
	ErrOTPDeliveryFailed       = errors.New("otp delivery failed")
)
