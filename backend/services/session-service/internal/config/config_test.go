package config

import (
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

func TestLoadUsesEnvironmentOverrides(t *testing.T) {
	t.Setenv("SESSION_INACTIVITY_TIMEOUT_SECONDS", "900")
	t.Setenv("SESSION_DEFAULT_CHANNEL", string(domain.ChannelUserAppWeb))
	t.Setenv("SESSION_MAX_USER_AGENT_LENGTH", "512")
	t.Setenv("SESSION_MONGO_DATABASE", "session_test_db")
	t.Setenv("SESSION_REDIS_DB", "6")
	t.Setenv("SESSION_ACTIVE_TTL_SECONDS", "1200")
	t.Setenv("SESSION_INGEST_MAX_BODY_BYTES", "32768")
	t.Setenv("SESSION_ALLOWED_CLOCK_SKEW_SECONDS", "120")
	t.Setenv("SESSION_MAX_EVENT_AGE_HOURS", "12")
	t.Setenv("SESSION_IP_HASH_SALT", "test-salt")
	t.Setenv("SESSION_DEVICE_TRACKING_ENABLED", "true")
	t.Setenv("SESSION_USER_AGENT_MAX_LENGTH", "768")
	t.Setenv("SESSION_GEOIP_ENABLED", "false")
	t.Setenv("SESSION_GEOIP_TIMEOUT_MS", "25")
	t.Setenv("SESSION_PRIVACY_HASH_PEPPER", "pepper")
	t.Setenv("SESSION_TRUSTED_PROXY_CIDRS", "10.0.0.0/8,192.168.0.0/16")
	t.Setenv("SESSION_STORE_USER_AGENT", "false")
	t.Setenv("SESSION_JOURNEY_DEFAULT_LIMIT", "250")
	t.Setenv("SESSION_JOURNEY_MAX_LIMIT", "750")
	t.Setenv("SESSION_JOURNEY_SUMMARY_TOP_PATHS_LIMIT", "7")
	t.Setenv("SESSION_JOURNEY_SUMMARY_UPSERT_ENABLED", "false")
	t.Setenv("SESSION_HEATMAP_CLICK_BUCKET_SIZE", "10")
	t.Setenv("SESSION_HEATMAP_MAX_DATE_RANGE_DAYS", "14")
	t.Setenv("SESSION_HEATMAP_MAX_POINTS", "2500")
	t.Setenv("SESSION_HEATMAP_AGGREGATION_ENABLED", "false")
	t.Setenv("SESSION_HEATMAP_AGGREGATION_INTERVAL", "30s")
	t.Setenv("SESSION_HEATMAP_AGGREGATION_BATCH_SIZE", "250")
	t.Setenv("SESSION_HEATMAP_AGGREGATION_INITIAL_LOOKBACK", "2h")
	t.Setenv("SESSION_HEATMAP_AGGREGATION_CHECKPOINT_LOOKBACK", "1m")
	t.Setenv("SESSION_HEATMAP_AGGREGATION_WORKER_NAME", "test_heatmap")
	t.Setenv("SESSION_METADATA_RETENTION_DAYS", "400")
	t.Setenv("SESSION_JOURNEY_RETENTION_DAYS", "400")
	t.Setenv("SESSION_HEATMAP_RETENTION_DAYS", "800")
	t.Setenv("SESSION_AGGREGATE_RETENTION_YEARS", "4")
	t.Setenv("SESSION_RETENTION_WORKER_BATCH_SIZE", "250")
	t.Setenv("SESSION_RETENTION_WORKER_INTERVAL_MINUTES", "30")
	t.Setenv("SESSION_RETENTION_DRY_RUN", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.SessionModel.InactivityTimeout != 15*time.Minute {
		t.Fatalf("unexpected inactivity timeout %s", cfg.SessionModel.InactivityTimeout)
	}
	if cfg.SessionModel.DefaultChannel != string(domain.ChannelUserAppWeb) {
		t.Fatalf("unexpected default channel %s", cfg.SessionModel.DefaultChannel)
	}
	if cfg.SessionModel.MaxUserAgentLength != 512 {
		t.Fatalf("unexpected user agent max length %d", cfg.SessionModel.MaxUserAgentLength)
	}
	if cfg.Storage.MongoDatabase != "session_test_db" {
		t.Fatalf("unexpected mongo database %s", cfg.Storage.MongoDatabase)
	}
	if cfg.Storage.RedisDB != 6 {
		t.Fatalf("unexpected redis db %d", cfg.Storage.RedisDB)
	}
	if cfg.Storage.ActiveSessionTTL != 20*time.Minute {
		t.Fatalf("unexpected active session ttl %s", cfg.Storage.ActiveSessionTTL)
	}
	if cfg.Ingest.MaxBodyBytes != 32768 {
		t.Fatalf("unexpected ingest max body bytes %d", cfg.Ingest.MaxBodyBytes)
	}
	if cfg.Ingest.AllowedClockSkew != 2*time.Minute {
		t.Fatalf("unexpected allowed clock skew %s", cfg.Ingest.AllowedClockSkew)
	}
	if cfg.Ingest.MaxEventAge != 12*time.Hour {
		t.Fatalf("unexpected max event age %s", cfg.Ingest.MaxEventAge)
	}
	if cfg.Ingest.IPHashSalt != "test-salt" {
		t.Fatalf("unexpected IP hash salt")
	}
	if !cfg.DeviceTracking.Enabled || cfg.DeviceTracking.UserAgentMaxLength != 768 {
		t.Fatalf("unexpected device tracking config: %+v", cfg.DeviceTracking)
	}
	if cfg.DeviceTracking.GeoIPTimeout != 25*time.Millisecond {
		t.Fatalf("unexpected geoip timeout %s", cfg.DeviceTracking.GeoIPTimeout)
	}
	if cfg.DeviceTracking.PrivacyHashPepper != "pepper" || cfg.DeviceTracking.StoreUserAgent {
		t.Fatalf("unexpected privacy config: %+v", cfg.DeviceTracking)
	}
	if len(cfg.DeviceTracking.TrustedProxyCIDRs) != 2 {
		t.Fatalf("unexpected trusted proxy cidrs: %+v", cfg.DeviceTracking.TrustedProxyCIDRs)
	}
	if cfg.Journey.DefaultLimit != 250 || cfg.Journey.MaxLimit != 750 {
		t.Fatalf("unexpected journey limits: %+v", cfg.Journey)
	}
	if cfg.Journey.SummaryTopPathsLimit != 7 || cfg.Journey.SummaryUpsertEnabled {
		t.Fatalf("unexpected journey summary config: %+v", cfg.Journey)
	}
	if cfg.Heatmap.ClickBucketSize != 10 || cfg.Heatmap.MaxDateRangeDays != 14 || cfg.Heatmap.MaxPoints != 2500 {
		t.Fatalf("unexpected heatmap query config: %+v", cfg.Heatmap)
	}
	if cfg.Heatmap.AggregationEnabled || cfg.Heatmap.AggregationInterval != 30*time.Second || cfg.Heatmap.AggregationBatchSize != 250 {
		t.Fatalf("unexpected heatmap worker config: %+v", cfg.Heatmap)
	}
	if cfg.Heatmap.AggregationInitialLookback != 2*time.Hour || cfg.Heatmap.AggregationCheckpointLookback != time.Minute || cfg.Heatmap.AggregationWorkerName != "test_heatmap" {
		t.Fatalf("unexpected heatmap aggregation config: %+v", cfg.Heatmap)
	}
	if cfg.Retention.SessionMetadataRetentionDays != 400 || cfg.Retention.JourneySummaryRetentionDays != 400 || cfg.Retention.HeatmapRetentionDays != 800 {
		t.Fatalf("unexpected retention windows: %+v", cfg.Retention)
	}
	if cfg.Retention.AggregateRetentionYears != 4 || cfg.Retention.WorkerBatchSize != 250 || cfg.Retention.WorkerInterval != 30*time.Minute || !cfg.Retention.WorkerDryRun {
		t.Fatalf("unexpected retention worker config: %+v", cfg.Retention)
	}
}

func TestConfigRejectsInvalidDefaults(t *testing.T) {
	cfg := Config{
		SessionModel: SessionModelConfig{
			SchemaVersion:           domain.CurrentSessionSchemaVersion,
			InactivityTimeout:       time.Minute,
			DefaultStatus:           "not-a-status",
			DefaultChannel:          string(domain.ChannelUnknown),
			DefaultDeviceType:       string(domain.DeviceTypeUnknown),
			DefaultIPVersion:        string(domain.IPVersionUnknown),
			DefaultRiskLevel:        string(domain.RiskLevelUnknown),
			MaxIDLength:             domain.DefaultMaxIDLength,
			MaxHashLength:           domain.DefaultMaxHashLength,
			MaxPageLength:           domain.DefaultMaxPageLength,
			MaxReferrerLength:       domain.DefaultMaxReferrerLength,
			MaxUserAgentLength:      domain.DefaultMaxUserAgentLength,
			MaxMetadataLength:       domain.DefaultMaxMetadataLength,
			MaxUTMValueLength:       domain.DefaultMaxUTMValueLength,
			MaxGeoValueLength:       domain.DefaultMaxGeoValueLength,
			MaxRiskReasons:          domain.DefaultMaxRiskReasons,
			MaxRiskReasonTextLength: domain.DefaultMaxRiskReasonTextLength,
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid default status error")
	}
}
