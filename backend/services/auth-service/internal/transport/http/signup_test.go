package httptransport

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/usecase"
)

func TestSignupRouteInvokesUsecase(t *testing.T) {
	signup := &fakeSignupUsecase{}
	handler := &Handler{
		signupUsecase: signup,
		logger:        slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
	}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", strings.NewReader(`{
		"email": "buyer@example.com",
		"phone": "+15551234567",
		"full_name": "Buyer One",
		"password": "StrongerPass123",
		"role": "buyer"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "req_signup")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if signup.input.Email != "buyer@example.com" ||
		signup.input.Phone != "+15551234567" ||
		signup.input.FullName != "Buyer One" ||
		signup.input.Role != "buyer" ||
		signup.input.TraceID != "req_signup" {
		t.Fatalf("signup input = %+v", signup.input)
	}

	var body authSessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.User.UserID != "user_123" || body.SessionID != "sess_123" || body.Tokens.AccessToken == "" {
		t.Fatalf("response = %+v", body)
	}
}

func TestLoginRouteAcceptsDeviceTimezone(t *testing.T) {
	auth := &fakeAuthUsecase{}
	handler := &Handler{
		authUsecase: auth,
		logger:      slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
	}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{
		"identifier": "buyer@example.com",
		"password": "StrongerPass123",
		"device": {
			"channel": "web",
			"locale": "en-IN",
			"timezone": "Asia/Kolkata",
			"user_agent": "TestBrowser/1.0"
		}
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if auth.input.Device.Timezone != "Asia/Kolkata" {
		t.Fatalf("timezone = %q", auth.input.Device.Timezone)
	}
	if auth.input.Device.UserAgent != "TestBrowser/1.0" {
		t.Fatalf("user agent = %q", auth.input.Device.UserAgent)
	}
	if auth.input.Device.Channel != "web" || auth.input.Device.Locale != "en-IN" {
		t.Fatalf("device input = %+v", auth.input.Device)
	}
}

func TestPublicResetPasswordRouteVerifiesResetOTPAndResetsPassword(t *testing.T) {
	accountID := "auth_123"
	passwords := &fakePasswordUsecase{}
	otps := &fakeOTPUsecase{
		challengeOutput: usecase.VerifyOTPOutput{
			ChallengeID: "otp_chal_test123",
			AccountID:   &accountID,
			Target:      "buyer@example.com",
			Channel:     domain.OTPChannelEmail,
			Purpose:     domain.OTPPurposePasswordReset,
		},
		verifyOutput: usecase.VerifyOTPOutput{
			ChallengeID: "otp_chal_test123",
			AccountID:   &accountID,
			Target:      "buyer@example.com",
			Channel:     domain.OTPChannelEmail,
			Purpose:     domain.OTPPurposePasswordReset,
		},
	}
	handler := &Handler{
		passwordUsecase: passwords,
		otpUsecase:      otps,
		logger:          slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
	}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/reset", strings.NewReader(`{
		"challenge_id": "otp_chal_test123",
		"otp": "123456",
		"new_password": "StrongerPass123"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if otps.verifyPurpose != domain.OTPPurposePasswordReset {
		t.Fatalf("verify purpose = %q", otps.verifyPurpose)
	}
	if otps.verifyInput.ChallengeID != "otp_chal_test123" || otps.verifyInput.OTP != "123456" {
		t.Fatalf("verify input = %+v", otps.verifyInput)
	}
	if passwords.resetInput.AccountID != "auth_123" ||
		passwords.resetInput.NewPassword != "StrongerPass123" {
		t.Fatalf("reset input = %+v", passwords.resetInput)
	}
	if passwords.resetByIdentifierInput.Identifier != "" {
		t.Fatalf("reset by identifier should not be used: %+v", passwords.resetByIdentifierInput)
	}
}

func TestPublicResetPasswordRouteResolvesLegacyChallengeBeforeVerifyingOTP(t *testing.T) {
	passwords := &fakePasswordUsecase{
		resolveOutput: usecase.PasswordResetAccount{
			AccountID:  "auth_123",
			Identifier: "buyer@example.com",
		},
	}
	otps := &fakeOTPUsecase{
		challengeOutput: usecase.VerifyOTPOutput{
			ChallengeID: "otp_chal_test123",
			Target:      "buyer@example.com",
			Channel:     domain.OTPChannelEmail,
			Purpose:     domain.OTPPurposePasswordReset,
		},
		verifyOutput: usecase.VerifyOTPOutput{
			ChallengeID: "otp_chal_test123",
			Target:      "buyer@example.com",
			Channel:     domain.OTPChannelEmail,
			Purpose:     domain.OTPPurposePasswordReset,
		},
	}
	handler := &Handler{
		passwordUsecase: passwords,
		otpUsecase:      otps,
		logger:          slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
	}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/reset", strings.NewReader(`{
		"challenge_id": "otp_chal_test123",
		"otp": "123456",
		"new_password": "StrongerPass123"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if passwords.resolveInput.Identifier != "buyer@example.com" {
		t.Fatalf("resolve input = %+v", passwords.resolveInput)
	}
	if otps.verifyInput.ChallengeID != "otp_chal_test123" || otps.verifyInput.OTP != "123456" {
		t.Fatalf("verify input = %+v", otps.verifyInput)
	}
	if passwords.resetInput.AccountID != "auth_123" {
		t.Fatalf("reset input = %+v", passwords.resetInput)
	}
}

func TestPublicResetPasswordRouteDoesNotVerifyOTPWhenChallengeTargetHasNoCredential(t *testing.T) {
	passwords := &fakePasswordUsecase{resolveErr: domain.ErrInvalidCredentials}
	otps := &fakeOTPUsecase{
		challengeOutput: usecase.VerifyOTPOutput{
			ChallengeID: "otp_chal_test123",
			Target:      "missing@example.com",
			Channel:     domain.OTPChannelEmail,
			Purpose:     domain.OTPPurposePasswordReset,
		},
	}
	handler := &Handler{
		passwordUsecase: passwords,
		otpUsecase:      otps,
		logger:          slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
	}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/reset", strings.NewReader(`{
		"challenge_id": "otp_chal_test123",
		"otp": "123456",
		"new_password": "StrongerPass123"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if otps.verifyInput.ChallengeID != "" {
		t.Fatalf("otp should not be verified when target has no credential: %+v", otps.verifyInput)
	}
	if passwords.resetInput.AccountID != "" {
		t.Fatalf("password should not be reset: %+v", passwords.resetInput)
	}
}

func TestForgotPasswordRouteBindsOTPToResolvedAccount(t *testing.T) {
	passwords := &fakePasswordUsecase{
		resolveOutput: usecase.PasswordResetAccount{
			AccountID:  "auth_123",
			Identifier: "buyer@example.com",
		},
	}
	otps := &fakeOTPUsecase{
		createOutput: usecase.CreateOTPChallengeOutput{
			ChallengeID:      "otp_chal_test123",
			ExpiresInSeconds: 300,
		},
	}
	handler := &Handler{
		passwordUsecase: passwords,
		otpUsecase:      otps,
		logger:          slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
	}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/forgot", strings.NewReader(`{
		"identifier": "Buyer@Example.COM"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if passwords.resolveInput.Identifier != "Buyer@Example.COM" {
		t.Fatalf("resolve input = %+v", passwords.resolveInput)
	}
	if otps.createInput.AccountID == nil || *otps.createInput.AccountID != "auth_123" {
		t.Fatalf("otp account id = %v", otps.createInput.AccountID)
	}
	if otps.createInput.Target != "buyer@example.com" ||
		otps.createInput.Channel != domain.OTPChannelEmail ||
		otps.createInput.Purpose != domain.OTPPurposePasswordReset {
		t.Fatalf("otp create input = %+v", otps.createInput)
	}
}

type fakeSignupUsecase struct {
	input usecase.SignupInput
}

func (u *fakeSignupUsecase) Signup(ctx context.Context, input usecase.SignupInput) (usecase.AuthSession, error) {
	u.input = input
	return usecase.AuthSession{
		User: usecase.AuthenticatedUser{
			UserID: "user_123",
			Roles:  []string{"buyer"},
		},
		Tokens: usecase.TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresIn:    900,
			SessionID:    "sess_123",
			UserID:       "user_123",
			Roles:        []string{"buyer"},
		},
		SessionID: "sess_123",
	}, nil
}

type fakeAuthUsecase struct {
	input usecase.LoginInput
}

func (u *fakeAuthUsecase) Login(ctx context.Context, input usecase.LoginInput) (usecase.AuthSession, error) {
	u.input = input
	return usecase.AuthSession{
		User: usecase.AuthenticatedUser{
			UserID: "user_123",
			Roles:  []string{"buyer"},
		},
		Tokens: usecase.TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresIn:    900,
			SessionID:    "sess_123",
			UserID:       "user_123",
			Roles:        []string{"buyer"},
		},
		SessionID: "sess_123",
	}, nil
}

type fakePasswordUsecase struct {
	resetInput             usecase.ResetPasswordInput
	resetByIdentifierInput usecase.ResetPasswordByIdentifierInput
	resolveInput           usecase.ResolvePasswordResetAccountInput
	resolveOutput          usecase.PasswordResetAccount
	resolveErr             error
}

func (u *fakePasswordUsecase) CreateCredential(ctx context.Context, input usecase.CreateCredentialInput) (usecase.CredentialSummary, error) {
	return usecase.CredentialSummary{AccountID: input.AccountID}, nil
}

func (u *fakePasswordUsecase) VerifyPassword(ctx context.Context, input usecase.VerifyPasswordInput) (usecase.VerifyPasswordOutput, error) {
	return usecase.VerifyPasswordOutput{AccountID: "auth_123"}, nil
}

func (u *fakePasswordUsecase) ResolvePasswordResetAccount(ctx context.Context, input usecase.ResolvePasswordResetAccountInput) (usecase.PasswordResetAccount, error) {
	u.resolveInput = input
	if u.resolveErr != nil {
		return usecase.PasswordResetAccount{}, u.resolveErr
	}
	if u.resolveOutput.AccountID != "" {
		return u.resolveOutput, nil
	}
	return usecase.PasswordResetAccount{AccountID: "auth_123", Identifier: input.Identifier}, nil
}

func (u *fakePasswordUsecase) ResetPassword(ctx context.Context, input usecase.ResetPasswordInput) (usecase.CredentialSummary, error) {
	u.resetInput = input
	return usecase.CredentialSummary{AccountID: input.AccountID}, nil
}

func (u *fakePasswordUsecase) ResetPasswordByIdentifier(ctx context.Context, input usecase.ResetPasswordByIdentifierInput) (usecase.CredentialSummary, error) {
	u.resetByIdentifierInput = input
	return usecase.CredentialSummary{AccountID: "auth_123"}, nil
}

type fakeOTPUsecase struct {
	createInput      usecase.CreateOTPChallengeInput
	createOutput     usecase.CreateOTPChallengeOutput
	challengeID      string
	challengePurpose domain.OTPPurpose
	challengeOutput  usecase.VerifyOTPOutput
	verifyInput      usecase.VerifyOTPInput
	verifyPurpose    domain.OTPPurpose
	verifyOutput     usecase.VerifyOTPOutput
}

func (u *fakeOTPUsecase) CreateOTPChallenge(ctx context.Context, input usecase.CreateOTPChallengeInput) (usecase.CreateOTPChallengeOutput, error) {
	u.createInput = input
	if u.createOutput.ChallengeID != "" {
		return u.createOutput, nil
	}
	return usecase.CreateOTPChallengeOutput{ChallengeID: "otp_chal_test123"}, nil
}

func (u *fakeOTPUsecase) GetOTPChallengeForPurpose(ctx context.Context, challengeID string, purpose domain.OTPPurpose) (usecase.VerifyOTPOutput, error) {
	u.challengeID = challengeID
	u.challengePurpose = purpose
	return u.challengeOutput, nil
}

func (u *fakeOTPUsecase) VerifyOTP(ctx context.Context, input usecase.VerifyOTPInput) error {
	u.verifyInput = input
	return nil
}

func (u *fakeOTPUsecase) VerifyOTPForPurpose(ctx context.Context, input usecase.VerifyOTPInput, purpose domain.OTPPurpose) (usecase.VerifyOTPOutput, error) {
	u.verifyInput = input
	u.verifyPurpose = purpose
	return u.verifyOutput, nil
}
