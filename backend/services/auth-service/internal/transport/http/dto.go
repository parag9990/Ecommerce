package httptransport

import "time"

type createCredentialRequest struct {
	AccountID string `json:"account_id"`
	Password  string `json:"password"`
}

type createCredentialResponse struct {
	AccountID         string    `json:"account_id"`
	PasswordAlgo      string    `json:"password_algo"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
}

type verifyPasswordRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type verifyPasswordResponse struct {
	AccountID        string `json:"account_id"`
	AccountStatus    string `json:"account_status"`
	EmailVerified    bool   `json:"email_verified"`
	PhoneVerified    bool   `json:"phone_verified"`
	PasswordValid    bool   `json:"password_valid"`
	PasswordRehashed bool   `json:"password_rehashed"`
}

type resetPasswordRequest struct {
	AccountID   string `json:"account_id"`
	NewPassword string `json:"new_password"`
}

type signupRequest struct {
	Email    string        `json:"email"`
	Phone    string        `json:"phone,omitempty"`
	FullName string        `json:"full_name"`
	Password string        `json:"password"`
	Role     string        `json:"role,omitempty"`
	Device   deviceRequest `json:"device,omitempty"`
}

type sendOTPRequest struct {
	AccountID *string `json:"account_id,omitempty"`
	Target    string  `json:"target"`
	Channel   string  `json:"channel"`
	Purpose   string  `json:"purpose"`
}

type forgotPasswordRequest struct {
	Identifier string `json:"identifier"`
	Target     string `json:"target"`
	Channel    string `json:"channel"`
}

type verifyOTPRequest struct {
	ChallengeID string `json:"challenge_id"`
	OTP         string `json:"otp"`
}

type otpChallengeResponse struct {
	ChallengeID        string    `json:"challenge_id"`
	ExpiresAt          time.Time `json:"expires_at"`
	ExpiresIn          int       `json:"expires_in"`
	ResendAfterSeconds int       `json:"resend_after_seconds"`
}

type loginRequest struct {
	Identifier string        `json:"identifier"`
	Password   string        `json:"password"`
	Device     deviceRequest `json:"device,omitempty"`
}

type deviceRequest struct {
	AnonymousID           string `json:"anonymous_id,omitempty"`
	Fingerprint           string `json:"fingerprint,omitempty"`
	DeviceFingerprintHash string `json:"device_fingerprint_hash,omitempty"`
	UserAgent             string `json:"user_agent,omitempty"`
	Channel               string `json:"channel,omitempty"`
	Locale                string `json:"locale,omitempty"`
	Timezone              string `json:"timezone,omitempty"`
}

type issueTokenPairRequest struct {
	AccountID string   `json:"account_id"`
	UserID    string   `json:"user_id"`
	SessionID string   `json:"session_id"`
	Roles     []string `json:"roles"`
	SellerID  string   `json:"seller_id"`
	TenantID  string   `json:"tenant_id"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
	AllDevices   bool   `json:"all_devices"`
}

type verifyAccessTokenRequest struct {
	AccessToken string `json:"access_token"`
}

type assignRoleRequest struct {
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	ScopeType string `json:"scope_type,omitempty"`
	ScopeID   string `json:"scope_id,omitempty"`
	Reason    string `json:"reason"`
}

type revokeRoleRequest struct {
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	ScopeType string `json:"scope_type,omitempty"`
	ScopeID   string `json:"scope_id,omitempty"`
	Reason    string `json:"reason"`
}

type authSessionResponse struct {
	User      authUserResponse `json:"user"`
	Tokens    tokenResponse    `json:"tokens"`
	SessionID string           `json:"session_id"`
}

type authUserResponse struct {
	UserID   string   `json:"user_id"`
	Roles    []string `json:"roles"`
	SellerID string   `json:"seller_id,omitempty"`
	TenantID string   `json:"tenant_id,omitempty"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type tokenClaimsResponse struct {
	UserID    string   `json:"user_id"`
	SessionID string   `json:"session_id"`
	Roles     []string `json:"roles"`
	SellerID  string   `json:"seller_id,omitempty"`
	TenantID  string   `json:"tenant_id,omitempty"`
}

type roleMutationResponse struct {
	Success bool                   `json:"success"`
	Role    roleAssignmentResponse `json:"role"`
}

type userRolesResponse struct {
	UserID string                   `json:"user_id"`
	Roles  []roleAssignmentResponse `json:"roles"`
}

type roleAssignmentResponse struct {
	Role       string     `json:"role"`
	ScopeType  string     `json:"scope_type,omitempty"`
	ScopeID    string     `json:"scope_id,omitempty"`
	AssignedAt time.Time  `json:"assigned_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

type successResponse struct {
	Success bool `json:"success"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
