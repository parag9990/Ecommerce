package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConfigValidateRequiresDownstreamGRPCTargets(t *testing.T) {
	tests := []struct {
		name    string
		clear   func(*Config)
		wantErr string
	}{
		{name: "auth", clear: func(cfg *Config) { cfg.AuthGRPCAddr = "" }, wantErr: "AUTH_GRPC_ADDR is required"},
		{name: "user", clear: func(cfg *Config) { cfg.UserGRPCAddr = "" }, wantErr: "USER_GRPC_ADDR is required"},
		{name: "product", clear: func(cfg *Config) { cfg.ProductGRPCAddr = "" }, wantErr: "PRODUCT_GRPC_ADDR is required"},
		{name: "cart", clear: func(cfg *Config) { cfg.CartGRPCAddr = "" }, wantErr: "CART_GRPC_ADDR is required"},
		{name: "wishlist", clear: func(cfg *Config) { cfg.WishlistGRPCAddr = "" }, wantErr: "WISHLIST_GRPC_ADDR is required"},
		{name: "order", clear: func(cfg *Config) { cfg.OrderGRPCAddr = "" }, wantErr: "ORDER_GRPC_ADDR is required"},
		{name: "payment", clear: func(cfg *Config) { cfg.PaymentGRPCAddr = "" }, wantErr: "PAYMENT_GRPC_ADDR is required"},
		{name: "search", clear: func(cfg *Config) { cfg.SearchGRPCAddr = "" }, wantErr: "SEARCH_GRPC_ADDR is required"},
		{name: "cms", clear: func(cfg *Config) { cfg.CMSGRPCAddr = "" }, wantErr: "CMS_GRPC_ADDR is required"},
		{name: "session", clear: func(cfg *Config) { cfg.SessionGRPCAddr = "" }, wantErr: "SESSION_GRPC_ADDR is required"},
		{name: "notification", clear: func(cfg *Config) { cfg.NotificationGRPCAddr = "" }, wantErr: "NOTIFICATION_GRPC_ADDR is required"},
		{name: "superadmin", clear: func(cfg *Config) { cfg.SuperadminGRPCAddr = "" }, wantErr: "SUPERADMIN_GRPC_ADDR is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig(t)
			tt.clear(&cfg)

			err := cfg.Validate()
			if err == nil {
				t.Fatalf("expected missing %s grpc address to fail validation", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected %q validation error, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestConfigLoadReadsCORSOrigins(t *testing.T) {
	contractPath := writeTestContract(t)
	t.Setenv("API_CONTRACT_PATH", contractPath)
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3003,https://admin.example.com")
	setGRPCTargetEnv(t)

	cfg, err := Load(context.Background())
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if got := strings.Join(cfg.CORS.AllowedOrigins, ","); got != "http://localhost:3003,https://admin.example.com" {
		t.Fatalf("unexpected CORS origins %q", got)
	}
}

func TestConfigValidateRejectsWildcardCORSOrigin(t *testing.T) {
	cfg := validConfig(t)
	cfg.CORS.AllowedOrigins = []string{"*"}

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "wildcard origin is not allowed") {
		t.Fatalf("expected wildcard CORS origin error, got %v", err)
	}
}

func TestConfigLoadReadsGRPCSettings(t *testing.T) {
	contractPath := writeTestContract(t)
	t.Setenv("API_CONTRACT_PATH", contractPath)
	t.Setenv("GRPC_TLS_ENABLED", "true")
	t.Setenv("GRPC_DIAL_TIMEOUT", "1500ms")
	setGRPCTargetEnv(t)

	cfg, err := Load(context.Background())
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.GRPCTLSEnabled {
		t.Fatal("expected grpc tls to be enabled")
	}
	if cfg.GRPCDialTimeout != 1500*time.Millisecond {
		t.Fatalf("expected grpc dial timeout 1500ms, got %s", cfg.GRPCDialTimeout)
	}
	if cfg.ProductGRPCAddr != "product-service:9090" {
		t.Fatalf("expected product grpc target from env, got %q", cfg.ProductGRPCAddr)
	}
}

func TestConfigLoadReadsGRPCWebSettings(t *testing.T) {
	contractPath := writeTestContract(t)
	policyPath := filepath.Join(t.TempDir(), "grpcweb-policies.json")
	if err := os.WriteFile(policyPath, []byte(`{"policies":[]}`), 0o600); err != nil {
		t.Fatalf("write grpc-web policy: %v", err)
	}
	t.Setenv("API_CONTRACT_PATH", contractPath)
	t.Setenv("GRPC_WEB_ENABLED", "true")
	t.Setenv("GRPC_ADDR", ":19090")
	t.Setenv("GRPC_WEB_POLICY_PATH", policyPath)
	t.Setenv("GRPC_WEB_EXPOSED_SERVICES", "ecommerce.session.v1.SessionService,ecommerce.superadmin.v1.SuperadminService")
	t.Setenv("GRPC_WEB_MAX_RECEIVE_MESSAGE_BYTES", "2048")
	t.Setenv("GRPC_WEB_MAX_SEND_MESSAGE_BYTES", "4096")
	setGRPCTargetEnv(t)

	cfg, err := Load(context.Background())
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.GRPCWeb.Enabled || cfg.GRPCWeb.Address != ":19090" {
		t.Fatalf("unexpected grpc-web config: %+v", cfg.GRPCWeb)
	}
	if cfg.GRPCWeb.PolicyPath != policyPath || cfg.GRPCWeb.MaxReceiveMsgBytes != 2048 || cfg.GRPCWeb.MaxSendMsgBytes != 4096 {
		t.Fatalf("unexpected grpc-web policy/message config: %+v", cfg.GRPCWeb)
	}
	if strings.Join(cfg.GRPCWeb.ExposedServices, ",") != "ecommerce.session.v1.SessionService,ecommerce.superadmin.v1.SuperadminService" {
		t.Fatalf("unexpected exposed services: %v", cfg.GRPCWeb.ExposedServices)
	}
}

func TestConfigValidateRejectsEnabledGRPCWebWithoutAllowlist(t *testing.T) {
	cfg := validConfig(t)
	cfg.GRPCWeb.Enabled = true
	cfg.GRPCWeb.Address = ":19090"
	cfg.GRPCWeb.PolicyPath = writeTestContract(t)
	cfg.GRPCWeb.MaxReceiveMsgBytes = 1024
	cfg.GRPCWeb.MaxSendMsgBytes = 1024

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "GRPC_WEB_EXPOSED_SERVICES") {
		t.Fatalf("expected missing grpc-web allowlist error, got %v", err)
	}
}

func TestConfigValidateRejectsGRPCWebHTTPAddressConflict(t *testing.T) {
	cfg := validConfig(t)
	cfg.GRPCWeb.Enabled = true
	cfg.GRPCWeb.Address = cfg.HTTPAddress
	cfg.GRPCWeb.PolicyPath = writeTestContract(t)
	cfg.GRPCWeb.ExposedServices = []string{"ecommerce.session.v1.SessionService"}
	cfg.GRPCWeb.MaxReceiveMsgBytes = 1024
	cfg.GRPCWeb.MaxSendMsgBytes = 1024

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "GRPC_ADDR must not equal HTTP_ADDR") {
		t.Fatalf("expected grpc/http address conflict error, got %v", err)
	}
}

func TestConfigLoadReadsJWTSettings(t *testing.T) {
	contractPath := writeTestContract(t)
	t.Setenv("API_CONTRACT_PATH", contractPath)
	t.Setenv("JWT_ISSUER", "issuer-test")
	t.Setenv("JWT_AUDIENCE", "audience-test")
	t.Setenv("JWT_ALLOWED_ALGS", "RS256,RS512")
	t.Setenv("JWT_JWKS_URL", "https://auth.example.test/.well-known/jwks.json")
	t.Setenv("JWT_JWKS_CACHE_TTL", "2m")
	t.Setenv("JWT_JWKS_FETCH_TIMEOUT", "750ms")
	t.Setenv("JWT_CLOCK_SKEW", "15s")
	t.Setenv("WEBHOOK_SIGNATURE_HEADER", "X-Test-Signature")
	setGRPCTargetEnv(t)

	cfg, err := Load(context.Background())
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.JWTIssuer != "issuer-test" || cfg.JWTAudience != "audience-test" {
		t.Fatalf("unexpected issuer/audience: %q %q", cfg.JWTIssuer, cfg.JWTAudience)
	}
	if strings.Join(cfg.JWTAllowedAlgs, ",") != "RS256,RS512" {
		t.Fatalf("unexpected allowed algorithms: %v", cfg.JWTAllowedAlgs)
	}
	if cfg.JWTJWKSURL != "https://auth.example.test/.well-known/jwks.json" {
		t.Fatalf("unexpected jwks url: %q", cfg.JWTJWKSURL)
	}
	if cfg.JWTJWKSCacheTTL != 2*time.Minute || cfg.JWTJWKSFetchTimeout != 750*time.Millisecond || cfg.JWTClockSkew != 15*time.Second {
		t.Fatalf("unexpected jwt durations: ttl=%s fetch=%s skew=%s", cfg.JWTJWKSCacheTTL, cfg.JWTJWKSFetchTimeout, cfg.JWTClockSkew)
	}
	if cfg.WebhookSignatureHeader != "X-Test-Signature" {
		t.Fatalf("unexpected webhook signature header: %q", cfg.WebhookSignatureHeader)
	}
}

func TestConfigLoadReadsRateLimitSettings(t *testing.T) {
	contractPath := writeTestContract(t)
	t.Setenv("API_CONTRACT_PATH", contractPath)
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("REDIS_PASSWORD", "secret")
	t.Setenv("REDIS_DB", "2")
	t.Setenv("REDIS_TLS_ENABLED", "true")
	t.Setenv("REDIS_DIAL_TIMEOUT", "2s")
	t.Setenv("RATE_LIMIT_ENABLED", "true")
	t.Setenv("RATE_LIMIT_KEY_PREFIX", "rl:test")
	t.Setenv("RATE_LIMIT_FAIL_OPEN", "true")
	t.Setenv("RATE_LIMIT_DEFAULT_IP_LIMIT", "42")
	t.Setenv("RATE_LIMIT_DEFAULT_IP_WINDOW", "30s")
	t.Setenv("RATE_LIMIT_TRUSTED_PROXY_CIDRS", "10.0.0.0/8,127.0.0.1/32")
	t.Setenv("RATE_LIMIT_TARGET_BODY_LIMIT_BYTES", "2048")
	setGRPCTargetEnv(t)

	cfg, err := Load(context.Background())
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Redis.Addr != "redis:6379" || cfg.Redis.Password != "secret" || cfg.Redis.DB != 2 || !cfg.Redis.TLSEnabled || cfg.Redis.DialTimeout != 2*time.Second {
		t.Fatalf("unexpected redis config: %+v", cfg.Redis)
	}
	if !cfg.RateLimit.Enabled || cfg.RateLimit.KeyPrefix != "rl:test" || !cfg.RateLimit.FailOpen {
		t.Fatalf("unexpected rate limit toggles: %+v", cfg.RateLimit)
	}
	if cfg.RateLimit.DefaultIPLimit != 42 || cfg.RateLimit.DefaultIPWindow != 30*time.Second {
		t.Fatalf("unexpected default IP policy config: %+v", cfg.RateLimit)
	}
	if strings.Join(cfg.RateLimit.TrustedProxyCIDRs, ",") != "10.0.0.0/8,127.0.0.1/32" {
		t.Fatalf("unexpected trusted proxies: %v", cfg.RateLimit.TrustedProxyCIDRs)
	}
	if cfg.RateLimit.TargetBodyLimitBytes != 2048 {
		t.Fatalf("unexpected target body limit: %d", cfg.RateLimit.TargetBodyLimitBytes)
	}
}

func TestConfigLoadReadsRequestValidationSettings(t *testing.T) {
	contractPath := writeTestContract(t)
	t.Setenv("API_CONTRACT_PATH", contractPath)
	t.Setenv("REQUEST_VALIDATION_ENABLED", "true")
	t.Setenv("REQUEST_VALIDATION_DEFAULT_MAX_BODY_BYTES", "4096")
	t.Setenv("REQUEST_VALIDATION_MAX_HEADER_BYTES", "8192")
	t.Setenv("REQUEST_VALIDATION_MAX_QUERY_BYTES", "1024")
	setGRPCTargetEnv(t)

	cfg, err := Load(context.Background())
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.Validation.Enabled {
		t.Fatal("expected request validation to be enabled")
	}
	if cfg.Validation.DefaultMaxBodyBytes != 4096 || cfg.Validation.MaxHeaderBytes != 8192 || cfg.Validation.MaxQueryBytes != 1024 {
		t.Fatalf("unexpected validation config: %+v", cfg.Validation)
	}
}

func TestConfigLoadReadsObservabilitySettings(t *testing.T) {
	contractPath := writeTestContract(t)
	t.Setenv("API_CONTRACT_PATH", contractPath)
	t.Setenv("METRICS_ENABLED", "true")
	t.Setenv("METRICS_ADDR", ":19090")
	t.Setenv("TRACE_ENABLED", "true")
	t.Setenv("TRACE_EXPORTER_OTLP_ENDPOINT", "otel.test:4317")
	t.Setenv("TRACE_EXPORTER_OTLP_INSECURE", "false")
	t.Setenv("TRACE_SAMPLE_RATIO", "0.25")
	t.Setenv("TRACE_SHUTDOWN_TIMEOUT", "2s")
	t.Setenv("OBSERVABILITY_HASH_SALT", "test-salt")
	setGRPCTargetEnv(t)

	cfg, err := Load(context.Background())
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.Observability.MetricsEnabled || cfg.Observability.MetricsAddress != ":19090" {
		t.Fatalf("unexpected metrics config: %+v", cfg.Observability)
	}
	if !cfg.Observability.TracingEnabled || cfg.Observability.TraceOTLPEndpoint != "otel.test:4317" || cfg.Observability.TraceOTLPInsecure {
		t.Fatalf("unexpected tracing config: %+v", cfg.Observability)
	}
	if cfg.Observability.TraceSampleRatio != 0.25 || cfg.Observability.TraceShutdownTimeout != 2*time.Second {
		t.Fatalf("unexpected trace tuning: %+v", cfg.Observability)
	}
	if cfg.Observability.UserHashSalt != "test-salt" {
		t.Fatalf("unexpected hash salt: %q", cfg.Observability.UserHashSalt)
	}
}

func TestConfigValidateRejectsInvalidTrustedProxyCIDR(t *testing.T) {
	cfg := validConfig(t)
	cfg.RateLimit.Enabled = true
	cfg.Redis.Addr = "redis:6379"
	cfg.Redis.DialTimeout = time.Second
	cfg.RateLimit.KeyPrefix = "rl:test"
	cfg.RateLimit.DefaultIPLimit = 10
	cfg.RateLimit.DefaultIPWindow = time.Minute
	cfg.RateLimit.TargetBodyLimitBytes = 1024
	cfg.RateLimit.TrustedProxyCIDRs = []string{"not-a-cidr"}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected invalid CIDR to fail validation")
	}
	if !strings.Contains(err.Error(), "RATE_LIMIT_TRUSTED_PROXY_CIDRS") {
		t.Fatalf("expected trusted proxy validation error, got %v", err)
	}
}

func TestConfigValidateRejectsInvalidRequestValidationLimits(t *testing.T) {
	cfg := validConfig(t)
	cfg.Validation.Enabled = true
	cfg.Validation.DefaultMaxBodyBytes = 0
	cfg.Validation.MaxHeaderBytes = 1024
	cfg.Validation.MaxQueryBytes = 1024

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected invalid validation limit to fail validation")
	}
	if !strings.Contains(err.Error(), "REQUEST_VALIDATION_DEFAULT_MAX_BODY_BYTES") {
		t.Fatalf("expected request validation limit error, got %v", err)
	}
}

func validConfig(t *testing.T) Config {
	t.Helper()
	return Config{
		ServiceName:            "api-gateway",
		Environment:            "test",
		HTTPAddress:            ":0",
		APIBasePath:            "/api/v1",
		APIContractPath:        writeTestContract(t),
		LogLevel:               "error",
		ReadHeaderTimeout:      time.Second,
		ShutdownTimeout:        time.Second,
		GRPCDialTimeout:        time.Second,
		JWTIssuer:              "ecommerce-auth",
		JWTAudience:            "ecommerce-api",
		JWTAllowedAlgs:         []string{"RS256"},
		JWTJWKSCacheTTL:        time.Minute,
		JWTJWKSFetchTimeout:    time.Second,
		JWTClockSkew:           30 * time.Second,
		WebhookSignatureHeader: "X-Provider-Signature",
		AuthGRPCAddr:           "auth-service:9090",
		UserGRPCAddr:           "user-service:9090",
		ProductGRPCAddr:        "product-service:9090",
		CartGRPCAddr:           "cart-service:9090",
		WishlistGRPCAddr:       "wishlist-service:9090",
		OrderGRPCAddr:          "order-service:9090",
		PaymentGRPCAddr:        "payment-service:9090",
		SearchGRPCAddr:         "search-service:9090",
		CMSGRPCAddr:            "cms-service:9090",
		SessionGRPCAddr:        "session-service:9090",
		NotificationGRPCAddr:   "notification-service:9090",
		SuperadminGRPCAddr:     "superadmin-service:9090",
	}
}

func writeTestContract(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "master-api.json")
	if err := os.WriteFile(path, []byte(`{"project":"test","version":"1.0.0","rest_endpoints":[]}`), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	return path
}

func setGRPCTargetEnv(t *testing.T) {
	t.Helper()
	targets := map[string]string{
		"AUTH_GRPC_ADDR":         "auth-service:9090",
		"USER_GRPC_ADDR":         "user-service:9090",
		"PRODUCT_GRPC_ADDR":      "product-service:9090",
		"CART_GRPC_ADDR":         "cart-service:9090",
		"WISHLIST_GRPC_ADDR":     "wishlist-service:9090",
		"ORDER_GRPC_ADDR":        "order-service:9090",
		"PAYMENT_GRPC_ADDR":      "payment-service:9090",
		"SEARCH_GRPC_ADDR":       "search-service:9090",
		"CMS_GRPC_ADDR":          "cms-service:9090",
		"SESSION_GRPC_ADDR":      "session-service:9090",
		"NOTIFICATION_GRPC_ADDR": "notification-service:9090",
		"SUPERADMIN_GRPC_ADDR":   "superadmin-service:9090",
	}
	for key, value := range targets {
		t.Setenv(key, value)
	}
}
