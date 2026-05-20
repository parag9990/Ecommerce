package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
	tokensecurity "github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/token"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/sessionlink"
)

type RefreshTokenRepository interface {
	InsertRefreshToken(ctx context.Context, refreshToken domain.RefreshToken) error
	FindRefreshTokenByHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error)
	RotateRefreshToken(ctx context.Context, oldTokenID string, next domain.RefreshToken, revokedAt time.Time) error
	RevokeSessionTokens(ctx context.Context, sessionID string, revokedAt time.Time) error
	RevokeAccountTokens(ctx context.Context, accountID string, revokedAt time.Time) error
}

type TokenSubjectRepository interface {
	FindTokenSubject(ctx context.Context, accountID string) (domain.TokenSubject, error)
}

type RoleRepository interface {
	ActiveRoles(ctx context.Context, accountID string) ([]string, error)
}

type AccessTokenIssuer interface {
	IssueAccessToken(input tokensecurity.AccessTokenInput) (tokensecurity.AccessTokenResult, error)
}

type AccessTokenVerifier interface {
	VerifyAccessToken(tokenString string) (domain.TokenClaims, error)
}

type TokenUsecaseConfig struct {
	RefreshTokenTTL    time.Duration
	RefreshTokenPepper string
}

type TokenUsecase struct {
	refreshTokens RefreshTokenRepository
	subjects      TokenSubjectRepository
	roles         RoleRepository
	issuer        AccessTokenIssuer
	verifier      AccessTokenVerifier
	sessionLink   SessionLinker
	privacyHasher SessionPrivacyHasher
	jwks          tokensecurity.JWKSet
	logger        *slog.Logger
	clock         Clock
	refreshTTL    time.Duration
	refreshPepper string
}

func NewTokenUsecase(
	refreshTokens RefreshTokenRepository,
	subjects TokenSubjectRepository,
	roles RoleRepository,
	issuer AccessTokenIssuer,
	verifier AccessTokenVerifier,
	sessionLink SessionLinker,
	privacyHasher SessionPrivacyHasher,
	jwks tokensecurity.JWKSet,
	cfg TokenUsecaseConfig,
	logger *slog.Logger,
) (*TokenUsecase, error) {
	if refreshTokens == nil {
		return nil, errors.New("refresh token repository is required")
	}
	if subjects == nil {
		return nil, errors.New("token subject repository is required")
	}
	if roles == nil {
		return nil, errors.New("role repository is required")
	}
	if issuer == nil {
		return nil, errors.New("access token issuer is required")
	}
	if verifier == nil {
		return nil, errors.New("access token verifier is required")
	}
	if sessionLink == nil {
		return nil, errors.New("session linker is required")
	}
	if privacyHasher == nil {
		return nil, errors.New("session privacy hasher is required")
	}
	if cfg.RefreshTokenTTL <= 0 {
		return nil, errors.New("refresh token ttl must be greater than zero")
	}
	if cfg.RefreshTokenPepper == "" {
		return nil, errors.New("refresh token pepper is required")
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &TokenUsecase{
		refreshTokens: refreshTokens,
		subjects:      subjects,
		roles:         roles,
		issuer:        issuer,
		verifier:      verifier,
		sessionLink:   sessionLink,
		privacyHasher: privacyHasher,
		jwks:          jwks,
		logger:        logger,
		clock:         realClock{},
		refreshTTL:    cfg.RefreshTokenTTL,
		refreshPepper: cfg.RefreshTokenPepper,
	}, nil
}

func (u *TokenUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

type IssueTokenPairInput struct {
	AccountID string
	UserID    string
	SessionID string
	Roles     []string
	SellerID  string
	TenantID  string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	SessionID    string
	UserID       string
	Roles        []string
	SellerID     string
	TenantID     string
}

type RefreshTokenInput struct {
	RefreshToken string
	TraceID      string
	Device       SessionDeviceInput
	Network      SessionNetworkInput
}

type LogoutInput struct {
	RefreshToken string
	AllDevices   bool
	Reason       string
	TraceID      string
	Device       SessionDeviceInput
	Network      SessionNetworkInput
}

func (u *TokenUsecase) IssueTokenPair(ctx context.Context, input IssueTokenPairInput) (TokenPair, error) {
	accountID := normalizeAccountID(input.AccountID)
	userID := strings.TrimSpace(input.UserID)
	sessionID := strings.TrimSpace(input.SessionID)
	roles := normalizeRoles(input.Roles)

	if accountID == "" || userID == "" {
		return TokenPair{}, domain.ErrTokenSubjectMissing
	}
	if len(roles) == 0 {
		return TokenPair{}, domain.ErrRoleRequired
	}
	if sessionID == "" {
		generated, err := tokensecurity.NewSessionID()
		if err != nil {
			return TokenPair{}, err
		}
		sessionID = generated
	}

	accessToken, err := u.issuer.IssueAccessToken(tokensecurity.AccessTokenInput{
		UserID:    userID,
		SessionID: sessionID,
		Roles:     roles,
		SellerID:  input.SellerID,
		TenantID:  input.TenantID,
	})
	if err != nil {
		return TokenPair{}, fmt.Errorf("issue access token: %w", err)
	}

	refreshToken, refreshRow, err := u.newRefreshTokenRow(accountID, sessionID, "", u.clock.Now())
	if err != nil {
		return TokenPair{}, err
	}
	if err := u.refreshTokens.InsertRefreshToken(ctx, refreshRow); err != nil {
		return TokenPair{}, fmt.Errorf("insert refresh token: %w", err)
	}

	u.logger.InfoContext(ctx, "auth.token_pair.issued",
		slog.String("account_id", accountID),
		slog.String("user_id", userID),
		slog.String("session_id", sessionID),
		slog.String("refresh_token_id", refreshRow.TokenID),
	)

	return TokenPair{
		AccessToken:  accessToken.Token,
		RefreshToken: refreshToken,
		ExpiresIn:    accessToken.ExpiresIn,
		SessionID:    sessionID,
		UserID:       userID,
		Roles:        roles,
		SellerID:     strings.TrimSpace(input.SellerID),
		TenantID:     strings.TrimSpace(input.TenantID),
	}, nil
}

func (u *TokenUsecase) IssueTokenPairForAccount(ctx context.Context, accountID string) (TokenPair, error) {
	subject, roles, err := u.loadTokenContext(ctx, accountID)
	if err != nil {
		return TokenPair{}, err
	}

	return u.IssueTokenPair(ctx, IssueTokenPairInput{
		AccountID: subject.AccountID,
		UserID:    subject.UserID,
		Roles:     roles,
		SellerID:  subject.SellerID,
		TenantID:  subject.TenantID,
	})
}

func (u *TokenUsecase) RefreshToken(ctx context.Context, input RefreshTokenInput) (TokenPair, error) {
	tokenHash, err := tokensecurity.HashRefreshToken(input.RefreshToken, u.refreshPepper)
	if err != nil {
		return TokenPair{}, domain.ErrInvalidRefreshToken
	}

	oldToken, err := u.refreshTokens.FindRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return TokenPair{}, domain.ErrInvalidRefreshToken
	}

	now := u.clock.Now().UTC()
	if oldToken.RevokedAt != nil {
		_ = u.refreshTokens.RevokeSessionTokens(ctx, oldToken.SessionID, now)
		u.recordRefreshReuseDetected(ctx, oldToken, input)
		u.logger.WarnContext(ctx, "auth.refresh.reuse_detected",
			slog.String("account_id", oldToken.AccountID),
			slog.String("session_id", oldToken.SessionID),
			slog.String("refresh_token_id", oldToken.TokenID),
		)
		return TokenPair{}, domain.ErrRefreshTokenReuse
	}
	if !oldToken.IsActive(now) {
		return TokenPair{}, domain.ErrInvalidRefreshToken
	}

	subject, roles, err := u.loadTokenContext(ctx, oldToken.AccountID)
	if err != nil {
		return TokenPair{}, err
	}

	nextPlain, nextRow, err := u.newRefreshTokenRow(oldToken.AccountID, oldToken.SessionID, oldToken.TokenID, now)
	if err != nil {
		return TokenPair{}, err
	}

	accessToken, err := u.issuer.IssueAccessToken(tokensecurity.AccessTokenInput{
		UserID:    subject.UserID,
		SessionID: oldToken.SessionID,
		Roles:     roles,
		SellerID:  subject.SellerID,
		TenantID:  subject.TenantID,
	})
	if err != nil {
		return TokenPair{}, fmt.Errorf("issue access token: %w", err)
	}

	if err := u.refreshTokens.RotateRefreshToken(ctx, oldToken.TokenID, nextRow, now); err != nil {
		if errors.Is(err, domain.ErrRefreshTokenReuse) {
			_ = u.refreshTokens.RevokeSessionTokens(ctx, oldToken.SessionID, now)
			u.recordRefreshReuseDetected(ctx, oldToken, input)
		}
		return TokenPair{}, err
	}

	u.logger.InfoContext(ctx, "auth.refresh.rotated",
		slog.String("account_id", oldToken.AccountID),
		slog.String("session_id", oldToken.SessionID),
		slog.String("old_refresh_token_id", oldToken.TokenID),
		slog.String("new_refresh_token_id", nextRow.TokenID),
	)

	return TokenPair{
		AccessToken:  accessToken.Token,
		RefreshToken: nextPlain,
		ExpiresIn:    accessToken.ExpiresIn,
		SessionID:    oldToken.SessionID,
		UserID:       subject.UserID,
		Roles:        roles,
		SellerID:     subject.SellerID,
		TenantID:     subject.TenantID,
	}, nil
}

func (u *TokenUsecase) Logout(ctx context.Context, input LogoutInput) error {
	tokenHash, err := tokensecurity.HashRefreshToken(input.RefreshToken, u.refreshPepper)
	if err != nil {
		return domain.ErrInvalidRefreshToken
	}

	refreshToken, err := u.refreshTokens.FindRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return domain.ErrInvalidRefreshToken
	}

	now := u.clock.Now().UTC()
	if !refreshToken.IsActive(now) {
		return domain.ErrInvalidRefreshToken
	}

	if input.AllDevices {
		if err := u.refreshTokens.RevokeAccountTokens(ctx, refreshToken.AccountID, now); err != nil {
			return err
		}
	} else if err := u.refreshTokens.RevokeSessionTokens(ctx, refreshToken.SessionID, now); err != nil {
		return err
	}

	subject, err := u.subjects.FindTokenSubject(ctx, refreshToken.AccountID)
	if err != nil {
		subject = domain.TokenSubject{AccountID: refreshToken.AccountID}
	}
	u.recordLogoutSucceeded(ctx, refreshToken, subject, input, now)

	u.logger.InfoContext(ctx, "auth.logout.revoked",
		slog.String("account_id", refreshToken.AccountID),
		slog.String("session_id", refreshToken.SessionID),
		slog.Bool("all_devices", input.AllDevices),
	)
	return nil
}

func (u *TokenUsecase) VerifyAccessToken(ctx context.Context, accessToken string) (domain.TokenClaims, error) {
	claims, err := u.verifier.VerifyAccessToken(accessToken)
	if err != nil {
		u.logger.InfoContext(ctx, "auth.access_token.verify_failed")
		return domain.TokenClaims{}, domain.ErrInvalidAccessToken
	}
	return claims, nil
}

func (u *TokenUsecase) JWKS(ctx context.Context) tokensecurity.JWKSet {
	u.logger.DebugContext(ctx, "auth.jwks.requested")
	return u.jwks
}

func (u *TokenUsecase) loadTokenContext(ctx context.Context, accountID string) (domain.TokenSubject, []string, error) {
	accountID = normalizeAccountID(accountID)
	if accountID == "" {
		return domain.TokenSubject{}, nil, domain.ErrTokenSubjectMissing
	}

	subject, err := u.subjects.FindTokenSubject(ctx, accountID)
	if err != nil {
		return domain.TokenSubject{}, nil, fmt.Errorf("load token subject: %w", err)
	}
	if subject.AccountID == "" || subject.UserID == "" {
		return domain.TokenSubject{}, nil, domain.ErrTokenSubjectMissing
	}
	if !subject.Status.CanAuthenticate() {
		return domain.TokenSubject{}, nil, domain.ErrAccountInactive
	}

	roles, err := u.roles.ActiveRoles(ctx, accountID)
	if err != nil {
		return domain.TokenSubject{}, nil, fmt.Errorf("load active roles: %w", err)
	}
	roles = normalizeRoles(roles)
	if len(roles) == 0 {
		return domain.TokenSubject{}, nil, domain.ErrRoleRequired
	}

	return subject, roles, nil
}

func (u *TokenUsecase) newRefreshTokenRow(accountID string, sessionID string, parentTokenID string, now time.Time) (string, domain.RefreshToken, error) {
	plainToken, err := tokensecurity.NewRefreshToken()
	if err != nil {
		return "", domain.RefreshToken{}, err
	}
	tokenHash, err := tokensecurity.HashRefreshToken(plainToken, u.refreshPepper)
	if err != nil {
		return "", domain.RefreshToken{}, err
	}
	tokenID, err := tokensecurity.NewRefreshTokenID()
	if err != nil {
		return "", domain.RefreshToken{}, err
	}

	now = now.UTC()
	return plainToken, domain.RefreshToken{
		TokenID:       tokenID,
		AccountID:     accountID,
		SessionID:     sessionID,
		TokenHash:     tokenHash,
		ParentTokenID: parentTokenID,
		ExpiresAt:     now.Add(u.refreshTTL),
		CreatedAt:     now,
	}, nil
}

func (u *TokenUsecase) recordLogoutSucceeded(ctx context.Context, refreshToken domain.RefreshToken, subject domain.TokenSubject, input LogoutInput, occurredAt time.Time) {
	eventID, err := sessionlink.NewEventID()
	if err != nil {
		u.logger.WarnContext(ctx, "auth.session_link.event_id_failed",
			slog.String("event_type", sessionlink.EventTypeAuthLogoutSucceeded),
			slog.String("account_id", refreshToken.AccountID),
			slog.String("session_id", refreshToken.SessionID),
			slog.String("error", err.Error()),
		)
		return
	}

	event := sessionlink.LogoutSucceededEvent{
		EventID:    eventID,
		TraceID:    input.TraceID,
		OccurredAt: occurredAt,
		AccountID:  refreshToken.AccountID,
		UserID:     subject.UserID,
		SessionID:  refreshToken.SessionID,
		AllDevices: input.AllDevices,
		Reason:     input.Reason,
	}
	if err := u.sessionLink.RecordLogoutSucceeded(ctx, event); err != nil {
		u.logger.WarnContext(ctx, "auth.session_link.enqueue_failed",
			slog.String("event_type", sessionlink.EventTypeAuthLogoutSucceeded),
			slog.String("account_id", refreshToken.AccountID),
			slog.String("session_id", refreshToken.SessionID),
			slog.String("error", err.Error()),
		)
	}
}

func (u *TokenUsecase) recordRefreshReuseDetected(ctx context.Context, refreshToken domain.RefreshToken, input RefreshTokenInput) {
	eventID, err := sessionlink.NewEventID()
	if err != nil {
		u.logger.WarnContext(ctx, "auth.session_link.event_id_failed",
			slog.String("event_type", sessionlink.EventTypeRefreshTokenReuseDetected),
			slog.String("account_id", refreshToken.AccountID),
			slog.String("session_id", refreshToken.SessionID),
			slog.String("error", err.Error()),
		)
		return
	}

	subject, err := u.subjects.FindTokenSubject(ctx, refreshToken.AccountID)
	if err != nil {
		subject = domain.TokenSubject{AccountID: refreshToken.AccountID}
	}

	event := sessionlink.RefreshReuseDetectedEvent{
		EventID:               eventID,
		TraceID:               input.TraceID,
		OccurredAt:            u.clock.Now(),
		AccountID:             refreshToken.AccountID,
		UserID:                subject.UserID,
		SessionID:             refreshToken.SessionID,
		TokenID:               refreshToken.TokenID,
		IPHash:                u.ipHash(input.Network),
		DeviceFingerprintHash: u.deviceFingerprintHash(input.Device),
		ActionTaken:           sessionlink.ActionRevokedSessionFamily,
	}
	if err := u.sessionLink.RecordRefreshReuseDetected(ctx, event); err != nil {
		u.logger.WarnContext(ctx, "auth.session_link.enqueue_failed",
			slog.String("event_type", sessionlink.EventTypeRefreshTokenReuseDetected),
			slog.String("account_id", refreshToken.AccountID),
			slog.String("session_id", refreshToken.SessionID),
			slog.String("refresh_token_id", refreshToken.TokenID),
			slog.String("error", err.Error()),
		)
	}
}

func (u *TokenUsecase) deviceFingerprintHash(device SessionDeviceInput) string {
	if hash := strings.TrimSpace(device.DeviceFingerprintHash); hash != "" {
		return hash
	}
	return u.privacyHasher.HashDeviceFingerprint(device.Fingerprint)
}

func (u *TokenUsecase) ipHash(network SessionNetworkInput) string {
	if hash := strings.TrimSpace(network.IPHash); hash != "" {
		return hash
	}
	return u.privacyHasher.HashIP(network.IPAddress)
}

func normalizeRoles(roles []string) []string {
	seen := make(map[string]struct{}, len(roles))
	result := make([]string, 0, len(roles))
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		if _, exists := seen[role]; exists {
			continue
		}
		seen[role] = struct{}{}
		result = append(result, role)
	}
	return result
}
