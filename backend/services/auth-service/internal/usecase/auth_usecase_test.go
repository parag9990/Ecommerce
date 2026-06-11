package usecase

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/sessionlink"
)

func TestAuthUsecaseLoginRecordsSessionLinkEvent(t *testing.T) {
	linker := &recordingSessionLinker{}
	hasher, err := sessionlink.NewPrivacyHasher("session-pepper")
	if err != nil {
		t.Fatalf("NewPrivacyHasher() error = %v", err)
	}

	uc, err := NewAuthUsecase(
		fakeCredentialVerifier{accountID: "auth_123"},
		fakeAccountTokenIssuer{pair: TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresIn:    900,
			SessionID:    "sess_123",
			UserID:       "user_123",
			Roles:        []string{"buyer"},
		}},
		linker,
		hasher,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatalf("NewAuthUsecase() error = %v", err)
	}

	session, err := uc.Login(context.Background(), LoginInput{
		Identifier: "buyer@example.com",
		Password:   "plain-password",
		TraceID:    "trace_123",
		Device: SessionDeviceInput{
			AnonymousID: "anon_456",
			Fingerprint: "raw-fingerprint",
			UserAgent:   "Mozilla/5.0",
			Channel:     "web",
			Locale:      "en-US",
		},
		Network: SessionNetworkInput{IPAddress: "203.0.113.10"},
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if session.SessionID != "sess_123" {
		t.Fatalf("session id = %q", session.SessionID)
	}
	if linker.login.SessionID != "sess_123" || linker.login.UserID != "user_123" {
		t.Fatalf("recorded login event = %+v", linker.login)
	}
	if linker.login.Device.DeviceFingerprintHash == "" || linker.login.Device.DeviceFingerprintHash == "raw-fingerprint" {
		t.Fatalf("device fingerprint hash = %q", linker.login.Device.DeviceFingerprintHash)
	}
	if linker.login.Network.IPHash == "" || linker.login.Network.IPHash == "203.0.113.10" {
		t.Fatalf("ip hash = %q", linker.login.Network.IPHash)
	}
	if linker.login.Auth.Method != sessionlink.AuthMethodPassword {
		t.Fatalf("auth method = %q", linker.login.Auth.Method)
	}
}

type fakeCredentialVerifier struct {
	accountID string
	err       error
}

func (v fakeCredentialVerifier) VerifyPassword(ctx context.Context, input VerifyPasswordInput) (VerifyPasswordOutput, error) {
	if v.err != nil {
		return VerifyPasswordOutput{}, v.err
	}
	return VerifyPasswordOutput{
		AccountID:     v.accountID,
		AccountStatus: domain.AccountStatusActive,
	}, nil
}

type fakeAccountTokenIssuer struct {
	pair TokenPair
	err  error
}

func (i fakeAccountTokenIssuer) IssueTokenPairForAccount(ctx context.Context, accountID string) (TokenPair, error) {
	if i.err != nil {
		return TokenPair{}, i.err
	}
	return i.pair, nil
}

type recordingSessionLinker struct {
	login sessionlink.LoginSucceededEvent
}

func (l *recordingSessionLinker) RecordLoginSucceeded(ctx context.Context, event sessionlink.LoginSucceededEvent) error {
	l.login = event
	return nil
}

func (l *recordingSessionLinker) RecordSignupSucceeded(ctx context.Context, event sessionlink.SignupSucceededEvent) error {
	return nil
}

func (l *recordingSessionLinker) RecordLogoutSucceeded(ctx context.Context, event sessionlink.LogoutSucceededEvent) error {
	return nil
}

func (l *recordingSessionLinker) RecordRefreshReuseDetected(ctx context.Context, event sessionlink.RefreshReuseDetectedEvent) error {
	return nil
}
