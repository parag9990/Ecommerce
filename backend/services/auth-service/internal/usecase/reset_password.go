package usecase

import (
	"context"
	"fmt"
	"log/slog"
)

type ResetPasswordInput struct {
	AccountID   string
	NewPassword string
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
