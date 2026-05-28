package domain

import (
	"fmt"
	"strings"
	"time"
)

type RecommendationType string

const (
	RecommendationTypeSimilarProducts          RecommendationType = "similar_products"
	RecommendationTypeTrending                 RecommendationType = "trending"
	RecommendationTypePersonalized             RecommendationType = "personalized"
	RecommendationTypeFrequentlyBoughtTogether RecommendationType = "frequently_bought_together"
)

type RecommendationContext string

const (
	ContextHomeFeed        RecommendationContext = "home_feed"
	ContextProductDetail   RecommendationContext = "product_detail"
	ContextCategoryListing RecommendationContext = "category_listing"
	ContextCart            RecommendationContext = "cart"
	ContextCheckout        RecommendationContext = "checkout"
	ContextSearchResults   RecommendationContext = "search_results"
	ContextSellerStore     RecommendationContext = "seller_store"
)

type SignalSource string

const (
	SignalSourceProductService  SignalSource = "product_service"
	SignalSourceSessionService  SignalSource = "session_service"
	SignalSourceOrderService    SignalSource = "order_service"
	SignalSourceWishlistService SignalSource = "wishlist_service"
)

type InputRequirement struct {
	Field       string
	Required    bool
	Description string
}

type AlternativeInputRequirement struct {
	Fields      []string
	Description string
}

type SignalDefinition struct {
	Source      SignalSource
	Name        string
	Description string
}

type ScoringFactor struct {
	Name        string
	Weight      string
	Description string
}

type FallbackStep struct {
	Priority    int
	Name        string
	Type        RecommendationType
	StrategyID  StrategyID
	Description string
}

type OutputField struct {
	Field       string
	Public      bool
	Description string
}

type TypeDefinition struct {
	Type                   RecommendationType
	DisplayName            string
	Purpose                string
	BestContexts           []RecommendationContext
	RequiredInputs         []InputRequirement
	RequiredAnyOf          []AlternativeInputRequirement
	OptionalInputs         []InputRequirement
	ProductSignals         []SignalDefinition
	BehaviorSignals        []SignalDefinition
	ScoringFactors         []ScoringFactor
	OutputFields           []OutputField
	DefaultStrategyID      StrategyID
	Fallbacks              []FallbackStep
	ColdStartFallbackNotes []string
}

type ContextDefinition struct {
	Context        RecommendationContext
	Description    string
	PreferredTypes []RecommendationType
	FallbackTypes  []RecommendationType
}

type RecommendationRequest struct {
	UserID         string
	AnonymousID    string
	SessionID      string
	ProductID      string
	CategoryID     string
	SellerID       string
	CartProductIDs []string
	Context        RecommendationContext
	Limit          int
	Type           RecommendationType
}

type RecommendationItem struct {
	ProductID string
	Score     float64
	Reason    string
}

type RecommendationResult struct {
	RecommendationID string
	Type             RecommendationType
	StrategyID       StrategyID
	Items            []RecommendationItem
	GeneratedAt      time.Time
}

func SupportedRecommendationTypes() []RecommendationType {
	return []RecommendationType{
		RecommendationTypeSimilarProducts,
		RecommendationTypeTrending,
		RecommendationTypePersonalized,
		RecommendationTypeFrequentlyBoughtTogether,
	}
}

func SupportedRecommendationContexts() []RecommendationContext {
	return []RecommendationContext{
		ContextHomeFeed,
		ContextProductDetail,
		ContextCategoryListing,
		ContextCart,
		ContextCheckout,
		ContextSearchResults,
		ContextSellerStore,
	}
}

func ParseRecommendationType(value string) (RecommendationType, error) {
	typ := RecommendationType(strings.TrimSpace(value))
	if typ.IsValid() {
		return typ, nil
	}
	return "", fmt.Errorf("%w: %s", ErrUnsupportedRecommendationType, value)
}

func ParseRecommendationContext(value string) (RecommendationContext, error) {
	context := RecommendationContext(strings.TrimSpace(value))
	if context.IsValid() {
		return context, nil
	}
	return "", fmt.Errorf("%w: %s", ErrUnsupportedContext, value)
}

func (t RecommendationType) IsValid() bool {
	switch t {
	case RecommendationTypeSimilarProducts,
		RecommendationTypeTrending,
		RecommendationTypePersonalized,
		RecommendationTypeFrequentlyBoughtTogether:
		return true
	default:
		return false
	}
}

func (c RecommendationContext) IsValid() bool {
	switch c {
	case ContextHomeFeed,
		ContextProductDetail,
		ContextCategoryListing,
		ContextCart,
		ContextCheckout,
		ContextSearchResults,
		ContextSellerStore:
		return true
	default:
		return false
	}
}

func (s SignalSource) IsValid() bool {
	switch s {
	case SignalSourceProductService,
		SignalSourceSessionService,
		SignalSourceOrderService,
		SignalSourceWishlistService:
		return true
	default:
		return false
	}
}

func ResolveRecommendationType(req RecommendationRequest) RecommendationType {
	if req.Type.IsValid() {
		return req.Type
	}
	if req.Context == ContextCart || req.Context == ContextCheckout {
		return RecommendationTypeFrequentlyBoughtTogether
	}
	if req.Context == ContextProductDetail && strings.TrimSpace(req.ProductID) != "" {
		return RecommendationTypeSimilarProducts
	}
	if strings.TrimSpace(req.UserID) != "" || strings.TrimSpace(req.AnonymousID) != "" {
		return RecommendationTypePersonalized
	}
	return RecommendationTypeTrending
}

func (r RecommendationRequest) Normalize() RecommendationRequest {
	normalized := RecommendationRequest{
		UserID:      strings.TrimSpace(r.UserID),
		AnonymousID: strings.TrimSpace(r.AnonymousID),
		SessionID:   strings.TrimSpace(r.SessionID),
		ProductID:   strings.TrimSpace(r.ProductID),
		CategoryID:  strings.TrimSpace(r.CategoryID),
		SellerID:    strings.TrimSpace(r.SellerID),
		Context:     RecommendationContext(strings.TrimSpace(string(r.Context))),
		Type:        RecommendationType(strings.TrimSpace(string(r.Type))),
		Limit:       r.Limit,
	}
	if len(r.CartProductIDs) > 0 {
		seen := make(map[string]struct{}, len(r.CartProductIDs))
		normalized.CartProductIDs = make([]string, 0, len(r.CartProductIDs))
		for _, productID := range r.CartProductIDs {
			productID = strings.TrimSpace(productID)
			if productID == "" {
				continue
			}
			if _, ok := seen[productID]; ok {
				continue
			}
			seen[productID] = struct{}{}
			normalized.CartProductIDs = append(normalized.CartProductIDs, productID)
		}
	}
	return normalized
}

func (r RecommendationRequest) Validate(maxIdentifierLength int) error {
	req := r.Normalize()
	if !req.Context.IsValid() {
		return fmt.Errorf("%w: context is required", ErrUnsupportedContext)
	}
	if req.Type != "" && !req.Type.IsValid() {
		return fmt.Errorf("%w: %s", ErrUnsupportedRecommendationType, req.Type)
	}
	if req.Limit < 0 {
		return fmt.Errorf("%w: limit cannot be negative", ErrInvalidLimit)
	}
	if maxIdentifierLength <= 0 {
		maxIdentifierLength = 128
	}
	for field, value := range map[string]string{
		"user_id":      req.UserID,
		"anonymous_id": req.AnonymousID,
		"session_id":   req.SessionID,
		"product_id":   req.ProductID,
		"category_id":  req.CategoryID,
		"seller_id":    req.SellerID,
	} {
		if len(value) > maxIdentifierLength {
			return fmt.Errorf("%w: %s is too long", ErrInvalidRecommendationRequest, field)
		}
	}
	for _, productID := range req.CartProductIDs {
		if len(productID) > maxIdentifierLength {
			return fmt.Errorf("%w: cart_product_ids contains an identifier that is too long", ErrInvalidRecommendationRequest)
		}
	}
	if req.Type.IsValid() {
		return validateTypeSpecificInputs(req)
	}
	return nil
}

func validateTypeSpecificInputs(req RecommendationRequest) error {
	switch req.Type {
	case RecommendationTypeSimilarProducts:
		if req.ProductID == "" {
			return fmt.Errorf("%w: product_id is required for %s", ErrInvalidRecommendationRequest, req.Type)
		}
	case RecommendationTypeTrending:
		return nil
	case RecommendationTypePersonalized:
		if req.UserID == "" && req.AnonymousID == "" {
			return fmt.Errorf("%w: user_id or anonymous_id is required for %s", ErrInvalidRecommendationRequest, req.Type)
		}
	case RecommendationTypeFrequentlyBoughtTogether:
		if req.ProductID == "" && len(req.CartProductIDs) == 0 {
			return fmt.Errorf("%w: product_id or cart_product_ids is required for %s", ErrInvalidRecommendationRequest, req.Type)
		}
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedRecommendationType, req.Type)
	}
	return nil
}
