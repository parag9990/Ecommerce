package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

type ResolveInput struct {
	RequestID      string
	UserID         string
	AnonymousID    string
	SessionID      string
	ProductID      string
	CategoryID     string
	SellerID       string
	CartProductIDs []string
	Context        domain.RecommendationContext
	Limit          int
	Type           domain.RecommendationType
}

type Resolution struct {
	Request        domain.RecommendationRequest
	Type           domain.RecommendationType
	Context        domain.ContextDefinition
	Definition     domain.TypeDefinition
	StrategyID     domain.StrategyID
	Fallbacks      []domain.FallbackStep
	StrategyFormat string
}

func (s *DefinitionService) ListTypes(ctx context.Context) ([]domain.TypeDefinition, error) {
	return s.repo.ListTypes(ctx)
}

func (s *DefinitionService) GetType(ctx context.Context, typ domain.RecommendationType) (domain.TypeDefinition, error) {
	if !typ.IsValid() {
		return domain.TypeDefinition{}, fmt.Errorf("%w: %s", domain.ErrUnsupportedRecommendationType, typ)
	}
	return s.repo.GetType(ctx, typ)
}

func (s *DefinitionService) ListContexts(ctx context.Context) ([]domain.ContextDefinition, error) {
	return s.repo.ListContexts(ctx)
}

func (s *DefinitionService) Resolve(ctx context.Context, input ResolveInput) (Resolution, error) {
	req := domain.RecommendationRequest{
		UserID:         input.UserID,
		AnonymousID:    input.AnonymousID,
		SessionID:      input.SessionID,
		ProductID:      input.ProductID,
		CategoryID:     input.CategoryID,
		SellerID:       input.SellerID,
		CartProductIDs: input.CartProductIDs,
		Context:        input.Context,
		Limit:          input.Limit,
		Type:           input.Type,
	}.Normalize()

	if err := req.Validate(s.maxIdentifierLength); err != nil {
		return Resolution{}, err
	}
	if req.Limit == 0 {
		req.Limit = s.defaultLimit
	}
	if req.Limit > s.maxLimit {
		return Resolution{}, fmt.Errorf("%w: limit cannot exceed %d", domain.ErrInvalidLimit, s.maxLimit)
	}

	resolvedType := domain.ResolveRecommendationType(req)
	resolvedRequest := req
	resolvedRequest.Type = resolvedType
	if err := resolvedRequest.Validate(s.maxIdentifierLength); err != nil {
		return Resolution{}, err
	}

	definition, err := s.repo.GetType(ctx, resolvedType)
	if err != nil {
		return Resolution{}, err
	}
	contextDefinition, err := s.repo.GetContext(ctx, req.Context)
	if err != nil {
		return Resolution{}, err
	}
	fallbacks := definition.Fallbacks
	if len(fallbacks) == 0 {
		fallbacks, err = s.repo.FallbackChain(ctx)
		if err != nil {
			return Resolution{}, err
		}
	}

	s.logger.InfoContext(ctx, "recommendation.type_resolved",
		slog.String("request_id", input.RequestID),
		slog.String("context", string(req.Context)),
		slog.String("type", string(resolvedType)),
		slog.String("strategy_id", string(definition.DefaultStrategyID)),
		slog.Int("limit", req.Limit),
	)

	return Resolution{
		Request:        resolvedRequest,
		Type:           resolvedType,
		Context:        contextDefinition,
		Definition:     definition,
		StrategyID:     definition.DefaultStrategyID,
		Fallbacks:      append([]domain.FallbackStep(nil), fallbacks...),
		StrategyFormat: domain.StrategyIDPattern,
	}, nil
}
