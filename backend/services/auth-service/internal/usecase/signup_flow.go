package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/sessionlink"
)

var ErrInvalidSignupRequest = errors.New("invalid signup request")

type SignupAccountRepository interface {
	CreateAccount(ctx context.Context, account domain.AuthAccount) error
	LinkUser(ctx context.Context, accountID string, userID string) error
}

type SignupCredentialCreator interface {
	CreateCredential(ctx context.Context, input CreateCredentialInput) (CredentialSummary, error)
}

type InitialRoleAssigner interface {
	AssignRole(ctx context.Context, assignment domain.RoleAssignment) error
}

type SignupTokenIssuer interface {
	IssueTokenPair(ctx context.Context, input IssueTokenPairInput) (TokenPair, error)
}

type SignupUserCreator interface {
	CreateUser(ctx context.Context, authAccountID string, email string, phone string, fullName string, requestID string) (string, error)
}

type SignupUsecase struct {
	accounts      SignupAccountRepository
	credentials   SignupCredentialCreator
	roles         InitialRoleAssigner
	tokens        SignupTokenIssuer
	users         SignupUserCreator
	sessionLink   SessionLinker
	privacyHasher SessionPrivacyHasher
	logger        *slog.Logger
	clock         Clock
	accountIDs    func(context.Context) (string, error)
}

type SignupInput struct {
	Email    string
	Phone    string
	FullName string
	Password string
	Role     string
	TraceID  string
	Device   SessionDeviceInput
	Network  SessionNetworkInput
}

func NewSignupUsecase(
	accounts SignupAccountRepository,
	credentials SignupCredentialCreator,
	roles InitialRoleAssigner,
	tokens SignupTokenIssuer,
	users SignupUserCreator,
	sessionLink SessionLinker,
	privacyHasher SessionPrivacyHasher,
	logger *slog.Logger,
) (*SignupUsecase, error) {
	if accounts == nil {
		return nil, errors.New("account repository is required")
	}
	if credentials == nil {
		return nil, errors.New("credential creator is required")
	}
	if roles == nil {
		return nil, errors.New("role assigner is required")
	}
	if tokens == nil {
		return nil, errors.New("token issuer is required")
	}
	if users == nil {
		return nil, errors.New("user creator is required")
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
	return &SignupUsecase{
		accounts:      accounts,
		credentials:   credentials,
		roles:         roles,
		tokens:        tokens,
		users:         users,
		sessionLink:   sessionLink,
		privacyHasher: privacyHasher,
		logger:        logger,
		clock:         realClock{},
		accountIDs:    newAccountID,
	}, nil
}

func (u *SignupUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func (u *SignupUsecase) WithAccountIDFactory(factory func(context.Context) (string, error)) {
	if factory != nil {
		u.accountIDs = factory
	}
}

func (u *SignupUsecase) Signup(ctx context.Context, input SignupInput) (AuthSession, error) {
	email, phone, fullName, role, err := normalizeSignupInput(input)
	if err != nil {
		return AuthSession{}, err
	}

	accountID, err := u.accountIDs(ctx)
	if err != nil {
		return AuthSession{}, fmt.Errorf("generate account id: %w", err)
	}

	now := u.clock.Now()
	account := domain.AuthAccount{
		AccountID: accountID,
		Email:     &email,
		Phone:     optionalStringPtr(phone),
		Status:    domain.AccountStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := u.accounts.CreateAccount(ctx, account); err != nil {
		return AuthSession{}, err
	}
	if _, err := u.credentials.CreateCredential(ctx, CreateCredentialInput{
		AccountID: accountID,
		Password:  input.Password,
	}); err != nil {
		return AuthSession{}, err
	}

	userID, err := u.users.CreateUser(ctx, accountID, email, phone, fullName, strings.TrimSpace(input.TraceID))
	if err != nil {
		return AuthSession{}, err
	}
	if err := u.accounts.LinkUser(ctx, accountID, userID); err != nil {
		return AuthSession{}, err
	}

	roles := []string{role.String()}
	if err := u.roles.AssignRole(ctx, domain.RoleAssignment{
		AccountID:  accountID,
		Role:       role,
		AssignedBy: "auth-service",
		Reason:     "buyer signup",
		AssignedAt: now,
	}); err != nil {
		return AuthSession{}, err
	}

	tokenPair, err := u.tokens.IssueTokenPair(ctx, IssueTokenPairInput{
		AccountID: accountID,
		UserID:    userID,
		Roles:     roles,
	})
	if err != nil {
		return AuthSession{}, err
	}
	u.recordSignupSessionLink(ctx, accountID, tokenPair, input)

	u.logger.InfoContext(ctx, "auth.signup.completed",
		slog.String("account_id", accountID),
		slog.String("user_id", userID),
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

func normalizeSignupInput(input SignupInput) (string, string, string, domain.Role, error) {
	email := normalizeIdentifier(input.Email)
	phone := strings.TrimSpace(input.Phone)
	fullName := strings.TrimSpace(input.FullName)
	if email == "" || strings.TrimSpace(input.Password) == "" || fullName == "" {
		return "", "", "", "", ErrInvalidSignupRequest
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || !strings.EqualFold(parsed.Address, email) {
		return "", "", "", "", ErrInvalidSignupRequest
	}

	roleValue := strings.ToLower(strings.TrimSpace(input.Role))
	if roleValue == "" {
		roleValue = domain.RoleBuyer.String()
	}
	role := domain.Role(roleValue)
	if role != domain.RoleBuyer {
		return "", "", "", "", domain.ErrInvalidRoleAssignment
	}
	return email, phone, fullName, role, nil
}

func (u *SignupUsecase) recordSignupSessionLink(ctx context.Context, accountID string, tokenPair TokenPair, input SignupInput) {
	eventID, err := sessionlink.NewEventID()
	if err != nil {
		u.logger.WarnContext(ctx, "auth.session_link.event_id_failed",
			slog.String("account_id", accountID),
			slog.String("session_id", tokenPair.SessionID),
			slog.String("error", err.Error()),
		)
		return
	}

	event := sessionlink.SignupSucceededEvent{
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
			Timezone:              input.Device.Timezone,
		},
		Network: sessionlink.NetworkInfo{
			IPHash: u.ipHash(input.Network),
		},
		Auth: sessionlink.AuthInfo{
			Method:  sessionlink.AuthMethodSignup,
			MFAUsed: false,
		},
	}

	if err := u.sessionLink.RecordSignupSucceeded(ctx, event); err != nil {
		u.logger.WarnContext(ctx, "auth.session_link.enqueue_failed",
			slog.String("event_type", sessionlink.EventTypeAuthSignupSucceeded),
			slog.String("account_id", accountID),
			slog.String("session_id", tokenPair.SessionID),
			slog.String("error", err.Error()),
		)
	}
}

func (u *SignupUsecase) deviceFingerprintHash(device SessionDeviceInput) string {
	if hash := strings.TrimSpace(device.DeviceFingerprintHash); hash != "" {
		return hash
	}
	return u.privacyHasher.HashDeviceFingerprint(device.Fingerprint)
}

func (u *SignupUsecase) ipHash(network SessionNetworkInput) string {
	if hash := strings.TrimSpace(network.IPHash); hash != "" {
		return hash
	}
	return u.privacyHasher.HashIP(network.IPAddress)
}

func newAccountID(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return "auth_" + hex.EncodeToString(raw), nil
}

func optionalStringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
