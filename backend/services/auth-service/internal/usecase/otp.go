package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/clients"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
	otpsec "github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/otp"
)

type OTPChallengeRepository interface {
	CreateOTPChallenge(ctx context.Context, challenge domain.OTPChallenge) error
	ExpireOlderActiveOTPChallenges(ctx context.Context, target string, purpose domain.OTPPurpose, now time.Time) error
	ExpireOTPChallenge(ctx context.Context, challengeID string, now time.Time) error
	MutateLockedOTPChallenge(ctx context.Context, challengeID string, mutate func(domain.OTPChallenge) (domain.OTPChallengeMutation, error)) error
}

type OTPRateRepository interface {
	ReserveSend(ctx context.Context, purpose domain.OTPPurpose, target string, now time.Time) (time.Duration, error)
	MarkVerifyAttempt(ctx context.Context, challengeID string, verificationSource string, now time.Time) error
}

type OTPCodeGenerator interface {
	Generate() (string, error)
}

type OTPCodeHasher interface {
	Hash(challengeID string, code string) (string, error)
	Compare(challengeID string, code string, expectedHash string) (bool, error)
}

type OTPUsecaseConfig struct {
	Policy otpsec.Policy
}

type OTPUsecase struct {
	challenges OTPChallengeRepository
	rates      OTPRateRepository
	notifier   clients.NotificationClient
	generator  OTPCodeGenerator
	hasher     OTPCodeHasher
	logger     *slog.Logger
	clock      Clock
	policy     otpsec.Policy
	idFactory  func() (string, error)
}

func NewOTPUsecase(
	challenges OTPChallengeRepository,
	rates OTPRateRepository,
	notifier clients.NotificationClient,
	generator OTPCodeGenerator,
	hasher OTPCodeHasher,
	cfg OTPUsecaseConfig,
	logger *slog.Logger,
) (*OTPUsecase, error) {
	if challenges == nil {
		return nil, errors.New("otp challenge repository is required")
	}
	if rates == nil {
		return nil, errors.New("otp rate repository is required")
	}
	if notifier == nil {
		return nil, errors.New("notification client is required")
	}
	if generator == nil {
		return nil, errors.New("otp generator is required")
	}
	if hasher == nil {
		return nil, errors.New("otp hasher is required")
	}
	if err := cfg.Policy.Validate(); err != nil {
		return nil, fmt.Errorf("invalid otp policy: %w", err)
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &OTPUsecase{
		challenges: challenges,
		rates:      rates,
		notifier:   notifier,
		generator:  generator,
		hasher:     hasher,
		logger:     logger,
		clock:      realClock{},
		policy:     cfg.Policy,
		idFactory:  otpsec.NewChallengeID,
	}, nil
}

func (u *OTPUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func (u *OTPUsecase) WithIDFactory(idFactory func() (string, error)) {
	if idFactory != nil {
		u.idFactory = idFactory
	}
}

type CreateOTPChallengeInput struct {
	AccountID *string
	Target    string
	Channel   domain.OTPChannel
	Purpose   domain.OTPPurpose
}

type CreateOTPChallengeOutput struct {
	ChallengeID        string
	ExpiresAt          time.Time
	ExpiresInSeconds   int
	ResendAfterSeconds int
}

func (u *OTPUsecase) CreateOTPChallenge(ctx context.Context, input CreateOTPChallengeInput) (CreateOTPChallengeOutput, error) {
	channel := domain.OTPChannel(strings.TrimSpace(string(input.Channel)))
	if !channel.Valid() {
		return CreateOTPChallengeOutput{}, fmt.Errorf("%w: unsupported otp channel", domain.ErrInvalidOTPRequest)
	}
	purpose := domain.OTPPurpose(strings.TrimSpace(string(input.Purpose)))
	if !purpose.Valid() {
		return CreateOTPChallengeOutput{}, fmt.Errorf("%w: unsupported otp purpose", domain.ErrInvalidOTPRequest)
	}

	target, err := otpsec.NormalizeTarget(channel, input.Target)
	if err != nil {
		return CreateOTPChallengeOutput{}, fmt.Errorf("%w: %v", domain.ErrInvalidOTPRequest, err)
	}

	var accountID *string
	if input.AccountID != nil {
		normalizedAccountID := normalizeAccountID(*input.AccountID)
		if normalizedAccountID == "" {
			return CreateOTPChallengeOutput{}, fmt.Errorf("%w: account_id cannot be blank when provided", domain.ErrInvalidOTPRequest)
		}
		accountID = &normalizedAccountID
	}

	now := u.clock.Now()
	resendAfter, err := u.rates.ReserveSend(ctx, purpose, target, now)
	if err != nil {
		u.logger.InfoContext(ctx, "auth.otp.create_rate_limited",
			slog.String("purpose", string(purpose)),
			slog.String("channel", string(channel)),
		)
		return CreateOTPChallengeOutput{}, err
	}

	code, err := u.generator.Generate()
	if err != nil {
		return CreateOTPChallengeOutput{}, fmt.Errorf("generate otp: %w", err)
	}

	challengeID, err := u.idFactory()
	if err != nil {
		return CreateOTPChallengeOutput{}, fmt.Errorf("generate otp challenge id: %w", err)
	}
	if !otpsec.ValidateChallengeID(challengeID) {
		return CreateOTPChallengeOutput{}, fmt.Errorf("generated invalid otp challenge id")
	}

	codeHash, err := u.hasher.Hash(challengeID, code)
	if err != nil {
		return CreateOTPChallengeOutput{}, fmt.Errorf("hash otp: %w", err)
	}

	expiresAt := now.Add(u.policy.TTL)
	challenge := domain.OTPChallenge{
		ChallengeID: challengeID,
		AccountID:   accountID,
		Target:      target,
		Channel:     channel,
		Purpose:     purpose,
		OTPHash:     codeHash,
		Attempts:    0,
		MaxAttempts: u.policy.MaxAttempts,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
	}

	if err := u.challenges.ExpireOlderActiveOTPChallenges(ctx, target, purpose, now); err != nil {
		return CreateOTPChallengeOutput{}, fmt.Errorf("expire older otp challenges: %w", err)
	}
	if err := u.challenges.CreateOTPChallenge(ctx, challenge); err != nil {
		return CreateOTPChallengeOutput{}, fmt.Errorf("create otp challenge: %w", err)
	}

	if err := u.notifier.SendOTP(ctx, clients.SendOTPRequest{
		Target:           target,
		Channel:          string(channel),
		Purpose:          string(purpose),
		OTP:              code,
		ChallengeID:      challengeID,
		ExpiresInSeconds: int(u.policy.TTL.Seconds()),
	}); err != nil {
		_ = u.challenges.ExpireOTPChallenge(ctx, challengeID, now)
		u.logger.WarnContext(ctx, "auth.otp.delivery_failed",
			slog.String("challenge_id", challengeID),
			slog.String("purpose", string(purpose)),
			slog.String("channel", string(channel)),
		)
		return CreateOTPChallengeOutput{}, fmt.Errorf("%w", domain.ErrOTPDeliveryFailed)
	}

	u.logger.InfoContext(ctx, "auth.otp.challenge_created",
		slog.String("challenge_id", challengeID),
		slog.String("purpose", string(purpose)),
		slog.String("channel", string(channel)),
		slog.Time("expires_at", expiresAt.UTC()),
	)

	return CreateOTPChallengeOutput{
		ChallengeID:        challengeID,
		ExpiresAt:          expiresAt.UTC(),
		ExpiresInSeconds:   int(u.policy.TTL.Seconds()),
		ResendAfterSeconds: int(resendAfter.Seconds()),
	}, nil
}

type VerifyOTPInput struct {
	ChallengeID        string
	OTP                string
	VerificationSource string
}

type VerifyOTPOutput struct {
	ChallengeID string
	AccountID   *string
	Target      string
	Channel     domain.OTPChannel
	Purpose     domain.OTPPurpose
}

func (u *OTPUsecase) VerifyOTP(ctx context.Context, input VerifyOTPInput) error {
	_, err := u.verifyOTP(ctx, input, nil)
	return err
}

func (u *OTPUsecase) VerifyOTPForPurpose(ctx context.Context, input VerifyOTPInput, purpose domain.OTPPurpose) (VerifyOTPOutput, error) {
	expectedPurpose, err := normalizeExpectedOTPPurpose(purpose)
	if err != nil {
		return VerifyOTPOutput{}, err
	}
	return u.verifyOTP(ctx, input, &expectedPurpose)
}

func (u *OTPUsecase) GetOTPChallengeForPurpose(ctx context.Context, challengeID string, purpose domain.OTPPurpose) (VerifyOTPOutput, error) {
	challengeID = strings.TrimSpace(challengeID)
	if !otpsec.ValidateChallengeID(challengeID) {
		return VerifyOTPOutput{}, fmt.Errorf("%w: invalid otp challenge", domain.ErrInvalidOTPRequest)
	}
	expectedPurpose, err := normalizeExpectedOTPPurpose(purpose)
	if err != nil {
		return VerifyOTPOutput{}, err
	}

	now := u.clock.Now()
	var output VerifyOTPOutput
	err = u.challenges.MutateLockedOTPChallenge(ctx, challengeID, func(challenge domain.OTPChallenge) (domain.OTPChallengeMutation, error) {
		if err := validateOTPChallengeState(challenge, now, &expectedPurpose); err != nil {
			return domain.OTPChallengeMutation{}, err
		}
		output = verifyOTPOutputFromChallenge(challenge)
		return domain.OTPChallengeMutation{}, nil
	})
	if err != nil {
		return VerifyOTPOutput{}, u.handleOTPChallengeError(ctx, challengeID, err)
	}

	return output, nil
}

func (u *OTPUsecase) verifyOTP(ctx context.Context, input VerifyOTPInput, expectedPurpose *domain.OTPPurpose) (VerifyOTPOutput, error) {
	challengeID := strings.TrimSpace(input.ChallengeID)
	code := strings.TrimSpace(input.OTP)
	if !otpsec.ValidateChallengeID(challengeID) || !otpsec.ValidateCode(code, u.policy.Length) {
		return VerifyOTPOutput{}, fmt.Errorf("%w: invalid otp verification request", domain.ErrInvalidOTPRequest)
	}

	now := u.clock.Now()
	if err := u.rates.MarkVerifyAttempt(ctx, challengeID, input.VerificationSource, now); err != nil {
		return VerifyOTPOutput{}, err
	}

	var output VerifyOTPOutput
	err := u.challenges.MutateLockedOTPChallenge(ctx, challengeID, func(challenge domain.OTPChallenge) (domain.OTPChallengeMutation, error) {
		if err := validateOTPChallengeState(challenge, now, expectedPurpose); err != nil {
			return domain.OTPChallengeMutation{}, err
		}

		matched, err := u.hasher.Compare(challenge.ChallengeID, code, challenge.OTPHash)
		if err != nil {
			return domain.OTPChallengeMutation{}, fmt.Errorf("compare otp: %w", err)
		}
		if !matched {
			return domain.OTPChallengeMutation{IncrementAttempts: true}, domain.ErrInvalidOTP
		}

		output = verifyOTPOutputFromChallenge(challenge)
		verifiedAt := now.UTC()
		mutation := domain.OTPChallengeMutation{
			MarkVerifiedAt: &verifiedAt,
		}
		if challenge.AccountID != nil {
			switch {
			case challenge.Purpose == domain.OTPPurposeEmailVerify:
				mutation.MarkEmailVerifiedAccount = *challenge.AccountID
			case challenge.Purpose == domain.OTPPurposePhoneVerify:
				mutation.MarkPhoneVerifiedAccount = *challenge.AccountID
			case challenge.Purpose == domain.OTPPurposeSignup && challenge.Channel == domain.OTPChannelEmail:
				mutation.MarkEmailVerifiedAccount = *challenge.AccountID
			case challenge.Purpose == domain.OTPPurposeSignup && challenge.Channel == domain.OTPChannelPhone:
				mutation.MarkPhoneVerifiedAccount = *challenge.AccountID
			}
		}
		return mutation, nil
	})
	if err != nil {
		return VerifyOTPOutput{}, u.handleOTPChallengeError(ctx, challengeID, err)
	}

	u.logger.InfoContext(ctx, "auth.otp.verified",
		slog.String("challenge_id", challengeID),
	)
	return output, nil
}

func normalizeExpectedOTPPurpose(purpose domain.OTPPurpose) (domain.OTPPurpose, error) {
	expectedPurpose := domain.OTPPurpose(strings.TrimSpace(string(purpose)))
	if !expectedPurpose.Valid() {
		return "", fmt.Errorf("%w: unsupported otp purpose", domain.ErrInvalidOTPRequest)
	}
	return expectedPurpose, nil
}

func validateOTPChallengeState(challenge domain.OTPChallenge, now time.Time, expectedPurpose *domain.OTPPurpose) error {
	if expectedPurpose != nil && challenge.Purpose != *expectedPurpose {
		return domain.ErrInvalidOTP
	}
	if challenge.IsVerified() {
		return domain.ErrOTPAlreadyUsed
	}
	if challenge.IsExpired(now) {
		return domain.ErrOTPExpired
	}
	if challenge.AttemptsExhausted() {
		return domain.ErrOTPAttemptsExceeded
	}
	return nil
}

func (u *OTPUsecase) handleOTPChallengeError(ctx context.Context, challengeID string, err error) error {
	if errors.Is(err, domain.ErrOTPChallengeNotFound) {
		return domain.ErrInvalidOTP
	}
	if errors.Is(err, domain.ErrInvalidOTP) ||
		errors.Is(err, domain.ErrOTPExpired) ||
		errors.Is(err, domain.ErrOTPAlreadyUsed) ||
		errors.Is(err, domain.ErrOTPAttemptsExceeded) {
		u.logger.InfoContext(ctx, "auth.otp.verify_failed",
			slog.String("challenge_id", challengeID),
			slog.String("reason", otpFailureReason(err)),
		)
	}
	return err
}

func verifyOTPOutputFromChallenge(challenge domain.OTPChallenge) VerifyOTPOutput {
	output := VerifyOTPOutput{
		ChallengeID: challenge.ChallengeID,
		Target:      challenge.Target,
		Channel:     challenge.Channel,
		Purpose:     challenge.Purpose,
	}
	if challenge.AccountID != nil {
		accountID := *challenge.AccountID
		output.AccountID = &accountID
	}
	return output
}

func otpFailureReason(err error) string {
	switch {
	case errors.Is(err, domain.ErrInvalidOTP):
		return "invalid_otp"
	case errors.Is(err, domain.ErrOTPExpired):
		return "expired"
	case errors.Is(err, domain.ErrOTPAlreadyUsed):
		return "already_used"
	case errors.Is(err, domain.ErrOTPAttemptsExceeded):
		return "attempts_exceeded"
	default:
		return "unknown"
	}
}
