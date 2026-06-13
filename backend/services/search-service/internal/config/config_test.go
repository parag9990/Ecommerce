package config

import (
	"testing"
	"time"
)

func TestLoadTypesenseConfig(t *testing.T) {
	t.Setenv("TYPESENSE_HOST", "typesense")
	t.Setenv("TYPESENSE_PORT", "8108")
	t.Setenv("TYPESENSE_PROTOCOL", "http")
	t.Setenv("TYPESENSE_API_KEY", "dev-typesense-key")
	t.Setenv("TYPESENSE_PRODUCTS_COLLECTION", "products")
	t.Setenv("TYPESENSE_POPULAR_QUERIES_COLLECTION", "popular_queries")
	t.Setenv("TYPESENSE_TIMEOUT_MS", "300")
	t.Setenv("PRODUCT_SERVICE_URL", "http://product-service:8082")
	t.Setenv("PRODUCT_SERVICE_SEARCH_EXPORT_PATH", "/internal/v1/products/search-export")
	t.Setenv("PRODUCT_SERVICE_TIMEOUT_MS", "250")
	t.Setenv("PRODUCT_SERVICE_SEARCH_EXPORT_TIMEOUT_MS", "2000")
	t.Setenv("SEARCH_DEFAULT_PAGE_SIZE", "20")
	t.Setenv("SEARCH_MAX_PAGE_SIZE", "100")
	t.Setenv("SEARCH_AUTOCOMPLETE_DEFAULT_LIMIT", "8")
	t.Setenv("SEARCH_AUTOCOMPLETE_MAX_LIMIT", "10")
	t.Setenv("SEARCH_AUTOCOMPLETE_PREFIX_CACHE_TTL", "1m")
	t.Setenv("SEARCH_AUTOCOMPLETE_EMPTY_CACHE_TTL", "5m")
	t.Setenv("SEARCH_AUTOCOMPLETE_TYPESENSE_TIMEOUT_MS", "150")
	t.Setenv("SEARCH_AUTOCOMPLETE_CACHE_TIMEOUT_MS", "50")
	t.Setenv("SEARCH_ZERO_RESULT_TRACKING_ENABLED", "true")
	t.Setenv("SEARCH_ZERO_RESULT_DEDUPE_TTL", "30m")
	t.Setenv("SEARCH_ZERO_RESULT_SEND_TIMEOUT_MS", "150")
	t.Setenv("SEARCH_ZERO_RESULT_QUEUE_SIZE", "256")
	t.Setenv("SEARCH_ZERO_RESULT_WORKERS", "3")
	t.Setenv("SESSION_SERVICE_URL", "http://session-service:8086")
	t.Setenv("SESSION_SERVICE_INGEST_PATH", "/api/v1/sessions/events")
	t.Setenv("SESSION_SERVICE_TIMEOUT_MS", "150")
	t.Setenv("SEARCH_ADMIN_TIMEOUT_MS", "500")
	t.Setenv("SEARCH_ADMIN_AUTH_ENABLED", "true")
	t.Setenv("SEARCH_ADMIN_MUTATION_RATE_LIMIT", "30")
	t.Setenv("SEARCH_ADMIN_MUTATION_RATE_WINDOW", "1m")
	t.Setenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	t.Setenv("SEARCH_REDIS_ADDR", "localhost:6379")
	t.Setenv("SEARCH_REINDEX_MODE", "alias")
	t.Setenv("SEARCH_REINDEX_BATCH_SIZE", "500")
	t.Setenv("SEARCH_REINDEX_MAX_BATCH_SIZE", "5000")
	t.Setenv("SEARCH_REINDEX_LOCK_TTL", "3h")
	t.Setenv("SEARCH_REINDEX_JOB_TIMEOUT", "2h")
	t.Setenv("SEARCH_REINDEX_COLLECTION_PREFIX", "products")
	t.Setenv("SEARCH_REINDEX_OLD_COLLECTION_RETENTION", "72h")
	t.Setenv("TYPESENSE_REINDEX_IMPORT_TIMEOUT_MS", "5000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Typesense.Endpoint() != "http://typesense:8108" {
		t.Fatalf("typesense endpoint = %q", cfg.Typesense.Endpoint())
	}
	if cfg.Typesense.APIKey != "dev-typesense-key" {
		t.Fatal("typesense api key was not loaded")
	}
	if cfg.Typesense.ProductsCollection != "products" {
		t.Fatalf("products collection = %q", cfg.Typesense.ProductsCollection)
	}
	if cfg.Typesense.PopularQueriesCollection != "popular_queries" {
		t.Fatalf("popular queries collection = %q", cfg.Typesense.PopularQueriesCollection)
	}
	if cfg.Typesense.RequestTimeout != 300*time.Millisecond {
		t.Fatalf("typesense timeout = %s", cfg.Typesense.RequestTimeout)
	}
	if cfg.Product.URL != "http://product-service:8082" {
		t.Fatalf("product service url = %q", cfg.Product.URL)
	}
	if cfg.Product.Timeout != 250*time.Millisecond {
		t.Fatalf("product timeout = %s", cfg.Product.Timeout)
	}
	if cfg.Product.SearchExportPath != "/internal/v1/products/search-export" || cfg.Product.SearchExportTimeout != 2*time.Second {
		t.Fatalf("product search export config = %#v", cfg.Product)
	}
	if cfg.Search.MaxPageSize != 100 {
		t.Fatalf("max page size = %d", cfg.Search.MaxPageSize)
	}
	if cfg.Search.AutocompleteDefaultLimit != 8 || cfg.Search.AutocompleteMaxLimit != 10 {
		t.Fatalf("autocomplete limits = %d/%d", cfg.Search.AutocompleteDefaultLimit, cfg.Search.AutocompleteMaxLimit)
	}
	if cfg.Search.AutocompletePrefixCacheTTL != time.Minute || cfg.Search.AutocompleteEmptyCacheTTL != 5*time.Minute {
		t.Fatalf("autocomplete cache ttls = %s/%s", cfg.Search.AutocompletePrefixCacheTTL, cfg.Search.AutocompleteEmptyCacheTTL)
	}
	if cfg.Search.AutocompleteSearchTimeout != 150*time.Millisecond || cfg.Search.AutocompleteCacheTimeout != 50*time.Millisecond {
		t.Fatalf("autocomplete timeouts = %s/%s", cfg.Search.AutocompleteSearchTimeout, cfg.Search.AutocompleteCacheTimeout)
	}
	if !cfg.Search.ZeroResultTrackingEnabled || cfg.Search.ZeroResultDedupeTTL != 30*time.Minute || cfg.Search.ZeroResultSendTimeout != 150*time.Millisecond {
		t.Fatalf("zero-result config = %#v", cfg.Search)
	}
	if cfg.Search.ZeroResultQueueSize != 256 || cfg.Search.ZeroResultWorkerCount != 3 {
		t.Fatalf("zero-result worker config = %#v", cfg.Search)
	}
	if cfg.Session.URL != "http://session-service:8086" || cfg.Session.IngestPath != "/api/v1/sessions/events" || cfg.Session.Timeout != 150*time.Millisecond {
		t.Fatalf("session config = %#v", cfg.Session)
	}
	if cfg.Admin.Timeout != 500*time.Millisecond || !cfg.Admin.AuthEnabled {
		t.Fatalf("admin config = %#v", cfg.Admin)
	}
	if cfg.Admin.MutationRateLimit != 30 || cfg.Admin.MutationRateWindow != time.Minute {
		t.Fatalf("admin rate limit config = %#v", cfg.Admin)
	}
	if cfg.Queue.ProductEventsExchange != defaultProductEventsExchange {
		t.Fatalf("product events exchange = %q", cfg.Queue.ProductEventsExchange)
	}
	if len(cfg.Queue.ProductIndexerRoutingKeys) != 7 {
		t.Fatalf("routing keys = %#v", cfg.Queue.ProductIndexerRoutingKeys)
	}
	if cfg.Redis.Addr != "localhost:6379" {
		t.Fatalf("redis addr = %q", cfg.Redis.Addr)
	}
	if cfg.Reindex.Mode != "alias" || cfg.Reindex.BatchSize != 500 || cfg.Reindex.CollectionPrefix != "products" {
		t.Fatalf("reindex config = %#v", cfg.Reindex)
	}
	if cfg.Reindex.LockTTL != 3*time.Hour || cfg.Reindex.JobTimeout != 2*time.Hour || cfg.Reindex.OldCollectionRetention != 72*time.Hour || cfg.Reindex.ImportTimeout != 5*time.Second {
		t.Fatalf("reindex durations = %#v", cfg.Reindex)
	}
}

func TestLoadTypesenseURLOverridesEndpointParts(t *testing.T) {
	t.Setenv("TYPESENSE_URL", "https://typesense.internal:443")
	t.Setenv("TYPESENSE_API_KEY", "dev-typesense-key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Typesense.Endpoint() != "https://typesense.internal:443" {
		t.Fatalf("typesense endpoint = %q", cfg.Typesense.Endpoint())
	}
}

func TestLoadRejectsInvalidSearchPageSizes(t *testing.T) {
	t.Setenv("TYPESENSE_HOST", "typesense")
	t.Setenv("TYPESENSE_PORT", "8108")
	t.Setenv("TYPESENSE_PROTOCOL", "http")
	t.Setenv("TYPESENSE_API_KEY", "dev-typesense-key")
	t.Setenv("SEARCH_DEFAULT_PAGE_SIZE", "200")
	t.Setenv("SEARCH_MAX_PAGE_SIZE", "100")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid search page size config to fail validation")
	}
}

func TestLoadRequiresTypesenseAPIKey(t *testing.T) {
	t.Setenv("TYPESENSE_HOST", "typesense")
	t.Setenv("TYPESENSE_PORT", "8108")
	t.Setenv("TYPESENSE_PROTOCOL", "http")
	t.Setenv("TYPESENSE_API_KEY", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected missing TYPESENSE_API_KEY to fail validation")
	}
}

func TestLoadRejectsInvalidTypesensePort(t *testing.T) {
	t.Setenv("TYPESENSE_HOST", "typesense")
	t.Setenv("TYPESENSE_PORT", "not-a-port")
	t.Setenv("TYPESENSE_PROTOCOL", "http")
	t.Setenv("TYPESENSE_API_KEY", "dev-typesense-key")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid TYPESENSE_PORT to fail validation")
	}
}

func TestLoadRejectsUnsupportedQueueProviderWhenIndexerEnabled(t *testing.T) {
	t.Setenv("TYPESENSE_HOST", "typesense")
	t.Setenv("TYPESENSE_PORT", "8108")
	t.Setenv("TYPESENSE_PROTOCOL", "http")
	t.Setenv("TYPESENSE_API_KEY", "dev-typesense-key")
	t.Setenv("QUEUE_PROVIDER", "kafka")

	if _, err := Load(); err == nil {
		t.Fatal("expected unsupported QUEUE_PROVIDER to fail validation")
	}
}

func TestLoadSkipsQueueValidationWhenIndexerDisabled(t *testing.T) {
	t.Setenv("TYPESENSE_HOST", "typesense")
	t.Setenv("TYPESENSE_PORT", "8108")
	t.Setenv("TYPESENSE_PROTOCOL", "http")
	t.Setenv("TYPESENSE_API_KEY", "dev-typesense-key")
	t.Setenv("SEARCH_INDEXER_ENABLED", "false")
	t.Setenv("QUEUE_PROVIDER", "kafka")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Indexer.Enabled {
		t.Fatal("expected indexer disabled")
	}
}

func TestLoadRejectsInvalidSessionURLWhenZeroResultTrackingEnabled(t *testing.T) {
	t.Setenv("TYPESENSE_HOST", "typesense")
	t.Setenv("TYPESENSE_PORT", "8108")
	t.Setenv("TYPESENSE_PROTOCOL", "http")
	t.Setenv("TYPESENSE_API_KEY", "dev-typesense-key")
	t.Setenv("SEARCH_ZERO_RESULT_TRACKING_ENABLED", "true")
	t.Setenv("SESSION_SERVICE_URL", "not-a-url")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid SESSION_SERVICE_URL to fail validation")
	}
}

func TestLoadRejectsInvalidReindexMode(t *testing.T) {
	t.Setenv("TYPESENSE_HOST", "typesense")
	t.Setenv("TYPESENSE_PORT", "8108")
	t.Setenv("TYPESENSE_PROTOCOL", "http")
	t.Setenv("TYPESENSE_API_KEY", "dev-typesense-key")
	t.Setenv("SEARCH_REINDEX_MODE", "invalid")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid SEARCH_REINDEX_MODE to fail validation")
	}
}

func TestLoadSkipsSessionValidationWhenZeroResultTrackingDisabled(t *testing.T) {
	t.Setenv("TYPESENSE_HOST", "typesense")
	t.Setenv("TYPESENSE_PORT", "8108")
	t.Setenv("TYPESENSE_PROTOCOL", "http")
	t.Setenv("TYPESENSE_API_KEY", "dev-typesense-key")
	t.Setenv("SEARCH_ZERO_RESULT_TRACKING_ENABLED", "false")
	t.Setenv("SESSION_SERVICE_URL", "not-a-url")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Search.ZeroResultTrackingEnabled {
		t.Fatal("expected zero-result tracking disabled")
	}
}

func TestTypesenseConfigValidateRejectsInvalidEndpointParts(t *testing.T) {
	tests := map[string]TypesenseConfig{
		"empty host": {
			Host:     "",
			Port:     8108,
			Protocol: "http",
			APIKey:   "key",
		},
		"host with protocol": {
			Host:     "http://typesense",
			Port:     8108,
			Protocol: "http",
			APIKey:   "key",
		},
		"bad port": {
			Host:     "typesense",
			Port:     70000,
			Protocol: "http",
			APIKey:   "key",
		},
		"bad protocol": {
			Host:     "typesense",
			Port:     8108,
			Protocol: "tcp",
			APIKey:   "key",
		},
		"bad url": {
			URL:    "ftp://typesense:8108",
			APIKey: "key",
		},
		"empty api key": {
			Host:     "typesense",
			Port:     8108,
			Protocol: "http",
			APIKey:   "",
		},
	}

	for name, cfg := range tests {
		t.Run(name, func(t *testing.T) {
			if err := cfg.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
