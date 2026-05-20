package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/sessionlink"
)

type SessionLinker interface {
	RecordLoginSucceeded(ctx context.Context, event sessionlink.LoginSucceededEvent) error
	RecordSignupSucceeded(ctx context.Context, event sessionlink.SignupSucceededEvent) error
	RecordLogoutSucceeded(ctx context.Context, event sessionlink.LogoutSucceededEvent) error
	RecordRefreshReuseDetected(ctx context.Context, event sessionlink.RefreshReuseDetectedEvent) error
}

type SessionPrivacyHasher interface {
	HashIP(value string) string
	HashDeviceFingerprint(value string) string
}

type CredentialVerifier interface {
	VerifyPassword(ctx context.Context, input VerifyPasswordInput) (VerifyPasswordOutput, error)
}

type AccountTokenIssuer interface {
	IssueTokenPairForAccount(ctx context.Context, accountID string) (TokenPair, error)
}

type AuthUsecase struct {
	credentials   CredentialVerifier
	tokens        AccountTokenIssuer
	sessionLink   SessionLinker
	privacyHasher SessionPrivacyHasher
	logger        *slog.Logger
	clock         Clock
}

func NewAuthUsecase(credentials CredentialVerifier, tokens AccountTokenIssuer, sessionLink SessionLinker, privacyHasher SessionPrivacyHasher, logger *slog.Logger) (*AuthUsecase, error) {
	if credentials == nil {
		return nil, errors.New("credential verifier is required")
	}
	if tokens == nil {
		return nil, errors.New("token issuer is required")
	}
	if sessionLink == nil {
		return nil, errors.New("session linker is required")
	}
	if privacyHasher == nil {
		return nil, errors.New("session privacy hasher is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &AuthUsecase{
		credentials:   credentials,
		tokens:        tokens,
		sessionLink:   sessionLink,
		privacyHasher: privacyHasher,
		logger:        logger,
		clock:         realClock{},
	}, nil
}

func (u *AuthUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

type LoginInput struct {
	Identifier string
	Password   string
	TraceID    string
	Device     SessionDeviceInput
	Network    SessionNetworkInput
}

type SessionDeviceInput struct {
	AnonymousID           string
	Fingerprint           string
	DeviceFingerprintHash string
	UserAgent             string
	Channel               string
	Locale                string
}

type SessionNetworkInput struct {
	IPAddress string
	IPHash    string
}

type AuthenticatedUser struct {
	UserID   string
	Roles    []string
	SellerID string
	TenantID string
}

type AuthSession struct {
	User      AuthenticatedUser
	Tokens    TokenPair
	SessionID string
}

func (u *AuthUsecase) Login(ctx context.Context, input LoginInput) (AuthSession, error) {
	verified, err := u.credentials.VerifyPassword(ctx, VerifyPasswordInput{
		Identifier: input.Identifier,
		Password:   input.Password,
	})
	if err != nil {
		return AuthSession{}, err
	}

	tokenPair, err := u.tokens.IssueTokenPairForAccount(ctx, verified.AccountID)
	if err != nil {
		return AuthSession{}, err
	}
	u.recordLoginSessionLink(ctx, verified.AccountID, tokenPair, input)

	u.logger.InfoContext(ctx, "auth.login.token_issued",
		slog.String("account_id", verified.AccountID),
		slog.String("user_id", tokenPair.UserID),
		slog.String("session_id", tokenPair.SessionID),
	)

	return AuthSession{
		User: AuthenticatedUser{
			UserID:   tokenPair.UserID,
			Roles:    tokenPair.Roles,
			SellerID: tokenPair.SellerID,
			TenantID: tokenPair.TenantID,
		},
		Tokens:    tokenPair,
		SessionID: tokenPair.SessionID,
	}, nil
}

func (u *AuthUsecase) recordLoginSessionLink(ctx context.Context, accountID string, tokenPair TokenPair, input LoginInput) {
	eventID, err := sessionlink.NewEventID()
	if err != nil {
		u.logger.WarnContext(ctx, "auth.session_link.event_id_failed",
			slog.String("account_id", accountID),
			slog.String("session_id", tokenPair.SessionID),
			slog.String("error", err.Error()),
		)
		return
	}

	event := sessionlink.LoginSucceededEvent{
		EventID:     eventID,
		TraceID:     input.TraceID,
		OccurredAt:  u.clock.Now(),
		AccountID:   accountID,
		UserID:      tokenPair.UserID,
		SessionID:   tokenPair.SessionID,
		AnonymousID: input.Device.AnonymousID,
		Roles:       append([]string(nil), tokenPair.Roles...),
		SellerID:    tokenPair.SellerID,
		TenantID:    tokenPair.TenantID,
		Device: sessionlink.DeviceInfo{
			DeviceFingerprintHash: u.deviceFingerprintHash(input.Device),
			UserAgent:             input.Device.UserAgent,
			Channel:               input.Device.Channel,
			Locale:                input.Device.Locale,
		},
		Network: sessionlink.NetworkInfo{
			IPHash: u.ipHash(input.Network),
		},
		Auth: sessionlink.AuthInfo{
			Method:  sessionlink.AuthMethodPassword,
			MFAUsed: false,
		},
	}

	if err := u.sessionLink.RecordLoginSucceeded(ctx, event); err != nil {
		u.logger.WarnContext(ctx, "auth.session_link.enqueue_failed",
			slog.String("event_type", sessionlink.EventTypeAuthLoginSucceeded),
			slog.String("account_id", accountID),
			slog.String("session_id", tokenPair.SessionID),
			slog.String("error", err.Error()),
		)
	}
}

func (u *AuthUsecase) deviceFingerprintHash(device SessionDeviceInput) string {
	if hash := strings.TrimSpace(device.DeviceFingerprintHash); hash != "" {
		return hash
	}
	return u.privacyHasher.HashDeviceFingerprint(device.Fingerprint)
}

func (u *AuthUsecase) ipHash(network SessionNetworkInput) string {
	if hash := strings.TrimSpace(network.IPHash); hash != "" {
		return hash
	}
	return u.privacyHasher.HashIP(network.IPAddress)
}
