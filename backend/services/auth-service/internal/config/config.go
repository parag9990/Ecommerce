package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	otpsecurity "github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/otp"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/password"
	tokensecurity "github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/token"
)

const (
	defaultHTTPAddress     = ":8081"
	defaultShutdownTimeout = 10 * time.Second
)

type Config struct {
	HTTP         HTTPConfig
	Database     DatabaseConfig
	Redis        RedisConfig
	Password     PasswordConfig
	Token        TokenConfig
	OTP          OTPConfig
	RBAC         RBACConfig
	Notification NotificationConfig
	SessionLink  SessionLinkConfig
}

type HTTPConfig struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type PasswordConfig struct {
	Policy            password.Policy
	Argon2id          password.Argon2idParams
	Bcrypt            password.BcryptParams
	MaxFailedAttempts int
	LockoutDuration   time.Duration
}

type TokenConfig struct {
	Issuer             string
	Audience           string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	ClockSkew          time.Duration
	SigningAlgorithm   string
	KeyID              string
	PrivateKeyPEMPath  string
	PublicKeyPEMPath   string
	RefreshTokenPepper string
}

type OTPConfig struct {
	HashPepper         string
	RateLimitPepper    string
	Length             int
	TTL                time.Duration
	MaxAttempts        int
	ResendCooldown     time.Duration
	SendLimitWindow    time.Duration
	SendLimitPerWindow int
	DailyLimit         int
	VerifyIPWindow     time.Duration
	VerifyIPLimit      int
}

type RBACConfig struct {
	RoleMutationReasonMaxLength int
}

type NotificationConfig struct {
	OTPEndpoint string
	Timeout     time.Duration
}

type SessionLinkConfig struct {
	Mode                string
	AuthEventsTopic     string
	HTTPPublishEndpoint string
	SessionGRPCAddr     string
	Timeout             time.Duration
	EventPepper         string
	OutboxWorker        OutboxWorkerConfig
}

type OutboxWorkerConfig struct {
	BatchSize        int
	Interval         time.Duration
	MaxAttempts      int
	InitialBackoff   time.Duration
	MaxBackoff       time.Duration
	StaleLockTimeout time.Duration
}

func Load() (Config, error) {
	defaultOTPPolicy := otpsecurity.DefaultPolicy()
	otpHashPepper := os.Getenv("OTP_HASH_PEPPER")
	otpRateLimitPepper := envString("OTP_RATE_LIMIT_PEPPER", otpHashPepper)

	cfg := Config{
		HTTP: HTTPConfig{
			Address:         envString("AUTH_HTTP_ADDR", defaultHTTPAddress),
			ReadTimeout:     envDuration("AUTH_HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:    envDuration("AUTH_HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     envDuration("AUTH_HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: envDuration("AUTH_SHUTDOWN_TIMEOUT", defaultShutdownTimeout),
		},
		Database: DatabaseConfig{
			DSN: os.Getenv("AUTH_MYSQL_DSN"),
		},
		Redis: RedisConfig{
			Addr:         envString("AUTH_REDIS_ADDR", "localhost:6379"),
			Password:     os.Getenv("AUTH_REDIS_PASSWORD"),
			DB:           envInt("AUTH_REDIS_DB", 0),
			DialTimeout:  envDuration("AUTH_REDIS_DIAL_TIMEOUT", 2*time.Second),
			ReadTimeout:  envDuration("AUTH_REDIS_READ_TIMEOUT", 2*time.Second),
			WriteTimeout: envDuration("AUTH_REDIS_WRITE_TIMEOUT", 2*time.Second),
		},
		Password: PasswordConfig{
			Policy: password.Policy{
				MinLength: envInt("AUTH_PASSWORD_MIN_LENGTH", password.DefaultMinLength),
				MaxLength: envInt("AUTH_PASSWORD_MAX_LENGTH", password.DefaultMaxLength),
			},
			Argon2id: password.Argon2idParams{
				MemoryKiB:   uint32(envInt("AUTH_PASSWORD_ARGON2_MEMORY_KIB", password.DefaultArgon2idMemoryKiB)),
				Iterations:  uint32(envInt("AUTH_PASSWORD_ARGON2_ITERATIONS", password.DefaultArgon2idIterations)),
				Parallelism: uint8(envInt("AUTH_PASSWORD_ARGON2_PARALLELISM", password.DefaultArgon2idParallelism)),
				SaltLength:  uint32(envInt("AUTH_PASSWORD_SALT_LENGTH", password.DefaultArgon2idSaltLength)),
				KeyLength:   uint32(envInt("AUTH_PASSWORD_KEY_LENGTH", password.DefaultArgon2idKeyLength)),
			},
			Bcrypt: password.BcryptParams{
				Cost: envInt("AUTH_PASSWORD_BCRYPT_COST", password.DefaultBcryptCost),
			},
			MaxFailedAttempts: envInt("AUTH_PASSWORD_MAX_FAILED_ATTEMPTS", 5),
			LockoutDuration:   envDuration("AUTH_PASSWORD_LOCKOUT_DURATION", 15*time.Minute),
		},
		Token: TokenConfig{
			Issuer:             envString("JWT_ISSUER", "ecommerce-auth"),
			Audience:           envString("JWT_AUDIENCE", "ecommerce-api"),
			AccessTokenTTL:     envDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTokenTTL:    envDuration("REFRESH_TOKEN_TTL", 720*time.Hour),
			ClockSkew:          envDuration("JWT_CLOCK_SKEW", 30*time.Second),
			SigningAlgorithm:   envString("JWT_SIGNING_ALG", tokensecurity.SigningAlgorithmRS256),
			KeyID:              os.Getenv("JWT_KEY_ID"),
			PrivateKeyPEMPath:  os.Getenv("JWT_PRIVATE_KEY_PEM_PATH"),
			PublicKeyPEMPath:   os.Getenv("JWT_PUBLIC_KEY_PEM_PATH"),
			RefreshTokenPepper: os.Getenv("REFRESH_TOKEN_PEPPER"),
		},
		OTP: OTPConfig{
			HashPepper:         otpHashPepper,
			RateLimitPepper:    otpRateLimitPepper,
			Length:             envInt("OTP_LENGTH", defaultOTPPolicy.Length),
			TTL:                envDurationWithSeconds("OTP_TTL", "OTP_TTL_SECONDS", defaultOTPPolicy.TTL),
			MaxAttempts:        envInt("OTP_MAX_ATTEMPTS", defaultOTPPolicy.MaxAttempts),
			ResendCooldown:     envDurationWithSeconds("OTP_RESEND_COOLDOWN", "OTP_RESEND_COOLDOWN_SECONDS", defaultOTPPolicy.ResendCooldown),
			SendLimitWindow:    envDurationWithSeconds("OTP_SEND_LIMIT_WINDOW", "OTP_SEND_LIMIT_WINDOW_SECONDS", defaultOTPPolicy.SendLimitWindow),
			SendLimitPerWindow: envInt("OTP_SEND_LIMIT_PER_WINDOW", defaultOTPPolicy.SendLimitPerWindow),
			DailyLimit:         envInt("OTP_DAILY_LIMIT", defaultOTPPolicy.DailyLimit),
			VerifyIPWindow:     envDurationWithSeconds("OTP_VERIFY_IP_WINDOW", "OTP_VERIFY_IP_WINDOW_SECONDS", defaultOTPPolicy.VerifyIPWindow),
			VerifyIPLimit:      envInt("OTP_VERIFY_IP_LIMIT", defaultOTPPolicy.VerifyIPLimit),
		},
		RBAC: RBACConfig{
			RoleMutationReasonMaxLength: envInt("AUTH_ROLE_REASON_MAX_LENGTH", 512),
		},
		Notification: NotificationConfig{
			OTPEndpoint: envString("NOTIFICATION_OTP_ENDPOINT", "http://localhost:8084/internal/v1/notifications/otp"),
			Timeout:     envDuration("NOTIFICATION_TIMEOUT", 3*time.Second),
		},
		SessionLink: SessionLinkConfig{
			Mode:                strings.ToLower(strings.TrimSpace(envString("SESSION_LINK_MODE", "outbox"))),
			AuthEventsTopic:     envString("AUTH_EVENTS_TOPIC", "auth.events"),
			HTTPPublishEndpoint: os.Getenv("AUTH_EVENTS_PUBLISH_ENDPOINT"),
			SessionGRPCAddr:     envString("SESSION_GRPC_ADDR", "session-service:9090"),
			Timeout:             envDurationWithMillis("SESSION_LINK_TIMEOUT", "SESSION_LINK_TIMEOUT_MS", 150*time.Millisecond),
			EventPepper:         os.Getenv("SESSION_EVENT_PEPPER"),
			OutboxWorker: OutboxWorkerConfig{
				BatchSize:        envInt("OUTBOX_WORKER_BATCH_SIZE", 100),
				Interval:         envDurationWithMillis("OUTBOX_WORKER_INTERVAL", "OUTBOX_WORKER_INTERVAL_MS", time.Second),
				MaxAttempts:      envInt("OUTBOX_MAX_ATTEMPTS", 5),
				InitialBackoff:   envDuration("OUTBOX_INITIAL_BACKOFF", 5*time.Second),
				MaxBackoff:       envDuration("OUTBOX_MAX_BACKOFF", 10*time.Minute),
				StaleLockTimeout: envDuration("OUTBOX_STALE_LOCK_TIMEOUT", 5*time.Minute),
			},
		},
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) PasswordRouterConfig() password.RouterConfig {
	return password.RouterConfig{
		PreferredAlgorithm: password.AlgorithmArgon2id,
		Policy:             c.Password.Policy,
		Argon2id:           c.Password.Argon2id,
		Bcrypt:             c.Password.Bcrypt,
	}
}

func (c Config) OTPPolicy() otpsecurity.Policy {
	return otpsecurity.Policy{
		Length:             c.OTP.Length,
		TTL:                c.OTP.TTL,
		MaxAttempts:        c.OTP.MaxAttempts,
		ResendCooldown:     c.OTP.ResendCooldown,
		SendLimitWindow:    c.OTP.SendLimitWindow,
		SendLimitPerWindow: c.OTP.SendLimitPerWindow,
		DailyLimit:         c.OTP.DailyLimit,
		VerifyIPWindow:     c.OTP.VerifyIPWindow,
		VerifyIPLimit:      c.OTP.VerifyIPLimit,
	}
}

func (c Config) Validate() error {
	if c.HTTP.Address == "" {
		return errors.New("AUTH_HTTP_ADDR cannot be empty")
	}
	if c.Database.DSN == "" {
		return errors.New("AUTH_MYSQL_DSN cannot be empty")
	}
	if err := c.Redis.Validate(); err != nil {
		return fmt.Errorf("invalid redis config: %w", err)
	}
	if c.Password.MaxFailedAttempts <= 0 {
		return errors.New("AUTH_PASSWORD_MAX_FAILED_ATTEMPTS must be greater than zero")
	}
	if c.Password.LockoutDuration <= 0 {
		return errors.New("AUTH_PASSWORD_LOCKOUT_DURATION must be greater than zero")
	}
	if err := c.Password.Policy.Validate(); err != nil {
		return fmt.Errorf("invalid password policy: %w", err)
	}
	if err := c.Password.Argon2id.Validate(); err != nil {
		return fmt.Errorf("invalid argon2id config: %w", err)
	}
	if err := c.Password.Bcrypt.Validate(); err != nil {
		return fmt.Errorf("invalid bcrypt config: %w", err)
	}
	if err := c.Token.Validate(); err != nil {
		return fmt.Errorf("invalid token config: %w", err)
	}
	if err := c.OTP.Validate(); err != nil {
		return fmt.Errorf("invalid otp config: %w", err)
	}
	if err := c.RBAC.Validate(); err != nil {
		return fmt.Errorf("invalid rbac config: %w", err)
	}
	if err := c.Notification.Validate(); err != nil {
		return fmt.Errorf("invalid notification config: %w", err)
	}
	if err := c.SessionLink.Validate(); err != nil {
		return fmt.Errorf("invalid session link config: %w", err)
	}
	return nil
}

func (c RedisConfig) Validate() error {
	if c.Addr == "" {
		return errors.New("AUTH_REDIS_ADDR cannot be empty")
	}
	if c.DB < 0 {
		return errors.New("AUTH_REDIS_DB cannot be negative")
	}
	if c.DialTimeout <= 0 {
		return errors.New("AUTH_REDIS_DIAL_TIMEOUT must be greater than zero")
	}
	if c.ReadTimeout <= 0 {
		return errors.New("AUTH_REDIS_READ_TIMEOUT must be greater than zero")
	}
	if c.WriteTimeout <= 0 {
		return errors.New("AUTH_REDIS_WRITE_TIMEOUT must be greater than zero")
	}
	return nil
}

func (c TokenConfig) Validate() error {
	if c.Issuer == "" {
		return errors.New("JWT_ISSUER cannot be empty")
	}
	if c.Audience == "" {
		return errors.New("JWT_AUDIENCE cannot be empty")
	}
	if c.AccessTokenTTL <= 0 {
		return errors.New("JWT_ACCESS_TTL must be greater than zero")
	}
	if c.RefreshTokenTTL <= 0 {
		return errors.New("REFRESH_TOKEN_TTL must be greater than zero")
	}
	if c.ClockSkew < 0 {
		return errors.New("JWT_CLOCK_SKEW cannot be negative")
	}
	if c.SigningAlgorithm != tokensecurity.SigningAlgorithmRS256 {
		return errors.New("JWT_SIGNING_ALG must be RS256")
	}
	if c.KeyID == "" {
		return errors.New("JWT_KEY_ID cannot be empty")
	}
	if c.PrivateKeyPEMPath == "" {
		return errors.New("JWT_PRIVATE_KEY_PEM_PATH cannot be empty")
	}
	if c.RefreshTokenPepper == "" {
		return errors.New("REFRESH_TOKEN_PEPPER cannot be empty")
	}
	return nil
}

func (c OTPConfig) Validate() error {
	if c.HashPepper == "" {
		return errors.New("OTP_HASH_PEPPER cannot be empty")
	}
	if c.RateLimitPepper == "" {
		return errors.New("OTP_RATE_LIMIT_PEPPER cannot be empty")
	}
	return otpsecurity.Policy{
		Length:             c.Length,
		TTL:                c.TTL,
		MaxAttempts:        c.MaxAttempts,
		ResendCooldown:     c.ResendCooldown,
		SendLimitWindow:    c.SendLimitWindow,
		SendLimitPerWindow: c.SendLimitPerWindow,
		DailyLimit:         c.DailyLimit,
		VerifyIPWindow:     c.VerifyIPWindow,
		VerifyIPLimit:      c.VerifyIPLimit,
	}.Validate()
}

func (c RBACConfig) Validate() error {
	if c.RoleMutationReasonMaxLength <= 0 {
		return errors.New("AUTH_ROLE_REASON_MAX_LENGTH must be greater than zero")
	}
	return nil
}

func (c NotificationConfig) Validate() error {
	if c.OTPEndpoint == "" {
		return errors.New("NOTIFICATION_OTP_ENDPOINT cannot be empty")
	}
	parsed, err := url.ParseRequestURI(c.OTPEndpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("NOTIFICATION_OTP_ENDPOINT must be an absolute URL")
	}
	if c.Timeout <= 0 {
		return errors.New("NOTIFICATION_TIMEOUT must be greater than zero")
	}
	return nil
}

func (c SessionLinkConfig) Validate() error {
	mode := strings.ToLower(strings.TrimSpace(c.Mode))
	switch mode {
	case "outbox", "disabled":
	default:
		return errors.New("SESSION_LINK_MODE must be outbox or disabled")
	}
	if mode == "disabled" {
		return nil
	}
	if c.AuthEventsTopic == "" {
		return errors.New("AUTH_EVENTS_TOPIC cannot be empty")
	}
	if c.Timeout <= 0 {
		return errors.New("SESSION_LINK_TIMEOUT must be greater than zero")
	}
	if c.EventPepper == "" {
		return errors.New("SESSION_EVENT_PEPPER cannot be empty")
	}
	if c.HTTPPublishEndpoint != "" {
		parsed, err := url.ParseRequestURI(c.HTTPPublishEndpoint)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return errors.New("AUTH_EVENTS_PUBLISH_ENDPOINT must be an absolute URL when set")
		}
	}
	if err := c.OutboxWorker.Validate(); err != nil {
		return err
	}
	return nil
}

func (c OutboxWorkerConfig) Validate() error {
	if c.BatchSize <= 0 {
		return errors.New("OUTBOX_WORKER_BATCH_SIZE must be greater than zero")
	}
	if c.Interval <= 0 {
		return errors.New("OUTBOX_WORKER_INTERVAL must be greater than zero")
	}
	if c.MaxAttempts <= 0 {
		return errors.New("OUTBOX_MAX_ATTEMPTS must be greater than zero")
	}
	if c.InitialBackoff <= 0 {
		return errors.New("OUTBOX_INITIAL_BACKOFF must be greater than zero")
	}
	if c.MaxBackoff < c.InitialBackoff {
		return errors.New("OUTBOX_MAX_BACKOFF must be greater than or equal to OUTBOX_INITIAL_BACKOFF")
	}
	if c.StaleLockTimeout <= 0 {
		return errors.New("OUTBOX_STALE_LOCK_TIMEOUT must be greater than zero")
	}
	return nil
}

func envString(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envDurationWithSeconds(durationKey string, secondsKey string, fallback time.Duration) time.Duration {
	if raw := os.Getenv(durationKey); raw != "" {
		value, err := time.ParseDuration(raw)
		if err == nil {
			return value
		}
		if seconds, err := strconv.Atoi(raw); err == nil {
			return time.Duration(seconds) * time.Second
		}
		return fallback
	}
	if raw := os.Getenv(secondsKey); raw != "" {
		seconds, err := strconv.Atoi(raw)
		if err != nil {
			return fallback
		}
		return time.Duration(seconds) * time.Second
	}
	return fallback
}

func envDurationWithMillis(durationKey string, millisKey string, fallback time.Duration) time.Duration {
	if raw := os.Getenv(durationKey); raw != "" {
		value, err := time.ParseDuration(raw)
		if err == nil {
			return value
		}
		if millis, err := strconv.Atoi(raw); err == nil {
			return time.Duration(millis) * time.Millisecond
		}
		return fallback
	}
	if raw := os.Getenv(millisKey); raw != "" {
		millis, err := strconv.Atoi(raw)
		if err != nil {
			return fallback
		}
		return time.Duration(millis) * time.Millisecond
	}
	return fallback
}
