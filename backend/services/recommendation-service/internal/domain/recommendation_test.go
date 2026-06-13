package domain

import (
	"errors"
	"testing"
)

func TestResolveRecommendationType(t *testing.T) {
	tests := []struct {
		name string
		req  RecommendationRequest
		want RecommendationType
	}{
		{
			name: "product detail with product uses similar products",
			req: RecommendationRequest{
				Context:   ContextProductDetail,
				ProductID: "prod_123",
			},
			want: RecommendationTypeSimilarProducts,
		},
		{
			name: "cart uses frequently bought together",
			req: RecommendationRequest{
				Context:        ContextCart,
				CartProductIDs: []string{"prod_1"},
			},
			want: RecommendationTypeFrequentlyBoughtTogether,
		},
		{
			name: "logged in homepage uses personalized",
			req: RecommendationRequest{
				Context: ContextHomeFeed,
				UserID:  "user_123",
			},
			want: RecommendationTypePersonalized,
		},
		{
			name: "guest homepage with anonymous id uses personalized",
			req: RecommendationRequest{
				Context:     ContextHomeFeed,
				AnonymousID: "anon_123",
			},
			want: RecommendationTypePersonalized,
		},
		{
			name: "category page without user uses trending",
			req: RecommendationRequest{
				Context:    ContextCategoryListing,
				CategoryID: "cat_shoes",
			},
			want: RecommendationTypeTrending,
		},
		{
			name: "explicit valid type wins",
			req: RecommendationRequest{
				Context: ContextProductDetail,
				Type:    RecommendationTypeTrending,
			},
			want: RecommendationTypeTrending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveRecommendationType(tt.req); got != tt.want {
				t.Fatalf("ResolveRecommendationType() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRecommendationRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     RecommendationRequest
		wantErr error
	}{
		{
			name: "valid trending request",
			req: RecommendationRequest{
				Context: ContextHomeFeed,
				Type:    RecommendationTypeTrending,
				Limit:   12,
			},
		},
		{
			name: "missing context",
			req: RecommendationRequest{
				Type: RecommendationTypeTrending,
			},
			wantErr: ErrUnsupportedContext,
		},
		{
			name: "negative limit",
			req: RecommendationRequest{
				Context: ContextHomeFeed,
				Limit:   -1,
			},
			wantErr: ErrInvalidLimit,
		},
		{
			name: "similar requires product id",
			req: RecommendationRequest{
				Context: ContextProductDetail,
				Type:    RecommendationTypeSimilarProducts,
			},
			wantErr: ErrInvalidRecommendationRequest,
		},
		{
			name: "personalized requires identity",
			req: RecommendationRequest{
				Context: ContextHomeFeed,
				Type:    RecommendationTypePersonalized,
			},
			wantErr: ErrInvalidRecommendationRequest,
		},
		{
			name: "frequently bought together accepts cart products",
			req: RecommendationRequest{
				Context:        ContextCart,
				Type:           RecommendationTypeFrequentlyBoughtTogether,
				CartProductIDs: []string{"prod_1", "prod_1", " "},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate(128)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestStrategyIDsValidate(t *testing.T) {
	valid := []StrategyID{
		StrategySimilarAttributes,
		StrategyTrendingRecentActivity,
		StrategyPersonalizedBehavior,
		StrategyFBTOrderCooccurrence,
		StrategyTrendingFallbackGlobal,
		StrategyTrendingFallbackCategory,
	}
	for _, strategyID := range valid {
		if err := ValidateStrategyID(strategyID); err != nil {
			t.Fatalf("ValidateStrategyID(%q) error = %v", strategyID, err)
		}
	}

	if err := ValidateStrategyID("similar_attributes_v1"); !errors.Is(err, ErrUnsupportedStrategy) {
		t.Fatalf("ValidateStrategyID(invalid) error = %v, want %v", err, ErrUnsupportedStrategy)
	}
}
