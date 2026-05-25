package config

import "testing"

func TestConfigValidateAcceptsMongoURI(t *testing.T) {
	cfg := validConfig()

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestConfigValidateRequiresCartDatabaseOwnership(t *testing.T) {
	cfg := validConfig()
	cfg.Mongo.Database = "shared_db"

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want database ownership error")
	}
}

func TestConfigValidateRequiresProductBaseURL(t *testing.T) {
	cfg := validConfig()
	cfg.Product.BaseURL = ""

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want product base url error")
	}
}

func TestConfigValidateRequiresCMSBaseURL(t *testing.T) {
	cfg := validConfig()
	cfg.CMS.BaseURL = ""

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want cms base url error")
	}
}

func TestConfigValidateRequiresRedisAddress(t *testing.T) {
	cfg := validConfig()
	cfg.Redis.Address = ""

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want redis address error")
	}
}

func validConfig() Config {
	return Config{
		ServiceName: "cart-service",
		Environment: "test",
		HTTP: HTTPConfig{
			Address:         ":8084",
			ReadTimeout:     defaultShutdownTimeout,
			WriteTimeout:    defaultShutdownTimeout,
			IdleTimeout:     defaultShutdownTimeout,
			ShutdownTimeout: defaultShutdownTimeout,
		},
		Mongo: MongoConfig{
			URI:            "mongodb://ecommerce_root:ecommerce_password@mongo:27017/cart_db?authSource=admin",
			Database:       defaultMongoDatabase,
			ConnectTimeout: defaultShutdownTimeout,
			PingTimeout:    defaultShutdownTimeout,
		},
		Redis: RedisConfig{
			Address:      "redis:6379",
			DB:           0,
			DialTimeout:  defaultShutdownTimeout,
			ReadTimeout:  defaultShutdownTimeout,
			WriteTimeout: defaultShutdownTimeout,
			PingTimeout:  defaultShutdownTimeout,
		},
		Product: ProductConfig{
			BaseURL:        "http://product-service:8080",
			RequestTimeout: defaultShutdownTimeout,
		},
		CMS: CMSConfig{
			BaseURL:            "http://cms-service:8080",
			ValidateCouponPath: defaultCMSValidatePath,
			RequestTimeout:     defaultShutdownTimeout,
		},
		Cart: CartConfig{
			ExpiryTTL:       90 * 24 * defaultShutdownTimeout,
			UserExpiryTTL:   90 * 24 * defaultShutdownTimeout,
			GuestExpiryTTL:  30 * 24 * defaultShutdownTimeout,
			DefaultCurrency: "INR",
			MaxSaveAttempts: 3,
		},
		CartExpiry: CartExpiryConfig{
			CleanupInterval:   defaultShutdownTimeout,
			CleanupBatchSize:  500,
			CleanupLockTTL:    defaultShutdownTimeout,
			CleanupRunTimeout: defaultShutdownTimeout,
			CleanupLockKey:    "cart:lock:expiry-cleanup",
		},
		Cache: CacheConfig{
			ActiveUserTTL:  defaultShutdownTimeout,
			ActiveGuestTTL: defaultShutdownTimeout,
			SummaryTTL:     defaultShutdownTimeout,
		},
	}
}
