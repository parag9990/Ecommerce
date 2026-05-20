package httptransport

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/password"
	tokensecurity "github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/token"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/usecase"
)

const maxRequestBodyBytes = 1 << 20

type PasswordUsecase interface {
	CreateCredential(ctx context.Context, input usecase.CreateCredentialInput) (usecase.CredentialSummary, error)
	VerifyPassword(ctx context.Context, input usecase.VerifyPasswordInput) (usecase.VerifyPasswordOutput, error)
	ResetPassword(ctx context.Context, input usecase.ResetPasswordInput) (usecase.CredentialSummary, error)
}

type AuthUsecase interface {
	Login(ctx context.Context, input usecase.LoginInput) (usecase.AuthSession, error)
}

type TokenUsecase interface {
	IssueTokenPair(ctx context.Context, input usecase.IssueTokenPairInput) (usecase.TokenPair, error)
	RefreshToken(ctx context.Context, input usecase.RefreshTokenInput) (usecase.TokenPair, error)
	Logout(ctx context.Context, input usecase.LogoutInput) error
	VerifyAccessToken(ctx context.Context, accessToken string) (domain.TokenClaims, error)
	JWKS(ctx context.Context) tokensecurity.JWKSet
}

type OTPUsecase interface {
	CreateOTPChallenge(ctx context.Context, input usecase.CreateOTPChallengeInput) (usecase.CreateOTPChallengeOutput, error)
	VerifyOTP(ctx context.Context, input usecase.VerifyOTPInput) error
}

type RoleUsecase interface {
	AssignRole(ctx context.Context, input usecase.AssignRoleInput) (domain.RoleAssignment, error)
	RevokeRole(ctx context.Context, input usecase.RevokeRoleInput) error
	GetUserRoles(ctx context.Context, input usecase.GetUserRolesInput) (usecase.UserRoles, error)
	RequireFreshRole(ctx context.Context, userID string, allowed ...domain.Role) error
}

type Handler struct {
	passwordUsecase PasswordUsecase
	authUsecase     AuthUsecase
	tokenUsecase    TokenUsecase
	otpUsecase      OTPUsecase
	roleUsecase     RoleUsecase
	logger          *slog.Logger
}

func NewHandler(passwordUsecase PasswordUsecase, authUsecase AuthUsecase, tokenUsecase TokenUsecase, otpUsecase OTPUsecase, roleUsecase RoleUsecase, logger *slog.Logger) (*Handler, error) {
	if passwordUsecase == nil {
		return nil, errors.New("password usecase is required")
	}
	if authUsecase == nil {
		return nil, errors.New("auth usecase is required")
	}
	if tokenUsecase == nil {
		return nil, errors.New("token usecase is required")
	}
	if otpUsecase == nil {
		return nil, errors.New("otp usecase is required")
	}
	if roleUsecase == nil {
		return nil, errors.New("role usecase is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		passwordUsecase: passwordUsecase,
		authUsecase:     authUsecase,
		tokenUsecase:    tokenUsecase,
		otpUsecase:      otpUsecase,
		roleUsecase:     roleUsecase,
		logger:          logger,
	}, nil
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/auth/login", h.handleLogin)
	mux.HandleFunc("/api/v1/auth/refresh", h.handleRefreshToken)
	mux.HandleFunc("/api/v1/auth/logout", h.handleLogout)
	mux.HandleFunc("/api/v1/auth/otp/send", h.handleSendOTP)
	mux.HandleFunc("/api/v1/auth/otp/verify", h.handleVerifyOTP)
	mux.HandleFunc("/api/v1/auth/password/forgot", h.handleForgotPassword)
	mux.HandleFunc("/internal/v1/auth/credentials", h.handleCreateCredential)
	mux.HandleFunc("/internal/v1/auth/password/verify", h.handleVerifyPassword)
	mux.HandleFunc("/internal/v1/auth/password/reset", h.handleResetPassword)
	mux.HandleFunc("/internal/v1/auth/tokens/issue", h.handleIssueTokenPair)
	mux.HandleFunc("/internal/v1/auth/tokens/verify", h.handleVerifyAccessToken)
	mux.Handle("/internal/v1/auth/roles", h.AuthRequired(http.HandlerFunc(h.handleGetUserRoles)))
	mux.Handle("/internal/v1/auth/roles/assign", h.AuthRequired(h.RequireRoles(roleMutationRoles()...)(http.HandlerFunc(h.handleAssignRole))))
	mux.Handle("/internal/v1/auth/roles/revoke", h.AuthRequired(h.RequireRoles(roleMutationRoles()...)(http.HandlerFunc(h.handleRevokeRole))))
	mux.HandleFunc("/.well-known/jwks.json", h.handleJWKS)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req loginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	session, err := h.authUsecase.Login(r.Context(), usecase.LoginInput{
		Identifier: req.Identifier,
		Password:   req.Password,
		TraceID:    traceID(r),
		Device:     sessionDeviceInput(r, req.Device),
		Network:    sessionNetworkInput(r),
	})
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, authSessionResponse{
		User: authUserResponse{
			UserID:   session.User.UserID,
			Roles:    session.User.Roles,
			SellerID: session.User.SellerID,
			TenantID: session.User.TenantID,
		},
		Tokens:    tokenResponseFromPair(session.Tokens),
		SessionID: session.SessionID,
	})
}

func (h *Handler) handleCreateCredential(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req createCredentialRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	summary, err := h.passwordUsecase.CreateCredential(r.Context(), usecase.CreateCredentialInput{
		AccountID: req.AccountID,
		Password:  req.Password,
	})
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, createCredentialResponse{
		AccountID:         summary.AccountID,
		PasswordAlgo:      summary.PasswordAlgo,
		PasswordChangedAt: summary.PasswordChangedAt,
	})
}

func (h *Handler) handleVerifyPassword(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req verifyPasswordRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	result, err := h.passwordUsecase.VerifyPassword(r.Context(), usecase.VerifyPasswordInput{
		Identifier: req.Identifier,
		Password:   req.Password,
	})
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, verifyPasswordResponse{
		AccountID:        result.AccountID,
		AccountStatus:    string(result.AccountStatus),
		EmailVerified:    result.EmailVerified,
		PhoneVerified:    result.PhoneVerified,
		PasswordValid:    true,
		PasswordRehashed: result.PasswordRehashed,
	})
}

func (h *Handler) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req resetPasswordRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if _, err := h.passwordUsecase.ResetPassword(r.Context(), usecase.ResetPasswordInput{
		AccountID:   req.AccountID,
		NewPassword: req.NewPassword,
	}); err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, successResponse{Success: true})
}

func (h *Handler) handleSendOTP(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req sendOTPRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	out, err := h.otpUsecase.CreateOTPChallenge(r.Context(), usecase.CreateOTPChallengeInput{
		AccountID: req.AccountID,
		Target:    req.Target,
		Channel:   domain.OTPChannel(req.Channel),
		Purpose:   domain.OTPPurpose(req.Purpose),
	})
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, otpChallengeResponse{
		ChallengeID:        out.ChallengeID,
		ExpiresAt:          out.ExpiresAt,
		ExpiresIn:          out.ExpiresInSeconds,
		ResendAfterSeconds: out.ResendAfterSeconds,
	})
}

func (h *Handler) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req forgotPasswordRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	target := req.Target
	if strings.TrimSpace(target) == "" {
		target = req.Identifier
	}
	channel := req.Channel
	if strings.TrimSpace(channel) == "" {
		channel = inferOTPChannel(target)
	}
	out, err := h.otpUsecase.CreateOTPChallenge(r.Context(), usecase.CreateOTPChallengeInput{
		Target:  target,
		Channel: domain.OTPChannel(channel),
		Purpose: domain.OTPPurposePasswordReset,
	})
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, otpChallengeResponse{
		ChallengeID:        out.ChallengeID,
		ExpiresAt:          out.ExpiresAt,
		ExpiresIn:          out.ExpiresInSeconds,
		ResendAfterSeconds: out.ResendAfterSeconds,
	})
}

func (h *Handler) handleVerifyOTP(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req verifyOTPRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	err := h.otpUsecase.VerifyOTP(r.Context(), usecase.VerifyOTPInput{
		ChallengeID:        req.ChallengeID,
		OTP:                req.OTP,
		VerificationSource: verificationSource(r),
	})
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, successResponse{Success: true})
}

func (h *Handler) handleIssueTokenPair(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req issueTokenPairRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	pair, err := h.tokenUsecase.IssueTokenPair(r.Context(), usecase.IssueTokenPairInput{
		AccountID: req.AccountID,
		UserID:    req.UserID,
		SessionID: req.SessionID,
		Roles:     req.Roles,
		SellerID:  req.SellerID,
		TenantID:  req.TenantID,
	})
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, authSessionResponse{
		User: authUserResponse{
			UserID:   pair.UserID,
			Roles:    pair.Roles,
			SellerID: pair.SellerID,
			TenantID: pair.TenantID,
		},
		Tokens:    tokenResponseFromPair(pair),
		SessionID: pair.SessionID,
	})
}

func (h *Handler) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req refreshTokenRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	pair, err := h.tokenUsecase.RefreshToken(r.Context(), usecase.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
		TraceID:      traceID(r),
		Device:       sessionDeviceInput(r, deviceRequest{}),
		Network:      sessionNetworkInput(r),
	})
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, tokenResponseFromPair(pair))
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req logoutRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if err := h.tokenUsecase.Logout(r.Context(), usecase.LogoutInput{
		RefreshToken: req.RefreshToken,
		AllDevices:   req.AllDevices,
		Reason:       "user_requested",
		TraceID:      traceID(r),
		Device:       sessionDeviceInput(r, deviceRequest{}),
		Network:      sessionNetworkInput(r),
	}); err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, successResponse{Success: true})
}

func (h *Handler) handleVerifyAccessToken(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req verifyAccessTokenRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	claims, err := h.tokenUsecase.VerifyAccessToken(r.Context(), req.AccessToken)
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, tokenClaimsResponse{
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
		Roles:     claims.Roles,
		SellerID:  claims.SellerID,
		TenantID:  claims.TenantID,
	})
}

func (h *Handler) handleAssignRole(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req assignRoleRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	assignment, err := h.roleUsecase.AssignRole(r.Context(), usecase.AssignRoleInput{
		TargetUserID: req.UserID,
		Role:         req.Role,
		ScopeType:    req.ScopeType,
		ScopeID:      req.ScopeID,
		Reason:       req.Reason,
		RequestID:    requestID(r),
	})
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, roleMutationResponse{
		Success: true,
		Role:    roleAssignmentResponseFromDomain(assignment),
	})
}

func (h *Handler) handleRevokeRole(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req revokeRoleRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if err := h.roleUsecase.RevokeRole(r.Context(), usecase.RevokeRoleInput{
		TargetUserID: req.UserID,
		Role:         req.Role,
		ScopeType:    req.ScopeType,
		ScopeID:      req.ScopeID,
		Reason:       req.Reason,
		RequestID:    requestID(r),
	}); err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, successResponse{Success: true})
}

func (h *Handler) handleGetUserRoles(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	query := r.URL.Query()
	userID := strings.TrimSpace(query.Get("user_id"))
	if userID == "" {
		userID = strings.TrimSpace(query.Get("target_user_id"))
	}

	includeRevoked := strings.EqualFold(strings.TrimSpace(query.Get("include_revoked")), "true")
	result, err := h.roleUsecase.GetUserRoles(r.Context(), usecase.GetUserRolesInput{
		UserID:         userID,
		IncludeRevoked: includeRevoked,
	})
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	roles := make([]roleAssignmentResponse, 0, len(result.Roles))
	for _, assignment := range result.Roles {
		roles = append(roles, roleAssignmentResponseFromDomain(assignment))
	}
	writeJSON(w, http.StatusOK, userRolesResponse{
		UserID: result.UserID,
		Roles:  roles,
	})
}

func (h *Handler) handleJWKS(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(w, http.StatusOK, h.tokenUsecase.JWKS(r.Context()))
}

func (h *Handler) writeUsecaseError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidOTPRequest):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, domain.ErrInvalidOTP),
		errors.Is(err, domain.ErrOTPExpired),
		errors.Is(err, domain.ErrOTPAlreadyUsed):
		writeAPIError(w, http.StatusUnauthorized, "INVALID_OTP", "Invalid or expired OTP")
	case errors.Is(err, domain.ErrOTPAttemptsExceeded), errors.Is(err, domain.ErrOTPRateLimited):
		var rateLimitErr *domain.OTPRateLimitError
		if errors.As(err, &rateLimitErr) {
			if retryAfter := rateLimitErr.RetryAfterSeconds(); retryAfter > 0 {
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			}
		}
		writeAPIError(w, http.StatusTooManyRequests, "OTP_RATE_LIMITED", "OTP request rate limit exceeded")
	case errors.Is(err, domain.ErrOTPRateLimitUnavailable):
		writeAPIError(w, http.StatusServiceUnavailable, "OTP_RATE_LIMIT_UNAVAILABLE", "OTP verification is temporarily unavailable")
	case errors.Is(err, domain.ErrOTPDeliveryFailed):
		writeAPIError(w, http.StatusServiceUnavailable, "OTP_DELIVERY_UNAVAILABLE", "OTP delivery is temporarily unavailable")
	case errors.Is(err, domain.ErrInvalidCredentials):
		writeAPIError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid credentials")
	case errors.Is(err, domain.ErrCredentialNotFound):
		writeAPIError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid credentials")
	case errors.Is(err, domain.ErrDuplicateCredential):
		writeAPIError(w, http.StatusConflict, "CREDENTIAL_EXISTS", "Credential already exists")
	case errors.Is(err, domain.ErrInvalidRefreshToken), errors.Is(err, domain.ErrRefreshTokenReuse):
		writeAPIError(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "Invalid refresh token")
	case errors.Is(err, domain.ErrInvalidAccessToken):
		writeAPIError(w, http.StatusUnauthorized, "INVALID_ACCESS_TOKEN", "Invalid access token")
	case errors.Is(err, domain.ErrAccountInactive):
		writeAPIError(w, http.StatusForbidden, "ACCOUNT_INACTIVE", "Account cannot authenticate")
	case errors.Is(err, domain.ErrUnauthenticated):
		writeAPIError(w, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication required")
	case errors.Is(err, domain.ErrForbidden):
		writeAPIError(w, http.StatusForbidden, "PERMISSION_DENIED", "Permission denied")
	case errors.Is(err, domain.ErrUnknownRole),
		errors.Is(err, domain.ErrInvalidRoleAssignment),
		errors.Is(err, domain.ErrInvalidRoleScope),
		errors.Is(err, domain.ErrAuditReasonRequired):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, domain.ErrRoleAlreadyAssigned):
		writeAPIError(w, http.StatusConflict, "ROLE_ALREADY_ASSIGNED", "Role is already assigned")
	case errors.Is(err, domain.ErrRoleAssignmentNotFound), errors.Is(err, domain.ErrAccountNotFound):
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Resource not found")
	case errors.Is(err, usecase.ErrInvalidAccountID), errors.Is(err, usecase.ErrInvalidIdentifier):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, domain.ErrTokenSubjectMissing), errors.Is(err, domain.ErrRoleRequired):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, password.ErrPasswordBlank),
		errors.Is(err, password.ErrPasswordTooShort),
		errors.Is(err, password.ErrPasswordTooLong),
		errors.Is(err, password.ErrPasswordTooLongForBcrypt):
		writeAPIError(w, http.StatusBadRequest, "PASSWORD_POLICY_FAILED", err.Error())
	default:
		h.logger.ErrorContext(r.Context(), "auth.http_error",
			slog.String("error", err.Error()),
		)
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

func verificationSource(r *http.Request) string {
	if forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwardedFor != "" {
		first, _, _ := strings.Cut(forwardedFor, ",")
		if ip := strings.TrimSpace(first); ip != "" {
			return ip
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func inferOTPChannel(target string) string {
	target = strings.TrimSpace(target)
	if strings.Contains(target, "@") {
		return string(domain.OTPChannelEmail)
	}
	if strings.HasPrefix(target, "+") {
		return string(domain.OTPChannelPhone)
	}
	return ""
}

func tokenResponseFromPair(pair usecase.TokenPair) tokenResponse {
	return tokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
	}
}

func roleAssignmentResponseFromDomain(assignment domain.RoleAssignment) roleAssignmentResponse {
	return roleAssignmentResponse{
		Role:       assignment.Role.String(),
		ScopeType:  assignment.ScopeType,
		ScopeID:    assignment.ScopeID,
		AssignedAt: assignment.AssignedAt,
		RevokedAt:  assignment.RevokedAt,
	}
}

func roleMutationRoles() []domain.Role {
	return []domain.Role{
		domain.RoleSeller,
		domain.RoleSellerManager,
		domain.RoleAdmin,
		domain.RoleOperationsAdmin,
		domain.RoleFinanceAdmin,
		domain.RoleCatalogAdmin,
		domain.RoleReadonlyAdmin,
		domain.RoleSuperadmin,
	}
}

func sessionDeviceInput(r *http.Request, device deviceRequest) usecase.SessionDeviceInput {
	anonymousID := strings.TrimSpace(device.AnonymousID)
	if anonymousID == "" {
		anonymousID = strings.TrimSpace(r.Header.Get("X-Anonymous-ID"))
	}

	fingerprint := strings.TrimSpace(device.Fingerprint)
	if fingerprint == "" {
		fingerprint = strings.TrimSpace(r.Header.Get("X-Device-Fingerprint"))
	}

	fingerprintHash := strings.TrimSpace(device.DeviceFingerprintHash)
	if fingerprintHash == "" {
		fingerprintHash = strings.TrimSpace(r.Header.Get("X-Device-Fingerprint-Hash"))
	}

	channel := strings.TrimSpace(device.Channel)
	if channel == "" {
		channel = strings.TrimSpace(r.Header.Get("X-Client-Channel"))
	}

	locale := strings.TrimSpace(device.Locale)
	if locale == "" {
		locale = firstLanguage(r.Header.Get("Accept-Language"))
	}

	return usecase.SessionDeviceInput{
		AnonymousID:           anonymousID,
		Fingerprint:           fingerprint,
		DeviceFingerprintHash: fingerprintHash,
		UserAgent:             r.UserAgent(),
		Channel:               channel,
		Locale:                locale,
	}
}

func sessionNetworkInput(r *http.Request) usecase.SessionNetworkInput {
	return usecase.SessionNetworkInput{
		IPAddress: verificationSource(r),
		IPHash:    strings.TrimSpace(r.Header.Get("X-IP-Hash")),
	}
}

func traceID(r *http.Request) string {
	if traceID := strings.TrimSpace(r.Header.Get("Traceparent")); traceID != "" {
		return traceID
	}
	if traceID := strings.TrimSpace(r.Header.Get("X-Trace-ID")); traceID != "" {
		return traceID
	}
	return requestID(r)
}

func firstLanguage(header string) string {
	first, _, _ := strings.Cut(strings.TrimSpace(header), ",")
	language, _, _ := strings.Cut(first, ";")
	return strings.TrimSpace(language)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	return false
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, errorResponse{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}
