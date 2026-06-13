package config

import (
	"errors"
	"fmt"
	neturl "net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPAddress                       = ":8085"
	defaultShutdownTimeout                   = 10 * time.Second
	defaultTypesenseHost                     = "localhost"
	defaultTypesensePort                     = 8108
	defaultTypesenseScheme                   = "http"
	defaultTypesenseProductsCollection       = "products"
	defaultTypesensePopularQueriesCollection = "popular_queries"
	defaultTypesenseRequestTimeout           = 300 * time.Millisecond
	defaultProductServiceURL                 = "http://localhost:8082"
	defaultProductServiceBatchGetPath        = "/internal/v1/products:batchGet"
	defaultProductServiceSearchExportPath    = "/internal/v1/products/search-export"
	defaultProductServiceTimeout             = 300 * time.Millisecond
	defaultProductServiceSearchExportTimeout = 2 * time.Second
	defaultSearchDefaultPageSize             = 20
	defaultSearchMaxPageSize                 = 100
	defaultAutocompleteDefaultLimit          = 8
	defaultAutocompleteMaxLimit              = 10
	defaultAutocompletePrefixCacheTTL        = time.Minute
	defaultAutocompleteEmptyCacheTTL         = 5 * time.Minute
	defaultAutocompleteTypesenseTimeout      = 150 * time.Millisecond
	defaultAutocompleteCacheTimeout          = 50 * time.Millisecond
	defaultZeroResultTrackingEnabled         = true
	defaultZeroResultDedupeTTL               = 30 * time.Minute
	defaultZeroResultSendTimeout             = 150 * time.Millisecond
	defaultZeroResultQueueSize               = 1024
	defaultZeroResultWorkerCount             = 2
	defaultSearchAdminTimeout                = 500 * time.Millisecond
	defaultSearchAdminMutationRateLimit      = 30
	defaultSearchAdminMutationRateWindow     = time.Minute
	defaultSessionServiceURL                 = "http://localhost:8086"
	defaultSessionServiceIngestPath          = "/api/v1/sessions/events"
	defaultSessionServiceTimeout             = 150 * time.Millisecond
	defaultQueueProvider                     = "rabbitmq"
	defaultRabbitMQURL                       = "amqp://ecommerce:ecommerce_password@localhost:5672/ecommerce"
	defaultProductEventsExchange             = "product.events"
	defaultProductIndexerQueue               = "search-service.product-indexer"
	defaultProductIndexerDLX                 = "product.events.dlx"
	defaultProductIndexerDLQ                 = "search-service.product-indexer.dlq"
	defaultProductIndexerConsumerTag         = "search-service-product-indexer"
	defaultProductIndexerPrefetch            = 20
	defaultProductIndexerMaxRetries          = 5
	defaultRedisAddr                         = "localhost:6379"
	defaultProcessedEventTTL                 = 30 * 24 * time.Hour
	defaultReindexMode                       = "alias"
	defaultReindexBatchSize                  = 500
	defaultReindexMaxBatchSize               = 5000
	defaultReindexLockTTL                    = 3 * time.Hour
	defaultReindexJobTimeout                 = 2 * time.Hour
	defaultReindexOldCollectionRetention     = 72 * time.Hour
	defaultReindexImportTimeout              = 5 * time.Second
)

type Config struct {
	HTTP      HTTPConfig
	Typesense TypesenseConfig
	Search    SearchConfig
	Admin     AdminConfig
	Product   ProductServiceConfig
	Session   SessionServiceConfig
	Indexer   IndexerConfig
	Queue     QueueConfig
	Redis     RedisConfig
	Reindex   ReindexConfig
}

type HTTPConfig struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type TypesenseConfig struct {
	URL                      string
	Host                     string
	Port                     int
	Protocol                 string
	APIKey                   string
	ProductsCollection       string
	PopularQueriesCollection string
	RequestTimeout           time.Duration
}

type SearchConfig struct {
	DefaultPageSize            int
	MaxPageSize                int
	AutocompleteDefaultLimit   int
	AutocompleteMaxLimit       int
	AutocompletePrefixCacheTTL time.Duration
	AutocompleteEmptyCacheTTL  time.Duration
	AutocompleteSearchTimeout  time.Duration
	AutocompleteCacheTimeout   time.Duration
	ZeroResultTrackingEnabled  bool
	ZeroResultDedupeTTL        time.Duration
	ZeroResultSendTimeout      time.Duration
	ZeroResultQueueSize        int
	ZeroResultWorkerCount      int
}

type AdminConfig struct {
	Timeout            time.Duration
	AuthEnabled        bool
	MutationRateLimit  int
	MutationRateWindow time.Duration
}

type ProductServiceConfig struct {
	URL                 string
	BatchGetPath        string
	SearchExportPath    string
	Timeout             time.Duration
	SearchExportTimeout time.Duration
}

type SessionServiceConfig struct {
	URL        string
	IngestPath string
	Timeout    time.Duration
}

type IndexerConfig struct {
	Enabled           bool
	MessageTimeout    time.Duration
	ProcessedEventTTL time.Duration
}

type QueueConfig struct {
	Provider                     string
	RabbitMQURL                  string
	ProductEventsExchange        string
	ProductIndexerQueue          string
	ProductIndexerDLX            string
	ProductIndexerDLQ            string
	ProductIndexerConsumerTag    string
	ProductIndexerRoutingKeys    []string
	ProductIndexerPrefetch       int
	ProductIndexerMaxRetries     int
	ProductIndexerReconnectDelay time.Duration
	ProductIndexerRetryBaseDelay time.Duration
	ProductIndexerRetryMaxDelay  time.Duration
}

type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type ReindexConfig struct {
	Mode                   string
	BatchSize              int
	MaxBatchSize           int
	LockTTL                time.Duration
	JobTimeout             time.Duration
	CollectionPrefix       string
	OldCollectionRetention time.Duration
	ImportTimeout          time.Duration
}

func Load() (Config, error) {
	typesensePort, err := envInt("TYPESENSE_PORT", defaultTypesensePort)
	if err != nil {
		return Config{}, err
	}
	redisDB, err := envInt("SEARCH_REDIS_DB", 0)
	if err != nil {
		return Config{}, err
	}
	prefetch, err := envInt("PRODUCT_INDEXER_PREFETCH", defaultProductIndexerPrefetch)
	if err != nil {
		return Config{}, err
	}
	maxRetries, err := envInt("PRODUCT_INDEXER_MAX_RETRIES", defaultProductIndexerMaxRetries)
	if err != nil {
		return Config{}, err
	}
	defaultPageSize, err := envInt("SEARCH_DEFAULT_PAGE_SIZE", defaultSearchDefaultPageSize)
	if err != nil {
		return Config{}, err
	}
	maxPageSize, err := envInt("SEARCH_MAX_PAGE_SIZE", defaultSearchMaxPageSize)
	if err != nil {
		return Config{}, err
	}
	autocompleteDefaultLimit, err := envInt("SEARCH_AUTOCOMPLETE_DEFAULT_LIMIT", defaultAutocompleteDefaultLimit)
	if err != nil {
		return Config{}, err
	}
	autocompleteMaxLimit, err := envInt("SEARCH_AUTOCOMPLETE_MAX_LIMIT", defaultAutocompleteMaxLimit)
	if err != nil {
		return Config{}, err
	}
	adminMutationRateLimit, err := envInt("SEARCH_ADMIN_MUTATION_RATE_LIMIT", defaultSearchAdminMutationRateLimit)
	if err != nil {
		return Config{}, err
	}
	zeroResultQueueSize, err := envInt("SEARCH_ZERO_RESULT_QUEUE_SIZE", defaultZeroResultQueueSize)
	if err != nil {
		return Config{}, err
	}
	zeroResultWorkerCount, err := envInt("SEARCH_ZERO_RESULT_WORKERS", defaultZeroResultWorkerCount)
	if err != nil {
		return Config{}, err
	}
	reindexBatchSize, err := envInt("SEARCH_REINDEX_BATCH_SIZE", defaultReindexBatchSize)
	if err != nil {
		return Config{}, err
	}
	reindexMaxBatchSize, err := envInt("SEARCH_REINDEX_MAX_BATCH_SIZE", defaultReindexMaxBatchSize)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTP: HTTPConfig{
			Address:         envString("SEARCH_HTTP_ADDR", defaultHTTPAddress),
			ReadTimeout:     envDuration("SEARCH_HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:    envDuration("SEARCH_HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     envDuration("SEARCH_HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: envDuration("SEARCH_SHUTDOWN_TIMEOUT", defaultShutdownTimeout),
		},
		Typesense: TypesenseConfig{
			URL:                      envString("TYPESENSE_URL", ""),
			Host:                     envString("TYPESENSE_HOST", defaultTypesenseHost),
			Port:                     typesensePort,
			Protocol:                 envString("TYPESENSE_PROTOCOL", defaultTypesenseScheme),
			APIKey:                   os.Getenv("TYPESENSE_API_KEY"),
			ProductsCollection:       envString("TYPESENSE_PRODUCTS_COLLECTION", defaultTypesenseProductsCollection),
			PopularQueriesCollection: envString("TYPESENSE_POPULAR_QUERIES_COLLECTION", defaultTypesensePopularQueriesCollection),
			RequestTimeout:           envDurationMS("TYPESENSE_TIMEOUT_MS", defaultTypesenseRequestTimeout),
		},
		Search: SearchConfig{
			DefaultPageSize:            defaultPageSize,
			MaxPageSize:                maxPageSize,
			AutocompleteDefaultLimit:   autocompleteDefaultLimit,
			AutocompleteMaxLimit:       autocompleteMaxLimit,
			AutocompletePrefixCacheTTL: envDuration("SEARCH_AUTOCOMPLETE_PREFIX_CACHE_TTL", defaultAutocompletePrefixCacheTTL),
			AutocompleteEmptyCacheTTL:  envDuration("SEARCH_AUTOCOMPLETE_EMPTY_CACHE_TTL", defaultAutocompleteEmptyCacheTTL),
			AutocompleteSearchTimeout:  envDurationMS("SEARCH_AUTOCOMPLETE_TYPESENSE_TIMEOUT_MS", defaultAutocompleteTypesenseTimeout),
			AutocompleteCacheTimeout:   envDurationMS("SEARCH_AUTOCOMPLETE_CACHE_TIMEOUT_MS", defaultAutocompleteCacheTimeout),
			ZeroResultTrackingEnabled:  envBool("SEARCH_ZERO_RESULT_TRACKING_ENABLED", defaultZeroResultTrackingEnabled),
			ZeroResultDedupeTTL:        envDuration("SEARCH_ZERO_RESULT_DEDUPE_TTL", defaultZeroResultDedupeTTL),
			ZeroResultSendTimeout:      envDurationMS("SEARCH_ZERO_RESULT_SEND_TIMEOUT_MS", defaultZeroResultSendTimeout),
			ZeroResultQueueSize:        zeroResultQueueSize,
			ZeroResultWorkerCount:      zeroResultWorkerCount,
		},
		Admin: AdminConfig{
			Timeout:            envDurationMS("SEARCH_ADMIN_TIMEOUT_MS", defaultSearchAdminTimeout),
			AuthEnabled:        envBool("SEARCH_ADMIN_AUTH_ENABLED", true),
			MutationRateLimit:  adminMutationRateLimit,
			MutationRateWindow: envDuration("SEARCH_ADMIN_MUTATION_RATE_WINDOW", defaultSearchAdminMutationRateWindow),
		},
		Product: ProductServiceConfig{
			URL:                 envString("PRODUCT_SERVICE_URL", defaultProductServiceURL),
			BatchGetPath:        envString("PRODUCT_SERVICE_BATCH_GET_PATH", defaultProductServiceBatchGetPath),
			SearchExportPath:    envString("PRODUCT_SERVICE_SEARCH_EXPORT_PATH", defaultProductServiceSearchExportPath),
			Timeout:             envDurationMS("PRODUCT_SERVICE_TIMEOUT_MS", defaultProductServiceTimeout),
			SearchExportTimeout: envDurationMS("PRODUCT_SERVICE_SEARCH_EXPORT_TIMEOUT_MS", defaultProductServiceSearchExportTimeout),
		},
		Session: SessionServiceConfig{
			URL:        envString("SESSION_SERVICE_URL", defaultSessionServiceURL),
			IngestPath: envString("SESSION_SERVICE_INGEST_PATH", defaultSessionServiceIngestPath),
			Timeout:    envDurationMS("SESSION_SERVICE_TIMEOUT_MS", defaultSessionServiceTimeout),
		},
		Indexer: IndexerConfig{
			Enabled:           envBool("SEARCH_INDEXER_ENABLED", true),
			MessageTimeout:    envDuration("PRODUCT_INDEXER_MESSAGE_TIMEOUT", 30*time.Second),
			ProcessedEventTTL: envDuration("PRODUCT_INDEXER_PROCESSED_EVENT_TTL", defaultProcessedEventTTL),
		},
		Queue: QueueConfig{
			Provider:                     strings.ToLower(strings.TrimSpace(envString("QUEUE_PROVIDER", defaultQueueProvider))),
			RabbitMQURL:                  envString("RABBITMQ_URL", defaultRabbitMQURL),
			ProductEventsExchange:        envString("PRODUCT_EVENTS_EXCHANGE", defaultProductEventsExchange),
			ProductIndexerQueue:          envString("PRODUCT_INDEXER_QUEUE", defaultProductIndexerQueue),
			ProductIndexerDLX:            envString("PRODUCT_INDEXER_DLX", defaultProductIndexerDLX),
			ProductIndexerDLQ:            envString("PRODUCT_INDEXER_DLQ", defaultProductIndexerDLQ),
			ProductIndexerConsumerTag:    envString("PRODUCT_INDEXER_CONSUMER_TAG", defaultProductIndexerConsumerTag),
			ProductIndexerRoutingKeys:    envCSV("PRODUCT_INDEXER_ROUTING_KEYS", defaultProductIndexerRoutingKeys()),
			ProductIndexerPrefetch:       prefetch,
			ProductIndexerMaxRetries:     maxRetries,
			ProductIndexerReconnectDelay: envDuration("PRODUCT_INDEXER_RECONNECT_DELAY", 5*time.Second),
			ProductIndexerRetryBaseDelay: envDuration("PRODUCT_INDEXER_RETRY_BASE_DELAY", time.Second),
			ProductIndexerRetryMaxDelay:  envDuration("PRODUCT_INDEXER_RETRY_MAX_DELAY", 5*time.Minute),
		},
		Redis: RedisConfig{
			Addr:         envString("SEARCH_REDIS_ADDR", envString("REDIS_ADDR", defaultRedisAddr)),
			Password:     envString("SEARCH_REDIS_PASSWORD", os.Getenv("REDIS_PASSWORD")),
			DB:           redisDB,
			DialTimeout:  envDuration("SEARCH_REDIS_DIAL_TIMEOUT", 2*time.Second),
			ReadTimeout:  envDuration("SEARCH_REDIS_READ_TIMEOUT", 2*time.Second),
			WriteTimeout: envDuration("SEARCH_REDIS_WRITE_TIMEOUT", 2*time.Second),
		},
		Reindex: ReindexConfig{
			Mode:                   strings.ToLower(strings.TrimSpace(envString("SEARCH_REINDEX_MODE", defaultReindexMode))),
			BatchSize:              reindexBatchSize,
			MaxBatchSize:           reindexMaxBatchSize,
			LockTTL:                envDuration("SEARCH_REINDEX_LOCK_TTL", defaultReindexLockTTL),
			JobTimeout:             envDuration("SEARCH_REINDEX_JOB_TIMEOUT", defaultReindexJobTimeout),
			CollectionPrefix:       envString("SEARCH_REINDEX_COLLECTION_PREFIX", defaultTypesenseProductsCollection),
			OldCollectionRetention: envDuration("SEARCH_REINDEX_OLD_COLLECTION_RETENTION", defaultReindexOldCollectionRetention),
			ImportTimeout:          envDurationMS("TYPESENSE_REINDEX_IMPORT_TIMEOUT_MS", defaultReindexImportTimeout),
		},
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.HTTP.Address == "" {
		return errors.New("SEARCH_HTTP_ADDR cannot be empty")
	}
	if c.HTTP.ReadTimeout <= 0 {
		return errors.New("SEARCH_HTTP_READ_TIMEOUT must be greater than zero")
	}
	if c.HTTP.WriteTimeout <= 0 {
		return errors.New("SEARCH_HTTP_WRITE_TIMEOUT must be greater than zero")
	}
	if c.HTTP.IdleTimeout <= 0 {
		return errors.New("SEARCH_HTTP_IDLE_TIMEOUT must be greater than zero")
	}
	if c.HTTP.ShutdownTimeout <= 0 {
		return errors.New("SEARCH_SHUTDOWN_TIMEOUT must be greater than zero")
	}
	if err := c.Typesense.Validate(); err != nil {
		return err
	}
	if err := c.Search.Validate(); err != nil {
		return err
	}
	if err := c.Admin.Validate(); err != nil {
		return err
	}
	if err := c.Product.Validate(); err != nil {
		return err
	}
	if c.Search.ZeroResultTrackingEnabled {
		if err := c.Session.Validate(); err != nil {
			return err
		}
	}
	if err := c.Indexer.Validate(); err != nil {
		return err
	}
	if err := c.Redis.Validate(); err != nil {
		return err
	}
	if err := c.Reindex.Validate(); err != nil {
		return err
	}
	if c.Indexer.Enabled {
		if err := c.Queue.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c TypesenseConfig) Validate() error {
	if strings.TrimSpace(c.URL) != "" {
		parsed, err := neturl.Parse(strings.TrimSpace(c.URL))
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return errors.New("TYPESENSE_URL must be a valid absolute URL")
		}
		switch strings.ToLower(parsed.Scheme) {
		case "http", "https":
		default:
			return errors.New("TYPESENSE_URL must use http or https")
		}
	} else {
		if strings.TrimSpace(c.Host) == "" {
			return errors.New("TYPESENSE_HOST cannot be empty")
		}
		if strings.Contains(c.Host, "://") {
			return errors.New("TYPESENSE_HOST must not include a protocol")
		}
		if c.Port <= 0 || c.Port > 65535 {
			return errors.New("TYPESENSE_PORT must be between 1 and 65535")
		}
		switch strings.ToLower(strings.TrimSpace(c.Protocol)) {
		case "http", "https":
		default:
			return errors.New("TYPESENSE_PROTOCOL must be http or https")
		}
	}
	if strings.TrimSpace(c.APIKey) == "" {
		return errors.New("TYPESENSE_API_KEY cannot be empty")
	}
	if strings.TrimSpace(c.ProductsCollection) == "" {
		return errors.New("TYPESENSE_PRODUCTS_COLLECTION cannot be empty")
	}
	if strings.TrimSpace(c.PopularQueriesCollection) == "" {
		return errors.New("TYPESENSE_POPULAR_QUERIES_COLLECTION cannot be empty")
	}
	if c.RequestTimeout <= 0 {
		return errors.New("TYPESENSE_TIMEOUT_MS must be greater than zero")
	}
	return nil
}

func (c TypesenseConfig) Endpoint() string {
	if url := strings.TrimRight(strings.TrimSpace(c.URL), "/"); url != "" {
		return url
	}
	protocol := strings.ToLower(strings.TrimSpace(c.Protocol))
	host := strings.TrimSpace(c.Host)
	return protocol + "://" + host + ":" + strconv.Itoa(c.Port)
}

func (c SearchConfig) Validate() error {
	if c.DefaultPageSize <= 0 {
		return errors.New("SEARCH_DEFAULT_PAGE_SIZE must be greater than zero")
	}
	if c.MaxPageSize <= 0 {
		return errors.New("SEARCH_MAX_PAGE_SIZE must be greater than zero")
	}
	if c.DefaultPageSize > c.MaxPageSize {
		return errors.New("SEARCH_DEFAULT_PAGE_SIZE must be less than or equal to SEARCH_MAX_PAGE_SIZE")
	}
	if c.AutocompleteDefaultLimit <= 0 {
		return errors.New("SEARCH_AUTOCOMPLETE_DEFAULT_LIMIT must be greater than zero")
	}
	if c.AutocompleteMaxLimit <= 0 {
		return errors.New("SEARCH_AUTOCOMPLETE_MAX_LIMIT must be greater than zero")
	}
	if c.AutocompleteMaxLimit > defaultAutocompleteMaxLimit {
		return errors.New("SEARCH_AUTOCOMPLETE_MAX_LIMIT must be less than or equal to 10")
	}
	if c.AutocompleteDefaultLimit > c.AutocompleteMaxLimit {
		return errors.New("SEARCH_AUTOCOMPLETE_DEFAULT_LIMIT must be less than or equal to SEARCH_AUTOCOMPLETE_MAX_LIMIT")
	}
	if c.AutocompletePrefixCacheTTL <= 0 {
		return errors.New("SEARCH_AUTOCOMPLETE_PREFIX_CACHE_TTL must be greater than zero")
	}
	if c.AutocompleteEmptyCacheTTL <= 0 {
		return errors.New("SEARCH_AUTOCOMPLETE_EMPTY_CACHE_TTL must be greater than zero")
	}
	if c.AutocompleteSearchTimeout <= 0 {
		return errors.New("SEARCH_AUTOCOMPLETE_TYPESENSE_TIMEOUT_MS must be greater than zero")
	}
	if c.AutocompleteCacheTimeout <= 0 {
		return errors.New("SEARCH_AUTOCOMPLETE_CACHE_TIMEOUT_MS must be greater than zero")
	}
	if c.ZeroResultDedupeTTL <= 0 {
		return errors.New("SEARCH_ZERO_RESULT_DEDUPE_TTL must be greater than zero")
	}
	if c.ZeroResultSendTimeout <= 0 {
		return errors.New("SEARCH_ZERO_RESULT_SEND_TIMEOUT_MS must be greater than zero")
	}
	if c.ZeroResultQueueSize <= 0 {
		return errors.New("SEARCH_ZERO_RESULT_QUEUE_SIZE must be greater than zero")
	}
	if c.ZeroResultWorkerCount <= 0 {
		return errors.New("SEARCH_ZERO_RESULT_WORKERS must be greater than zero")
	}
	return nil
}

func (c AdminConfig) Validate() error {
	if c.Timeout <= 0 {
		return errors.New("SEARCH_ADMIN_TIMEOUT_MS must be greater than zero")
	}
	if c.MutationRateLimit <= 0 {
		return errors.New("SEARCH_ADMIN_MUTATION_RATE_LIMIT must be greater than zero")
	}
	if c.MutationRateWindow <= 0 {
		return errors.New("SEARCH_ADMIN_MUTATION_RATE_WINDOW must be greater than zero")
	}
	return nil
}

func (c ProductServiceConfig) Validate() error {
	if strings.TrimSpace(c.URL) == "" {
		return errors.New("PRODUCT_SERVICE_URL cannot be empty")
	}
	if strings.TrimSpace(c.BatchGetPath) == "" {
		return errors.New("PRODUCT_SERVICE_BATCH_GET_PATH cannot be empty")
	}
	if strings.TrimSpace(c.SearchExportPath) == "" {
		return errors.New("PRODUCT_SERVICE_SEARCH_EXPORT_PATH cannot be empty")
	}
	if c.Timeout <= 0 {
		return errors.New("PRODUCT_SERVICE_TIMEOUT_MS must be greater than zero")
	}
	if c.SearchExportTimeout <= 0 {
		return errors.New("PRODUCT_SERVICE_SEARCH_EXPORT_TIMEOUT_MS must be greater than zero")
	}
	return nil
}

func (c SessionServiceConfig) Validate() error {
	parsed, err := neturl.Parse(strings.TrimSpace(c.URL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("SESSION_SERVICE_URL must be a valid absolute URL")
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
	default:
		return errors.New("SESSION_SERVICE_URL must use http or https")
	}
	if strings.TrimSpace(c.IngestPath) == "" {
		return errors.New("SESSION_SERVICE_INGEST_PATH cannot be empty")
	}
	if c.Timeout <= 0 {
		return errors.New("SESSION_SERVICE_TIMEOUT_MS must be greater than zero")
	}
	return nil
}

func (c IndexerConfig) Validate() error {
	if c.MessageTimeout <= 0 {
		return errors.New("PRODUCT_INDEXER_MESSAGE_TIMEOUT must be greater than zero")
	}
	if c.ProcessedEventTTL <= 0 {
		return errors.New("PRODUCT_INDEXER_PROCESSED_EVENT_TTL must be greater than zero")
	}
	return nil
}

func (c QueueConfig) Validate() error {
	if strings.TrimSpace(c.Provider) == "" {
		return errors.New("QUEUE_PROVIDER cannot be empty")
	}
	if c.Provider != "rabbitmq" {
		return errors.New("QUEUE_PROVIDER must be rabbitmq for the product indexer")
	}
	if strings.TrimSpace(c.RabbitMQURL) == "" {
		return errors.New("RABBITMQ_URL cannot be empty")
	}
	if strings.TrimSpace(c.ProductEventsExchange) == "" {
		return errors.New("PRODUCT_EVENTS_EXCHANGE cannot be empty")
	}
	if strings.TrimSpace(c.ProductIndexerQueue) == "" {
		return errors.New("PRODUCT_INDEXER_QUEUE cannot be empty")
	}
	if strings.TrimSpace(c.ProductIndexerDLX) == "" {
		return errors.New("PRODUCT_INDEXER_DLX cannot be empty")
	}
	if strings.TrimSpace(c.ProductIndexerDLQ) == "" {
		return errors.New("PRODUCT_INDEXER_DLQ cannot be empty")
	}
	if len(c.ProductIndexerRoutingKeys) == 0 {
		return errors.New("PRODUCT_INDEXER_ROUTING_KEYS cannot be empty")
	}
	if c.ProductIndexerPrefetch <= 0 {
		return errors.New("PRODUCT_INDEXER_PREFETCH must be greater than zero")
	}
	if c.ProductIndexerMaxRetries < 0 {
		return errors.New("PRODUCT_INDEXER_MAX_RETRIES cannot be negative")
	}
	if c.ProductIndexerReconnectDelay <= 0 {
		return errors.New("PRODUCT_INDEXER_RECONNECT_DELAY must be greater than zero")
	}
	if c.ProductIndexerRetryBaseDelay <= 0 {
		return errors.New("PRODUCT_INDEXER_RETRY_BASE_DELAY must be greater than zero")
	}
	if c.ProductIndexerRetryMaxDelay < c.ProductIndexerRetryBaseDelay {
		return errors.New("PRODUCT_INDEXER_RETRY_MAX_DELAY must be greater than or equal to base delay")
	}
	return nil
}

func (c RedisConfig) Validate() error {
	if strings.TrimSpace(c.Addr) == "" {
		return errors.New("SEARCH_REDIS_ADDR cannot be empty")
	}
	if c.DB < 0 {
		return errors.New("SEARCH_REDIS_DB cannot be negative")
	}
	if c.DialTimeout <= 0 {
		return errors.New("SEARCH_REDIS_DIAL_TIMEOUT must be greater than zero")
	}
	if c.ReadTimeout <= 0 {
		return errors.New("SEARCH_REDIS_READ_TIMEOUT must be greater than zero")
	}
	if c.WriteTimeout <= 0 {
		return errors.New("SEARCH_REDIS_WRITE_TIMEOUT must be greater than zero")
	}
	return nil
}

func (c ReindexConfig) Validate() error {
	switch strings.ToLower(strings.TrimSpace(c.Mode)) {
	case "alias", "in_place":
	default:
		return errors.New("SEARCH_REINDEX_MODE must be alias or in_place")
	}
	if c.BatchSize <= 0 {
		return errors.New("SEARCH_REINDEX_BATCH_SIZE must be greater than zero")
	}
	if c.MaxBatchSize <= 0 {
		return errors.New("SEARCH_REINDEX_MAX_BATCH_SIZE must be greater than zero")
	}
	if c.BatchSize > c.MaxBatchSize {
		return errors.New("SEARCH_REINDEX_BATCH_SIZE must be less than or equal to SEARCH_REINDEX_MAX_BATCH_SIZE")
	}
	if c.LockTTL <= 0 {
		return errors.New("SEARCH_REINDEX_LOCK_TTL must be greater than zero")
	}
	if c.JobTimeout <= 0 {
		return errors.New("SEARCH_REINDEX_JOB_TIMEOUT must be greater than zero")
	}
	if strings.TrimSpace(c.CollectionPrefix) == "" {
		return errors.New("SEARCH_REINDEX_COLLECTION_PREFIX cannot be empty")
	}
	if c.OldCollectionRetention < 0 {
		return errors.New("SEARCH_REINDEX_OLD_COLLECTION_RETENTION cannot be negative")
	}
	if c.ImportTimeout <= 0 {
		return errors.New("TYPESENSE_REINDEX_IMPORT_TIMEOUT_MS must be greater than zero")
	}
	return nil
}

func envString(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return value, nil
}

func envBool(key string, fallback bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envCSV(key string, fallback []string) []string {
	raw := os.Getenv(key)
	if raw == "" {
		return append([]string(nil), fallback...)
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	if len(values) == 0 {
		return append([]string(nil), fallback...)
	}
	return values
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err == nil {
		return value
	}
	if seconds, err := strconv.Atoi(raw); err == nil {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}

func envDurationMS(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err == nil {
		return value
	}
	if milliseconds, err := strconv.Atoi(raw); err == nil {
		return time.Duration(milliseconds) * time.Millisecond
	}
	return fallback
}

func defaultProductIndexerRoutingKeys() []string {
	return []string{
		"product.published",
		"product.updated",
		"product.price_changed",
		"product.inventory_changed",
		"product.unpublished",
		"product.deleted",
		"product.blocked",
	}
}
