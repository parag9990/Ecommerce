package domain

import (
	"errors"
	"testing"
)

func TestScoreAndSortRuleBasedProductsFiltersUnsafeProductsAndUsesWeights(t *testing.T) {
	products := []ProductFeature{
		rankableProduct("views", ProductCounters{Views24h: 100}),
		rankableProduct("paid", ProductCounters{Purchases7d: 20}),
		{ProductID: "inactive", Status: "disabled", StockStatus: ProductStockStatusInStock, QualityFlags: ProductQualityFlags{IsRecommendable: true}, Counters: ProductCounters{Purchases7d: 1000}},
		{ProductID: "sold_out", Status: ProductStatusActive, StockStatus: "out_of_stock", QualityFlags: ProductQualityFlags{IsRecommendable: true}, Counters: ProductCounters{Purchases7d: 1000}},
		{ProductID: "blocked", Status: ProductStatusActive, StockStatus: ProductStockStatusInStock, Counters: ProductCounters{Purchases7d: 1000}},
	}

	got := ScoreAndSortRuleBasedProducts(products, DefaultPopularityWeights())
	if len(got) != 2 {
		t.Fatalf("ranked products = %d, want 2", len(got))
	}
	if got[0].ProductID != "paid" || got[0].Score != 160 {
		t.Fatalf("top product = %+v, want paid with score 160", got[0])
	}
}

func TestScoreAndSortRuleBasedProductsUsesStableTieBreaks(t *testing.T) {
	products := []ProductFeature{
		rankableProduct("product_b", ProductCounters{Views24h: 8}),
		rankableProduct("product_a", ProductCounters{Purchases7d: 1}),
		rankableProduct("product_c", ProductCounters{Views24h: 8}),
	}

	got := ScoreAndSortRuleBasedProducts(products, DefaultPopularityWeights())
	want := []string{"product_a", "product_b", "product_c"}
	for i, productID := range want {
		if got[i].ProductID != productID {
			t.Fatalf("rank %d product = %s, want %s", i+1, got[i].ProductID, productID)
		}
	}
}

func TestRankRequestValidatesScopedIdentity(t *testing.T) {
	if err := (RankRequest{Scope: RankingScopeCategory}).Validate(128); !errors.Is(err, ErrInvalidRecommendationRequest) {
		t.Fatalf("Validate() error = %v, want invalid recommendation request", err)
	}
	if err := (RankRequest{Scope: RankingScopeGlobal, ScopeID: "cat_1"}).Validate(128); !errors.Is(err, ErrInvalidRecommendationRequest) {
		t.Fatalf("Validate() error = %v, want invalid recommendation request", err)
	}
}

func rankableProduct(productID string, counters ProductCounters) ProductFeature {
	return ProductFeature{
		ProductID:   productID,
		Status:      ProductStatusActive,
		StockStatus: ProductStockStatusInStock,
		QualityFlags: ProductQualityFlags{
			IsRecommendable: true,
		},
		Counters: counters,
	}
}
