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
