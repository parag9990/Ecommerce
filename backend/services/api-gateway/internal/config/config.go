package config

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ecommerce/api-gateway/internal/observability"
)

type RedisConfig struct {
	Addr        string
	Password    string
	DB          int
	TLSEnabled  bool
	DialTimeout time.Duration
}

type RateLimitConfig struct {
	Enabled              bool
	KeyPrefix            string
	FailOpen             bool
	DefaultIPLimit       int64
	DefaultIPWindow      time.Duration
	TrustedProxyCIDRs    []string
	TargetBodyLimitBytes int64
}

type RequestValidationConfig struct {
	Enabled             bool
	DefaultMaxBodyBytes int64
	MaxHeaderBytes      int64
	MaxQueryBytes       int
}

type GRPCWebConfig struct {
	Enabled            bool
	Address            string
	PolicyPath         string
	ExposedServices    []string
	MaxReceiveMsgBytes int
	MaxSendMsgBytes    int
}

type Config struct {
	ServiceName       string
	Environment       string
	HTTPAddress       string
	APIBasePath       string
	APIContractPath   string
	LogLevel          string
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration

	GRPCTLSEnabled                  bool
	GRPCDialTimeout                 time.Duration
	GRPCAllowUnavailableDownstreams bool

	Redis         RedisConfig
	RateLimit     RateLimitConfig
	Validation    RequestValidationConfig
	GRPCWeb       GRPCWebConfig
	Observability observability.Config

	JWTIssuer           string
	JWTAudience         string
	JWTAllowedAlgs      []string
	JWTJWKSURL          string
	JWTJWKSCacheTTL     time.Duration
	JWTJWKSFetchTimeout time.Duration
	JWTClockSkew        time.Duration

	WebhookSignatureHeader string

	AuthGRPCAddr          string
	AuthHTTPURL           string
	AuthHTTPTimeout       time.Duration
	UserGRPCAddr          string
	ProductGRPCAddr       string
	CartGRPCAddr          string
	WishlistGRPCAddr      string
	WishlistHTTPURL       string
	WishlistHTTPTimeout   time.Duration
	OrderGRPCAddr         string
	PaymentGRPCAddr       string
	SearchGRPCAddr        string
	CMSGRPCAddr           string
	SessionGRPCAddr       string
	SessionHTTPURL        string
	SessionHTTPTimeout    time.Duration
	NotificationGRPCAddr  string
	SuperadminGRPCAddr    string
	SuperadminHTTPURL     string
	SuperadminHTTPTimeout time.Duration
}

func Load(ctx context.Context) (Config, error) {
	if err := ctx.Err(); err != nil {
		return Config{}, err
	}

	contractPath, err := resolveContractPath(getenv("API_CONTRACT_PATH", ""))
	if err != nil {
		return Config{}, err
	}
	readHeaderTimeout, err := getDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := getDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	grpcDialTimeout, err := getDuration("GRPC_DIAL_TIMEOUT", 3*time.Second)
	if err != nil {
		return Config{}, err
	}
	sessionHTTPTimeout, err := getDuration("SESSION_HTTP_TIMEOUT", 30*time.Second)
	if err != nil {
		return Config{}, err
	}
	superadminHTTPTimeout, err := getDuration("SUPERADMIN_HTTP_TIMEOUT", 30*time.Second)
	if err != nil {
		return Config{}, err
	}
	authHTTPTimeout, err := getDuration("AUTH_HTTP_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	wishlistHTTPTimeout, err := getDuration("WISHLIST_HTTP_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	grpcTLSEnabled, err := getBool("GRPC_TLS_ENABLED", false)
	if err != nil {
		return Config{}, err
	}
	grpcAllowUnavailableDownstreams, err := getBool("GRPC_ALLOW_UNAVAILABLE_DOWNSTREAMS", false)
	if err != nil {
		return Config{}, err
	}
	grpcWebEnabled, err := getBool("GRPC_WEB_ENABLED", false)
	if err != nil {
		return Config{}, err
	}
	grpcWebPolicyPath, err := resolveGRPCWebPolicyPath(getenv("GRPC_WEB_POLICY_PATH", ""))
	if err != nil {
		return Config{}, err
	}
	grpcWebMaxReceiveMsgBytes, err := getInt("GRPC_WEB_MAX_RECEIVE_MESSAGE_BYTES", 4*1024*1024)
	if err != nil {
		return Config{}, err
	}
	grpcWebMaxSendMsgBytes, err := getInt("GRPC_WEB_MAX_SEND_MESSAGE_BYTES", 4*1024*1024)
	if err != nil {
		return Config{}, err
	}
	jwksCacheTTL, err := getDuration("JWT_JWKS_CACHE_TTL", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}
	jwksFetchTimeout, err := getDuration("JWT_JWKS_FETCH_TIMEOUT", 3*time.Second)
	if err != nil {
		return Config{}, err
	}
	jwtClockSkew, err := getDuration("JWT_CLOCK_SKEW", 30*time.Second)
	if err != nil {
		return Config{}, err
	}
	redisDB, err := getInt("REDIS_DB", 0)
	if err != nil {
		return Config{}, err
	}
	redisTLSEnabled, err := getBool("REDIS_TLS_ENABLED", false)
	if err != nil {
		return Config{}, err
	}
	redisDialTimeout, err := getDuration("REDIS_DIAL_TIMEOUT", 3*time.Second)
	if err != nil {
		return Config{}, err
	}
	rateLimitEnabled, err := getBool("RATE_LIMIT_ENABLED", true)
	if err != nil {
		return Config{}, err
	}
	rateLimitFailOpen, err := getBool("RATE_LIMIT_FAIL_OPEN", false)
	if err != nil {
		return Config{}, err
	}
	defaultIPLimit, err := getInt64("RATE_LIMIT_DEFAULT_IP_LIMIT", 600)
	if err != nil {
		return Config{}, err
	}
	defaultIPWindow, err := getDuration("RATE_LIMIT_DEFAULT_IP_WINDOW", time.Minute)
	if err != nil {
		return Config{}, err
	}
	targetBodyLimitBytes, err := getInt64("RATE_LIMIT_TARGET_BODY_LIMIT_BYTES", 64*1024)
	if err != nil {
		return Config{}, err
	}
	validationEnabled, err := getBool("REQUEST_VALIDATION_ENABLED", true)
	if err != nil {
		return Config{}, err
	}
	defaultMaxBodyBytes, err := getInt64("REQUEST_VALIDATION_DEFAULT_MAX_BODY_BYTES", 128*1024)
	if err != nil {
		return Config{}, err
	}
	maxHeaderBytes, err := getInt64("REQUEST_VALIDATION_MAX_HEADER_BYTES", 32*1024)
	if err != nil {
		return Config{}, err
	}
	maxQueryBytes, err := getInt("REQUEST_VALIDATION_MAX_QUERY_BYTES", 8*1024)
	if err != nil {
		return Config{}, err
	}
	metricsEnabled, err := getBool("METRICS_ENABLED", true)
	if err != nil {
		return Config{}, err
	}
	tracingEnabled, err := getBool("TRACE_ENABLED", true)
	if err != nil {
		return Config{}, err
	}
	traceOTLPInsecure, err := getBool("TRACE_EXPORTER_OTLP_INSECURE", true)
	if err != nil {
		return Config{}, err
	}
	traceSampleRatio, err := getFloat64("TRACE_SAMPLE_RATIO", 1.0)
	if err != nil {
		return Config{}, err
	}
	traceShutdownTimeout, err := getDuration("TRACE_SHUTDOWN_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}

	serviceName := getenv("SERVICE_NAME", "api-gateway")
	environment := getenv("APP_ENV", "local")
	observabilityConfig := observability.DefaultConfig(serviceName, environment)
	observabilityConfig.RequestIDHeader = getenv("REQUEST_ID_HEADER", observability.HeaderRequestID)
	observabilityConfig.MetricsEnabled = metricsEnabled
	observabilityConfig.MetricsAddress = metricsAddress()
	observabilityConfig.TracingEnabled = tracingEnabled
	observabilityConfig.TraceOTLPEndpoint = getenv("TRACE_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317")
	observabilityConfig.TraceOTLPInsecure = traceOTLPInsecure
	observabilityConfig.TraceSampleRatio = traceSampleRatio
	observabilityConfig.TraceShutdownTimeout = traceShutdownTimeout
	observabilityConfig.UserHashSalt = getenv("OBSERVABILITY_HASH_SALT", "")

	cfg := Config{
		ServiceName:                     serviceName,
		Environment:                     environment,
		HTTPAddress:                     getenv("HTTP_ADDR", ":8080"),
		APIBasePath:                     getenv("API_BASE_PATH", "/api/v1"),
		APIContractPath:                 contractPath,
		LogLevel:                        getenv("LOG_LEVEL", "info"),
		ReadHeaderTimeout:               readHeaderTimeout,
		ShutdownTimeout:                 shutdownTimeout,
		GRPCTLSEnabled:                  grpcTLSEnabled,
		GRPCDialTimeout:                 grpcDialTimeout,
		GRPCAllowUnavailableDownstreams: grpcAllowUnavailableDownstreams,
		Redis: RedisConfig{
			Addr:        getenv("REDIS_ADDR", "localhost:6379"),
			Password:    getenv("REDIS_PASSWORD", ""),
			DB:          redisDB,
			TLSEnabled:  redisTLSEnabled,
			DialTimeout: redisDialTimeout,
		},
		RateLimit: RateLimitConfig{
			Enabled:              rateLimitEnabled,
			KeyPrefix:            getenv("RATE_LIMIT_KEY_PREFIX", "rl:v1"),
			FailOpen:             rateLimitFailOpen,
			DefaultIPLimit:       defaultIPLimit,
			DefaultIPWindow:      defaultIPWindow,
			TrustedProxyCIDRs:    getCSV("RATE_LIMIT_TRUSTED_PROXY_CIDRS", nil),
			TargetBodyLimitBytes: targetBodyLimitBytes,
		},
		Validation: RequestValidationConfig{
			Enabled:             validationEnabled,
			DefaultMaxBodyBytes: defaultMaxBodyBytes,
			MaxHeaderBytes:      maxHeaderBytes,
			MaxQueryBytes:       maxQueryBytes,
		},
		GRPCWeb: GRPCWebConfig{
			Enabled:            grpcWebEnabled,
			Address:            getenv("GRPC_ADDR", ":9090"),
			PolicyPath:         grpcWebPolicyPath,
			ExposedServices:    getCSV("GRPC_WEB_EXPOSED_SERVICES", nil),
			MaxReceiveMsgBytes: grpcWebMaxReceiveMsgBytes,
			MaxSendMsgBytes:    grpcWebMaxSendMsgBytes,
		},
		Observability:          observabilityConfig,
		JWTIssuer:              getenv("JWT_ISSUER", "ecommerce-auth"),
		JWTAudience:            getenv("JWT_AUDIENCE", "ecommerce-api"),
		JWTAllowedAlgs:         getCSV("JWT_ALLOWED_ALGS", []string{"RS256"}),
		JWTJWKSURL:             getenv("JWT_JWKS_URL", ""),
		JWTJWKSCacheTTL:        jwksCacheTTL,
		JWTJWKSFetchTimeout:    jwksFetchTimeout,
		JWTClockSkew:           jwtClockSkew,
		WebhookSignatureHeader: getenv("WEBHOOK_SIGNATURE_HEADER", "X-Provider-Signature"),
		AuthGRPCAddr:           getenv("AUTH_GRPC_ADDR", ""),
		AuthHTTPURL:            getenv("AUTH_HTTP_URL", "http://auth-service:8081"),
		AuthHTTPTimeout:        authHTTPTimeout,
		UserGRPCAddr:           getenv("USER_GRPC_ADDR", ""),
		ProductGRPCAddr:        getenv("PRODUCT_GRPC_ADDR", ""),
		CartGRPCAddr:           getenv("CART_GRPC_ADDR", ""),
		WishlistGRPCAddr:       getenv("WISHLIST_GRPC_ADDR", ""),
		WishlistHTTPURL:        getenv("WISHLIST_HTTP_URL", "http://wishlist-service:8084"),
		WishlistHTTPTimeout:    wishlistHTTPTimeout,
		OrderGRPCAddr:          getenv("ORDER_GRPC_ADDR", ""),
		PaymentGRPCAddr:        getenv("PAYMENT_GRPC_ADDR", ""),
		SearchGRPCAddr:         getenv("SEARCH_GRPC_ADDR", ""),
		CMSGRPCAddr:            getenv("CMS_GRPC_ADDR", ""),
		SessionGRPCAddr:        getenv("SESSION_GRPC_ADDR", ""),
		SessionHTTPURL:         getenv("SESSION_HTTP_URL", "http://session-service:8086"),
		SessionHTTPTimeout:     sessionHTTPTimeout,
		NotificationGRPCAddr:   getenv("NOTIFICATION_GRPC_ADDR", ""),
		SuperadminGRPCAddr:     getenv("SUPERADMIN_GRPC_ADDR", ""),
		SuperadminHTTPURL:      getenv("SUPERADMIN_HTTP_URL", "http://superadmin-service:8088"),
		SuperadminHTTPTimeout:  superadminHTTPTimeout,
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var errs []error
	if strings.TrimSpace(c.ServiceName) == "" {
		errs = append(errs, errors.New("SERVICE_NAME is required"))
	}
	if strings.TrimSpace(c.HTTPAddress) == "" {
		errs = append(errs, errors.New("HTTP_ADDR is required"))
	}
	if !strings.HasPrefix(c.APIBasePath, "/") {
		errs = append(errs, errors.New("API_BASE_PATH must start with /"))
	}
	if c.APIBasePath != "/" && strings.HasSuffix(c.APIBasePath, "/") {
		errs = append(errs, errors.New("API_BASE_PATH must not end with /"))
	}
	if strings.TrimSpace(c.APIContractPath) == "" {
		errs = append(errs, errors.New("API_CONTRACT_PATH is required"))
	} else if stat, err := os.Stat(c.APIContractPath); err != nil {
		errs = append(errs, fmt.Errorf("API_CONTRACT_PATH is not readable: %w", err))
	} else if stat.IsDir() {
		errs = append(errs, errors.New("API_CONTRACT_PATH must be a file"))
	}
	if c.ReadHeaderTimeout <= 0 {
		errs = append(errs, errors.New("HTTP_READ_HEADER_TIMEOUT must be positive"))
	}
	if c.ShutdownTimeout <= 0 {
		errs = append(errs, errors.New("HTTP_SHUTDOWN_TIMEOUT must be positive"))
	}
	if c.GRPCDialTimeout <= 0 {
		errs = append(errs, errors.New("GRPC_DIAL_TIMEOUT must be positive"))
	}
	if c.RateLimit.Enabled {
		if strings.TrimSpace(c.Redis.Addr) == "" {
			errs = append(errs, errors.New("REDIS_ADDR is required when rate limiting is enabled"))
		}
		if strings.ContainsAny(c.Redis.Addr, " \t\r\n") {
			errs = append(errs, errors.New("REDIS_ADDR must not contain whitespace"))
		}
		if c.Redis.DB < 0 {
			errs = append(errs, errors.New("REDIS_DB must not be negative"))
		}
		if c.Redis.DialTimeout <= 0 {
			errs = append(errs, errors.New("REDIS_DIAL_TIMEOUT must be positive"))
		}
		if strings.TrimSpace(c.RateLimit.KeyPrefix) == "" {
			errs = append(errs, errors.New("RATE_LIMIT_KEY_PREFIX is required when rate limiting is enabled"))
		}
		if strings.ContainsAny(c.RateLimit.KeyPrefix, " \t\r\n") {
			errs = append(errs, errors.New("RATE_LIMIT_KEY_PREFIX must not contain whitespace"))
		}
		if c.RateLimit.DefaultIPLimit <= 0 {
			errs = append(errs, errors.New("RATE_LIMIT_DEFAULT_IP_LIMIT must be positive"))
		}
		if c.RateLimit.DefaultIPWindow <= 0 {
			errs = append(errs, errors.New("RATE_LIMIT_DEFAULT_IP_WINDOW must be positive"))
		}
		if c.RateLimit.TargetBodyLimitBytes <= 0 {
			errs = append(errs, errors.New("RATE_LIMIT_TARGET_BODY_LIMIT_BYTES must be positive"))
		}
		for _, cidr := range c.RateLimit.TrustedProxyCIDRs {
			if _, err := netip.ParsePrefix(cidr); err != nil {
				errs = append(errs, fmt.Errorf("RATE_LIMIT_TRUSTED_PROXY_CIDRS contains invalid CIDR %q: %w", cidr, err))
			}
		}
	}
	if c.Validation.Enabled {
		if c.Validation.DefaultMaxBodyBytes <= 0 {
			errs = append(errs, errors.New("REQUEST_VALIDATION_DEFAULT_MAX_BODY_BYTES must be positive"))
		}
		if c.Validation.MaxHeaderBytes <= 0 {
			errs = append(errs, errors.New("REQUEST_VALIDATION_MAX_HEADER_BYTES must be positive"))
		}
		if c.Validation.MaxQueryBytes <= 0 {
			errs = append(errs, errors.New("REQUEST_VALIDATION_MAX_QUERY_BYTES must be positive"))
		}
	}
	if c.GRPCWeb.Enabled {
		if strings.TrimSpace(c.GRPCWeb.Address) == "" {
			errs = append(errs, errors.New("GRPC_ADDR is required when gRPC-Web is enabled"))
		} else if strings.ContainsAny(c.GRPCWeb.Address, " \t\r\n") {
			errs = append(errs, errors.New("GRPC_ADDR must not contain whitespace"))
		}
		if strings.TrimSpace(c.GRPCWeb.PolicyPath) == "" {
			errs = append(errs, errors.New("GRPC_WEB_POLICY_PATH is required when gRPC-Web is enabled"))
		} else if stat, err := os.Stat(c.GRPCWeb.PolicyPath); err != nil {
			errs = append(errs, fmt.Errorf("GRPC_WEB_POLICY_PATH is not readable: %w", err))
		} else if stat.IsDir() {
			errs = append(errs, errors.New("GRPC_WEB_POLICY_PATH must be a file"))
		}
		if len(c.GRPCWeb.ExposedServices) == 0 {
			errs = append(errs, errors.New("GRPC_WEB_EXPOSED_SERVICES must contain at least one service when gRPC-Web is enabled"))
		}
		for _, service := range c.GRPCWeb.ExposedServices {
			if strings.TrimSpace(service) == "" || strings.ContainsAny(service, " \t\r\n/") {
				errs = append(errs, fmt.Errorf("GRPC_WEB_EXPOSED_SERVICES contains invalid service %q", service))
			}
		}
		if c.GRPCWeb.MaxReceiveMsgBytes <= 0 {
			errs = append(errs, errors.New("GRPC_WEB_MAX_RECEIVE_MESSAGE_BYTES must be positive"))
		}
		if c.GRPCWeb.MaxSendMsgBytes <= 0 {
			errs = append(errs, errors.New("GRPC_WEB_MAX_SEND_MESSAGE_BYTES must be positive"))
		}
		observabilityConfig := c.Observability.Normalize(c.ServiceName, c.Environment)
		if observabilityConfig.MetricsEnabled && strings.TrimSpace(c.GRPCWeb.Address) == strings.TrimSpace(observabilityConfig.MetricsAddress) {
			errs = append(errs, errors.New("GRPC_ADDR must not equal METRICS_ADDR"))
		}
		if strings.TrimSpace(c.GRPCWeb.Address) == strings.TrimSpace(c.HTTPAddress) {
			errs = append(errs, errors.New("GRPC_ADDR must not equal HTTP_ADDR"))
		}
	}
	observabilityConfig := c.Observability.Normalize(c.ServiceName, c.Environment)
	if err := observabilityConfig.Validate(); err != nil {
		errs = append(errs, err)
	}
	if strings.TrimSpace(c.JWTIssuer) == "" {
		errs = append(errs, errors.New("JWT_ISSUER is required"))
	}
	if strings.TrimSpace(c.JWTAudience) == "" {
		errs = append(errs, errors.New("JWT_AUDIENCE is required"))
	}
	if len(c.JWTAllowedAlgs) == 0 {
		errs = append(errs, errors.New("JWT_ALLOWED_ALGS must contain at least one algorithm"))
	}
	for _, alg := range c.JWTAllowedAlgs {
		if !isSupportedJWTAlgorithm(alg) {
			errs = append(errs, fmt.Errorf("JWT_ALLOWED_ALGS contains unsupported algorithm %q", alg))
		}
	}
	if strings.TrimSpace(c.JWTJWKSURL) != "" {
		if err := validateHTTPURL(c.JWTJWKSURL); err != nil {
			errs = append(errs, fmt.Errorf("JWT_JWKS_URL is invalid: %w", err))
		}
	}
	if strings.TrimSpace(c.SessionHTTPURL) != "" {
		if err := validateHTTPURL(c.SessionHTTPURL); err != nil {
			errs = append(errs, fmt.Errorf("SESSION_HTTP_URL is invalid: %w", err))
		}
	}
	if c.SessionHTTPTimeout <= 0 {
		errs = append(errs, errors.New("SESSION_HTTP_TIMEOUT must be positive"))
	}
	if strings.TrimSpace(c.SuperadminHTTPURL) != "" {
		if err := validateHTTPURL(c.SuperadminHTTPURL); err != nil {
			errs = append(errs, fmt.Errorf("SUPERADMIN_HTTP_URL is invalid: %w", err))
		}
	}
	if c.SuperadminHTTPTimeout <= 0 {
		errs = append(errs, errors.New("SUPERADMIN_HTTP_TIMEOUT must be positive"))
	}
	if strings.TrimSpace(c.AuthHTTPURL) != "" {
		if err := validateHTTPURL(c.AuthHTTPURL); err != nil {
			errs = append(errs, fmt.Errorf("AUTH_HTTP_URL is invalid: %w", err))
		}
	}
	if c.AuthHTTPTimeout <= 0 {
		errs = append(errs, errors.New("AUTH_HTTP_TIMEOUT must be positive"))
	}
	if strings.TrimSpace(c.WishlistHTTPURL) != "" {
		if err := validateHTTPURL(c.WishlistHTTPURL); err != nil {
			errs = append(errs, fmt.Errorf("WISHLIST_HTTP_URL is invalid: %w", err))
		}
		if c.WishlistHTTPTimeout <= 0 {
			errs = append(errs, errors.New("WISHLIST_HTTP_TIMEOUT must be positive"))
		}
	}
	if c.JWTJWKSCacheTTL <= 0 {
		errs = append(errs, errors.New("JWT_JWKS_CACHE_TTL must be positive"))
	}
	if c.JWTJWKSFetchTimeout <= 0 {
		errs = append(errs, errors.New("JWT_JWKS_FETCH_TIMEOUT must be positive"))
	}
	if c.JWTClockSkew < 0 {
		errs = append(errs, errors.New("JWT_CLOCK_SKEW must not be negative"))
	}
	if c.JWTClockSkew > 5*time.Minute {
		errs = append(errs, errors.New("JWT_CLOCK_SKEW must not exceed 5m"))
	}
	if strings.TrimSpace(c.WebhookSignatureHeader) == "" {
		errs = append(errs, errors.New("WEBHOOK_SIGNATURE_HEADER is required"))
	} else if strings.ContainsAny(c.WebhookSignatureHeader, " \t\r\n") {
		errs = append(errs, errors.New("WEBHOOK_SIGNATURE_HEADER must not contain whitespace"))
	} else if http.CanonicalHeaderKey(c.WebhookSignatureHeader) == "" {
		errs = append(errs, errors.New("WEBHOOK_SIGNATURE_HEADER must be a valid header name"))
	}
	requiredTargets := map[string]string{
		"AUTH_GRPC_ADDR":         c.AuthGRPCAddr,
		"USER_GRPC_ADDR":         c.UserGRPCAddr,
		"PRODUCT_GRPC_ADDR":      c.ProductGRPCAddr,
		"CART_GRPC_ADDR":         c.CartGRPCAddr,
		"ORDER_GRPC_ADDR":        c.OrderGRPCAddr,
		"PAYMENT_GRPC_ADDR":      c.PaymentGRPCAddr,
		"SEARCH_GRPC_ADDR":       c.SearchGRPCAddr,
		"CMS_GRPC_ADDR":          c.CMSGRPCAddr,
		"SESSION_GRPC_ADDR":      c.SessionGRPCAddr,
		"NOTIFICATION_GRPC_ADDR": c.NotificationGRPCAddr,
		"SUPERADMIN_GRPC_ADDR":   c.SuperadminGRPCAddr,
	}
	for name, target := range requiredTargets {
		if strings.TrimSpace(target) == "" {
			errs = append(errs, fmt.Errorf("%s is required", name))
		} else if strings.ContainsAny(target, " \t\r\n") {
			errs = append(errs, fmt.Errorf("%s must not contain whitespace", name))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}
	return parsed, nil
}

func getBool(key string, fallback bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	switch strings.ToLower(value) {
	case "1", "t", "true", "y", "yes", "on":
		return true, nil
	case "0", "f", "false", "n", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("%s must be a valid boolean", key)
	}
}

func getInt(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}
	return parsed, nil
}

func getInt64(key string, fallback int64) (int64, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}
	return parsed, nil
}

func getFloat64(key string, fallback float64) (float64, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid decimal: %w", key, err)
	}
	return parsed, nil
}

func getCSV(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		out := make([]string, len(fallback))
		copy(out, fallback)
		return out
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func metricsAddress() string {
	if addr := getenv("METRICS_ADDR", ""); addr != "" {
		return addr
	}
	port := getenv("METRICS_PORT", "9090")
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}

func isSupportedJWTAlgorithm(algorithm string) bool {
	switch strings.TrimSpace(algorithm) {
	case "RS256", "RS384", "RS512":
		return true
	default:
		return false
	}
}

func validateHTTPURL(value string) error {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil {
		return err
	}
	switch parsed.Scheme {
	case "http", "https":
		return nil
	default:
		return fmt.Errorf("scheme must be http or https")
	}
}

func resolveContractPath(configured string) (string, error) {
	if configured != "" {
		return filepath.Abs(configured)
	}
	candidates := []string{
		"api/master-api.json",
		"../api/master-api.json",
		"../../api/master-api.json",
		"../../../api/master-api.json",
		"../../../../api/master-api.json",
		"../../../../../api/master-api.json",
	}
	for _, candidate := range candidates {
		if stat, err := os.Stat(candidate); err == nil && !stat.IsDir() {
			return filepath.Abs(candidate)
		}
	}
	return "", errors.New("API_CONTRACT_PATH is required when api/master-api.json cannot be discovered")
}

func resolveGRPCWebPolicyPath(configured string) (string, error) {
	if configured != "" {
		return filepath.Abs(configured)
	}
	candidates := []string{
		"config/grpcweb-policies.json",
		"backend/services/api-gateway/config/grpcweb-policies.json",
		"services/api-gateway/config/grpcweb-policies.json",
		"../config/grpcweb-policies.json",
		"../../config/grpcweb-policies.json",
		"../../../config/grpcweb-policies.json",
		"../../../../backend/services/api-gateway/config/grpcweb-policies.json",
	}
	for _, candidate := range candidates {
		if stat, err := os.Stat(candidate); err == nil && !stat.IsDir() {
			return filepath.Abs(candidate)
		}
	}
	return "", nil
}
