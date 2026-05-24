package config

import (
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPAddress               = ":8087"
	defaultGRPCAddress               = ":9098"
	defaultShutdownTimeout           = 10 * time.Second
	defaultDBHost                    = "localhost"
	defaultDBPort                    = 3306
	defaultDBName                    = "cms_db"
	defaultDBUser                    = "cms_user"
	defaultDBTimezone                = "UTC"
	defaultProductBaseURL            = "http://localhost:8080"
	defaultCampaignMaxDays           = 90
	defaultAnalyticsRangeDays        = 30
	defaultAnalyticsMaxRangeDays     = 366
	defaultAnalyticsTopProductsLimit = 5
	defaultAnalyticsMaxTopProducts   = 20
	defaultAuditRangeDays            = 30
	defaultAuditPageSize             = 20
	defaultAuditMaxPageSize          = 100
)

type StaffStatusSource string

const (
	StaffStatusSourceGateway StaffStatusSource = "gateway"
	StaffStatusSourceMySQL   StaffStatusSource = "mysql"
)

type Config struct {
	HTTP           HTTPConfig
	GRPC           GRPCConfig
	Security       SecurityConfig
	Database       DatabaseConfig
	Authorization  AuthorizationConfig
	ProductService ProductServiceConfig
	Campaign       CampaignConfig
	Analytics      AnalyticsConfig
	Audit          AuditConfig
}

type HTTPConfig struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	MaxBodyBytes    int64
}

type GRPCConfig struct {
	Address                string
	MaxRecvMsgBytes        int
	MaxSendMsgBytes        int
	AllowedInternalCallers []string
}

type SecurityConfig struct {
	Environment        string
	InternalAuthHeader string
	InternalAuthToken  string
}

type DatabaseConfig struct {
	DSN             string
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	Timezone        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type AuthorizationConfig struct {
	StaffStatusSource StaffStatusSource
}

type ProductServiceConfig struct {
	BaseURL            string
	Timeout            time.Duration
	InternalAuthHeader string
	InternalAuthToken  string
}

type CampaignConfig struct {
	MaxDuration time.Duration
}

type AnalyticsConfig struct {
	DefaultCurrency         string
	DefaultRangeDays        int
	MaxRangeDays            int
	DefaultTopProductsLimit int
	MaxTopProductsLimit     int
}

type AuditConfig struct {
	DefaultRangeDays int
	DefaultPageSize  int
	MaxPageSize      int
}

func Load() (Config, error) {
	cfg := Config{
		HTTP: HTTPConfig{
			Address:         envString("CMS_HTTP_ADDR", defaultHTTPAddress),
			ReadTimeout:     envDuration("CMS_HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:    envDuration("CMS_HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     envDuration("CMS_HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: envDuration("CMS_SHUTDOWN_TIMEOUT", defaultShutdownTimeout),
			MaxBodyBytes:    int64(envInt("CMS_MAX_BODY_BYTES", 1<<20)),
		},
		GRPC: GRPCConfig{
			Address:                envString("CMS_GRPC_ADDR", defaultGRPCAddress),
			MaxRecvMsgBytes:        envInt("CMS_GRPC_MAX_RECV_MSG_BYTES", 1<<20),
			MaxSendMsgBytes:        envInt("CMS_GRPC_MAX_SEND_MSG_BYTES", 1<<20),
			AllowedInternalCallers: envCSV("CMS_GRPC_ALLOWED_INTERNAL_CALLERS", []string{"cart-service", "order-service", "api-gateway"}),
		},
		Security: SecurityConfig{
			Environment:        strings.ToLower(strings.TrimSpace(envString("CMS_ENV", "local"))),
			InternalAuthHeader: envString("CMS_INTERNAL_AUTH_HEADER", "X-Internal-Token"),
			InternalAuthToken:  os.Getenv("CMS_INTERNAL_AUTH_TOKEN"),
		},
		Database: DatabaseConfig{
			DSN:             strings.TrimSpace(os.Getenv("CMS_MYSQL_DSN")),
			Host:            envString("CMS_DB_HOST", defaultDBHost),
			Port:            envInt("CMS_DB_PORT", defaultDBPort),
			Name:            envString("CMS_DB_NAME", defaultDBName),
			User:            envString("CMS_DB_USER", defaultDBUser),
			Password:        os.Getenv("CMS_DB_PASSWORD"),
			Timezone:        envString("CMS_DB_TIMEZONE", defaultDBTimezone),
			MaxOpenConns:    envInt("CMS_DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    envInt("CMS_DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: time.Duration(envInt("CMS_DB_CONN_MAX_LIFETIME_SECONDS", 300)) * time.Second,
		},
		Authorization: AuthorizationConfig{
			StaffStatusSource: StaffStatusSource(strings.ToLower(strings.TrimSpace(envString("CMS_STAFF_STATUS_SOURCE", string(StaffStatusSourceGateway))))),
		},
		ProductService: ProductServiceConfig{
			BaseURL:            envString("CMS_PRODUCT_SERVICE_BASE_URL", defaultProductBaseURL),
			Timeout:            envDuration("CMS_PRODUCT_SERVICE_TIMEOUT", 5*time.Second),
			InternalAuthHeader: envString("CMS_PRODUCT_SERVICE_AUTH_HEADER", "X-Internal-Token"),
			InternalAuthToken:  os.Getenv("CMS_PRODUCT_SERVICE_AUTH_TOKEN"),
		},
		Campaign: CampaignConfig{
			MaxDuration: time.Duration(envInt("CMS_CAMPAIGN_MAX_DURATION_DAYS", defaultCampaignMaxDays)) * 24 * time.Hour,
		},
		Analytics: AnalyticsConfig{
			DefaultCurrency:         strings.ToUpper(strings.TrimSpace(envString("CMS_ANALYTICS_DEFAULT_CURRENCY", "INR"))),
			DefaultRangeDays:        envInt("CMS_ANALYTICS_DEFAULT_RANGE_DAYS", defaultAnalyticsRangeDays),
			MaxRangeDays:            envInt("CMS_ANALYTICS_MAX_RANGE_DAYS", defaultAnalyticsMaxRangeDays),
			DefaultTopProductsLimit: envInt("CMS_ANALYTICS_DEFAULT_TOP_PRODUCTS_LIMIT", defaultAnalyticsTopProductsLimit),
			MaxTopProductsLimit:     envInt("CMS_ANALYTICS_MAX_TOP_PRODUCTS_LIMIT", defaultAnalyticsMaxTopProducts),
		},
		Audit: AuditConfig{
			DefaultRangeDays: envInt("CMS_AUDIT_DEFAULT_RANGE_DAYS", defaultAuditRangeDays),
			DefaultPageSize:  envInt("CMS_AUDIT_DEFAULT_PAGE_SIZE", defaultAuditPageSize),
			MaxPageSize:      envInt("CMS_AUDIT_MAX_PAGE_SIZE", defaultAuditMaxPageSize),
		},
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.HTTP.Address == "" {
		return errors.New("CMS_HTTP_ADDR cannot be empty")
	}
	if c.HTTP.ReadTimeout <= 0 {
		return errors.New("CMS_HTTP_READ_TIMEOUT must be greater than zero")
	}
	if c.HTTP.WriteTimeout <= 0 {
		return errors.New("CMS_HTTP_WRITE_TIMEOUT must be greater than zero")
	}
	if c.HTTP.IdleTimeout <= 0 {
		return errors.New("CMS_HTTP_IDLE_TIMEOUT must be greater than zero")
	}
	if c.HTTP.ShutdownTimeout <= 0 {
		return errors.New("CMS_SHUTDOWN_TIMEOUT must be greater than zero")
	}
	if c.HTTP.MaxBodyBytes <= 0 {
		return errors.New("CMS_MAX_BODY_BYTES must be greater than zero")
	}
	if err := c.GRPC.Validate(); err != nil {
		return err
	}
	if c.Security.InternalAuthHeader == "" {
		return errors.New("CMS_INTERNAL_AUTH_HEADER cannot be empty")
	}
	if c.Security.Environment == "production" && c.Security.InternalAuthToken == "" {
		return errors.New("CMS_INTERNAL_AUTH_TOKEN cannot be empty in production")
	}
	if err := c.Database.Validate(c.Security.Environment); err != nil {
		return err
	}
	switch c.Authorization.StaffStatusSource {
	case StaffStatusSourceGateway:
	case StaffStatusSourceMySQL:
	default:
		return errors.New("CMS_STAFF_STATUS_SOURCE must be gateway or mysql")
	}
	if err := c.ProductService.Validate(c.Security.Environment); err != nil {
		return err
	}
	if c.Campaign.MaxDuration < 0 {
		return errors.New("CMS_CAMPAIGN_MAX_DURATION_DAYS cannot be negative")
	}
	analytics := c.Analytics
	if analytics == (AnalyticsConfig{}) {
		analytics = defaultAnalyticsConfig()
	}
	if err := analytics.Validate(); err != nil {
		return err
	}
	audit := c.Audit
	if audit == (AuditConfig{}) {
		audit = defaultAuditConfig()
	}
	if err := audit.Validate(); err != nil {
		return err
	}
	return nil
}

func (c GRPCConfig) Validate() error {
	if strings.TrimSpace(c.Address) == "" && c.MaxRecvMsgBytes == 0 && c.MaxSendMsgBytes == 0 && len(c.AllowedInternalCallers) == 0 {
		return nil
	}
	if strings.TrimSpace(c.Address) == "" {
		return errors.New("CMS_GRPC_ADDR cannot be empty")
	}
	if c.MaxRecvMsgBytes <= 0 {
		return errors.New("CMS_GRPC_MAX_RECV_MSG_BYTES must be greater than zero")
	}
	if c.MaxSendMsgBytes <= 0 {
		return errors.New("CMS_GRPC_MAX_SEND_MSG_BYTES must be greater than zero")
	}
	return nil
}

func (c DatabaseConfig) Validate(environment string) error {
	if strings.TrimSpace(c.DSN) == "" {
		if strings.TrimSpace(c.Host) == "" {
			return errors.New("CMS_DB_HOST cannot be empty when CMS_MYSQL_DSN is not set")
		}
		if c.Port <= 0 || c.Port > 65535 {
			return errors.New("CMS_DB_PORT must be between 1 and 65535")
		}
		if strings.TrimSpace(c.Name) == "" {
			return errors.New("CMS_DB_NAME cannot be empty when CMS_MYSQL_DSN is not set")
		}
		if strings.TrimSpace(c.User) == "" {
			return errors.New("CMS_DB_USER cannot be empty when CMS_MYSQL_DSN is not set")
		}
		if strings.EqualFold(strings.TrimSpace(environment), "production") && c.Password == "" {
			return errors.New("CMS_DB_PASSWORD cannot be empty in production when CMS_MYSQL_DSN is not set")
		}
	}
	if strings.TrimSpace(c.Timezone) == "" {
		return errors.New("CMS_DB_TIMEZONE cannot be empty")
	}
	if c.MaxOpenConns <= 0 {
		return errors.New("CMS_DB_MAX_OPEN_CONNS must be greater than zero")
	}
	if c.MaxIdleConns < 0 {
		return errors.New("CMS_DB_MAX_IDLE_CONNS cannot be negative")
	}
	if c.MaxIdleConns > c.MaxOpenConns {
		return errors.New("CMS_DB_MAX_IDLE_CONNS cannot be greater than CMS_DB_MAX_OPEN_CONNS")
	}
	if c.ConnMaxLifetime <= 0 {
		return errors.New("CMS_DB_CONN_MAX_LIFETIME_SECONDS must be greater than zero")
	}
	return nil
}

func (c ProductServiceConfig) Validate(environment string) error {
	if strings.TrimSpace(c.BaseURL) == "" {
		if strings.EqualFold(strings.TrimSpace(environment), "production") {
			return errors.New("CMS_PRODUCT_SERVICE_BASE_URL cannot be empty in production")
		}
		return nil
	}
	parsed, err := url.Parse(strings.TrimSpace(c.BaseURL))
	if err != nil {
		return errors.New("CMS_PRODUCT_SERVICE_BASE_URL must be a valid URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("CMS_PRODUCT_SERVICE_BASE_URL must use http or https")
	}
	if parsed.Host == "" {
		return errors.New("CMS_PRODUCT_SERVICE_BASE_URL host cannot be empty")
	}
	if c.Timeout <= 0 {
		return errors.New("CMS_PRODUCT_SERVICE_TIMEOUT must be greater than zero")
	}
	if strings.TrimSpace(c.InternalAuthHeader) == "" {
		return errors.New("CMS_PRODUCT_SERVICE_AUTH_HEADER cannot be empty")
	}
	return nil
}

func (c AnalyticsConfig) Validate() error {
	if !validConfigCurrency(c.DefaultCurrency) {
		return errors.New("CMS_ANALYTICS_DEFAULT_CURRENCY must be a 3-letter uppercase currency code")
	}
	if c.DefaultRangeDays <= 0 {
		return errors.New("CMS_ANALYTICS_DEFAULT_RANGE_DAYS must be greater than zero")
	}
	if c.MaxRangeDays <= 0 || c.MaxRangeDays > defaultAnalyticsMaxRangeDays {
		return errors.New("CMS_ANALYTICS_MAX_RANGE_DAYS must be between 1 and 366")
	}
	if c.DefaultRangeDays > c.MaxRangeDays {
		return errors.New("CMS_ANALYTICS_DEFAULT_RANGE_DAYS cannot be greater than CMS_ANALYTICS_MAX_RANGE_DAYS")
	}
	if c.DefaultTopProductsLimit <= 0 {
		return errors.New("CMS_ANALYTICS_DEFAULT_TOP_PRODUCTS_LIMIT must be greater than zero")
	}
	if c.MaxTopProductsLimit <= 0 || c.MaxTopProductsLimit > defaultAnalyticsMaxTopProducts {
		return errors.New("CMS_ANALYTICS_MAX_TOP_PRODUCTS_LIMIT must be between 1 and 20")
	}
	if c.DefaultTopProductsLimit > c.MaxTopProductsLimit {
		return errors.New("CMS_ANALYTICS_DEFAULT_TOP_PRODUCTS_LIMIT cannot be greater than CMS_ANALYTICS_MAX_TOP_PRODUCTS_LIMIT")
	}
	return nil
}

func (c AuditConfig) Validate() error {
	if c.DefaultRangeDays <= 0 {
		return errors.New("CMS_AUDIT_DEFAULT_RANGE_DAYS must be greater than zero")
	}
	if c.DefaultPageSize <= 0 {
		return errors.New("CMS_AUDIT_DEFAULT_PAGE_SIZE must be greater than zero")
	}
	if c.MaxPageSize <= 0 || c.MaxPageSize > defaultAuditMaxPageSize {
		return errors.New("CMS_AUDIT_MAX_PAGE_SIZE must be between 1 and 100")
	}
	if c.DefaultPageSize > c.MaxPageSize {
		return errors.New("CMS_AUDIT_DEFAULT_PAGE_SIZE cannot be greater than CMS_AUDIT_MAX_PAGE_SIZE")
	}
	return nil
}

func defaultAnalyticsConfig() AnalyticsConfig {
	return AnalyticsConfig{
		DefaultCurrency:         "INR",
		DefaultRangeDays:        defaultAnalyticsRangeDays,
		MaxRangeDays:            defaultAnalyticsMaxRangeDays,
		DefaultTopProductsLimit: defaultAnalyticsTopProductsLimit,
		MaxTopProductsLimit:     defaultAnalyticsMaxTopProducts,
	}
}

func defaultAuditConfig() AuditConfig {
	return AuditConfig{
		DefaultRangeDays: defaultAuditRangeDays,
		DefaultPageSize:  defaultAuditPageSize,
		MaxPageSize:      defaultAuditMaxPageSize,
	}
}

func validConfigCurrency(currency string) bool {
	currency = strings.TrimSpace(currency)
	if len(currency) != 3 {
		return false
	}
	for _, ch := range currency {
		if ch < 'A' || ch > 'Z' {
			return false
		}
	}
	return true
}

func envString(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
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

func envCSV(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return append([]string(nil), fallback...)
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
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
