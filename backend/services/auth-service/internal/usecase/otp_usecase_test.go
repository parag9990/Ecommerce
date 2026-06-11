package usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/clients"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
	otpsec "github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/otp"
)

type fakeOTPChallengeRepository struct {
	mu        sync.Mutex
	challenge domain.OTPChallenge
	createErr error
	mutateErr error

	expiredTarget  string
	expiredPurpose domain.OTPPurpose
	expiredAt      time.Time
	expiredID      string
	created        domain.OTPChallenge
	incremented    bool
	verifiedAt     *time.Time
	emailAccountID string
	phoneAccountID string
}

func (r *fakeOTPChallengeRepository) CreateOTPChallenge(ctx context.Context, challenge domain.OTPChallenge) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.created = challenge
	r.challenge = challenge
	return r.createErr
}

func (r *fakeOTPChallengeRepository) ExpireOlderActiveOTPChallenges(ctx context.Context, target string, purpose domain.OTPPurpose, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.expiredTarget = target
	r.expiredPurpose = purpose
	r.expiredAt = now
	return nil
}

func (r *fakeOTPChallengeRepository) ExpireOTPChallenge(ctx context.Context, challengeID string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.expiredID = challengeID
	r.expiredAt = now
	return nil
}

func (r *fakeOTPChallengeRepository) MutateLockedOTPChallenge(ctx context.Context, challengeID string, mutate func(domain.OTPChallenge) (domain.OTPChallengeMutation, error)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.mutateErr != nil {
		return r.mutateErr
	}
	if r.challenge.ChallengeID == "" || r.challenge.ChallengeID != challengeID {
		return domain.ErrOTPChallengeNotFound
	}

	mutation, err := mutate(r.challenge)
	if mutation.IncrementAttempts {
		r.incremented = true
		r.challenge.Attempts++
	}
	if mutation.MarkVerifiedAt != nil {
		verifiedAt := mutation.MarkVerifiedAt.UTC()
		r.verifiedAt = &verifiedAt
		r.challenge.VerifiedAt = &verifiedAt
	}
	if mutation.MarkEmailVerifiedAccount != "" {
		r.emailAccountID = mutation.MarkEmailVerifiedAccount
	}
	if mutation.MarkPhoneVerifiedAccount != "" {
		r.phoneAccountID = mutation.MarkPhoneVerifiedAccount
	}
	return err
}

type fakeOTPRateRepository struct {
	mu          sync.Mutex
	resendAfter time.Duration
	reserveErr  error
	verifyErr   error

	reservedTarget  string
	reservedPurpose domain.OTPPurpose
	verifyChallenge string
	verifySource    string
}

func (r *fakeOTPRateRepository) ReserveSend(ctx context.Context, purpose domain.OTPPurpose, target string, now time.Time) (time.Duration, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.reservedPurpose = purpose
	r.reservedTarget = target
	if r.reserveErr != nil {
		return 0, r.reserveErr
	}
	if r.resendAfter > 0 {
		return r.resendAfter, nil
	}
	return time.Minute, nil
}

func (r *fakeOTPRateRepository) MarkVerifyAttempt(ctx context.Context, challengeID string, verificationSource string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.verifyChallenge = challengeID
	r.verifySource = verificationSource
	return r.verifyErr
}

type fakeOTPNotifier struct {
	req clients.SendOTPRequest
	err error
}

func (n *fakeOTPNotifier) SendOTP(ctx context.Context, req clients.SendOTPRequest) error {
	n.req = req
	return n.err
}

type fixedOTPGenerator struct {
	code string
	err  error
}

func (g fixedOTPGenerator) Generate() (string, error) {
	if g.err != nil {
		return "", g.err
	}
	return g.code, nil
}

func newTestOTPUsecase(t *testing.T, repo *fakeOTPChallengeRepository, rates *fakeOTPRateRepository, notifier *fakeOTPNotifier, now time.Time) *OTPUsecase {
	t.Helper()

	hasher, err := otpsec.NewHasher("test-otp-pepper")
	if err != nil {
		t.Fatalf("NewHasher() error = %v", err)
	}
	policy := otpsec.DefaultPolicy()
	uc, err := NewOTPUsecase(
		repo,
		rates,
		notifier,
		fixedOTPGenerator{code: "123456"},
		hasher,
		OTPUsecaseConfig{Policy: policy},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatalf("NewOTPUsecase() error = %v", err)
	}
	uc.WithClock(fixedClock{now: now})
	uc.WithIDFactory(func() (string, error) { return "otp_chal_test123", nil })
	return uc
}

func TestCreateOTPChallengeHashesAndNotifies(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	repo := &fakeOTPChallengeRepository{}
	rates := &fakeOTPRateRepository{resendAfter: time.Minute}
	notifier := &fakeOTPNotifier{}
	uc := newTestOTPUsecase(t, repo, rates, notifier, now)

	accountID := " acct_123 "
	out, err := uc.CreateOTPChallenge(context.Background(), CreateOTPChallengeInput{
		AccountID: &accountID,
		Target:    " User@Example.COM ",
		Channel:   domain.OTPChannelEmail,
		Purpose:   domain.OTPPurposeSignup,
	})
	if err != nil {
		t.Fatalf("CreateOTPChallenge() error = %v", err)
	}

	if out.ChallengeID != "otp_chal_test123" || out.ResendAfterSeconds != 60 {
		t.Fatalf("output = %+v", out)
	}
	if repo.created.Target != "user@example.com" || rates.reservedTarget != "user@example.com" {
		t.Fatalf("target normalization failed: repo=%q rate=%q", repo.created.Target, rates.reservedTarget)
	}
	if repo.created.OTPHash == "123456" || repo.created.OTPHash == "" {
		t.Fatalf("otp hash was not stored safely: %q", repo.created.OTPHash)
	}
	if repo.created.AccountID == nil || *repo.created.AccountID != "acct_123" {
		t.Fatalf("stored account_id = %v", repo.created.AccountID)
	}
	if notifier.req.OTP != "123456" || notifier.req.ChallengeID != "otp_chal_test123" {
		t.Fatalf("notification request = %+v", notifier.req)
	}
	if repo.expiredTarget != "user@example.com" || repo.expiredPurpose != domain.OTPPurposeSignup {
		t.Fatalf("older challenge expiry = target %q purpose %q", repo.expiredTarget, repo.expiredPurpose)
	}
}

func TestVerifyOTPIncrementsAttemptsOnWrongCode(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	hasher, _ := otpsec.NewHasher("test-otp-pepper")
	hash, _ := hasher.Hash("otp_chal_test123", "123456")
	repo := &fakeOTPChallengeRepository{
		challenge: domain.OTPChallenge{
			ChallengeID: "otp_chal_test123",
			OTPHash:     hash,
			MaxAttempts: 5,
			ExpiresAt:   now.Add(time.Minute),
		},
	}
	rates := &fakeOTPRateRepository{}
	uc := newTestOTPUsecase(t, repo, rates, &fakeOTPNotifier{}, now)

	err := uc.VerifyOTP(context.Background(), VerifyOTPInput{
		ChallengeID:        "otp_chal_test123",
		OTP:                "000000",
		VerificationSource: "203.0.113.10",
	})
	if !errors.Is(err, domain.ErrInvalidOTP) {
		t.Fatalf("VerifyOTP() error = %v, want invalid otp", err)
	}
	if !repo.incremented || repo.challenge.Attempts != 1 {
		t.Fatalf("attempts not incremented: incremented=%v attempts=%d", repo.incremented, repo.challenge.Attempts)
	}
	if rates.verifySource != "203.0.113.10" {
		t.Fatalf("verify source = %q", rates.verifySource)
	}
}

func TestVerifyOTPMarksVerifiedAndAccountEmailVerified(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	hasher, _ := otpsec.NewHasher("test-otp-pepper")
	hash, _ := hasher.Hash("otp_chal_test123", "123456")
	accountID := "acct_123"
	repo := &fakeOTPChallengeRepository{
		challenge: domain.OTPChallenge{
			ChallengeID: "otp_chal_test123",
			AccountID:   &accountID,
			Channel:     domain.OTPChannelEmail,
			Purpose:     domain.OTPPurposeEmailVerify,
			OTPHash:     hash,
			MaxAttempts: 5,
			ExpiresAt:   now.Add(time.Minute),
		},
	}
	uc := newTestOTPUsecase(t, repo, &fakeOTPRateRepository{}, &fakeOTPNotifier{}, now)

	if err := uc.VerifyOTP(context.Background(), VerifyOTPInput{
		ChallengeID:        "otp_chal_test123",
		OTP:                "123456",
		VerificationSource: "203.0.113.10",
	}); err != nil {
		t.Fatalf("VerifyOTP() error = %v", err)
	}
	if repo.verifiedAt == nil || !repo.verifiedAt.Equal(now) {
		t.Fatalf("verified_at = %v", repo.verifiedAt)
	}
	if repo.emailAccountID != "acct_123" || repo.phoneAccountID != "" {
		t.Fatalf("account verification email=%q phone=%q", repo.emailAccountID, repo.phoneAccountID)
	}
}

func TestVerifyOTPRejectsReplay(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	hasher, _ := otpsec.NewHasher("test-otp-pepper")
	hash, _ := hasher.Hash("otp_chal_test123", "123456")
	verifiedAt := now.Add(-time.Minute)
	repo := &fakeOTPChallengeRepository{
		challenge: domain.OTPChallenge{
			ChallengeID: "otp_chal_test123",
			OTPHash:     hash,
			MaxAttempts: 5,
			VerifiedAt:  &verifiedAt,
			ExpiresAt:   now.Add(time.Minute),
		},
	}
	uc := newTestOTPUsecase(t, repo, &fakeOTPRateRepository{}, &fakeOTPNotifier{}, now)

	err := uc.VerifyOTP(context.Background(), VerifyOTPInput{
		ChallengeID:        "otp_chal_test123",
		OTP:                "123456",
		VerificationSource: "203.0.113.10",
	})
	if !errors.Is(err, domain.ErrOTPAlreadyUsed) {
		t.Fatalf("VerifyOTP() error = %v, want already used", err)
	}
	if repo.incremented {
		t.Fatal("replay must not increment attempts")
	}
}

func TestVerifyOTPRejectsReplayAfterSuccessfulVerification(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	hasher, _ := otpsec.NewHasher("test-otp-pepper")
	hash, _ := hasher.Hash("otp_chal_test123", "123456")
	repo := &fakeOTPChallengeRepository{
		challenge: domain.OTPChallenge{
			ChallengeID: "otp_chal_test123",
			OTPHash:     hash,
			MaxAttempts: 5,
			ExpiresAt:   now.Add(time.Minute),
		},
	}
	uc := newTestOTPUsecase(t, repo, &fakeOTPRateRepository{}, &fakeOTPNotifier{}, now)

	if err := uc.VerifyOTP(context.Background(), VerifyOTPInput{
		ChallengeID:        "otp_chal_test123",
		OTP:                "123456",
		VerificationSource: "203.0.113.10",
	}); err != nil {
		t.Fatalf("first VerifyOTP() error = %v", err)
	}

	err := uc.VerifyOTP(context.Background(), VerifyOTPInput{
		ChallengeID:        "otp_chal_test123",
		OTP:                "123456",
		VerificationSource: "203.0.113.10",
	})
	if !errors.Is(err, domain.ErrOTPAlreadyUsed) {
		t.Fatalf("replayed VerifyOTP() error = %v, want already used", err)
	}
	if repo.challenge.Attempts != 0 {
		t.Fatalf("replayed otp changed attempts = %d", repo.challenge.Attempts)
	}
}

func TestVerifyOTPRejectsExpiredChallenge(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	hasher, _ := otpsec.NewHasher("test-otp-pepper")
	hash, _ := hasher.Hash("otp_chal_test123", "123456")
	repo := &fakeOTPChallengeRepository{
		challenge: domain.OTPChallenge{
			ChallengeID: "otp_chal_test123",
			OTPHash:     hash,
			MaxAttempts: 5,
			ExpiresAt:   now.Add(-time.Second),
		},
	}
	uc := newTestOTPUsecase(t, repo, &fakeOTPRateRepository{}, &fakeOTPNotifier{}, now)

	err := uc.VerifyOTP(context.Background(), VerifyOTPInput{
		ChallengeID:        "otp_chal_test123",
		OTP:                "123456",
		VerificationSource: "203.0.113.10",
	})
	if !errors.Is(err, domain.ErrOTPExpired) {
		t.Fatalf("VerifyOTP() error = %v, want expired", err)
	}
	if repo.incremented || repo.verifiedAt != nil {
		t.Fatalf("expired challenge mutated: incremented=%v verified_at=%v", repo.incremented, repo.verifiedAt)
	}
}

func TestVerifyOTPMaxAttemptsReachedRejectsCorrectCode(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	hasher, _ := otpsec.NewHasher("test-otp-pepper")
	hash, _ := hasher.Hash("otp_chal_test123", "123456")
	repo := &fakeOTPChallengeRepository{
		challenge: domain.OTPChallenge{
			ChallengeID: "otp_chal_test123",
			OTPHash:     hash,
			Attempts:    5,
			MaxAttempts: 5,
			ExpiresAt:   now.Add(time.Minute),
		},
	}
	uc := newTestOTPUsecase(t, repo, &fakeOTPRateRepository{}, &fakeOTPNotifier{}, now)

	err := uc.VerifyOTP(context.Background(), VerifyOTPInput{
		ChallengeID:        "otp_chal_test123",
		OTP:                "123456",
		VerificationSource: "203.0.113.10",
	})
	if !errors.Is(err, domain.ErrOTPAttemptsExceeded) {
		t.Fatalf("VerifyOTP() error = %v, want attempts exceeded", err)
	}
	if repo.verifiedAt != nil {
		t.Fatalf("exhausted challenge was verified at %v", repo.verifiedAt)
	}
}

func TestVerifyOTPConcurrentReplayOnlyOneSucceeds(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	hasher, _ := otpsec.NewHasher("test-otp-pepper")
	hash, _ := hasher.Hash("otp_chal_test123", "123456")
	repo := &fakeOTPChallengeRepository{
		challenge: domain.OTPChallenge{
			ChallengeID: "otp_chal_test123",
			OTPHash:     hash,
			MaxAttempts: 5,
			ExpiresAt:   now.Add(time.Minute),
		},
	}
	uc := newTestOTPUsecase(t, repo, &fakeOTPRateRepository{}, &fakeOTPNotifier{}, now)

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- uc.VerifyOTP(context.Background(), VerifyOTPInput{
				ChallengeID:        "otp_chal_test123",
				OTP:                "123456",
				VerificationSource: "203.0.113.10",
			})
		}()
	}
	wg.Wait()
	close(errs)

	successes := 0
	replays := 0
	for err := range errs {
		if err == nil {
			successes++
			continue
		}
		if errors.Is(err, domain.ErrOTPAlreadyUsed) {
			replays++
			continue
		}
		t.Fatalf("VerifyOTP() unexpected concurrent error = %v", err)
	}
	if successes != 1 || replays != 1 {
		t.Fatalf("concurrent otp results: successes=%d replays=%d, want 1/1", successes, replays)
	}
}

func TestVerifyOTPFailureLogRedactsSubmittedCode(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	hasher, _ := otpsec.NewHasher("test-otp-pepper")
	hash, _ := hasher.Hash("otp_chal_test123", "123456")
	repo := &fakeOTPChallengeRepository{
		challenge: domain.OTPChallenge{
			ChallengeID: "otp_chal_test123",
			OTPHash:     hash,
			MaxAttempts: 5,
			ExpiresAt:   now.Add(time.Minute),
		},
	}
	logs := &bytes.Buffer{}
	policy := otpsec.DefaultPolicy()
	uc, err := NewOTPUsecase(
		repo,
		&fakeOTPRateRepository{},
		&fakeOTPNotifier{},
		fixedOTPGenerator{code: "123456"},
		hasher,
		OTPUsecaseConfig{Policy: policy},
		slog.New(slog.NewTextHandler(logs, nil)),
	)
	if err != nil {
		t.Fatalf("NewOTPUsecase() error = %v", err)
	}
	uc.WithClock(fixedClock{now: now})

	err = uc.VerifyOTP(context.Background(), VerifyOTPInput{
		ChallengeID:        "otp_chal_test123",
		OTP:                "000000",
		VerificationSource: "203.0.113.10",
	})
	if !errors.Is(err, domain.ErrInvalidOTP) {
		t.Fatalf("VerifyOTP() error = %v, want invalid otp", err)
	}

	output := logs.String()
	if strings.Contains(output, "000000") || strings.Contains(output, "123456") || strings.Contains(output, hash) {
		t.Fatalf("otp secret leaked into logs: %s", output)
	}
	if !strings.Contains(output, "auth.otp.verify_failed") {
		t.Fatalf("expected failure event in logs, got %s", output)
	}
}

func TestCreatePasswordResetOTPStoresHashOnly(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	repo := &fakeOTPChallengeRepository{}
	rates := &fakeOTPRateRepository{}
	notifier := &fakeOTPNotifier{}
	uc := newTestOTPUsecase(t, repo, rates, notifier, now)

	_, err := uc.CreateOTPChallenge(context.Background(), CreateOTPChallengeInput{
		Target:  "buyer@example.com",
		Channel: domain.OTPChannelEmail,
		Purpose: domain.OTPPurposePasswordReset,
	})
	if err != nil {
		t.Fatalf("CreateOTPChallenge() error = %v", err)
	}
	if repo.created.Purpose != domain.OTPPurposePasswordReset {
		t.Fatalf("otp purpose = %q, want password_reset", repo.created.Purpose)
	}
	if repo.created.OTPHash == "" || repo.created.OTPHash == notifier.req.OTP {
		t.Fatalf("password reset otp stored unsafely: hash=%q plain=%q", repo.created.OTPHash, notifier.req.OTP)
	}
}

func TestCreateOTPChallengeRateLimitedDoesNotCreateOrDeliverCode(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	repo := &fakeOTPChallengeRepository{}
	rates := &fakeOTPRateRepository{
		reserveErr: &domain.OTPRateLimitError{
			Reason:     "resend_cooldown",
			RetryAfter: time.Minute,
		},
	}
	notifier := &fakeOTPNotifier{}
	uc := newTestOTPUsecase(t, repo, rates, notifier, now)

	_, err := uc.CreateOTPChallenge(context.Background(), CreateOTPChallengeInput{
		Target:  "buyer@example.com",
		Channel: domain.OTPChannelEmail,
		Purpose: domain.OTPPurposeLogin,
	})
	if !errors.Is(err, domain.ErrOTPRateLimited) {
		t.Fatalf("CreateOTPChallenge() error = %v, want rate limited", err)
	}
	if repo.created.ChallengeID != "" {
		t.Fatalf("rate-limited otp challenge was created: %+v", repo.created)
	}
	if notifier.req.OTP != "" {
		t.Fatalf("rate-limited otp was delivered: %+v", notifier.req)
	}
}
