package config

import (
	"strings"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

func TestLoadRejectsFeatureClaimRetentionShorterThanRawRetention(t *testing.T) {
	t.Setenv("RECOMMENDATION_INTERACTION_RETENTION_SECONDS", "15552000")
	t.Setenv("RECOMMENDATION_FEATURE_EVENT_RETENTION_SECONDS", "2592000")
	t.Setenv("RECOMMENDATION_FEATURES_ENABLED", "true")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "RECOMMENDATION_FEATURE_EVENT_RETENTION_SECONDS") {
		t.Fatalf("Load() error = %v, want feature claim retention validation error", err)
	}
}

func TestLoadProvidesRuleBasedRankingDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.Ranking.Enabled || cfg.Ranking.Weights != domain.DefaultPopularityWeights() {
		t.Fatalf("ranking config = %+v, want enabled default popularity formula", cfg.Ranking)
	}
}

func TestLoadProvidesGRPCDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.GRPC.Address != ":9088" {
		t.Fatalf("grpc address = %s, want :9088", cfg.GRPC.Address)
	}
	if cfg.GRPC.MaxRecvBytes != 65536 || cfg.GRPC.MaxSendBytes != 262144 {
		t.Fatalf("grpc message sizes = %d/%d, want defaults", cfg.GRPC.MaxRecvBytes, cfg.GRPC.MaxSendBytes)
	}
	if cfg.GRPC.DefaultDeadline != 500*time.Millisecond {
		t.Fatalf("grpc default deadline = %s, want 500ms", cfg.GRPC.DefaultDeadline)
	}
}

func TestLoadProvidesPersonalizationDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.Personalization.Enabled {
		t.Fatal("personalization should be enabled by default")
	}
	if cfg.Personalization.FormulaVersion != domain.PersonalizationFormulaVersion {
		t.Fatalf("formula version = %s, want %s", cfg.Personalization.FormulaVersion, domain.PersonalizationFormulaVersion)
	}
	if cfg.Personalization.Weights != domain.DefaultPersonalizationWeights() {
		t.Fatalf("personalization weights = %+v, want defaults", cfg.Personalization.Weights)
	}
}

func TestLoadProvidesABTestingDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ABTesting.Enabled {
		t.Fatal("ab testing should be disabled by default")
	}
	if cfg.ABTesting.AssignmentTTL != domain.DefaultABAssignmentTTL {
		t.Fatalf("assignment ttl = %s, want %s", cfg.ABTesting.AssignmentTTL, domain.DefaultABAssignmentTTL)
	}
	if cfg.ABTesting.DefaultSalt != domain.DefaultABTestSalt {
		t.Fatalf("salt = %s, want %s", cfg.ABTesting.DefaultSalt, domain.DefaultABTestSalt)
	}
	if !cfg.ABTesting.FailOpen {
		t.Fatal("ab testing should fail open by default")
	}
}

func TestLoadParsesABExperiments(t *testing.T) {
	t.Setenv("RECOMMENDATION_AB_TESTING_ENABLED", "true")
	t.Setenv("RECOMMENDATION_AB_EXPERIMENTS_JSON", `[{
		"experiment_id":"reco_home_strategy_2026_05",
		"context":"home_feed",
		"type":"personalized",
		"status":"active",
		"variants":[
			{"variant_id":"control","strategy_id":"trending_v1_recent_activity","weight":50},
			{"variant_id":"treatment","strategy_id":"personalized_v1_behavior","weight":50}
		]
	}]`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.ABTesting.Experiments) != 1 {
		t.Fatalf("experiments = %d, want 1", len(cfg.ABTesting.Experiments))
	}
	if cfg.ABTesting.Experiments[0].Salt != "" {
		t.Fatalf("raw config salt = %s, want empty before service normalization", cfg.ABTesting.Experiments[0].Salt)
	}
}

func TestLoadRejectsInvalidABExperiments(t *testing.T) {
	t.Setenv("RECOMMENDATION_AB_EXPERIMENTS_JSON", `[{"experiment_id":"bad exp"}]`)

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "invalid ab testing config") {
		t.Fatalf("Load() error = %v, want ab testing validation error", err)
	}
}

func TestLoadRejectsInvalidPersonalizationWeights(t *testing.T) {
	t.Setenv("RECOMMENDATION_PERSONALIZED_CATEGORY_WEIGHT", "0.50")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "weights must total") {
		t.Fatalf("Load() error = %v, want personalization weight validation error", err)
	}
}

func TestLoadRejectsAllZeroRankingWeights(t *testing.T) {
	t.Setenv("RECOMMENDATION_RANKING_VIEW_24H_WEIGHT", "0")
	t.Setenv("RECOMMENDATION_RANKING_WISHLIST_7D_WEIGHT", "0")
	t.Setenv("RECOMMENDATION_RANKING_CART_7D_WEIGHT", "0")
	t.Setenv("RECOMMENDATION_RANKING_PURCHASE_7D_WEIGHT", "0")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "at least one popularity weight") {
		t.Fatalf("Load() error = %v, want invalid ranking weights", err)
	}
}

func TestLoadRequiresNewFormulaVersionForTunedRankingWeights(t *testing.T) {
	t.Setenv("RECOMMENDATION_RANKING_VIEW_24H_WEIGHT", "2")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "FORMULA_VERSION must change") {
		t.Fatalf("Load() error = %v, want formula version validation error", err)
	}

	t.Setenv("RECOMMENDATION_RANKING_FORMULA_VERSION", "popularity_v2")
	if _, err := Load(); err != nil {
		t.Fatalf("Load() error with versioned tuned formula = %v", err)
	}
}
