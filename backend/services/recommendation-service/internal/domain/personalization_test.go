package domain

import (
	"errors"
	"testing"
)

func TestBehaviorV1ScoreRewardsPersonalSignalsOverPopularity(t *testing.T) {
	preferred := BehaviorV1Score(PersonalizedSignals{
		CategoryAffinity:      1,
		DirectProductAffinity: 1,
		PopularityBaseline:    0.2,
	}, DefaultPersonalizationWeights())
	unrelatedPopular := BehaviorV1Score(PersonalizedSignals{
		PopularityBaseline: 1,
	}, DefaultPersonalizationWeights())

	if preferred <= unrelatedPopular {
		t.Fatalf("preferred score = %v, unrelated popular score = %v", preferred, unrelatedPopular)
	}
}

func TestPersonalizationWeightsValidateTotal(t *testing.T) {
	err := (PersonalizationWeights{Category: 1}).Validate()
	if err != nil {
		t.Fatalf("single full category weight should be valid: %v", err)
	}

	err = (PersonalizationWeights{Category: 0.5}).Validate()
	if !errors.Is(err, ErrInvalidFeature) {
		t.Fatalf("Validate() error = %v, want invalid feature", err)
	}
}

func TestScoreAndSortPersonalizedProductsFiltersUnsafeAndPenalizesNegativeFeedback(t *testing.T) {
	profile := UserFeatureProfile{
		CategoryScores: map[string]float64{"cat_a": 10},
	}
	counters := []UserProductCounter{
		{
			ProductID:          "kept",
			WeightedScoreInput: 10,
			Counters:           InteractionCounters{WishlistAdds: 1, CartAdds: 1},
		},
		{
			ProductID:          "removed",
			WeightedScoreInput: 10,
			Counters:           InteractionCounters{WishlistRemoves: 2},
		},
	}
	products := []ProductFeature{
		personalizedProduct("removed", "cat_a", ProductCounters{Purchases7d: 10}),
		personalizedProduct("kept", "cat_a", ProductCounters{Purchases7d: 1}),
		{ProductID: "sold_out", CategoryID: "cat_a", Status: ProductStatusActive, StockStatus: "out_of_stock", QualityFlags: ProductQualityFlags{IsRecommendable: true}},
	}

	got := ScoreAndSortPersonalizedProducts(profile, counters, products, DefaultPersonalizationWeights())
	if len(got) != 2 {
		t.Fatalf("ranked products = %d, want 2 safe products", len(got))
	}
	if got[0].ProductID != "kept" {
		t.Fatalf("top product = %+v, want kept before negative-feedback product", got[0])
	}
}

func TestScoreAndSortPersonalizedProductsUsesStableTieBreak(t *testing.T) {
	profile := UserFeatureProfile{CategoryScores: map[string]float64{"cat_a": 1}}
	products := []ProductFeature{
		personalizedProduct("product_b", "cat_a", ProductCounters{}),
		personalizedProduct("product_a", "cat_a", ProductCounters{}),
	}

	got := ScoreAndSortPersonalizedProducts(profile, nil, products, DefaultPersonalizationWeights())
	if got[0].ProductID != "product_a" {
		t.Fatalf("first product = %s, want product_a stable tie-break", got[0].ProductID)
	}
}

func personalizedProduct(productID string, categoryID string, counters ProductCounters) ProductFeature {
	return ProductFeature{
		ProductID:   productID,
		CategoryID:  categoryID,
		Status:      ProductStatusActive,
		StockStatus: ProductStockStatusInStock,
		QualityFlags: ProductQualityFlags{
			IsRecommendable: true,
		},
		Counters: counters,
	}
}
