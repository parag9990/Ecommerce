package repository

import (
	"context"
	"fmt"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

type StaticDefinitionRepository struct {
	types         map[domain.RecommendationType]domain.TypeDefinition
	contexts      map[domain.RecommendationContext]domain.ContextDefinition
	fallbackChain []domain.FallbackStep
}

func NewStaticDefinitionRepository() (*StaticDefinitionRepository, error) {
	repo := &StaticDefinitionRepository{
		types:         buildTypeDefinitions(),
		contexts:      buildContextDefinitions(),
		fallbackChain: buildGlobalFallbackChain(),
	}
	if err := repo.validate(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *StaticDefinitionRepository) ListTypes(ctx context.Context) ([]domain.TypeDefinition, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	supported := domain.SupportedRecommendationTypes()
	types := make([]domain.TypeDefinition, 0, len(supported))
	for _, typ := range supported {
		if definition, ok := r.types[typ]; ok {
			types = append(types, cloneTypeDefinition(definition))
		}
	}
	return types, nil
}

func (r *StaticDefinitionRepository) GetType(ctx context.Context, typ domain.RecommendationType) (domain.TypeDefinition, error) {
	if err := ctx.Err(); err != nil {
		return domain.TypeDefinition{}, err
	}
	definition, ok := r.types[typ]
	if !ok {
		return domain.TypeDefinition{}, fmt.Errorf("%w: type %s", domain.ErrDefinitionNotFound, typ)
	}
	return cloneTypeDefinition(definition), nil
}

func (r *StaticDefinitionRepository) ListContexts(ctx context.Context) ([]domain.ContextDefinition, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	supported := domain.SupportedRecommendationContexts()
	contexts := make([]domain.ContextDefinition, 0, len(supported))
	for _, context := range supported {
		if definition, ok := r.contexts[context]; ok {
			contexts = append(contexts, cloneContextDefinition(definition))
		}
	}
	return contexts, nil
}

func (r *StaticDefinitionRepository) GetContext(ctx context.Context, context domain.RecommendationContext) (domain.ContextDefinition, error) {
	if err := ctx.Err(); err != nil {
		return domain.ContextDefinition{}, err
	}
	definition, ok := r.contexts[context]
	if !ok {
		return domain.ContextDefinition{}, fmt.Errorf("%w: context %s", domain.ErrDefinitionNotFound, context)
	}
	return cloneContextDefinition(definition), nil
}

func (r *StaticDefinitionRepository) FallbackChain(ctx context.Context) ([]domain.FallbackStep, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return cloneFallbackSteps(r.fallbackChain), nil
}

func (r *StaticDefinitionRepository) validate() error {
	for _, typ := range domain.SupportedRecommendationTypes() {
		if _, ok := r.types[typ]; !ok {
			return fmt.Errorf("missing definition for recommendation type %s", typ)
		}
	}
	for typ, definition := range r.types {
		if !typ.IsValid() {
			return fmt.Errorf("invalid recommendation type definition %s", typ)
		}
		if definition.DefaultStrategyID == "" {
			return fmt.Errorf("missing default strategy for recommendation type %s", typ)
		}
		if err := domain.ValidateStrategyID(definition.DefaultStrategyID); err != nil {
			return fmt.Errorf("invalid default strategy for %s: %w", typ, err)
		}
		for _, context := range definition.BestContexts {
			if !context.IsValid() {
				return fmt.Errorf("invalid best context %s for type %s", context, typ)
			}
		}
	}
	for _, context := range domain.SupportedRecommendationContexts() {
		if _, ok := r.contexts[context]; !ok {
			return fmt.Errorf("missing definition for recommendation context %s", context)
		}
	}
	for _, step := range r.fallbackChain {
		if !step.Type.IsValid() {
			return fmt.Errorf("invalid fallback type %s", step.Type)
		}
		if err := domain.ValidateStrategyID(step.StrategyID); err != nil {
			return fmt.Errorf("invalid fallback strategy %s: %w", step.StrategyID, err)
		}
	}
	return nil
}

func buildTypeDefinitions() map[domain.RecommendationType]domain.TypeDefinition {
	return map[domain.RecommendationType]domain.TypeDefinition{
		domain.RecommendationTypeSimilarProducts: {
			Type:        domain.RecommendationTypeSimilarProducts,
			DisplayName: "Similar products",
			Purpose:     "Suggest active, in-stock products similar to the current product.",
			BestContexts: []domain.RecommendationContext{
				domain.ContextProductDetail,
			},
			RequiredInputs: []domain.InputRequirement{
				{Field: "product_id", Required: true, Description: "Base product used to find alternatives."},
			},
			OptionalInputs: []domain.InputRequirement{
				{Field: "category_id", Description: "Narrows similarity to the same catalog category."},
				{Field: "user_id", Description: "Future tie-break signal for known users."},
				{Field: "limit", Description: "Maximum number of products requested."},
			},
			ProductSignals: []domain.SignalDefinition{
				{Source: domain.SignalSourceProductService, Name: "product_id", Description: "Stable product identifier."},
				{Source: domain.SignalSourceProductService, Name: "category_id", Description: "Primary category used as a high-signal similarity boundary."},
				{Source: domain.SignalSourceProductService, Name: "brand_id", Description: "Brand affinity and alternative matching."},
				{Source: domain.SignalSourceProductService, Name: "attributes", Description: "Comparable attributes such as color, size group, material, and style."},
				{Source: domain.SignalSourceProductService, Name: "price", Description: "Price bucket for near-price alternatives."},
				{Source: domain.SignalSourceProductService, Name: "stock_status", Description: "Filters unavailable products from final candidates."},
			},
			ScoringFactors: []domain.ScoringFactor{
				{Name: "same_category", Weight: "high", Description: "Products in the same category are favored."},
				{Name: "shared_attributes", Weight: "high", Description: "Matching attributes increase similarity."},
				{Name: "same_brand", Weight: "medium", Description: "Same-brand alternatives can be boosted."},
				{Name: "similar_price_range", Weight: "medium", Description: "Nearby price bucket keeps recommendations relevant."},
				{Name: "in_stock_active", Weight: "mandatory_filter", Description: "Inactive or out-of-stock products are excluded."},
			},
			OutputFields: []domain.OutputField{
				{Field: "product_id", Public: true, Description: "Product identifier returned to the caller."},
				{Field: "score", Public: false, Description: "Internal similarity confidence."},
				{Field: "reason", Public: false, Description: "Internal explanation such as same_category_shared_attributes."},
			},
			DefaultStrategyID: domain.StrategySimilarAttributes,
			Fallbacks: []domain.FallbackStep{
				{Priority: 1, Name: "same_category_popular", Type: domain.RecommendationTypeTrending, StrategyID: domain.StrategyTrendingFallbackCategory, Description: "Popular products from the same category."},
				{Priority: 2, Name: "same_seller_popular", Type: domain.RecommendationTypeTrending, StrategyID: domain.StrategyTrendingRecentActivity, Description: "Popular products from the same seller."},
				{Priority: 3, Name: "global_trending", Type: domain.RecommendationTypeTrending, StrategyID: domain.StrategyTrendingFallbackGlobal, Description: "Global trending products."},
			},
			ColdStartFallbackNotes: []string{
				"Use same category popular products when similar candidates are sparse.",
				"Use global trending only after scoped fallbacks are empty.",
			},
		},
		domain.RecommendationTypeTrending: {
			Type:        domain.RecommendationTypeTrending,
			DisplayName: "Trending",
			Purpose:     "Show products that are currently popular for anonymous or broad discovery contexts.",
			BestContexts: []domain.RecommendationContext{
				domain.ContextHomeFeed,
				domain.ContextCategoryListing,
				domain.ContextSearchResults,
				domain.ContextSellerStore,
			},
			RequiredInputs: []domain.InputRequirement{
				{Field: "context", Required: true, Description: "Page or journey context requesting the list."},
			},
			OptionalInputs: []domain.InputRequirement{
				{Field: "category_id", Description: "Scopes popularity to one category."},
				{Field: "seller_id", Description: "Scopes popularity to one seller storefront."},
				{Field: "limit", Description: "Maximum number of products requested."},
			},
			ProductSignals: []domain.SignalDefinition{
				{Source: domain.SignalSourceProductService, Name: "category_id", Description: "Supports category-specific trending lists."},
				{Source: domain.SignalSourceProductService, Name: "seller_id", Description: "Supports seller-store trending lists."},
				{Source: domain.SignalSourceProductService, Name: "stock_status", Description: "Filters unavailable products."},
			},
			BehaviorSignals: []domain.SignalDefinition{
				{Source: domain.SignalSourceSessionService, Name: "product_view", Description: "Views in the last 24 hours supply a fresh attention signal."},
				{Source: domain.SignalSourceWishlistService, Name: "wishlist_add", Description: "Wishlist adds in the last 7 days supply a preference signal."},
				{Source: domain.SignalSourceSessionService, Name: "add_to_cart", Description: "Cart additions in the last 7 days supply a strong intent signal."},
				{Source: domain.SignalSourceOrderService, Name: "purchase", Description: "Purchases in the last 7 days supply the conversion signal."},
			},
			ScoringFactors: []domain.ScoringFactor{
				{Name: "views_24h", Weight: "1", Description: "Recent views create base demand."},
				{Name: "wishlist_adds_7d", Weight: "3", Description: "Wishlist adds indicate future purchase interest."},
				{Name: "cart_adds_7d", Weight: "4", Description: "Cart additions indicate strong purchase intent."},
				{Name: "purchases_7d", Weight: "8", Description: "Purchases indicate proven conversion."},
			},
			OutputFields: []domain.OutputField{
				{Field: "product_id", Public: true, Description: "Product identifier returned to the caller."},
				{Field: "score", Public: false, Description: "Internal trending score."},
				{Field: "reason", Public: false, Description: "Internal reason such as recent_activity."},
			},
			DefaultStrategyID: domain.StrategyTrendingRecentActivity,
			Fallbacks: []domain.FallbackStep{
				{Priority: 1, Name: "category_popular", Type: domain.RecommendationTypeTrending, StrategyID: domain.StrategyTrendingFallbackCategory, Description: "Category popular products."},
				{Priority: 2, Name: "global_popular", Type: domain.RecommendationTypeTrending, StrategyID: domain.StrategyTrendingFallbackGlobal, Description: "Global popular products."},
				{Priority: 3, Name: "curated_default", Type: domain.RecommendationTypeTrending, StrategyID: domain.StrategyTrendingFallbackGlobal, Description: "Manually curated default list."},
			},
			ColdStartFallbackNotes: []string{
				"Trending works for anonymous users because it does not require personal history.",
				"Use curated defaults when recent event data is unavailable.",
			},
		},
		domain.RecommendationTypePersonalized: {
			Type:        domain.RecommendationTypePersonalized,
			DisplayName: "Personalized",
			Purpose:     "Rank products for a known user or guest based on behavior and current context.",
			BestContexts: []domain.RecommendationContext{
				domain.ContextHomeFeed,
				domain.ContextCategoryListing,
				domain.ContextSearchResults,
			},
			RequiredInputs: []domain.InputRequirement{
				{Field: "context", Required: true, Description: "Page or journey context requesting the personalized list."},
			},
			RequiredAnyOf: []domain.AlternativeInputRequirement{
				{Fields: []string{"user_id", "anonymous_id"}, Description: "At least one identity reference is needed for personalized mode."},
			},
			OptionalInputs: []domain.InputRequirement{
				{Field: "category_id", Description: "Boosts the current browse category."},
				{Field: "limit", Description: "Maximum number of products requested."},
			},
			ProductSignals: []domain.SignalDefinition{
				{Source: domain.SignalSourceProductService, Name: "category_id", Description: "Maps interaction history to categories."},
				{Source: domain.SignalSourceProductService, Name: "brand_id", Description: "Supports brand-affinity boosts."},
				{Source: domain.SignalSourceProductService, Name: "attributes", Description: "Supports preference matching."},
				{Source: domain.SignalSourceProductService, Name: "stock_status", Description: "Filters unavailable products."},
			},
			BehaviorSignals: []domain.SignalDefinition{
				{Source: domain.SignalSourceSessionService, Name: "product_view", Description: "Low-strength interest signal."},
				{Source: domain.SignalSourceSessionService, Name: "product_click", Description: "Medium-strength interest signal."},
				{Source: domain.SignalSourceWishlistService, Name: "wishlist_add", Description: "High-strength preference signal."},
				{Source: domain.SignalSourceSessionService, Name: "add_to_cart", Description: "Very high-strength purchase intent signal."},
				{Source: domain.SignalSourceOrderService, Name: "purchase", Description: "Very high-strength conversion signal."},
				{Source: domain.SignalSourceSessionService, Name: "search_query", Description: "Captures explicit current intent."},
			},
			ScoringFactors: []domain.ScoringFactor{
				{Name: "recent_product_views", Weight: "low", Description: "Recent browsing interests seed personalization."},
				{Name: "product_clicks", Weight: "medium", Description: "Clicked products influence categories and attributes."},
				{Name: "wishlist_adds", Weight: "high", Description: "Wishlist products shape preference signals."},
				{Name: "add_to_cart", Weight: "very_high", Description: "Cart behavior strongly influences ranking."},
				{Name: "purchases", Weight: "very_high", Description: "Purchase history is the strongest conversion signal."},
			},
			OutputFields: []domain.OutputField{
				{Field: "product_id", Public: true, Description: "Product identifier returned to the caller."},
				{Field: "score", Public: false, Description: "Internal personalized ranking score."},
				{Field: "reason", Public: false, Description: "Internal reason such as behavior_category_affinity."},
			},
			DefaultStrategyID: domain.StrategyPersonalizedBehavior,
			Fallbacks: []domain.FallbackStep{
				{Priority: 1, Name: "current_category_popular", Type: domain.RecommendationTypeTrending, StrategyID: domain.StrategyTrendingFallbackCategory, Description: "Popular products from the current category."},
				{Priority: 2, Name: "recent_trending", Type: domain.RecommendationTypeTrending, StrategyID: domain.StrategyTrendingRecentActivity, Description: "Recent trending products."},
				{Priority: 3, Name: "homepage_curated", Type: domain.RecommendationTypeTrending, StrategyID: domain.StrategyTrendingFallbackGlobal, Description: "Curated products for cold-start users."},
			},
			ColdStartFallbackNotes: []string{
				"Do not use raw sensitive personal data.",
				"Fall back to category popular and trending when user behavior is missing.",
			},
		},
		domain.RecommendationTypeFrequentlyBoughtTogether: {
			Type:        domain.RecommendationTypeFrequentlyBoughtTogether,
			DisplayName: "Frequently bought together",
			Purpose:     "Suggest complementary products for product detail, cart, and checkout journeys.",
			BestContexts: []domain.RecommendationContext{
				domain.ContextProductDetail,
				domain.ContextCart,
				domain.ContextCheckout,
			},
			RequiredAnyOf: []domain.AlternativeInputRequirement{
				{Fields: []string{"product_id", "cart_product_ids"}, Description: "Product page can send product_id; cart and checkout can send cart_product_ids."},
			},
			OptionalInputs: []domain.InputRequirement{
				{Field: "category_id", Description: "Restricts complements to relevant categories."},
				{Field: "limit", Description: "Maximum number of products requested."},
			},
			ProductSignals: []domain.SignalDefinition{
				{Source: domain.SignalSourceProductService, Name: "product_id", Description: "Base product for complements."},
				{Source: domain.SignalSourceProductService, Name: "category_id", Description: "Supports accessory/category add-on rules."},
				{Source: domain.SignalSourceProductService, Name: "price", Description: "Supports low-risk add-on selection near checkout."},
				{Source: domain.SignalSourceProductService, Name: "stock_status", Description: "Filters unavailable add-ons."},
			},
			BehaviorSignals: []domain.SignalDefinition{
				{Source: domain.SignalSourceSessionService, Name: "add_to_cart", Description: "Cart co-occurrence seed signal."},
				{Source: domain.SignalSourceOrderService, Name: "purchase", Description: "Order co-occurrence signal."},
			},
			ScoringFactors: []domain.ScoringFactor{
				{Name: "times_bought_with_base_product", Weight: "numerator", Description: "Number of orders containing both products."},
				{Name: "total_orders_with_base_product", Weight: "denominator", Description: "Normalizes complement strength by base-product order volume."},
				{Name: "low_risk_checkout_addon", Weight: "context_boost", Description: "Checkout should prefer useful, lower-friction add-ons."},
			},
			OutputFields: []domain.OutputField{
				{Field: "product_id", Public: true, Description: "Product identifier returned to the caller."},
				{Field: "score", Public: false, Description: "Internal co-occurrence confidence."},
				{Field: "reason", Public: false, Description: "Internal reason such as bought_with_base_product."},
			},
			DefaultStrategyID: domain.StrategyFBTOrderCooccurrence,
			Fallbacks: []domain.FallbackStep{
				{Priority: 1, Name: "same_category_accessories", Type: domain.RecommendationTypeTrending, StrategyID: domain.StrategyTrendingFallbackCategory, Description: "Accessory-like products from the same category."},
				{Priority: 2, Name: "similar_products", Type: domain.RecommendationTypeSimilarProducts, StrategyID: domain.StrategySimilarAttributes, Description: "Similar products when complement data is unavailable."},
				{Priority: 3, Name: "category_popular_low_price_addons", Type: domain.RecommendationTypeTrending, StrategyID: domain.StrategyTrendingFallbackCategory, Description: "Popular low-risk add-ons in the category."},
			},
			ColdStartFallbackNotes: []string{
				"Use same-category accessories when co-occurrence data is sparse.",
				"Keep checkout recommendations low-risk and available.",
			},
		},
	}
}

func buildContextDefinitions() map[domain.RecommendationContext]domain.ContextDefinition {
	return map[domain.RecommendationContext]domain.ContextDefinition{
		domain.ContextHomeFeed: {
			Context:        domain.ContextHomeFeed,
			Description:    "Homepage feed and recommended-for-you surfaces.",
			PreferredTypes: []domain.RecommendationType{domain.RecommendationTypePersonalized, domain.RecommendationTypeTrending},
			FallbackTypes:  []domain.RecommendationType{domain.RecommendationTypeTrending},
		},
		domain.ContextProductDetail: {
			Context:        domain.ContextProductDetail,
			Description:    "Product detail page recommendations near the current product.",
			PreferredTypes: []domain.RecommendationType{domain.RecommendationTypeSimilarProducts, domain.RecommendationTypeFrequentlyBoughtTogether},
			FallbackTypes:  []domain.RecommendationType{domain.RecommendationTypeTrending},
		},
		domain.ContextCategoryListing: {
			Context:        domain.ContextCategoryListing,
			Description:    "Category listing or browse page.",
			PreferredTypes: []domain.RecommendationType{domain.RecommendationTypeTrending, domain.RecommendationTypePersonalized},
			FallbackTypes:  []domain.RecommendationType{domain.RecommendationTypeTrending},
		},
		domain.ContextCart: {
			Context:        domain.ContextCart,
			Description:    "Cart page add-ons and complementary products.",
			PreferredTypes: []domain.RecommendationType{domain.RecommendationTypeFrequentlyBoughtTogether},
			FallbackTypes:  []domain.RecommendationType{domain.RecommendationTypeTrending},
		},
		domain.ContextCheckout: {
			Context:        domain.ContextCheckout,
			Description:    "Checkout page low-risk add-ons.",
			PreferredTypes: []domain.RecommendationType{domain.RecommendationTypeFrequentlyBoughtTogether},
			FallbackTypes:  []domain.RecommendationType{domain.RecommendationTypeTrending},
		},
		domain.ContextSearchResults: {
			Context:        domain.ContextSearchResults,
			Description:    "Search results page recommendations informed by query and category intent.",
			PreferredTypes: []domain.RecommendationType{domain.RecommendationTypeTrending, domain.RecommendationTypePersonalized},
			FallbackTypes:  []domain.RecommendationType{domain.RecommendationTypeTrending},
		},
		domain.ContextSellerStore: {
			Context:        domain.ContextSellerStore,
			Description:    "Seller storefront popular products.",
			PreferredTypes: []domain.RecommendationType{domain.RecommendationTypeTrending},
			FallbackTypes:  []domain.RecommendationType{domain.RecommendationTypeTrending},
		},
	}
}

func buildGlobalFallbackChain() []domain.FallbackStep {
	return []domain.FallbackStep{
		{Priority: 1, Name: "same_context_recommendation", Type: domain.RecommendationTypeTrending, StrategyID: domain.StrategyTrendingRecentActivity, Description: "Try the most appropriate recommendation for the same page context."},
		{Priority: 2, Name: "same_category_popular", Type: domain.RecommendationTypeTrending, StrategyID: domain.StrategyTrendingFallbackCategory, Description: "Use popular products from the current category."},
		{Priority: 3, Name: "seller_popular", Type: domain.RecommendationTypeTrending, StrategyID: domain.StrategyTrendingRecentActivity, Description: "Use popular products from the current seller when seller_id is available."},
		{Priority: 4, Name: "global_trending", Type: domain.RecommendationTypeTrending, StrategyID: domain.StrategyTrendingFallbackGlobal, Description: "Use global trending products."},
		{Priority: 5, Name: "curated_default", Type: domain.RecommendationTypeTrending, StrategyID: domain.StrategyTrendingFallbackGlobal, Description: "Use a curated default list as the final fallback."},
	}
}

func cloneTypeDefinition(in domain.TypeDefinition) domain.TypeDefinition {
	out := in
	out.BestContexts = append([]domain.RecommendationContext(nil), in.BestContexts...)
	out.RequiredInputs = append([]domain.InputRequirement(nil), in.RequiredInputs...)
	out.RequiredAnyOf = cloneAlternativeInputRequirements(in.RequiredAnyOf)
	out.OptionalInputs = append([]domain.InputRequirement(nil), in.OptionalInputs...)
	out.ProductSignals = append([]domain.SignalDefinition(nil), in.ProductSignals...)
	out.BehaviorSignals = append([]domain.SignalDefinition(nil), in.BehaviorSignals...)
	out.ScoringFactors = append([]domain.ScoringFactor(nil), in.ScoringFactors...)
	out.OutputFields = append([]domain.OutputField(nil), in.OutputFields...)
	out.Fallbacks = cloneFallbackSteps(in.Fallbacks)
	out.ColdStartFallbackNotes = append([]string(nil), in.ColdStartFallbackNotes...)
	return out
}

func cloneContextDefinition(in domain.ContextDefinition) domain.ContextDefinition {
	out := in
	out.PreferredTypes = append([]domain.RecommendationType(nil), in.PreferredTypes...)
	out.FallbackTypes = append([]domain.RecommendationType(nil), in.FallbackTypes...)
	return out
}

func cloneAlternativeInputRequirements(in []domain.AlternativeInputRequirement) []domain.AlternativeInputRequirement {
	out := make([]domain.AlternativeInputRequirement, 0, len(in))
	for _, requirement := range in {
		cloned := requirement
		cloned.Fields = append([]string(nil), requirement.Fields...)
		out = append(out, cloned)
	}
	return out
}

func cloneFallbackSteps(in []domain.FallbackStep) []domain.FallbackStep {
	out := make([]domain.FallbackStep, len(in))
	copy(out, in)
	return out
}
