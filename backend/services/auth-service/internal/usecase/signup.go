package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

type CreateCredentialInput struct {
	AccountID string
	Password  string
}

type CredentialSummary struct {
	AccountID         string
	PasswordAlgo      string
	PasswordChangedAt time.Time
}

func (u *PasswordUsecase) CreateCredential(ctx context.Context, input CreateCredentialInput) (CredentialSummary, error) {
	accountID := normalizeAccountID(input.AccountID)
	if accountID == "" {
		return CredentialSummary{}, ErrInvalidAccountID
	}

	hashResult, err := u.hasher.Hash(ctx, input.Password)
	if err != nil {
		return CredentialSummary{}, err
	}

	now := u.clock.Now()
	credential := domain.Credential{
		AccountID:         accountID,
		PasswordHash:      hashResult.EncodedHash,
		PasswordAlgo:      hashResult.Algorithm,
		PasswordChangedAt: &now,
		FailedAttempts:    0,
	}

	if err := u.repo.CreateCredential(ctx, credential); err != nil {
		return CredentialSummary{}, fmt.Errorf("create credential: %w", err)
	}

	u.logger.InfoContext(ctx, "auth.password.credential_created",
		slog.String("account_id", accountID),
		slog.String("algorithm", hashResult.Algorithm),
	)

	return CredentialSummary{
		AccountID:         accountID,
		PasswordAlgo:      hashResult.Algorithm,
		PasswordChangedAt: now,
	}, nil
}
