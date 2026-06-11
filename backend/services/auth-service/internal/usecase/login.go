package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

type VerifyPasswordInput struct {
	Identifier string
	Password   string
}

type VerifyPasswordOutput struct {
	AccountID        string
	AccountStatus    domain.AccountStatus
	EmailVerified    bool
	PhoneVerified    bool
	PasswordRehashed bool
}

func (u *PasswordUsecase) VerifyPassword(ctx context.Context, input VerifyPasswordInput) (VerifyPasswordOutput, error) {
	identifier := normalizeIdentifier(input.Identifier)
	if identifier == "" {
		return VerifyPasswordOutput{}, domain.ErrInvalidCredentials
	}

	accountCredential, err := u.repo.FindByIdentifier(ctx, identifier)
	if err != nil {
		if errors.Is(err, domain.ErrCredentialNotFound) {
			u.logger.InfoContext(ctx, "auth.password.verify_failed",
				slog.String("reason", "invalid_credentials"),
			)
			return VerifyPasswordOutput{}, domain.ErrInvalidCredentials
		}
		return VerifyPasswordOutput{}, fmt.Errorf("load credential: %w", err)
	}

	if !accountCredential.AccountStatus.CanAuthenticate() {
		u.logger.InfoContext(ctx, "auth.password.verify_failed",
			slog.String("account_id", accountCredential.AccountID),
			slog.String("reason", "account_inactive"),
		)
		return VerifyPasswordOutput{}, domain.ErrInvalidCredentials
	}

	now := u.clock.Now()
	if accountCredential.Credential.IsLocked(now) {
		u.logger.InfoContext(ctx, "auth.password.verify_failed",
			slog.String("account_id", accountCredential.AccountID),
			slog.String("reason", "account_locked"),
		)
		return VerifyPasswordOutput{}, domain.ErrInvalidCredentials
	}

	result, err := u.hasher.Verify(
		ctx,
		input.Password,
		accountCredential.Credential.PasswordHash,
		accountCredential.Credential.PasswordAlgo,
	)
	if err != nil {
		u.logger.WarnContext(ctx, "auth.password.verify_error",
			slog.String("account_id", accountCredential.AccountID),
			slog.String("algorithm", accountCredential.Credential.PasswordAlgo),
			slog.String("reason", "hash_verification_error"),
		)
		return VerifyPasswordOutput{}, domain.ErrInvalidCredentials
	}

	if !result.Valid {
		if err := u.recordFailedAttempt(ctx, accountCredential.AccountID); err != nil {
			return VerifyPasswordOutput{}, err
		}
		return VerifyPasswordOutput{}, domain.ErrInvalidCredentials
	}

	if err := u.repo.ResetFailedAttempts(ctx, accountCredential.AccountID); err != nil {
		return VerifyPasswordOutput{}, fmt.Errorf("reset failed attempts: %w", err)
	}

	passwordRehashed := false
	if result.NeedsRehash {
		passwordRehashed = u.rehashPassword(ctx, accountCredential.AccountID, input.Password)
	}

	u.logger.InfoContext(ctx, "auth.password.verify_succeeded",
		slog.String("account_id", accountCredential.AccountID),
		slog.Bool("password_rehashed", passwordRehashed),
	)

	return VerifyPasswordOutput{
		AccountID:        accountCredential.AccountID,
		AccountStatus:    accountCredential.AccountStatus,
		EmailVerified:    accountCredential.EmailVerified,
		PhoneVerified:    accountCredential.PhoneVerified,
		PasswordRehashed: passwordRehashed,
	}, nil
}

func (u *PasswordUsecase) recordFailedAttempt(ctx context.Context, accountID string) error {
	lockUntil := u.clock.Now().Add(u.lockoutDuration)
	if err := u.repo.IncrementFailedAttempts(ctx, accountID, u.maxFailedAttempts, lockUntil); err != nil {
		return fmt.Errorf("increment failed attempts: %w", err)
	}

	u.logger.InfoContext(ctx, "auth.password.verify_failed",
		slog.String("account_id", accountID),
		slog.String("reason", "invalid_credentials"),
	)
	return nil
}

func (u *PasswordUsecase) rehashPassword(ctx context.Context, accountID string, plainPassword string) bool {
	hashResult, err := u.hasher.Hash(ctx, plainPassword)
	if err != nil {
		u.logger.WarnContext(ctx, "auth.password.rehash_failed",
			slog.String("account_id", accountID),
			slog.String("reason", "hash_error"),
		)
		return false
	}
	if err := u.repo.RehashPassword(ctx, accountID, hashResult.EncodedHash, hashResult.Algorithm); err != nil {
		u.logger.WarnContext(ctx, "auth.password.rehash_failed",
			slog.String("account_id", accountID),
			slog.String("reason", "repository_error"),
		)
		return false
	}

	u.logger.InfoContext(ctx, "auth.password.rehashed",
		slog.String("account_id", accountID),
		slog.String("algorithm", hashResult.Algorithm),
	)
	return true
}
