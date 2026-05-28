package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	RuleBasedRankingFormulaVersion = "popularity_v1"
	ProductStatusActive            = "active"
	ProductStockStatusInStock      = "in_stock"
	TrendingGlobalContextKey       = "home:global"
)

type RankingScope string

const (
	RankingScopeGlobal   RankingScope = "global"
	RankingScopeCategory RankingScope = "category"
	RankingScopeSeller   RankingScope = "seller"
)

type PopularityWeights struct {
	Views24h       int64
	WishlistAdds7d int64
	CartAdds7d     int64
	Purchases7d    int64
}

type RankRequest struct {
	Scope   RankingScope
	ScopeID string
	Limit   int
	Now     time.Time
}

type ScoredProduct struct {
	ProductID   string
	CategoryID  string
	SellerID    string
	Score       float64
	Purchases7d int64
}

func DefaultPopularityWeights() PopularityWeights {
	return PopularityWeights{
		Views24h:       1,
		WishlistAdds7d: 3,
		CartAdds7d:     4,
		Purchases7d:    8,
	}
}

func (w PopularityWeights) Validate() error {
	if w.Views24h < 0 || w.WishlistAdds7d < 0 || w.CartAdds7d < 0 || w.Purchases7d < 0 {
		return fmt.Errorf("%w: popularity weights cannot be negative", ErrInvalidFeature)
	}
	if w.Views24h == 0 && w.WishlistAdds7d == 0 && w.CartAdds7d == 0 && w.Purchases7d == 0 {
		return fmt.Errorf("%w: at least one popularity weight must be greater than zero", ErrInvalidFeature)
	}
	return nil
}

func (s RankingScope) IsValid() bool {
	switch s {
	case RankingScopeGlobal, RankingScopeCategory, RankingScopeSeller:
		return true
	default:
		return false
	}
}

func (r RankRequest) Validate(maxIdentifierLength int) error {
	if !r.Scope.IsValid() {
		return fmt.Errorf("%w: invalid ranking scope", ErrInvalidRecommendationRequest)
	}
	if r.Limit < 0 {
		return fmt.Errorf("%w: limit cannot be negative", ErrInvalidLimit)
	}
	if maxIdentifierLength <= 0 {
		maxIdentifierLength = 128
	}
	scopeID := strings.TrimSpace(r.ScopeID)
	if r.Scope == RankingScopeGlobal {
		if scopeID != "" {
			return fmt.Errorf("%w: global ranking scope must not have scope_id", ErrInvalidRecommendationRequest)
		}
		return nil
	}
	if scopeID == "" {
		return fmt.Errorf("%w: scope_id is required for %s ranking scope", ErrInvalidRecommendationRequest, r.Scope)
	}
	if len(scopeID) > maxIdentifierLength {
		return fmt.Errorf("%w: scope_id is too long", ErrInvalidRecommendationRequest)
	}
	return nil
}

func (p ProductFeature) EligibleForRuleBasedRanking() bool {
	return strings.TrimSpace(p.ProductID) != "" &&
		strings.EqualFold(strings.TrimSpace(p.Status), ProductStatusActive) &&
		strings.EqualFold(strings.TrimSpace(p.StockStatus), ProductStockStatusInStock) &&
		p.QualityFlags.IsRecommendable &&
		!p.QualityFlags.IsDeleted
}

func (p ProductFeature) PopularityScore(weights PopularityWeights) float64 {
	return float64(nonNegative(p.Counters.Views24h))*float64(weights.Views24h) +
		float64(nonNegative(p.Counters.WishlistAdds7d))*float64(weights.WishlistAdds7d) +
		float64(nonNegative(p.Counters.CartAdds7d))*float64(weights.CartAdds7d) +
		float64(nonNegative(p.Counters.Purchases7d))*float64(weights.Purchases7d)
}

func ScoreAndSortRuleBasedProducts(products []ProductFeature, weights PopularityWeights) []ScoredProduct {
	items := make([]ScoredProduct, 0, len(products))
	for _, product := range products {
		if !product.EligibleForRuleBasedRanking() {
			continue
		}
		items = append(items, ScoredProduct{
			ProductID:   strings.TrimSpace(product.ProductID),
			CategoryID:  strings.TrimSpace(product.CategoryID),
			SellerID:    strings.TrimSpace(product.SellerID),
			Score:       product.PopularityScore(weights),
			Purchases7d: nonNegative(product.Counters.Purchases7d),
		})
	}

	sort.SliceStable(items, func(i int, j int) bool {
		if items[i].Score != items[j].Score {
			return items[i].Score > items[j].Score
		}
		if items[i].Purchases7d != items[j].Purchases7d {
			return items[i].Purchases7d > items[j].Purchases7d
		}
		return items[i].ProductID < items[j].ProductID
	})
	return items
}

func nonNegative(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}
