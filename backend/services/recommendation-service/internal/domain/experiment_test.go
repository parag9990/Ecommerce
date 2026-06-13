package domain

import (
	"errors"
	"testing"
	"time"
)

func TestExperimentDefinitionValidateAllowsTrendingControlForPersonalizedExperiment(t *testing.T) {
	experiment := ExperimentDefinition{
		ExperimentID: "reco_home_strategy_2026_05",
		Context:      ContextHomeFeed,
		Type:         RecommendationTypePersonalized,
		Status:       ExperimentActive,
		Salt:         "salt",
		Variants: []ExperimentVariant{
			{VariantID: "control", StrategyID: StrategyTrendingRecentActivity, Weight: 50},
			{VariantID: "treatment", StrategyID: StrategyPersonalizedBehavior, Weight: 50},
		},
	}
	if err := experiment.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !experiment.ActiveAt(time.Now().UTC()) {
		t.Fatal("experiment should be active")
	}
}

func TestExperimentDefinitionValidateRejectsInvalidWeights(t *testing.T) {
	experiment := ExperimentDefinition{
		ExperimentID: "reco_home_strategy_2026_05",
		Context:      ContextHomeFeed,
		Type:         RecommendationTypePersonalized,
		Status:       ExperimentActive,
		Salt:         "salt",
		Variants: []ExperimentVariant{
			{VariantID: "control", StrategyID: StrategyTrendingRecentActivity, Weight: 10},
			{VariantID: "treatment", StrategyID: StrategyPersonalizedBehavior, Weight: 10},
		},
	}
	if err := experiment.Validate(); !errors.Is(err, ErrInvalidExperiment) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrInvalidExperiment)
	}
}

func TestAssignmentIdentityPriority(t *testing.T) {
	got, ok := NewAssignmentIdentity("user_123", "anon_123", "sess_123")
	if !ok {
		t.Fatal("identity should exist")
	}
	if got.AssignmentKey != "user:user_123" {
		t.Fatalf("assignment key = %s, want user:user_123", got.AssignmentKey)
	}
	got, ok = NewAssignmentIdentity("", "", "sess_123")
	if !ok || got.AssignmentKey != "session:sess_123" {
		t.Fatalf("session fallback = %+v ok=%t", got, ok)
	}
}

func TestBucketPercentIsStableAndVariantSelectionUsesWeights(t *testing.T) {
	got1 := BucketPercent("reco_home_strategy_2026_05", "user:user_123", "salt")
	got2 := BucketPercent("reco_home_strategy_2026_05", "user:user_123", "salt")
	if got1 != got2 {
		t.Fatalf("bucket changed: %d != %d", got1, got2)
	}
	variant, err := ChooseExperimentVariant(49, []ExperimentVariant{
		{VariantID: "control", StrategyID: StrategyTrendingRecentActivity, Weight: 50},
		{VariantID: "treatment", StrategyID: StrategyPersonalizedBehavior, Weight: 50},
	})
	if err != nil {
		t.Fatalf("ChooseExperimentVariant() error = %v", err)
	}
	if variant.VariantID != "control" {
		t.Fatalf("variant = %s, want control", variant.VariantID)
	}
	variant, err = ChooseExperimentVariant(50, []ExperimentVariant{
		{VariantID: "control", StrategyID: StrategyTrendingRecentActivity, Weight: 50},
		{VariantID: "treatment", StrategyID: StrategyPersonalizedBehavior, Weight: 50},
	})
	if err != nil {
		t.Fatalf("ChooseExperimentVariant() error = %v", err)
	}
	if variant.VariantID != "treatment" {
		t.Fatalf("variant = %s, want treatment", variant.VariantID)
	}
}

func TestExperimentAssignmentValidate(t *testing.T) {
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	assignment := NewExperimentAssignment(
		ExperimentDefinition{
			ExperimentID: "reco_home_strategy_2026_05",
			Context:      ContextHomeFeed,
			Type:         RecommendationTypePersonalized,
			Status:       ExperimentActive,
			Salt:         "salt",
		},
		AssignmentIdentity{AssignmentKey: "user:user_123", UserID: "user_123"},
		ExperimentVariant{VariantID: "treatment", StrategyID: StrategyPersonalizedBehavior, Weight: 50},
		now,
		24*time.Hour,
	)
	if err := assignment.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
