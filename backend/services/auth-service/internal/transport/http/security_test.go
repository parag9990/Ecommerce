package httptransport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/authctx"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
	tokensecurity "github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/token"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/usecase"
)

func TestAuthRequiredMissingTokenReturnsUnauthorized(t *testing.T) {
	handler := &Handler{
		tokenUsecase: fakeHTTPTokenUsecase{},
		logger:       slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
	}
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/internal/v1/auth/roles", nil)
	rr := httptest.NewRecorder()
	handler.AuthRequired(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
	if nextCalled {
		t.Fatal("protected handler ran without bearer token")
	}
	assertAPIError(t, rr.Body.Bytes(), "AUTHENTICATION_REQUIRED", "Authentication required")
}

func TestAuthRequiredInvalidTokenRedactsAuthorizationHeader(t *testing.T) {
	logs := &bytes.Buffer{}
	handler := &Handler{
		tokenUsecase: fakeHTTPTokenUsecase{verifyErr: domain.ErrInvalidAccessToken},
		logger:       slog.New(slog.NewTextHandler(logs, nil)),
	}
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	rawToken := "raw.jwt.secret"
	req := httptest.NewRequest(http.MethodGet, "/internal/v1/auth/roles", nil)
	req.Header.Set("Authorization", "Bearer "+rawToken)
	req.Header.Set("X-Request-ID", "req_123")
	rr := httptest.NewRecorder()
	handler.AuthRequired(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
	if nextCalled {
		t.Fatal("protected handler ran with invalid bearer token")
	}
	assertAPIError(t, rr.Body.Bytes(), "INVALID_ACCESS_TOKEN", "Invalid access token")

	output := logs.String()
	if strings.Contains(output, rawToken) || strings.Contains(output, "Authorization") || strings.Contains(output, "Bearer") {
		t.Fatalf("authorization header leaked into logs: %s", output)
	}
	if !strings.Contains(output, "req_123") {
		t.Fatalf("request id missing from denial log: %s", output)
	}
}

func TestRequireRolesBuyerCannotAccessSellerHandler(t *testing.T) {
	logs := &bytes.Buffer{}
	handler := &Handler{
		logger: slog.New(slog.NewTextHandler(logs, nil)),
	}
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/seller/products", nil)
	req = req.WithContext(authctx.WithClaims(req.Context(), authctx.Claims{
		UserID: "user_buyer",
		Roles:  []string{"buyer"},
	}))
	req.Header.Set("X-Request-ID", "req_rbac")
	rr := httptest.NewRecorder()

	handler.RequireRoles(domain.RoleSeller, domain.RoleSellerCatalogEditor, domain.RoleSuperadmin)(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
	if nextCalled {
		t.Fatal("seller handler ran for buyer role")
	}
	assertAPIError(t, rr.Body.Bytes(), "PERMISSION_DENIED", "Permission denied")
	if !strings.Contains(logs.String(), "auth.http.rbac_denied") {
		t.Fatalf("rbac denial was not logged: %s", logs.String())
	}
}

func TestWriteUsecaseErrorMapsSecurityFailuresToGenericResponses(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
		wantMsg    string
		secret     string
	}{
		{
			name:       "otp replay",
			err:        domain.ErrOTPAlreadyUsed,
			wantStatus: http.StatusUnauthorized,
			wantCode:   "INVALID_OTP",
			wantMsg:    "Invalid or expired OTP",
			secret:     "already used",
		},
		{
			name:       "refresh reuse",
			err:        domain.ErrRefreshTokenReuse,
			wantStatus: http.StatusUnauthorized,
			wantCode:   "INVALID_REFRESH_TOKEN",
			wantMsg:    "Invalid refresh token",
			secret:     "reuse",
		},
		{
			name:       "role bypass",
			err:        domain.ErrForbidden,
			wantStatus: http.StatusForbidden,
			wantCode:   "PERMISSION_DENIED",
			wantMsg:    "Permission denied",
			secret:     "role",
		},
	}

	handler := &Handler{logger: slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			rr := httptest.NewRecorder()

			handler.writeUsecaseError(rr, req, tt.err)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
			body := rr.Body.String()
			assertAPIError(t, rr.Body.Bytes(), tt.wantCode, tt.wantMsg)
			if strings.Contains(strings.ToLower(body), tt.secret) {
				t.Fatalf("response leaked internal detail %q: %s", tt.secret, body)
			}
		})
	}
}

func assertAPIError(t *testing.T, body []byte, wantCode string, wantMessage string) {
	t.Helper()

	var got errorResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal(errorResponse) error = %v; body=%s", err, string(body))
	}
	if got.Error.Code != wantCode || got.Error.Message != wantMessage {
		t.Fatalf("api error = %+v, want code=%q message=%q", got.Error, wantCode, wantMessage)
	}
}

type fakeHTTPTokenUsecase struct {
	verifyClaims domain.TokenClaims
	verifyErr    error
}

func (u fakeHTTPTokenUsecase) IssueTokenPair(ctx context.Context, input usecase.IssueTokenPairInput) (usecase.TokenPair, error) {
	return usecase.TokenPair{}, errors.New("not implemented in test fake")
}

func (u fakeHTTPTokenUsecase) RefreshToken(ctx context.Context, input usecase.RefreshTokenInput) (usecase.TokenPair, error) {
	return usecase.TokenPair{}, errors.New("not implemented in test fake")
}

func (u fakeHTTPTokenUsecase) Logout(ctx context.Context, input usecase.LogoutInput) error {
	return errors.New("not implemented in test fake")
}

func (u fakeHTTPTokenUsecase) VerifyAccessToken(ctx context.Context, accessToken string) (domain.TokenClaims, error) {
	if u.verifyErr != nil {
		return domain.TokenClaims{}, u.verifyErr
	}
	if u.verifyClaims.UserID != "" {
		return u.verifyClaims, nil
	}
	return domain.TokenClaims{
		UserID:    "user_123",
		SessionID: "sess_123",
		Roles:     []string{"buyer"},
		ExpiresAt: time.Now().Add(time.Minute),
	}, nil
}

func (u fakeHTTPTokenUsecase) JWKS(ctx context.Context) tokensecurity.JWKSet {
	return tokensecurity.JWKSet{}
}
