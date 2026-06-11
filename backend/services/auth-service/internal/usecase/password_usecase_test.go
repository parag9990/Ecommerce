package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/password"
)

type fakeCredentialRepository struct {
	accountCredential domain.AccountCredential
	findErr           error

	createdCredential domain.Credential
	updatedAccountID  string
	updatedHash       string
	updatedAlgorithm  string
	updatedChangedAt  time.Time

	rehashedAccountID string
	rehashedHash      string
	rehashedAlgorithm string

	incrementAccountID string
	incrementMax       int
	incrementLockUntil time.Time
	resetAccountID     string
}

func (r *fakeCredentialRepository) CreateCredential(ctx context.Context, credential domain.Credential) error {
	r.createdCredential = credential
	return nil
}

func (r *fakeCredentialRepository) FindByIdentifier(ctx context.Context, identifier string) (domain.AccountCredential, error) {
	if r.findErr != nil {
		return domain.AccountCredential{}, r.findErr
	}
	return r.accountCredential, nil
}

func (r *fakeCredentialRepository) UpdatePassword(ctx context.Context, accountID string, passwordHash string, algorithm string, changedAt time.Time) error {
	r.updatedAccountID = accountID
	r.updatedHash = passwordHash
	r.updatedAlgorithm = algorithm
	r.updatedChangedAt = changedAt
	return nil
}

func (r *fakeCredentialRepository) RehashPassword(ctx context.Context, accountID string, passwordHash string, algorithm string) error {
	r.rehashedAccountID = accountID
	r.rehashedHash = passwordHash
	r.rehashedAlgorithm = algorithm
	return nil
}

func (r *fakeCredentialRepository) IncrementFailedAttempts(ctx context.Context, accountID string, maxAttempts int, lockUntil time.Time) error {
	r.incrementAccountID = accountID
	r.incrementMax = maxAttempts
	r.incrementLockUntil = lockUntil
	return nil
}

func (r *fakeCredentialRepository) ResetFailedAttempts(ctx context.Context, accountID string) error {
	r.resetAccountID = accountID
	return nil
}

type stubHasher struct {
	hashResult   password.HashResult
	hashErr      error
	verifyResult password.VerificationResult
	verifyErr    error
}

func (h stubHasher) Hash(ctx context.Context, plain string) (password.HashResult, error) {
	if h.hashErr != nil {
		return password.HashResult{}, h.hashErr
	}
	if h.hashResult.EncodedHash != "" {
		return h.hashResult, nil
	}
	return password.HashResult{EncodedHash: "hash:" + plain, Algorithm: password.AlgorithmArgon2id}, nil
}

func (h stubHasher) Verify(ctx context.Context, plain string, encodedHash string, algorithm string) (password.VerificationResult, error) {
	return h.verifyResult, h.verifyErr
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func newTestUsecase(t *testing.T, repo *fakeCredentialRepository, hasher password.Hasher, now time.Time) *PasswordUsecase {
	t.Helper()

	uc, err := NewPasswordUsecase(
		repo,
		hasher,
		PasswordSecurityConfig{
			MaxFailedAttempts: 5,
			LockoutDuration:   15 * time.Minute,
		},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatalf("NewPasswordUsecase() error = %v", err)
	}
	uc.WithClock(fixedClock{now: now})
	return uc
}

func TestCreateCredentialHashesBeforeStoring(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	repo := &fakeCredentialRepository{}
	uc := newTestUsecase(t, repo, stubHasher{}, now)

	summary, err := uc.CreateCredential(context.Background(), CreateCredentialInput{
		AccountID: " auth_123 ",
		Password:  "correct-horse-battery",
	})
	if err != nil {
		t.Fatalf("CreateCredential() error = %v", err)
	}

	if repo.createdCredential.AccountID != "auth_123" {
		t.Fatalf("stored account id = %q", repo.createdCredential.AccountID)
	}
	if repo.createdCredential.PasswordHash == "correct-horse-battery" {
		t.Fatal("plain password was stored")
	}
	if repo.createdCredential.PasswordAlgo != password.AlgorithmArgon2id {
		t.Fatalf("stored algorithm = %q", repo.createdCredential.PasswordAlgo)
	}
	if summary.PasswordChangedAt != now {
		t.Fatalf("summary changed_at = %v, want %v", summary.PasswordChangedAt, now)
	}
}

func TestVerifyPasswordRecordsFailedAttempt(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	repo := &fakeCredentialRepository{
		accountCredential: domain.AccountCredential{
			AccountID:     "auth_123",
			AccountStatus: domain.AccountStatusActive,
			Credential: domain.Credential{
				AccountID:    "auth_123",
				PasswordHash: "stored-hash",
				PasswordAlgo: password.AlgorithmArgon2id,
			},
		},
	}
	uc := newTestUsecase(t, repo, stubHasher{
		verifyResult: password.VerificationResult{Valid: false, Algorithm: password.AlgorithmArgon2id},
	}, now)

	_, err := uc.VerifyPassword(context.Background(), VerifyPasswordInput{
		Identifier: "User@Example.com",
		Password:   "wrong-password",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("VerifyPassword() error = %v, want invalid credentials", err)
	}
	if repo.incrementAccountID != "auth_123" || repo.incrementMax != 5 {
		t.Fatalf("increment call = account %q max %d", repo.incrementAccountID, repo.incrementMax)
	}
	if !repo.incrementLockUntil.Equal(now.Add(15 * time.Minute)) {
		t.Fatalf("lock_until = %v", repo.incrementLockUntil)
	}
	if repo.resetAccountID != "" {
		t.Fatal("failed login must not reset failed attempts")
	}
}

func TestVerifyPasswordResetsAndRehashesOnSuccess(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	repo := &fakeCredentialRepository{
		accountCredential: domain.AccountCredential{
			AccountID:     "auth_123",
			AccountStatus: domain.AccountStatusActive,
			EmailVerified: true,
			Credential: domain.Credential{
				AccountID:    "auth_123",
				PasswordHash: "legacy-hash",
				PasswordAlgo: password.AlgorithmBcrypt,
			},
		},
	}
	uc := newTestUsecase(t, repo, stubHasher{
		hashResult:   password.HashResult{EncodedHash: "new-argon-hash", Algorithm: password.AlgorithmArgon2id},
		verifyResult: password.VerificationResult{Valid: true, NeedsRehash: true, Algorithm: password.AlgorithmBcrypt},
	}, now)

	output, err := uc.VerifyPassword(context.Background(), VerifyPasswordInput{
		Identifier: "user@example.com",
		Password:   "correct-horse-battery",
	})
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if repo.resetAccountID != "auth_123" {
		t.Fatalf("reset account id = %q", repo.resetAccountID)
	}
	if repo.rehashedAccountID != "auth_123" || repo.rehashedHash != "new-argon-hash" {
		t.Fatalf("rehash call = account %q hash %q", repo.rehashedAccountID, repo.rehashedHash)
	}
	if !output.PasswordRehashed || !output.EmailVerified {
		t.Fatalf("output = %+v", output)
	}
}

func TestResetPasswordStoresNewHashAndChangedAt(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	repo := &fakeCredentialRepository{}
	uc := newTestUsecase(t, repo, stubHasher{
		hashResult: password.HashResult{EncodedHash: "new-password-hash", Algorithm: password.AlgorithmArgon2id},
	}, now)

	summary, err := uc.ResetPassword(context.Background(), ResetPasswordInput{
		AccountID:   "auth_123",
		NewPassword: "new-correct-horse-battery",
	})
	if err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}
	if repo.updatedAccountID != "auth_123" || repo.updatedHash != "new-password-hash" {
		t.Fatalf("update call = account %q hash %q", repo.updatedAccountID, repo.updatedHash)
	}
	if repo.updatedAlgorithm != password.AlgorithmArgon2id {
		t.Fatalf("updated algorithm = %q", repo.updatedAlgorithm)
	}
	if !repo.updatedChangedAt.Equal(now) || !summary.PasswordChangedAt.Equal(now) {
		t.Fatalf("changed_at update = %v summary = %v", repo.updatedChangedAt, summary.PasswordChangedAt)
	}
}
