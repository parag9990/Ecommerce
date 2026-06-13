package grpctransport

import (
	"fmt"
	"time"

	recommendationv1 "github.com/example/ecommerce-platform/backend/proto-gen/go/ecommerce/recommendation/v1"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/usecase"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func inputFromProto(req *recommendationv1.GetRecommendationsRequest, requestID string, sessionID string) (usecase.GetRecommendationsInput, error) {
	if req == nil {
		return usecase.GetRecommendationsInput{}, fmt.Errorf("%w: request is required", domain.ErrInvalidRecommendationRequest)
	}
	if req.GetLimit() < 0 {
		return usecase.GetRecommendationsInput{}, fmt.Errorf("%w: limit cannot be negative", domain.ErrInvalidLimit)
	}

	contextValue, err := contextFromProto(req.GetContext())
	if err != nil {
		return usecase.GetRecommendationsInput{}, err
	}
	typeValue, err := typeFromProto(req.GetType())
	if err != nil {
		return usecase.GetRecommendationsInput{}, err
	}

	return usecase.GetRecommendationsInput{
		RequestID:      requestID,
		UserID:         req.GetUserId(),
		AnonymousID:    req.GetAnonymousId(),
		SessionID:      sessionID,
		ProductID:      req.GetProductId(),
		CategoryID:     req.GetCategoryId(),
		SellerID:       req.GetSellerId(),
		CartProductIDs: append([]string(nil), req.GetCartProductIds()...),
		Context:        contextValue,
		Type:           typeValue,
		Limit:          int(req.GetLimit()),
	}, nil
}

func responseFromDomain(output usecase.GetRecommendationsOutput) *recommendationv1.GetRecommendationsResponse {
	result := output.Result
	items := make([]*recommendationv1.RecommendationItem, 0, len(result.Items))
	for index, item := range result.Items {
		items = append(items, &recommendationv1.RecommendationItem{
			ProductId: item.ProductID,
			Rank:      int32(index + 1),
			Score:     item.Score,
			Reason:    item.Reason,
		})
	}

	var generatedAt *timestamppb.Timestamp
	if !result.GeneratedAt.IsZero() {
		generatedAt = timestamppb.New(result.GeneratedAt.UTC())
	}
	return &recommendationv1.GetRecommendationsResponse{
		RecommendationId: result.RecommendationID,
		Type:             typeToProto(result.Type),
		StrategyId:       string(result.StrategyID),
		Items:            items,
		GeneratedAt:      generatedAt,
		CacheTtlSeconds:  int64(nonNegativeDuration(output.CacheTTL).Seconds()),
	}
}

func contextFromProto(value recommendationv1.RecommendationContext) (domain.RecommendationContext, error) {
	switch value {
	case recommendationv1.RecommendationContext_RECOMMENDATION_CONTEXT_HOME_FEED:
		return domain.ContextHomeFeed, nil
	case recommendationv1.RecommendationContext_RECOMMENDATION_CONTEXT_PRODUCT_DETAIL:
		return domain.ContextProductDetail, nil
	case recommendationv1.RecommendationContext_RECOMMENDATION_CONTEXT_CATEGORY_LISTING:
		return domain.ContextCategoryListing, nil
	case recommendationv1.RecommendationContext_RECOMMENDATION_CONTEXT_CART:
		return domain.ContextCart, nil
	case recommendationv1.RecommendationContext_RECOMMENDATION_CONTEXT_CHECKOUT:
		return domain.ContextCheckout, nil
	case recommendationv1.RecommendationContext_RECOMMENDATION_CONTEXT_SEARCH_RESULTS:
		return domain.ContextSearchResults, nil
	case recommendationv1.RecommendationContext_RECOMMENDATION_CONTEXT_SELLER_STORE:
		return domain.ContextSellerStore, nil
	case recommendationv1.RecommendationContext_RECOMMENDATION_CONTEXT_UNSPECIFIED:
		return "", fmt.Errorf("%w: context is required", domain.ErrUnsupportedContext)
	default:
		return "", fmt.Errorf("%w: unknown context enum %d", domain.ErrUnsupportedContext, value)
	}
}

func typeFromProto(value recommendationv1.RecommendationType) (domain.RecommendationType, error) {
	switch value {
	case recommendationv1.RecommendationType_RECOMMENDATION_TYPE_UNSPECIFIED:
		return "", nil
	case recommendationv1.RecommendationType_RECOMMENDATION_TYPE_SIMILAR_PRODUCTS:
		return domain.RecommendationTypeSimilarProducts, nil
	case recommendationv1.RecommendationType_RECOMMENDATION_TYPE_TRENDING:
		return domain.RecommendationTypeTrending, nil
	case recommendationv1.RecommendationType_RECOMMENDATION_TYPE_PERSONALIZED:
		return domain.RecommendationTypePersonalized, nil
	case recommendationv1.RecommendationType_RECOMMENDATION_TYPE_FREQUENTLY_BOUGHT_TOGETHER:
		return domain.RecommendationTypeFrequentlyBoughtTogether, nil
	default:
		return "", fmt.Errorf("%w: unknown type enum %d", domain.ErrUnsupportedRecommendationType, value)
	}
}

func typeToProto(value domain.RecommendationType) recommendationv1.RecommendationType {
	switch value {
	case domain.RecommendationTypeSimilarProducts:
		return recommendationv1.RecommendationType_RECOMMENDATION_TYPE_SIMILAR_PRODUCTS
	case domain.RecommendationTypeTrending:
		return recommendationv1.RecommendationType_RECOMMENDATION_TYPE_TRENDING
	case domain.RecommendationTypePersonalized:
		return recommendationv1.RecommendationType_RECOMMENDATION_TYPE_PERSONALIZED
	case domain.RecommendationTypeFrequentlyBoughtTogether:
		return recommendationv1.RecommendationType_RECOMMENDATION_TYPE_FREQUENTLY_BOUGHT_TOGETHER
	default:
		return recommendationv1.RecommendationType_RECOMMENDATION_TYPE_UNSPECIFIED
	}
}

func nonNegativeDuration(duration time.Duration) time.Duration {
	if duration < 0 {
		return 0
	}
	return duration
}
