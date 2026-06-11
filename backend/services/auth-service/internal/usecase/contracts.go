package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/password"
)

var (
	ErrInvalidAccountID  = errors.New("account_id is required")
	ErrInvalidIdentifier = errors.New("identifier is required")
)

type CredentialRepository interface {
	CreateCredential(ctx context.Context, credential domain.Credential) error
	FindByIdentifier(ctx context.Context, identifier string) (domain.AccountCredential, error)
	UpdatePassword(ctx context.Context, accountID string, passwordHash string, algorithm string, changedAt time.Time) error
	RehashPassword(ctx context.Context, accountID string, passwordHash string, algorithm string) error
	IncrementFailedAttempts(ctx context.Context, accountID string, maxAttempts int, lockUntil time.Time) error
	ResetFailedAttempts(ctx context.Context, accountID string) error
}

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now().UTC()
}

type PasswordSecurityConfig struct {
	MaxFailedAttempts int
	LockoutDuration   time.Duration
}

type PasswordUsecase struct {
	repo              CredentialRepository
	hasher            password.Hasher
	logger            *slog.Logger
	clock             Clock
	maxFailedAttempts int
	lockoutDuration   time.Duration
}

func NewPasswordUsecase(repo CredentialRepository, hasher password.Hasher, cfg PasswordSecurityConfig, logger *slog.Logger) (*PasswordUsecase, error) {
	if repo == nil {
		return nil, errors.New("credential repository is required")
	}
	if hasher == nil {
		return nil, errors.New("password hasher is required")
	}
	if cfg.MaxFailedAttempts <= 0 {
		return nil, errors.New("max failed attempts must be greater than zero")
	}
	if cfg.LockoutDuration <= 0 {
		return nil, errors.New("lockout duration must be greater than zero")
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &PasswordUsecase{
		repo:              repo,
		hasher:            hasher,
		logger:            logger,
		clock:             realClock{},
		maxFailedAttempts: cfg.MaxFailedAttempts,
		lockoutDuration:   cfg.LockoutDuration,
	}, nil
}

func (u *PasswordUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func normalizeAccountID(accountID string) string {
	return strings.TrimSpace(accountID)
}

func normalizeIdentifier(identifier string) string {
	return strings.ToLower(strings.TrimSpace(identifier))
}
