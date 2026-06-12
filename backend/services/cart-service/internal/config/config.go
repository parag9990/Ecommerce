package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPAddress     = ":8084"
	defaultMongoDatabase   = "cart_db"
	defaultShutdownTimeout = 10 * time.Second
	defaultRedisAddress    = "localhost:6379"
	defaultCMSValidatePath = "/internal/v1/coupons/validate"
)

type Config struct {
	ServiceName        string
	Environment        string
	HTTP               HTTPConfig
	Mongo              MongoConfig
	Redis              RedisConfig
	Product            ProductConfig
	CMS                CMSConfig
	Cart               CartConfig
	CartExpiry         CartExpiryConfig
	Cache              CacheConfig
	BootstrapOnStartup bool
}

type HTTPConfig struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type MongoConfig struct {
	URI            string
	Database       string
	ConnectTimeout time.Duration
	PingTimeout    time.Duration
}

type RedisConfig struct {
	Address      string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PingTimeout  time.Duration
}

type ProductConfig struct {
	BaseURL        string
	RequestTimeout time.Duration
}

type CMSConfig struct {
	BaseURL            string
	ValidateCouponPath string
	RequestTimeout     time.Duration
}

type CartConfig struct {
	ExpiryTTL                 time.Duration
	UserExpiryTTL             time.Duration
	GuestExpiryTTL            time.Duration
	DefaultCurrency           string
	MaxSaveAttempts           int
	AllowGuestCartIDOnlyMerge bool
}

type CartExpiryConfig struct {
	CleanupInterval   time.Duration
	CleanupBatchSize  int64
	CleanupLockTTL    time.Duration
	CleanupRunTimeout time.Duration
	CleanupLockKey    string
}

type CacheConfig struct {
	ActiveUserTTL  time.Duration
	ActiveGuestTTL time.Duration
	SummaryTTL     time.Duration
}

func Load() (Config, error) {
	legacyExpiryTTL := envDuration("CART_EXPIRY_TTL", 90*24*time.Hour)
	cfg := Config{
		ServiceName: envString("CART_SERVICE_NAME", "cart-service"),
		Environment: envString("ENVIRONMENT", envString("APP_ENV", "local")),
		HTTP: HTTPConfig{
			Address:         envString("CART_HTTP_ADDR", defaultHTTPAddress),
			ReadTimeout:     envDuration("CART_HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:    envDuration("CART_HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     envDuration("CART_HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: envDuration("CART_SHUTDOWN_TIMEOUT", defaultShutdownTimeout),
		},
		Mongo: MongoConfig{
			URI:            envString("CART_MONGO_URI", os.Getenv("MONGO_URI")),
			Database:       envString("CART_MONGO_DATABASE", defaultMongoDatabase),
			ConnectTimeout: envDuration("CART_MONGO_CONNECT_TIMEOUT", 5*time.Second),
			PingTimeout:    envDuration("CART_MONGO_PING_TIMEOUT", 5*time.Second),
		},
		Redis: RedisConfig{
			Address:      envString("CART_REDIS_ADDR", envString("REDIS_ADDR", defaultRedisAddress)),
			Password:     envString("CART_REDIS_PASSWORD", envString("REDIS_PASSWORD", "")),
			DB:           envInt("CART_REDIS_DB", envInt("REDIS_DB", 0)),
			DialTimeout:  envDuration("CART_REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  envDuration("CART_REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: envDuration("CART_REDIS_WRITE_TIMEOUT", 3*time.Second),
			PingTimeout:  envDuration("CART_REDIS_PING_TIMEOUT", 3*time.Second),
		},
		Product: ProductConfig{
			BaseURL:        envString("CART_PRODUCT_BASE_URL", envString("PRODUCT_SERVICE_BASE_URL", "")),
			RequestTimeout: envDuration("CART_PRODUCT_REQUEST_TIMEOUT", 2*time.Second),
		},
		CMS: CMSConfig{
			BaseURL:            envString("CART_CMS_BASE_URL", envString("CMS_SERVICE_BASE_URL", "")),
			ValidateCouponPath: envString("CART_CMS_VALIDATE_COUPON_PATH", defaultCMSValidatePath),
			RequestTimeout:     envDuration("CART_COUPON_PREVIEW_TIMEOUT", envDuration("CART_CMS_REQUEST_TIMEOUT", 700*time.Millisecond)),
		},
		Cart: CartConfig{
			ExpiryTTL:                 legacyExpiryTTL,
			UserExpiryTTL:             envDuration("CART_USER_EXPIRY_TTL", legacyExpiryTTL),
			GuestExpiryTTL:            envDuration("CART_GUEST_EXPIRY_TTL", 30*24*time.Hour),
			DefaultCurrency:           strings.ToUpper(envString("CART_DEFAULT_CURRENCY", "INR")),
			MaxSaveAttempts:           envInt("CART_MAX_SAVE_ATTEMPTS", 3),
			AllowGuestCartIDOnlyMerge: envBool("CART_ALLOW_GUEST_CART_ID_ONLY_MERGE", false),
		},
		CartExpiry: CartExpiryConfig{
			CleanupInterval:   envDuration("CART_CLEANUP_INTERVAL", 15*time.Minute),
			CleanupBatchSize:  int64(envInt("CART_CLEANUP_BATCH_SIZE", 500)),
			CleanupLockTTL:    envDuration("CART_CLEANUP_LOCK_TTL", 10*time.Minute),
			CleanupRunTimeout: envDuration("CART_CLEANUP_RUN_TIMEOUT", 2*time.Minute),
			CleanupLockKey:    envString("CART_CLEANUP_LOCK_KEY", "cart:lock:expiry-cleanup"),
		},
		Cache: CacheConfig{
			ActiveUserTTL:  envDuration("CART_ACTIVE_CACHE_TTL", 15*time.Minute),
			ActiveGuestTTL: envDuration("CART_GUEST_ACTIVE_CACHE_TTL", envDuration("GUEST_CART_CACHE_TTL", 30*time.Minute)),
			SummaryTTL:     envDuration("CART_SUMMARY_CACHE_TTL", 5*time.Minute),
		},
		BootstrapOnStartup: envBool("CART_BOOTSTRAP_COLLECTIONS_ON_STARTUP", false),
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.ServiceName) == "" {
		return errors.New("CART_SERVICE_NAME is required")
	}
	if strings.TrimSpace(c.HTTP.Address) == "" {
		return errors.New("CART_HTTP_ADDR is required")
	}
	if strings.TrimSpace(c.Mongo.URI) == "" {
		return errors.New("CART_MONGO_URI or MONGO_URI is required")
	}
	if _, err := url.ParseRequestURI(c.Mongo.URI); err != nil {
		return fmt.Errorf("CART_MONGO_URI is invalid: %w", err)
	}
	if strings.TrimSpace(c.Mongo.Database) != defaultMongoDatabase {
		return fmt.Errorf("CART_MONGO_DATABASE must be %q for Cart Service ownership", defaultMongoDatabase)
	}
	if c.Mongo.ConnectTimeout <= 0 {
		return errors.New("CART_MONGO_CONNECT_TIMEOUT must be greater than zero")
	}
	if c.Mongo.PingTimeout <= 0 {
		return errors.New("CART_MONGO_PING_TIMEOUT must be greater than zero")
	}
	if strings.TrimSpace(c.Redis.Address) == "" {
		return errors.New("CART_REDIS_ADDR or REDIS_ADDR is required")
	}
	if c.Redis.DB < 0 {
		return errors.New("CART_REDIS_DB must be greater than or equal to zero")
	}
	if c.Redis.DialTimeout <= 0 || c.Redis.ReadTimeout <= 0 || c.Redis.WriteTimeout <= 0 || c.Redis.PingTimeout <= 0 {
		return errors.New("Redis timeouts must be greater than zero")
	}
	if strings.TrimSpace(c.Product.BaseURL) == "" {
		return errors.New("CART_PRODUCT_BASE_URL or PRODUCT_SERVICE_BASE_URL is required")
	}
	productURL, err := url.ParseRequestURI(c.Product.BaseURL)
	if err != nil {
		return fmt.Errorf("CART_PRODUCT_BASE_URL is invalid: %w", err)
	}
	if productURL.Scheme != "http" && productURL.Scheme != "https" {
		return errors.New("CART_PRODUCT_BASE_URL must use http or https")
	}
	if c.Product.RequestTimeout <= 0 {
		return errors.New("CART_PRODUCT_REQUEST_TIMEOUT must be greater than zero")
	}
	if strings.TrimSpace(c.CMS.BaseURL) == "" {
		return errors.New("CART_CMS_BASE_URL or CMS_SERVICE_BASE_URL is required")
	}
	cmsURL, err := url.ParseRequestURI(c.CMS.BaseURL)
	if err != nil {
		return fmt.Errorf("CART_CMS_BASE_URL is invalid: %w", err)
	}
	if cmsURL.Scheme != "http" && cmsURL.Scheme != "https" {
		return errors.New("CART_CMS_BASE_URL must use http or https")
	}
	if strings.TrimSpace(c.CMS.ValidateCouponPath) == "" {
		return errors.New("CART_CMS_VALIDATE_COUPON_PATH is required")
	}
	if c.CMS.RequestTimeout <= 0 {
		return errors.New("CART_COUPON_PREVIEW_TIMEOUT must be greater than zero")
	}
	if c.Cart.ExpiryTTL <= 0 {
		return errors.New("CART_EXPIRY_TTL must be greater than zero")
	}
	if c.Cart.UserExpiryTTL <= 0 {
		return errors.New("CART_USER_EXPIRY_TTL must be greater than zero")
	}
	if c.Cart.GuestExpiryTTL <= 0 {
		return errors.New("CART_GUEST_EXPIRY_TTL must be greater than zero")
	}
	if !validCurrency(c.Cart.DefaultCurrency) {
		return errors.New("CART_DEFAULT_CURRENCY must be a 3-letter uppercase ISO currency code")
	}
	if c.Cart.MaxSaveAttempts <= 0 {
		return errors.New("CART_MAX_SAVE_ATTEMPTS must be greater than zero")
	}
	if c.CartExpiry.CleanupInterval <= 0 {
		return errors.New("CART_CLEANUP_INTERVAL must be greater than zero")
	}
	if c.CartExpiry.CleanupBatchSize <= 0 {
		return errors.New("CART_CLEANUP_BATCH_SIZE must be greater than zero")
	}
	if c.CartExpiry.CleanupLockTTL <= 0 {
		return errors.New("CART_CLEANUP_LOCK_TTL must be greater than zero")
	}
	if c.CartExpiry.CleanupRunTimeout <= 0 {
		return errors.New("CART_CLEANUP_RUN_TIMEOUT must be greater than zero")
	}
	if strings.TrimSpace(c.CartExpiry.CleanupLockKey) == "" {
		return errors.New("CART_CLEANUP_LOCK_KEY is required")
	}
	if c.Cache.ActiveUserTTL <= 0 || c.Cache.ActiveGuestTTL <= 0 || c.Cache.SummaryTTL <= 0 {
		return errors.New("cart cache TTL values must be greater than zero")
	}
	if c.HTTP.ReadTimeout <= 0 || c.HTTP.WriteTimeout <= 0 || c.HTTP.IdleTimeout <= 0 || c.HTTP.ShutdownTimeout <= 0 {
		return errors.New("HTTP timeouts must be greater than zero")
	}
	return nil
}

func envString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func validCurrency(currency string) bool {
	if len(currency) != 3 {
		return false
	}
	for _, char := range currency {
		if char < 'A' || char > 'Z' {
			return false
		}
	}
	return true
}
