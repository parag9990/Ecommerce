package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

const (
	defaultServiceName                  = "recommendation-service"
	defaultHTTPAddress                  = ":8088"
	defaultGRPCAddress                  = ":9088"
	defaultReadTimeout                  = 5 * time.Second
	defaultWriteTimeout                 = 10 * time.Second
	defaultIdleTimeout                  = 60 * time.Second
	defaultShutdownTimeout              = 10 * time.Second
	defaultMaxBodyBytes                 = int64(64 << 10)
	defaultGRPCMaxRecvBytes             = 64 << 10
	defaultGRPCMaxSendBytes             = 256 << 10
	defaultGRPCDefaultDeadline          = 500 * time.Millisecond
	defaultRecommendationLimit          = 12
	defaultMaxLimit                     = 100
	defaultMaxIdentifierLength          = 128
	defaultMongoConnectTimeout          = 10 * time.Second
	defaultMongoPingTimeout             = 5 * time.Second
	defaultRedisDialTimeout             = 5 * time.Second
	defaultRedisReadTimeout             = 3 * time.Second
	defaultRedisWriteTimeout            = 3 * time.Second
	defaultRedisPingTimeout             = 3 * time.Second
	defaultRebuildLockTTL               = time.Minute
	defaultCacheDirtyTTL                = 15 * time.Minute
	defaultQueueProvider                = "kafka"
	defaultEventsTopic                  = "recommendation.events"
	defaultEventsRetryTopic             = "recommendation.events.retry"
	defaultEventsDLQTopic               = "recommendation.events.dlq"
	defaultEventsGroup                  = "recommendation-service-v1"
	defaultEventMaxBodyBytes            = int64(256 << 10)
	defaultEventRetryAttempts           = 4
	defaultEventVersion                 = 1
	defaultKafkaMinBytes                = 1
	defaultKafkaMaxBytes                = 10 << 20
	defaultFeatureReconcileBatchSize    = 200
	defaultFeatureReconcileInterval     = time.Minute
	defaultFeatureWindowRebuildInterval = 5 * time.Minute
	defaultRankingRebuildInterval       = 15 * time.Minute
	defaultPersonalizedProfileMaxAge    = 30 * 24 * time.Hour
	defaultABAssignmentTTL              = domain.DefaultABAssignmentTTL
	defaultABSalt                       = domain.DefaultABTestSalt
)

type Config struct {
	Service         ServiceConfig
	HTTP            HTTPConfig
	GRPC            GRPCConfig
	Recommendation  RecommendationConfig
	Storage         StorageConfig
	Events          EventConfig
	Features        FeatureConfig
	Ranking         RankingConfig
	Personalization PersonalizationConfig
	ABTesting       ABTestingConfig
}

type ServiceConfig struct {
	Name     string
	Env      string
	LogLevel slog.Level
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
	Address         string
	MaxRecvBytes    int
	MaxSendBytes    int
	DefaultDeadline time.Duration
}

type RecommendationConfig struct {
	DefaultLimit        int
	MaxLimit            int
	MaxIdentifierLength int
}

type StorageConfig struct {
	FailFast bool
	Mongo    MongoConfig
	Redis    RedisConfig
	Cache    CacheConfig
}

type EventConfig struct {
	Enabled            bool
	Provider           string
	SupportedVersion   int
	MaxMessageBytes    int64
	MaxRetryAttempts   int
	RetryBackoffs      []time.Duration
	UnknownEventPolicy string
	Kafka              KafkaEventConfig
}

type FeatureConfig struct {
	Enabled                 bool
	GuestProfileRetention   time.Duration
	UserProductRetention    time.Duration
	ProcessedEventRetention time.Duration
	RecentProductsLimit     int
	ReconcileBatchSize      int
	ReconcileInterval       time.Duration
	WindowRebuildInterval   time.Duration
}

type RankingConfig struct {
	Enabled         bool
	RebuildInterval time.Duration
	FormulaVersion  string
	Weights         domain.PopularityWeights
}

type PersonalizationConfig struct {
	Enabled                 bool
	FormulaVersion          string
	MinPositiveInteractions int
	ProfileMaxAge           time.Duration
	MaxCandidates           int
	MaxItemsPerSeller       int
	TopCategories           int
	TopSellers              int
	TopBrands               int
	DirectProductLimit      int
	Weights                 domain.PersonalizationWeights
}

type ABTestingConfig struct {
	Enabled             bool
	AssignmentTTL       time.Duration
	DefaultSalt         string
	FailOpen            bool
	Experiments         []domain.ExperimentDefinition
	MaxIdentifierLength int
}

type KafkaEventConfig struct {
	Brokers    []string
	Topic      string
	RetryTopic string
	DLQTopic   string
	GroupID    string
	MinBytes   int
	MaxBytes   int
}

type MongoConfig struct {
	URI                  string
	Database             string
	ConnectTimeout       time.Duration
	PingTimeout          time.Duration
	EnsureIndexes        bool
	InteractionRetention time.Duration
}

type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PingTimeout  time.Duration
}

type CacheConfig struct {
	KeyPrefix       string
	DefaultTTL      time.Duration
	PersonalizedTTL time.Duration
	GuestTTL        time.Duration
	RebuildLockTTL  time.Duration
	DirtyTTL        time.Duration
}

func Load() (Config, error) {
	experiments, err := envExperimentDefinitions("RECOMMENDATION_AB_EXPERIMENTS_JSON")
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Service: ServiceConfig{
			Name:     envString("SERVICE_NAME", defaultServiceName),
			Env:      envString("APP_ENV", "local"),
			LogLevel: envLogLevel("RECOMMENDATION_LOG_LEVEL", slog.LevelInfo),
		},
		HTTP: HTTPConfig{
			Address:         envString("RECOMMENDATION_HTTP_ADDR", defaultHTTPAddress),
			ReadTimeout:     envDuration("RECOMMENDATION_HTTP_READ_TIMEOUT", defaultReadTimeout),
			WriteTimeout:    envDuration("RECOMMENDATION_HTTP_WRITE_TIMEOUT", defaultWriteTimeout),
			IdleTimeout:     envDuration("RECOMMENDATION_HTTP_IDLE_TIMEOUT", defaultIdleTimeout),
			ShutdownTimeout: envDuration("RECOMMENDATION_SHUTDOWN_TIMEOUT", defaultShutdownTimeout),
			MaxBodyBytes:    envInt64("RECOMMENDATION_MAX_BODY_BYTES", defaultMaxBodyBytes),
		},
		GRPC: GRPCConfig{
			Address:         envString("RECOMMENDATION_GRPC_ADDR", defaultGRPCAddress),
			MaxRecvBytes:    envInt("RECOMMENDATION_GRPC_MAX_RECV_BYTES", defaultGRPCMaxRecvBytes),
			MaxSendBytes:    envInt("RECOMMENDATION_GRPC_MAX_SEND_BYTES", defaultGRPCMaxSendBytes),
			DefaultDeadline: envDuration("RECOMMENDATION_GRPC_DEFAULT_DEADLINE", defaultGRPCDefaultDeadline),
		},
		Recommendation: RecommendationConfig{
			DefaultLimit:        envInt("RECOMMENDATION_DEFAULT_LIMIT", defaultRecommendationLimit),
			MaxLimit:            envInt("RECOMMENDATION_MAX_LIMIT", defaultMaxLimit),
			MaxIdentifierLength: envInt("RECOMMENDATION_MAX_IDENTIFIER_LENGTH", defaultMaxIdentifierLength),
		},
		Storage: StorageConfig{
			FailFast: envBool("RECOMMENDATION_STORAGE_FAIL_FAST", false),
			Mongo: MongoConfig{
				URI:                  envString("RECOMMENDATION_MONGO_URI", ""),
				Database:             envString("RECOMMENDATION_MONGO_DATABASE", domain.RecommendationDatabaseName),
				ConnectTimeout:       envDuration("RECOMMENDATION_MONGO_CONNECT_TIMEOUT", defaultMongoConnectTimeout),
				PingTimeout:          envDuration("RECOMMENDATION_MONGO_PING_TIMEOUT", defaultMongoPingTimeout),
				EnsureIndexes:        envBool("RECOMMENDATION_MONGO_ENSURE_INDEXES", true),
				InteractionRetention: envDurationSeconds("RECOMMENDATION_INTERACTION_RETENTION_SECONDS", domain.DefaultInteractionRetention),
			},
			Redis: RedisConfig{
				Addr:         envString("RECOMMENDATION_REDIS_ADDR", ""),
				Password:     envString("RECOMMENDATION_REDIS_PASSWORD", ""),
				DB:           envInt("RECOMMENDATION_REDIS_DB", 0),
				DialTimeout:  envDuration("RECOMMENDATION_REDIS_DIAL_TIMEOUT", defaultRedisDialTimeout),
				ReadTimeout:  envDuration("RECOMMENDATION_REDIS_READ_TIMEOUT", defaultRedisReadTimeout),
				WriteTimeout: envDuration("RECOMMENDATION_REDIS_WRITE_TIMEOUT", defaultRedisWriteTimeout),
				PingTimeout:  envDuration("RECOMMENDATION_REDIS_PING_TIMEOUT", defaultRedisPingTimeout),
			},
			Cache: CacheConfig{
				KeyPrefix:       envString("RECOMMENDATION_CACHE_KEY_PREFIX", domain.DefaultCacheKeyPrefix),
				DefaultTTL:      envDurationSeconds("RECOMMENDATION_CACHE_TTL_SECONDS", domain.DefaultRecommendationCacheTTL),
				PersonalizedTTL: envDurationSeconds("RECOMMENDATION_PERSONALIZED_CACHE_TTL_SECONDS", domain.DefaultPersonalizedCacheTTL),
				GuestTTL:        envDurationSeconds("RECOMMENDATION_GUEST_CACHE_TTL_SECONDS", domain.DefaultGuestCacheTTL),
				RebuildLockTTL:  envDuration("RECOMMENDATION_CACHE_REBUILD_LOCK_TTL", defaultRebuildLockTTL),
				DirtyTTL:        envDurationSeconds("RECOMMENDATION_CACHE_DIRTY_TTL_SECONDS", defaultCacheDirtyTTL),
			},
		},
		Events: EventConfig{
			Enabled:            envBool("RECOMMENDATION_EVENTS_ENABLED", false),
			Provider:           envString("QUEUE_PROVIDER", envString("RECOMMENDATION_QUEUE_PROVIDER", defaultQueueProvider)),
			SupportedVersion:   envInt("EVENT_SUPPORTED_VERSION", defaultEventVersion),
			MaxMessageBytes:    envInt64("RECOMMENDATION_EVENTS_MAX_MESSAGE_BYTES", defaultEventMaxBodyBytes),
			MaxRetryAttempts:   envInt("EVENT_MAX_RETRY_ATTEMPTS", defaultEventRetryAttempts),
			RetryBackoffs:      envDurationListSeconds("EVENT_RETRY_BACKOFF_SECONDS", []time.Duration{5 * time.Second, 30 * time.Second, 2 * time.Minute}),
			UnknownEventPolicy: strings.ToLower(envString("RECOMMENDATION_EVENTS_UNKNOWN_POLICY", "dlq")),
			Kafka: KafkaEventConfig{
				Brokers:    envCSV("KAFKA_BROKERS"),
				Topic:      envString("RECOMMENDATION_EVENTS_TOPIC", defaultEventsTopic),
				RetryTopic: envString("RECOMMENDATION_EVENTS_RETRY_TOPIC", defaultEventsRetryTopic),
				DLQTopic:   envString("RECOMMENDATION_EVENTS_DLQ_TOPIC", defaultEventsDLQTopic),
				GroupID:    envString("RECOMMENDATION_EVENTS_GROUP", defaultEventsGroup),
				MinBytes:   envInt("RECOMMENDATION_KAFKA_MIN_BYTES", defaultKafkaMinBytes),
				MaxBytes:   envInt("RECOMMENDATION_KAFKA_MAX_BYTES", defaultKafkaMaxBytes),
			},
		},
		Features: FeatureConfig{
			Enabled:                 envBool("RECOMMENDATION_FEATURES_ENABLED", true),
			GuestProfileRetention:   envDurationSeconds("RECOMMENDATION_GUEST_PROFILE_RETENTION_SECONDS", domain.DefaultGuestProfileRetention),
			UserProductRetention:    envDurationSeconds("RECOMMENDATION_USER_PRODUCT_RETENTION_SECONDS", domain.DefaultUserProductRetention),
			ProcessedEventRetention: envDurationSeconds("RECOMMENDATION_FEATURE_EVENT_RETENTION_SECONDS", domain.DefaultFeatureEventRetention),
			RecentProductsLimit:     envInt("RECOMMENDATION_FEATURE_RECENT_PRODUCTS_LIMIT", domain.DefaultRecentProductsLimit),
			ReconcileBatchSize:      envInt("RECOMMENDATION_FEATURE_RECONCILE_BATCH_SIZE", defaultFeatureReconcileBatchSize),
			ReconcileInterval:       envDuration("RECOMMENDATION_FEATURE_RECONCILE_INTERVAL", defaultFeatureReconcileInterval),
			WindowRebuildInterval:   envDuration("RECOMMENDATION_FEATURE_WINDOW_REBUILD_INTERVAL", defaultFeatureWindowRebuildInterval),
		},
		Ranking: RankingConfig{
			Enabled:         envBool("RECOMMENDATION_RANKING_ENABLED", true),
			RebuildInterval: envDuration("RECOMMENDATION_RANKING_REBUILD_INTERVAL", defaultRankingRebuildInterval),
			FormulaVersion:  envString("RECOMMENDATION_RANKING_FORMULA_VERSION", domain.RuleBasedRankingFormulaVersion),
			Weights: domain.PopularityWeights{
				Views24h:       envInt64("RECOMMENDATION_RANKING_VIEW_24H_WEIGHT", domain.DefaultPopularityWeights().Views24h),
				WishlistAdds7d: envInt64("RECOMMENDATION_RANKING_WISHLIST_7D_WEIGHT", domain.DefaultPopularityWeights().WishlistAdds7d),
				CartAdds7d:     envInt64("RECOMMENDATION_RANKING_CART_7D_WEIGHT", domain.DefaultPopularityWeights().CartAdds7d),
				Purchases7d:    envInt64("RECOMMENDATION_RANKING_PURCHASE_7D_WEIGHT", domain.DefaultPopularityWeights().Purchases7d),
			},
		},
		Personalization: PersonalizationConfig{
			Enabled:                 envBool("RECOMMENDATION_PERSONALIZATION_ENABLED", true),
			FormulaVersion:          envString("RECOMMENDATION_PERSONALIZED_FORMULA_VERSION", domain.PersonalizationFormulaVersion),
			MinPositiveInteractions: envInt("RECOMMENDATION_PERSONALIZED_MIN_POSITIVE_INTERACTIONS", domain.DefaultPersonalizedMinInteractions),
			ProfileMaxAge:           envDurationDays("RECOMMENDATION_PERSONALIZED_PROFILE_MAX_AGE_DAYS", defaultPersonalizedProfileMaxAge),
			MaxCandidates:           envInt("RECOMMENDATION_PERSONALIZED_MAX_CANDIDATES", domain.DefaultPersonalizedMaxCandidates),
			MaxItemsPerSeller:       envInt("RECOMMENDATION_PERSONALIZED_MAX_ITEMS_PER_SELLER", domain.DefaultPersonalizedMaxItemsPerSeller),
			TopCategories:           envInt("RECOMMENDATION_PERSONALIZED_TOP_CATEGORIES", domain.DefaultPersonalizedTopCategories),
			TopSellers:              envInt("RECOMMENDATION_PERSONALIZED_TOP_SELLERS", domain.DefaultPersonalizedTopSellers),
			TopBrands:               envInt("RECOMMENDATION_PERSONALIZED_TOP_BRANDS", domain.DefaultPersonalizedTopBrands),
			DirectProductLimit:      envInt("RECOMMENDATION_PERSONALIZED_DIRECT_PRODUCT_LIMIT", domain.DefaultPersonalizedDirectProductLimit),
			Weights: domain.PersonalizationWeights{
				Category:       envFloat("RECOMMENDATION_PERSONALIZED_CATEGORY_WEIGHT", domain.DefaultPersonalizationWeights().Category),
				Seller:         envFloat("RECOMMENDATION_PERSONALIZED_SELLER_WEIGHT", domain.DefaultPersonalizationWeights().Seller),
				Brand:          envFloat("RECOMMENDATION_PERSONALIZED_BRAND_WEIGHT", domain.DefaultPersonalizationWeights().Brand),
				Price:          envFloat("RECOMMENDATION_PERSONALIZED_PRICE_WEIGHT", domain.DefaultPersonalizationWeights().Price),
				DirectAffinity: envFloat("RECOMMENDATION_PERSONALIZED_DIRECT_AFFINITY_WEIGHT", domain.DefaultPersonalizationWeights().DirectAffinity),
				Popularity:     envFloat("RECOMMENDATION_PERSONALIZED_POPULARITY_WEIGHT", domain.DefaultPersonalizationWeights().Popularity),
			},
		},
		ABTesting: ABTestingConfig{
			Enabled:             envBool("RECOMMENDATION_AB_TESTING_ENABLED", false),
			AssignmentTTL:       envDurationDays("RECOMMENDATION_AB_ASSIGNMENT_TTL_DAYS", defaultABAssignmentTTL),
			DefaultSalt:         envString("RECOMMENDATION_AB_DEFAULT_SALT", defaultABSalt),
			FailOpen:            envBool("RECOMMENDATION_AB_FAIL_OPEN", true),
			Experiments:         experiments,
			MaxIdentifierLength: envInt("RECOMMENDATION_MAX_IDENTIFIER_LENGTH", defaultMaxIdentifierLength),
		},
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.Service.Name == "" {
		return errors.New("SERVICE_NAME cannot be empty")
	}
	if c.Service.Env == "" {
		return errors.New("APP_ENV cannot be empty")
	}
	if err := c.HTTP.Validate(); err != nil {
		return fmt.Errorf("invalid http config: %w", err)
	}
	if err := c.GRPC.Validate(); err != nil {
		return fmt.Errorf("invalid grpc config: %w", err)
	}
	if err := c.Recommendation.Validate(); err != nil {
		return fmt.Errorf("invalid recommendation config: %w", err)
	}
	if err := c.Storage.Validate(); err != nil {
		return fmt.Errorf("invalid storage config: %w", err)
	}
	if err := c.Events.Validate(); err != nil {
		return fmt.Errorf("invalid event config: %w", err)
	}
	if err := c.Features.Validate(); err != nil {
		return fmt.Errorf("invalid feature config: %w", err)
	}
	if err := c.Ranking.Validate(); err != nil {
		return fmt.Errorf("invalid ranking config: %w", err)
	}
	if err := c.Personalization.Validate(c.Recommendation.DefaultLimit); err != nil {
		return fmt.Errorf("invalid personalization config: %w", err)
	}
	if err := c.ABTesting.Validate(); err != nil {
		return fmt.Errorf("invalid ab testing config: %w", err)
	}
	if c.Features.Enabled && c.Features.ProcessedEventRetention < c.Storage.Mongo.InteractionRetention {
		return errors.New("RECOMMENDATION_FEATURE_EVENT_RETENTION_SECONDS must be greater than or equal to RECOMMENDATION_INTERACTION_RETENTION_SECONDS")
	}
	return nil
}

func (c HTTPConfig) Validate() error {
	if strings.TrimSpace(c.Address) == "" {
		return errors.New("RECOMMENDATION_HTTP_ADDR cannot be empty")
	}
	if c.ReadTimeout <= 0 {
		return errors.New("RECOMMENDATION_HTTP_READ_TIMEOUT must be greater than zero")
	}
	if c.WriteTimeout <= 0 {
		return errors.New("RECOMMENDATION_HTTP_WRITE_TIMEOUT must be greater than zero")
	}
	if c.IdleTimeout <= 0 {
		return errors.New("RECOMMENDATION_HTTP_IDLE_TIMEOUT must be greater than zero")
	}
	if c.ShutdownTimeout <= 0 {
		return errors.New("RECOMMENDATION_SHUTDOWN_TIMEOUT must be greater than zero")
	}
	if c.MaxBodyBytes <= 0 {
		return errors.New("RECOMMENDATION_MAX_BODY_BYTES must be greater than zero")
	}
	return nil
}

func (c GRPCConfig) Validate() error {
	if strings.TrimSpace(c.Address) == "" {
		return errors.New("RECOMMENDATION_GRPC_ADDR cannot be empty")
	}
	if c.MaxRecvBytes <= 0 {
		return errors.New("RECOMMENDATION_GRPC_MAX_RECV_BYTES must be greater than zero")
	}
	if c.MaxSendBytes <= 0 {
		return errors.New("RECOMMENDATION_GRPC_MAX_SEND_BYTES must be greater than zero")
	}
	if c.DefaultDeadline <= 0 {
		return errors.New("RECOMMENDATION_GRPC_DEFAULT_DEADLINE must be greater than zero")
	}
	return nil
}

func (c RecommendationConfig) Validate() error {
	if c.DefaultLimit <= 0 {
		return errors.New("RECOMMENDATION_DEFAULT_LIMIT must be greater than zero")
	}
	if c.MaxLimit <= 0 {
		return errors.New("RECOMMENDATION_MAX_LIMIT must be greater than zero")
	}
	if c.DefaultLimit > c.MaxLimit {
		return errors.New("RECOMMENDATION_DEFAULT_LIMIT must be less than or equal to RECOMMENDATION_MAX_LIMIT")
	}
	if c.MaxIdentifierLength <= 0 {
		return errors.New("RECOMMENDATION_MAX_IDENTIFIER_LENGTH must be greater than zero")
	}
	return nil
}

func (c StorageConfig) Validate() error {
	if err := c.Mongo.Validate(); err != nil {
		return fmt.Errorf("invalid mongo config: %w", err)
	}
	if err := c.Redis.Validate(); err != nil {
		return fmt.Errorf("invalid redis config: %w", err)
	}
	if err := c.Cache.Validate(); err != nil {
		return fmt.Errorf("invalid cache config: %w", err)
	}
	return nil
}

func (c MongoConfig) Enabled() bool {
	return strings.TrimSpace(c.URI) != ""
}

func (c MongoConfig) Validate() error {
	if strings.TrimSpace(c.Database) == "" {
		return errors.New("RECOMMENDATION_MONGO_DATABASE cannot be empty")
	}
	if c.ConnectTimeout <= 0 {
		return errors.New("RECOMMENDATION_MONGO_CONNECT_TIMEOUT must be greater than zero")
	}
	if c.PingTimeout <= 0 {
		return errors.New("RECOMMENDATION_MONGO_PING_TIMEOUT must be greater than zero")
	}
	if c.InteractionRetention <= 0 {
		return errors.New("RECOMMENDATION_INTERACTION_RETENTION_SECONDS must be greater than zero")
	}
	return nil
}

func (c RedisConfig) Enabled() bool {
	return strings.TrimSpace(c.Addr) != ""
}

func (c RedisConfig) Validate() error {
	if c.DB < 0 {
		return errors.New("RECOMMENDATION_REDIS_DB cannot be negative")
	}
	if c.DialTimeout <= 0 {
		return errors.New("RECOMMENDATION_REDIS_DIAL_TIMEOUT must be greater than zero")
	}
	if c.ReadTimeout <= 0 {
		return errors.New("RECOMMENDATION_REDIS_READ_TIMEOUT must be greater than zero")
	}
	if c.WriteTimeout <= 0 {
		return errors.New("RECOMMENDATION_REDIS_WRITE_TIMEOUT must be greater than zero")
	}
	if c.PingTimeout <= 0 {
		return errors.New("RECOMMENDATION_REDIS_PING_TIMEOUT must be greater than zero")
	}
	return nil
}

func (c CacheConfig) Validate() error {
	if err := domain.ValidateCacheKeyPrefix(c.KeyPrefix); err != nil {
		return err
	}
	if c.DefaultTTL <= 0 {
		return errors.New("RECOMMENDATION_CACHE_TTL_SECONDS must be greater than zero")
	}
	if c.PersonalizedTTL <= 0 {
		return errors.New("RECOMMENDATION_PERSONALIZED_CACHE_TTL_SECONDS must be greater than zero")
	}
	if c.GuestTTL <= 0 {
		return errors.New("RECOMMENDATION_GUEST_CACHE_TTL_SECONDS must be greater than zero")
	}
	if c.RebuildLockTTL <= 0 {
		return errors.New("RECOMMENDATION_CACHE_REBUILD_LOCK_TTL must be greater than zero")
	}
	if c.DirtyTTL <= 0 {
		return errors.New("RECOMMENDATION_CACHE_DIRTY_TTL_SECONDS must be greater than zero")
	}
	return nil
}

func (c EventConfig) Active() bool {
	return c.Enabled && strings.ToLower(strings.TrimSpace(c.Provider)) != "disabled"
}

func (c FeatureConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.GuestProfileRetention <= 0 {
		return errors.New("RECOMMENDATION_GUEST_PROFILE_RETENTION_SECONDS must be greater than zero")
	}
	if c.UserProductRetention <= 0 {
		return errors.New("RECOMMENDATION_USER_PRODUCT_RETENTION_SECONDS must be greater than zero")
	}
	if c.ProcessedEventRetention <= 0 {
		return errors.New("RECOMMENDATION_FEATURE_EVENT_RETENTION_SECONDS must be greater than zero")
	}
	if c.RecentProductsLimit <= 0 {
		return errors.New("RECOMMENDATION_FEATURE_RECENT_PRODUCTS_LIMIT must be greater than zero")
	}
	if c.ReconcileBatchSize <= 0 {
		return errors.New("RECOMMENDATION_FEATURE_RECONCILE_BATCH_SIZE must be greater than zero")
	}
	if c.ReconcileInterval < 0 || c.WindowRebuildInterval < 0 {
		return errors.New("feature job intervals cannot be negative")
	}
	return nil
}

func (c RankingConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.RebuildInterval <= 0 {
		return errors.New("RECOMMENDATION_RANKING_REBUILD_INTERVAL must be greater than zero")
	}
	if strings.TrimSpace(c.FormulaVersion) == "" {
		return errors.New("RECOMMENDATION_RANKING_FORMULA_VERSION cannot be empty")
	}
	if err := c.Weights.Validate(); err != nil {
		return err
	}
	if c.Weights != domain.DefaultPopularityWeights() && c.FormulaVersion == domain.RuleBasedRankingFormulaVersion {
		return errors.New("RECOMMENDATION_RANKING_FORMULA_VERSION must change when ranking weights differ from popularity_v1")
	}
	return nil
}

func (c PersonalizationConfig) Validate(defaultLimit int) error {
	if !c.Enabled {
		return nil
	}
	if strings.TrimSpace(c.FormulaVersion) == "" {
		return errors.New("RECOMMENDATION_PERSONALIZED_FORMULA_VERSION cannot be empty")
	}
	if c.MinPositiveInteractions < 0 {
		return errors.New("RECOMMENDATION_PERSONALIZED_MIN_POSITIVE_INTERACTIONS cannot be negative")
	}
	if c.ProfileMaxAge <= 0 {
		return errors.New("RECOMMENDATION_PERSONALIZED_PROFILE_MAX_AGE_DAYS must be greater than zero")
	}
	if c.MaxCandidates <= 0 {
		return errors.New("RECOMMENDATION_PERSONALIZED_MAX_CANDIDATES must be greater than zero")
	}
	if defaultLimit > 0 && c.MaxCandidates < defaultLimit {
		return errors.New("RECOMMENDATION_PERSONALIZED_MAX_CANDIDATES must be greater than or equal to RECOMMENDATION_DEFAULT_LIMIT")
	}
	if c.MaxItemsPerSeller <= 0 {
		return errors.New("RECOMMENDATION_PERSONALIZED_MAX_ITEMS_PER_SELLER must be greater than zero")
	}
	if c.TopCategories <= 0 {
		return errors.New("RECOMMENDATION_PERSONALIZED_TOP_CATEGORIES must be greater than zero")
	}
	if c.TopSellers <= 0 {
		return errors.New("RECOMMENDATION_PERSONALIZED_TOP_SELLERS must be greater than zero")
	}
	if c.TopBrands <= 0 {
		return errors.New("RECOMMENDATION_PERSONALIZED_TOP_BRANDS must be greater than zero")
	}
	if c.DirectProductLimit <= 0 {
		return errors.New("RECOMMENDATION_PERSONALIZED_DIRECT_PRODUCT_LIMIT must be greater than zero")
	}
	if err := c.Weights.Validate(); err != nil {
		return err
	}
	if c.Weights != domain.DefaultPersonalizationWeights() && c.FormulaVersion == domain.PersonalizationFormulaVersion {
		return errors.New("RECOMMENDATION_PERSONALIZED_FORMULA_VERSION must change when behavior_v1 weights differ")
	}
	return nil
}

func (c ABTestingConfig) Validate() error {
	if c.AssignmentTTL <= 0 {
		return errors.New("RECOMMENDATION_AB_ASSIGNMENT_TTL_DAYS must be greater than zero")
	}
	if strings.TrimSpace(c.DefaultSalt) == "" {
		return errors.New("RECOMMENDATION_AB_DEFAULT_SALT cannot be empty")
	}
	if c.MaxIdentifierLength <= 0 {
		return errors.New("RECOMMENDATION_MAX_IDENTIFIER_LENGTH must be greater than zero")
	}
	seen := make(map[string]struct{}, len(c.Experiments))
	for _, experiment := range c.Experiments {
		experiment = experiment.Normalize(c.DefaultSalt)
		if err := experiment.Validate(); err != nil {
			return err
		}
		if _, ok := seen[experiment.ExperimentID]; ok {
			return fmt.Errorf("%w: duplicate experiment_id %q", domain.ErrInvalidExperiment, experiment.ExperimentID)
		}
		seen[experiment.ExperimentID] = struct{}{}
	}
	return nil
}

func (c EventConfig) Validate() error {
	if strings.TrimSpace(c.Provider) == "" {
		return errors.New("QUEUE_PROVIDER cannot be empty")
	}
	provider := strings.ToLower(strings.TrimSpace(c.Provider))
	if !c.Active() {
		return nil
	}
	if provider != "kafka" {
		return fmt.Errorf("QUEUE_PROVIDER %q is not supported by recommendation event ingestion", c.Provider)
	}
	if c.SupportedVersion <= 0 {
		return errors.New("EVENT_SUPPORTED_VERSION must be greater than zero")
	}
	if c.MaxMessageBytes <= 0 {
		return errors.New("RECOMMENDATION_EVENTS_MAX_MESSAGE_BYTES must be greater than zero")
	}
	if c.MaxRetryAttempts <= 0 {
		return errors.New("EVENT_MAX_RETRY_ATTEMPTS must be greater than zero")
	}
	for _, backoff := range c.RetryBackoffs {
		if backoff < 0 {
			return errors.New("EVENT_RETRY_BACKOFF_SECONDS cannot contain negative durations")
		}
	}
	if c.UnknownEventPolicy != "dlq" && c.UnknownEventPolicy != "skip" {
		return errors.New("RECOMMENDATION_EVENTS_UNKNOWN_POLICY must be dlq or skip")
	}
	return c.Kafka.Validate()
}

func (c KafkaEventConfig) Validate() error {
	if len(c.Brokers) == 0 {
		return errors.New("KAFKA_BROKERS cannot be empty when event ingestion is enabled")
	}
	if strings.TrimSpace(c.Topic) == "" {
		return errors.New("RECOMMENDATION_EVENTS_TOPIC cannot be empty")
	}
	if strings.TrimSpace(c.DLQTopic) == "" {
		return errors.New("RECOMMENDATION_EVENTS_DLQ_TOPIC cannot be empty")
	}
	if strings.TrimSpace(c.GroupID) == "" {
		return errors.New("RECOMMENDATION_EVENTS_GROUP cannot be empty")
	}
	if c.MinBytes <= 0 {
		return errors.New("RECOMMENDATION_KAFKA_MIN_BYTES must be greater than zero")
	}
	if c.MaxBytes <= 0 {
		return errors.New("RECOMMENDATION_KAFKA_MAX_BYTES must be greater than zero")
	}
	return nil
}

func envString(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envInt64(key string, fallback int64) int64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}
	return value
}

func envFloat(key string, fallback float64) float64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envDurationDays(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	days, err := strconv.ParseInt(raw, 10, 64)
	if err == nil && days > 0 {
		return time.Duration(days) * 24 * time.Hour
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func envDurationSeconds(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	seconds, err := strconv.ParseInt(raw, 10, 64)
	if err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envCSV(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}

func envExperimentDefinitions(key string) ([]domain.ExperimentDefinition, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil, nil
	}
	var experiments []domain.ExperimentDefinition
	if err := json.Unmarshal([]byte(raw), &experiments); err != nil {
		return nil, fmt.Errorf("%s must be valid experiment JSON: %w", key, err)
	}
	return experiments, nil
}

func envDurationListSeconds(key string, fallback []time.Duration) []time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return append([]time.Duration(nil), fallback...)
	}
	parts := strings.Split(raw, ",")
	values := make([]time.Duration, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if seconds, err := strconv.ParseInt(part, 10, 64); err == nil {
			values = append(values, time.Duration(seconds)*time.Second)
			continue
		}
		if duration, err := time.ParseDuration(part); err == nil {
			values = append(values, duration)
		}
	}
	if len(values) == 0 {
		return append([]time.Duration(nil), fallback...)
	}
	return values
}

func envLogLevel(key string, fallback slog.Level) slog.Level {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "debug":
		return slog.LevelDebug
	case "info", "":
		return fallback
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return fallback
	}
}
