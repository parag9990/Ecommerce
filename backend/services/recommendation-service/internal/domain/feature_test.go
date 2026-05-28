package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewProfileKeyUsesKnownUserBeforeAnonymousIdentity(t *testing.T) {
	key, err := NewProfileKey("user_123", "anon_123")
	if err != nil {
		t.Fatalf("NewProfileKey() error = %v", err)
	}
	if key != "user:user_123" {
		t.Fatalf("profile key = %s, want user:user_123", key)
	}
}

func TestNewProductPairKeyOrdersProducts(t *testing.T) {
	key, source, related, err := NewProductPairKey("prod_b", "prod_a")
	if err != nil {
		t.Fatalf("NewProductPairKey() error = %v", err)
	}
	if key != "prod_a:prod_b" || source != "prod_a" || related != "prod_b" {
		t.Fatalf("pair = %s %s %s, want canonical ordered pair", key, source, related)
	}
}

func TestFeatureBatchRejectsUnsafeScoreMapDimension(t *testing.T) {
	interaction := UserInteraction{
		ID:                  "interaction_1",
		DedupeKey:           "evt#product_view#prod_1#",
		EventID:             "evt",
		SourceEventType:     EventProductViewed,
		NormalizedEventType: InteractionProductView,
		Version:             1,
		Producer:            "session-service",
		UserID:              "user_1",
		ProductID:           "prod_1",
		CategoryID:          "cat.unsafe",
		Weight:              1,
		Quantity:            1,
		OccurredAt:          time.Now().UTC(),
		ReceivedAt:          time.Now().UTC(),
	}
	batch := FeatureBatch{
		ID:                      "feature_job_1",
		Interactions:            []UserInteraction{interaction},
		StartedAt:               time.Now().UTC(),
		ProcessedAt:             time.Now().UTC(),
		GuestProfileRetention:   DefaultGuestProfileRetention,
		UserProductRetention:    DefaultUserProductRetention,
		ProcessedEventRetention: DefaultFeatureEventRetention,
		RecentProductsLimit:     DefaultRecentProductsLimit,
	}
	if err := batch.Validate(); !errors.Is(err, ErrInvalidFeature) {
		t.Fatalf("Validate() error = %v, want invalid feature error", err)
	}
}
