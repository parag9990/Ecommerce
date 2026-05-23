package httptransport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"ecommerce/api-gateway/internal/clients"
	"ecommerce/api-gateway/internal/config"
	"ecommerce/api-gateway/internal/repository"
	"ecommerce/api-gateway/internal/usecase"
	"github.com/golang-jwt/jwt/v5"
)

func TestNewRouterRegistersDefinedRoute(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/p_123", nil)
	req.Header.Set("X-Request-Id", "req_test")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501 for defined route bridge, got %d", rec.Code)
	}
	var envelope ResponseEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.RequestID != "req_test" {
		t.Fatalf("expected request id propagation, got %q", envelope.RequestID)
	}
	if envelope.Error == nil || envelope.Error.Code != "ROUTE_BRIDGE_NOT_CONFIGURED" {
		t.Fatalf("expected bridge error envelope, got %+v", envelope.Error)
	}
}

func TestNewRouterHealthReady(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected ready status 200, got %d", rec.Code)
	}
	var envelope ResponseEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := envelope.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected health data map, got %T", envelope.Data)
	}
	if data["status"] != "ready" {
		t.Fatalf("expected ready status data, got %+v", data)
	}
}

func TestNewRouterHealthReadyChecksDownstreams(t *testing.T) {
	router := newTestRouterWithHealth(t, stubDownstreamHealth{
		report: clients.HealthReport{
			clients.DownstreamAuth: {
				Service: "auth",
				Status:  clients.HealthStatusUnavailable,
				Error:   "connection refused",
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected readiness failure 503, got %d", rec.Code)
	}
	var envelope ResponseEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Error == nil || envelope.Error.Code != "DOWNSTREAM_SERVICES_UNAVAILABLE" {
		t.Fatalf("expected downstream readiness error, got %+v", envelope.Error)
	}
}

func TestProtectedRouteRequiresBearerToken(t *testing.T) {
	router := newAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "UNAUTHORIZED")
}

func TestProtectedRouteAllowsMatchingRole(t *testing.T) {
	router := newAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer buyer-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected route bridge response after auth, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "ROUTE_BRIDGE_NOT_CONFIGURED")
}

func TestProtectedRouteRejectsInsufficientRole(t *testing.T) {
	router := newAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/seller/products", nil)
	req.Header.Set("Authorization", "Bearer buyer-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for insufficient role, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "FORBIDDEN")
}

func TestWebhookRouteRequiresSignatureInsteadOfJWT(t *testing.T) {
	router := newAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/payments/testpay", nil)
	req.Header.Set("Authorization", "Bearer buyer-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing webhook signature, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "UNAUTHORIZED")
}

func TestWebhookRouteAllowsConfiguredSignatureHeader(t *testing.T) {
	router := newAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/payments/testpay", nil)
	req.Header.Set("X-Provider-Signature", "signed")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected route bridge response after webhook guard, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "ROUTE_BRIDGE_NOT_CONFIGURED")
}

func TestRouterValidationRejectsInvalidPublicBodyBeforeBridge(t *testing.T) {
	router := newValidationTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", strings.NewReader(`{
		"email":"not-an-email",
		"password":"short",
		"full_name":"Buyer One"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected validation 400, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "VALIDATION_ERROR")
}

func TestRouterValidationRejectsProtectedBodyAfterAuth(t *testing.T) {
	router := newValidationTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", strings.NewReader(`{
		"product_id":"prod_12345678",
		"variant_id":"var_12345678",
		"quantity":0
	}`))
	req.Header.Set("Authorization", "Bearer buyer-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected validation 400, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "VALIDATION_ERROR")
}

func TestProtectedRoutesFailClosedWithoutVerifierConfiguration(t *testing.T) {
	path := writeAuthTestContract(t)
	cfg := authTestConfig(path)
	cfg.JWTJWKSURL = ""
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	_, err := NewRouter(context.Background(), cfg, catalog, logger, nil)
	if err == nil {
		t.Fatal("expected protected routes without verifier configuration to fail")
	}
	if !strings.Contains(err.Error(), "JWT_JWKS_URL is required") {
		t.Fatalf("expected missing jwks url error, got %v", err)
	}
}

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	return newTestRouterWithHealth(t, nil)
}

func newTestRouterWithHealth(t *testing.T, downstreamHealth DownstreamHealthChecker) http.Handler {
	t.Helper()
	path := filepath.Join(t.TempDir(), "master-api.json")
	body := `{
		"project": "scalable-ecommerce-platform",
		"version": "1.0.0",
		"rest_endpoints": [
			{
				"id": "product.detail",
				"method": "GET",
				"path": "/api/v1/products/{product_id}",
				"service": "product-service",
				"grpc": "ProductService.GetProduct",
				"auth": "public",
				"request_schema": "IdPathRequest",
				"response_schema": "Product"
			}
		]
	}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}

	cfg := config.Config{
		ServiceName:       "api-gateway",
		Environment:       "test",
		HTTPAddress:       ":0",
		APIBasePath:       "/api/v1",
		APIContractPath:   path,
		LogLevel:          "error",
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second,
	}
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router, err := NewRouter(context.Background(), cfg, catalog, logger, downstreamHealth)
	if err != nil {
		t.Fatalf("new router: %v", err)
	}
	return router
}

func newAuthTestRouter(t *testing.T) http.Handler {
	t.Helper()
	path := writeAuthTestContract(t)
	cfg := authTestConfig(path)
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router, err := NewRouterWithOptions(context.Background(), cfg, catalog, logger, nil, RouterOptions{
		TokenVerifier: stubTokenVerifier{},
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}
	return router
}

func newValidationTestRouter(t *testing.T) http.Handler {
	t.Helper()
	path := writeValidationTestContract(t)
	cfg := authTestConfig(path)
	cfg.Validation = config.RequestValidationConfig{
		Enabled:             true,
		DefaultMaxBodyBytes: 128 * 1024,
		MaxHeaderBytes:      32 * 1024,
		MaxQueryBytes:       8 * 1024,
	}
	repo := repository.NewJSONRouteRepository(path)
	catalog := usecase.NewRouteCatalogService(repo, cfg.APIBasePath)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router, err := NewRouterWithOptions(context.Background(), cfg, catalog, logger, nil, RouterOptions{
		TokenVerifier: stubTokenVerifier{},
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}
	return router
}

func writeAuthTestContract(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "master-api.json")
	body := `{
		"project": "scalable-ecommerce-platform",
		"version": "1.0.0",
		"rest_endpoints": [
			{
				"id": "user.me_get",
				"method": "GET",
				"path": "/api/v1/me",
				"service": "user-service",
				"grpc": "UserService.GetUser",
				"auth": "buyer",
				"request_schema": "Empty",
				"response_schema": "UserProfile"
			},
			{
				"id": "seller.product_create",
				"method": "POST",
				"path": "/api/v1/seller/products",
				"service": "product-service",
				"grpc": "ProductService.CreateProduct",
				"auth": "seller",
				"request_schema": "ProductInput",
				"response_schema": "Product"
			},
			{
				"id": "payment.webhook",
				"method": "POST",
				"path": "/api/v1/webhooks/payments/{provider}",
				"service": "payment-service",
				"grpc": "PaymentService.HandleWebhook",
				"auth": "webhook",
				"request_schema": "PaymentWebhookRequest",
				"response_schema": "SuccessResponse"
			}
		]
	}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	return path
}

func writeValidationTestContract(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "master-api.json")
	body := `{
		"project": "scalable-ecommerce-platform",
		"version": "1.0.0",
		"rest_endpoints": [
			{
				"id": "auth.signup",
				"method": "POST",
				"path": "/api/v1/auth/signup",
				"service": "auth-service",
				"grpc": "AuthService.Signup",
				"auth": "public",
				"request_schema": "SignupRequest",
				"response_schema": "TokenResponse"
			},
			{
				"id": "cart.add_item",
				"method": "POST",
				"path": "/api/v1/cart/items",
				"service": "cart-service",
				"grpc": "CartService.AddItem",
				"auth": "buyer",
				"request_schema": "CartItemInput",
				"response_schema": "Cart"
			}
		],
		"schemas": {
			"SignupRequest": {
				"type": "object",
				"required": ["email", "password", "full_name"],
				"properties": {
					"email": {"type": "string", "format": "email"},
					"password": {"type": "string", "minLength": 8},
					"full_name": {"type": "string"}
				}
			},
			"CartItemInput": {
				"type": "object",
				"required": ["product_id", "variant_id", "quantity"],
				"properties": {
					"product_id": {"type": "string"},
					"variant_id": {"type": "string"},
					"quantity": {"type": "integer", "minimum": 1}
				}
			}
		}
	}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	return path
}

func authTestConfig(path string) config.Config {
	return config.Config{
		ServiceName:            "api-gateway",
		Environment:            "test",
		HTTPAddress:            ":0",
		APIBasePath:            "/api/v1",
		APIContractPath:        path,
		LogLevel:               "error",
		ReadHeaderTimeout:      time.Second,
		ShutdownTimeout:        time.Second,
		JWTIssuer:              "ecommerce-auth",
		JWTAudience:            "ecommerce-api",
		JWTAllowedAlgs:         []string{"RS256"},
		JWTJWKSURL:             "http://auth-service/.well-known/jwks.json",
		JWTJWKSCacheTTL:        time.Minute,
		JWTJWKSFetchTimeout:    time.Second,
		JWTClockSkew:           30 * time.Second,
		WebhookSignatureHeader: "X-Provider-Signature",
	}
}

func assertErrorCode(t *testing.T, rec *httptest.ResponseRecorder, code string) {
	t.Helper()
	var envelope ResponseEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Error == nil || envelope.Error.Code != code {
		t.Fatalf("expected error code %q, got %+v", code, envelope.Error)
	}
}

type stubDownstreamHealth struct {
	report clients.HealthReport
}

func (s stubDownstreamHealth) Check(context.Context) clients.HealthReport {
	return s.report
}

type stubTokenVerifier struct{}

func (stubTokenVerifier) Verify(_ context.Context, raw string) (gatewayauth.AccessClaims, error) {
	switch raw {
	case "buyer-token":
		return gatewayauth.AccessClaims{
			SessionID: "sess_buyer",
			Roles:     []string{"buyer"},
			TokenType: gatewayauth.AccessTokenType,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user_buyer",
			},
		}, nil
	case "seller-token":
		return gatewayauth.AccessClaims{
			SessionID: "sess_seller",
			Roles:     []string{"seller"},
			TokenType: gatewayauth.AccessTokenType,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user_seller",
			},
		}, nil
	default:
		return gatewayauth.AccessClaims{}, errors.New("invalid token")
	}
}
