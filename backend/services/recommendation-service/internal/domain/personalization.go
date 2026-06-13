package domain

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

const (
	PersonalizationFormulaVersion         = "behavior_v1"
	DefaultPersonalizedMinInteractions    = 3
	DefaultPersonalizedProfileMaxAge      = 30 * 24 * time.Hour
	DefaultPersonalizedMaxCandidates      = 250
	DefaultPersonalizedMaxItemsPerSeller  = 3
	DefaultPersonalizedTopCategories      = 3
	DefaultPersonalizedTopSellers         = 2
	DefaultPersonalizedTopBrands          = 2
	DefaultPersonalizedDirectProductLimit = 20

	FallbackReasonCategoryPopular = "fallback_category_popular"
	FallbackReasonGlobalTrending  = "fallback_global_trending"
)

type PersonalizationWeights struct {
	Category       float64
	Seller         float64
	Brand          float64
	Price          float64
	DirectAffinity float64
	Popularity     float64
}

type PersonalizedCandidateQuery struct {
	CategoryIDs []string
	SellerIDs   []string
	BrandIDs    []string
	ProductIDs  []string
	Limit       int
}

type PersonalizedSignals struct {
	CategoryAffinity      float64
	SellerAffinity        float64
	BrandAffinity         float64
	PriceMatch            float64
	DirectProductAffinity float64
	PopularityBaseline    float64
	NegativePenalty       float64
}

type ScoredPersonalizedProduct struct {
	ProductID             string
	CategoryID            string
	SellerID              string
	BrandID               string
	Score                 float64
	Reason                string
	Signals               PersonalizedSignals
	DirectProductAffinity float64
	PopularityBaseline    float64
	WishlistAdds          int64
	WishlistRemoves       int64
	CartAdds              int64
	Purchases             int64
}

func DefaultPersonalizationWeights() PersonalizationWeights {
	return PersonalizationWeights{
		Category:       0.35,
		Seller:         0.15,
		Brand:          0.10,
		Price:          0.10,
		DirectAffinity: 0.20,
		Popularity:     0.10,
	}
}

func (w PersonalizationWeights) Validate() error {
	values := []struct {
		name  string
		value float64
	}{
		{name: "category", value: w.Category},
		{name: "seller", value: w.Seller},
		{name: "brand", value: w.Brand},
		{name: "price", value: w.Price},
		{name: "direct_affinity", value: w.DirectAffinity},
		{name: "popularity", value: w.Popularity},
	}
	total := 0.0
	for _, item := range values {
		if math.IsNaN(item.value) || math.IsInf(item.value, 0) || item.value < 0 {
			return fmt.Errorf("%w: personalization weight %s must be a non-negative finite value", ErrInvalidFeature, item.name)
		}
		total += item.value
	}
	if math.Abs(total-1.0) > 0.000001 {
		return fmt.Errorf("%w: personalization weights must total 1.00", ErrInvalidFeature)
	}
	return nil
}

func (q PersonalizedCandidateQuery) Normalize() PersonalizedCandidateQuery {
	normalized := PersonalizedCandidateQuery{
		CategoryIDs: uniqueTrimmed(q.CategoryIDs),
		SellerIDs:   uniqueTrimmed(q.SellerIDs),
		BrandIDs:    uniqueTrimmed(q.BrandIDs),
		ProductIDs:  uniqueTrimmed(q.ProductIDs),
		Limit:       q.Limit,
	}
	if normalized.Limit <= 0 {
		normalized.Limit = DefaultPersonalizedMaxCandidates
	}
	return normalized
}

func (q PersonalizedCandidateQuery) Empty() bool {
	q = q.Normalize()
	return len(q.CategoryIDs) == 0 &&
		len(q.SellerIDs) == 0 &&
		len(q.BrandIDs) == 0 &&
		len(q.ProductIDs) == 0
}

func NormalizePersonalizedSignal(value float64, maxValue float64) float64 {
	if maxValue <= 0 || value <= 0 {
		return 0
	}
	score := value / maxValue
	if score > 1 {
		return 1
	}
	return score
}

func BehaviorV1Score(signals PersonalizedSignals, weights PersonalizationWeights) float64 {
	if err := weights.Validate(); err != nil {
		weights = DefaultPersonalizationWeights()
	}
	base := 100 * (weights.Category*signals.CategoryAffinity +
		weights.Seller*signals.SellerAffinity +
		weights.Brand*signals.BrandAffinity +
		weights.Price*signals.PriceMatch +
		weights.DirectAffinity*signals.DirectProductAffinity +
		weights.Popularity*signals.PopularityBaseline)

	score := base - signals.NegativePenalty
	if score < 0 {
		return 0
	}
	return score
}

func (p ProductFeature) EligibleForPersonalization() bool {
	return p.EligibleForRuleBasedRanking()
}

func PositiveUserProductInteractions(counters []UserProductCounter) int {
	total := 0
	for _, counter := range counters {
		if counter.Counters.Views > 0 {
			total += int(counter.Counters.Views)
		}
		if counter.Counters.WishlistAdds > 0 {
			total += int(counter.Counters.WishlistAdds)
		}
		if counter.Counters.CartAdds > 0 {
			total += int(counter.Counters.CartAdds)
		}
		if counter.Counters.Purchases > 0 {
			total += int(counter.Counters.Purchases)
		}
	}
	return total
}

func HasPositiveProfileAffinity(profile UserFeatureProfile) bool {
	return maxScore(profile.CategoryScores) > 0 ||
		maxScore(profile.SellerScores) > 0 ||
		maxScore(profile.BrandScores) > 0
}

func NegativeFeedbackPenalty(counter UserProductCounter) float64 {
	if counter.Counters.WishlistRemoves > counter.Counters.WishlistAdds {
		return 20
	}
	return 0
}

func ScoreAndSortPersonalizedProducts(
	profile UserFeatureProfile,
	counters []UserProductCounter,
	products []ProductFeature,
	weights PersonalizationWeights,
) []ScoredPersonalizedProduct {
	counterByProduct := make(map[string]UserProductCounter, len(counters))
	for _, counter := range counters {
		productID := strings.TrimSpace(counter.ProductID)
		if productID == "" {
			continue
		}
		counterByProduct[productID] = counter
	}

	eligible := make([]ProductFeature, 0, len(products))
	maxDirect := 0.0
	maxPopularity := 0.0
	for _, product := range products {
		if !product.EligibleForPersonalization() {
			continue
		}
		product.ProductID = strings.TrimSpace(product.ProductID)
		if product.ProductID == "" {
			continue
		}
		eligible = append(eligible, product)
		if counter, ok := counterByProduct[product.ProductID]; ok && counter.WeightedScoreInput > 0 {
			maxDirect = math.Max(maxDirect, float64(counter.WeightedScoreInput))
		}
		maxPopularity = math.Max(maxPopularity, product.PopularityScore(DefaultPopularityWeights()))
	}

	maxCategory := maxScore(profile.CategoryScores)
	maxSeller := maxScore(profile.SellerScores)
	maxBrand := maxScore(profile.BrandScores)
	preferredPrices := preferredPriceSet(profile)

	scored := make([]ScoredPersonalizedProduct, 0, len(eligible))
	for _, product := range eligible {
		counter := counterByProduct[product.ProductID]
		priceMatch := 0.0
		if product.Price != nil && preferredPrices[strings.TrimSpace(product.Price.Bucket)] {
			priceMatch = 1
		}
		signals := PersonalizedSignals{
			CategoryAffinity:      NormalizePersonalizedSignal(scoreFor(profile.CategoryScores, product.CategoryID), maxCategory),
			SellerAffinity:        NormalizePersonalizedSignal(scoreFor(profile.SellerScores, product.SellerID), maxSeller),
			BrandAffinity:         NormalizePersonalizedSignal(scoreFor(profile.BrandScores, product.BrandID), maxBrand),
			PriceMatch:            priceMatch,
			DirectProductAffinity: NormalizePersonalizedSignal(float64(maxInt64(counter.WeightedScoreInput, 0)), maxDirect),
			PopularityBaseline:    NormalizePersonalizedSignal(product.PopularityScore(DefaultPopularityWeights()), maxPopularity),
			NegativePenalty:       NegativeFeedbackPenalty(counter),
		}
		score := BehaviorV1Score(signals, weights)
		if score <= 0 {
			continue
		}
		scored = append(scored, ScoredPersonalizedProduct{
			ProductID:             product.ProductID,
			CategoryID:            strings.TrimSpace(product.CategoryID),
			SellerID:              strings.TrimSpace(product.SellerID),
			BrandID:               strings.TrimSpace(product.BrandID),
			Score:                 score,
			Reason:                personalizationReason(signals, counter),
			Signals:               signals,
			DirectProductAffinity: signals.DirectProductAffinity,
			PopularityBaseline:    signals.PopularityBaseline,
			WishlistAdds:          counter.Counters.WishlistAdds,
			WishlistRemoves:       counter.Counters.WishlistRemoves,
			CartAdds:              counter.Counters.CartAdds,
			Purchases:             counter.Counters.Purchases,
		})
	}

	sort.SliceStable(scored, func(i int, j int) bool {
		if scored[i].Score != scored[j].Score {
			return scored[i].Score > scored[j].Score
		}
		if scored[i].DirectProductAffinity != scored[j].DirectProductAffinity {
			return scored[i].DirectProductAffinity > scored[j].DirectProductAffinity
		}
		if scored[i].PopularityBaseline != scored[j].PopularityBaseline {
			return scored[i].PopularityBaseline > scored[j].PopularityBaseline
		}
		return scored[i].ProductID < scored[j].ProductID
	})
	return scored
}

func personalizationReason(signals PersonalizedSignals, counter UserProductCounter) string {
	if signals.CategoryAffinity > 0 && (counter.Counters.CartAdds > 0 || counter.Counters.WishlistAdds > 0 || signals.DirectProductAffinity >= 0.5) {
		return "preferred_category_and_cart_affinity"
	}
	if signals.SellerAffinity >= signals.CategoryAffinity && signals.SellerAffinity > 0 {
		return "preferred_seller"
	}
	if signals.BrandAffinity >= signals.CategoryAffinity && signals.BrandAffinity > 0 {
		return "preferred_brand"
	}
	if signals.PriceMatch > 0 {
		return "preferred_price_range"
	}
	if signals.DirectProductAffinity > 0 {
		return "direct_product_affinity"
	}
	return "personalized_behavior"
}

func preferredPriceSet(profile UserFeatureProfile) map[string]bool {
	values := map[string]bool{}
	if profile.PriceAffinity == nil {
		return values
	}
	for _, bucket := range profile.PriceAffinity.PreferredBuckets {
		bucket = strings.TrimSpace(bucket)
		if bucket != "" {
			values[bucket] = true
		}
	}
	return values
}

func scoreFor(scores map[string]float64, key string) float64 {
	key = strings.TrimSpace(key)
	if key == "" || len(scores) == 0 {
		return 0
	}
	return scores[key]
}

func maxScore(scores map[string]float64) float64 {
	maxValue := 0.0
	for _, score := range scores {
		if score > maxValue {
			maxValue = score
		}
	}
	return maxValue
}

func uniqueTrimmed(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func maxInt64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
