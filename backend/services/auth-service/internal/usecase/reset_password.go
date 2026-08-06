package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

type ResetPasswordInput struct {
	AccountID   string
	NewPassword string
}

type ResolvePasswordResetAccountInput struct {
	Identifier string
}

type PasswordResetAccount struct {
	AccountID  string
	Identifier string
}

type ResetPasswordByIdentifierInput struct {
	Identifier  string
	NewPassword string
}

func (u *PasswordUsecase) ResolvePasswordResetAccount(ctx context.Context, input ResolvePasswordResetAccountInput) (PasswordResetAccount, error) {
	identifier := normalizeIdentifier(input.Identifier)
	if identifier == "" {
		return PasswordResetAccount{}, ErrInvalidIdentifier
	}

	accountCredential, err := u.repo.FindByIdentifier(ctx, identifier)
	if err != nil {
		if errors.Is(err, domain.ErrCredentialNotFound) {
			return PasswordResetAccount{}, domain.ErrInvalidCredentials
		}
		return PasswordResetAccount{}, fmt.Errorf("find credential by identifier: %w", err)
	}
	if !accountCredential.AccountStatus.CanAuthenticate() {
		return PasswordResetAccount{}, domain.ErrInvalidCredentials
	}
	if accountCredential.AccountID == "" {
		return PasswordResetAccount{}, domain.ErrInvalidCredentials
	}

	return PasswordResetAccount{
		AccountID:  accountCredential.AccountID,
		Identifier: identifier,
	}, nil
}

func (u *PasswordUsecase) ResetPassword(ctx context.Context, input ResetPasswordInput) (CredentialSummary, error) {
	accountID := normalizeAccountID(input.AccountID)
	if accountID == "" {
		return CredentialSummary{}, ErrInvalidAccountID
	}

	hashResult, err := u.hasher.Hash(ctx, input.NewPassword)
	if err != nil {
		return CredentialSummary{}, err
	}

	changedAt := u.clock.Now()
	if err := u.repo.UpdatePassword(ctx, accountID, hashResult.EncodedHash, hashResult.Algorithm, changedAt); err != nil {
		return CredentialSummary{}, fmt.Errorf("update password hash: %w", err)
	}

	u.logger.InfoContext(ctx, "auth.password.reset_completed",
		slog.String("account_id", accountID),
		slog.String("algorithm", hashResult.Algorithm),
	)

	return CredentialSummary{
		AccountID:         accountID,
		PasswordAlgo:      hashResult.Algorithm,
		PasswordChangedAt: changedAt,
	}, nil
}

func (u *PasswordUsecase) ResetPasswordByIdentifier(ctx context.Context, input ResetPasswordByIdentifierInput) (CredentialSummary, error) {
	identifier := normalizeIdentifier(input.Identifier)
	if identifier == "" {
		return CredentialSummary{}, ErrInvalidIdentifier
	}

	accountCredential, err := u.repo.FindByIdentifier(ctx, identifier)
	if err != nil {
		return CredentialSummary{}, fmt.Errorf("find credential by identifier: %w", err)
	}

	return u.ResetPassword(ctx, ResetPasswordInput{
		AccountID:   accountCredential.AccountID,
		NewPassword: input.NewPassword,
	})
}
