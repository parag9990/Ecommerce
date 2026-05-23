package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"product-service/internal/client"
	"product-service/internal/domain"
)

const DefaultMongoDatabaseName = "product_db"

type Config struct {
	ServiceName string
	Environment string
	LogLevel    slog.Level
	Catalog     CatalogConfig
	Read        ReadConfig
	CMS         CMSConfig
	Mongo       MongoConfig
}

type CatalogConfig struct {
	DefaultCurrency               string
	StrictAttributeSchema         bool
	RequirePrimaryImageForPublish bool
	MaxImagesPerProduct           int
	MaxVariantsPerProduct         int
}

type ReadConfig struct {
	DefaultPageSize int
	MaxPageSize     int
	MaxBatchSize    int
}

type MongoConfig struct {
	URI                   string
	DatabaseName          string
	AutoCreateCollections bool
}

type CMSConfig struct {
	CatalogManagementAllowed        bool
	ModerationDecision              client.ModerationDecision
	AllowDraftWritesWhenUnavailable bool
}

func Load() (Config, error) {
	cfg := Default()
	cfg.ServiceName = stringEnv("SERVICE_NAME", cfg.ServiceName)
	cfg.Environment = stringEnv("ENVIRONMENT", cfg.Environment)
	cfg.Catalog.DefaultCurrency = stringEnv("PRODUCT_DEFAULT_CURRENCY", cfg.Catalog.DefaultCurrency)
	cfg.Catalog.StrictAttributeSchema = boolEnv("PRODUCT_STRICT_ATTRIBUTE_SCHEMA", cfg.Catalog.StrictAttributeSchema)
	cfg.Catalog.RequirePrimaryImageForPublish = boolEnv("PRODUCT_REQUIRE_PRIMARY_IMAGE_FOR_PUBLISH", cfg.Catalog.RequirePrimaryImageForPublish)
	cfg.Catalog.MaxImagesPerProduct = intEnv("PRODUCT_MAX_IMAGES_PER_PRODUCT", cfg.Catalog.MaxImagesPerProduct)
	cfg.Catalog.MaxVariantsPerProduct = intEnv("PRODUCT_MAX_VARIANTS_PER_PRODUCT", cfg.Catalog.MaxVariantsPerProduct)
	cfg.Read.DefaultPageSize = intEnv("PRODUCT_READ_DEFAULT_PAGE_SIZE", cfg.Read.DefaultPageSize)
	cfg.Read.MaxPageSize = intEnv("PRODUCT_READ_MAX_PAGE_SIZE", cfg.Read.MaxPageSize)
	cfg.Read.MaxBatchSize = intEnv("PRODUCT_READ_MAX_BATCH_SIZE", cfg.Read.MaxBatchSize)
	cfg.CMS.CatalogManagementAllowed = boolEnv("PRODUCT_CMS_CATALOG_MANAGEMENT_ALLOWED", cfg.CMS.CatalogManagementAllowed)
	cfg.CMS.ModerationDecision = client.ModerationDecision(stringEnv("PRODUCT_CMS_MODERATION_DECISION", string(cfg.CMS.ModerationDecision)))
	cfg.CMS.AllowDraftWritesWhenUnavailable = boolEnv("PRODUCT_CMS_ALLOW_DRAFT_WRITES_WHEN_UNAVAILABLE", cfg.CMS.AllowDraftWritesWhenUnavailable)
	cfg.Mongo.URI = stringEnv("PRODUCT_MONGO_URI", cfg.Mongo.URI)
	cfg.Mongo.DatabaseName = stringEnv("PRODUCT_MONGO_DATABASE", cfg.Mongo.DatabaseName)
	cfg.Mongo.AutoCreateCollections = boolEnv("PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS", cfg.Mongo.AutoCreateCollections)

	level, err := parseLogLevel(stringEnv("LOG_LEVEL", cfg.LogLevel.String()))
	if err != nil {
		return Config{}, err
	}
	cfg.LogLevel = level

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Default() Config {
	options := domain.DefaultValidationOptions()
	return Config{
		ServiceName: "product-service",
		Environment: "local",
		LogLevel:    slog.LevelInfo,
		Catalog: CatalogConfig{
			DefaultCurrency:               options.DefaultCurrency,
			StrictAttributeSchema:         options.StrictAttributeSchema,
			RequirePrimaryImageForPublish: options.RequirePrimaryImageForPublish,
			MaxImagesPerProduct:           options.MaxImagesPerProduct,
			MaxVariantsPerProduct:         options.MaxVariantsPerProduct,
		},
		Read: ReadConfig{
			DefaultPageSize: 20,
			MaxPageSize:     100,
			MaxBatchSize:    100,
		},
		CMS: CMSConfig{
			CatalogManagementAllowed:        true,
			ModerationDecision:              client.ModerationAutoPublish,
			AllowDraftWritesWhenUnavailable: true,
		},
		Mongo: MongoConfig{
			URI:                   "mongodb://localhost:27017",
			DatabaseName:          DefaultMongoDatabaseName,
			AutoCreateCollections: false,
		},
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.ServiceName) == "" {
		return fmt.Errorf("SERVICE_NAME is required")
	}
	if strings.TrimSpace(c.Environment) == "" {
		return fmt.Errorf("ENVIRONMENT is required")
	}
	if c.Catalog.MaxImagesPerProduct <= 0 {
		return fmt.Errorf("PRODUCT_MAX_IMAGES_PER_PRODUCT must be greater than zero")
	}
	if c.Catalog.MaxVariantsPerProduct <= 0 {
		return fmt.Errorf("PRODUCT_MAX_VARIANTS_PER_PRODUCT must be greater than zero")
	}
	if c.Read.DefaultPageSize <= 0 {
		return fmt.Errorf("PRODUCT_READ_DEFAULT_PAGE_SIZE must be greater than zero")
	}
	if c.Read.MaxPageSize <= 0 {
		return fmt.Errorf("PRODUCT_READ_MAX_PAGE_SIZE must be greater than zero")
	}
	if c.Read.DefaultPageSize > c.Read.MaxPageSize {
		return fmt.Errorf("PRODUCT_READ_DEFAULT_PAGE_SIZE cannot exceed PRODUCT_READ_MAX_PAGE_SIZE")
	}
	if c.Read.MaxBatchSize <= 0 {
		return fmt.Errorf("PRODUCT_READ_MAX_BATCH_SIZE must be greater than zero")
	}
	if strings.TrimSpace(c.Mongo.URI) == "" {
		return fmt.Errorf("PRODUCT_MONGO_URI is required")
	}
	if strings.TrimSpace(c.Mongo.DatabaseName) == "" {
		return fmt.Errorf("PRODUCT_MONGO_DATABASE is required")
	}
	if !client.NormalizeModerationDecision(c.CMS.ModerationDecision).Valid() {
		return fmt.Errorf("PRODUCT_CMS_MODERATION_DECISION must be one of %s or %s", client.ModerationAutoPublish, client.ModerationReviewRequired)
	}
	report := domain.NewMoney(1, c.Catalog.DefaultCurrency).Validate("PRODUCT_DEFAULT_CURRENCY")
	if report.HasErrors() {
		return fmt.Errorf("PRODUCT_DEFAULT_CURRENCY must be a valid three-letter currency code")
	}
	return nil
}

func (c Config) ValidationOptions() domain.ValidationOptions {
	return domain.ValidationOptions{
		StrictAttributeSchema:         c.Catalog.StrictAttributeSchema,
		RequirePrimaryImageForPublish: c.Catalog.RequirePrimaryImageForPublish,
		MaxImagesPerProduct:           c.Catalog.MaxImagesPerProduct,
		MaxVariantsPerProduct:         c.Catalog.MaxVariantsPerProduct,
		DefaultCurrency:               strings.ToUpper(c.Catalog.DefaultCurrency),
	}
}

func stringEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return strings.TrimSpace(value)
}

func boolEnv(key string, fallback bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return parsed
}

func intEnv(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return parsed
}

func parseLogLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("LOG_LEVEL must be one of debug, info, warn, error")
	}
}
