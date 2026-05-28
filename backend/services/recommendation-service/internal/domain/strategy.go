package domain

import (
	"fmt"
	"regexp"
	"strings"
)

type StrategyID string

const (
	StrategySimilarAttributes        StrategyID = "similar_v1_attributes"
	StrategyTrendingRecentActivity   StrategyID = "trending_v1_recent_activity"
	StrategyPersonalizedBehavior     StrategyID = "personalized_v1_behavior"
	StrategyFBTOrderCooccurrence     StrategyID = "fbt_v1_order_cooccurrence"
	StrategyTrendingFallbackGlobal   StrategyID = "trending_v1_fallback_global"
	StrategyTrendingFallbackCategory StrategyID = "trending_v1_fallback_category"
)

const StrategyIDPattern = "<type>_v<version>_<main_logic>"

var strategyIDPattern = regexp.MustCompile(`^[a-z]+_v[0-9]+_[a-z0-9_]+$`)

func DefaultStrategyForType(typ RecommendationType) (StrategyID, error) {
	switch typ {
	case RecommendationTypeSimilarProducts:
		return StrategySimilarAttributes, nil
	case RecommendationTypeTrending:
		return StrategyTrendingRecentActivity, nil
	case RecommendationTypePersonalized:
		return StrategyPersonalizedBehavior, nil
	case RecommendationTypeFrequentlyBoughtTogether:
		return StrategyFBTOrderCooccurrence, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedRecommendationType, typ)
	}
}

func StrategyPrefixForType(typ RecommendationType) (string, error) {
	switch typ {
	case RecommendationTypeSimilarProducts:
		return "similar", nil
	case RecommendationTypeTrending:
		return "trending", nil
	case RecommendationTypePersonalized:
		return "personalized", nil
	case RecommendationTypeFrequentlyBoughtTogether:
		return "fbt", nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedRecommendationType, typ)
	}
}

func RecommendationTypeForStrategy(strategyID StrategyID) (RecommendationType, error) {
	if err := ValidateStrategyID(strategyID); err != nil {
		return "", err
	}
	prefix, _, _ := strings.Cut(strings.TrimSpace(string(strategyID)), "_v")
	switch prefix {
	case "similar":
		return RecommendationTypeSimilarProducts, nil
	case "trending":
		return RecommendationTypeTrending, nil
	case "personalized":
		return RecommendationTypePersonalized, nil
	case "fbt":
		return RecommendationTypeFrequentlyBoughtTogether, nil
	default:
		return "", fmt.Errorf("%w: unknown strategy prefix %s", ErrUnsupportedStrategy, prefix)
	}
}

func ValidateStrategyID(strategyID StrategyID) error {
	value := strings.TrimSpace(string(strategyID))
	if value == "" {
		return fmt.Errorf("%w: strategy_id is required", ErrUnsupportedStrategy)
	}
	if !strategyIDPattern.MatchString(value) {
		return fmt.Errorf("%w: strategy_id must match %s", ErrUnsupportedStrategy, StrategyIDPattern)
	}
	prefix, _, ok := strings.Cut(value, "_v")
	if !ok {
		return fmt.Errorf("%w: strategy_id must include version", ErrUnsupportedStrategy)
	}
	switch prefix {
	case "similar", "trending", "personalized", "fbt":
		return nil
	default:
		return fmt.Errorf("%w: unknown strategy prefix %s", ErrUnsupportedStrategy, prefix)
	}
}
