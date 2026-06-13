package httptransport

import (
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/usecase"
)

type resolveTypeRequest struct {
	UserID         string   `json:"user_id,omitempty"`
	AnonymousID    string   `json:"anonymous_id,omitempty"`
	SessionID      string   `json:"session_id,omitempty"`
	ProductID      string   `json:"product_id,omitempty"`
	CategoryID     string   `json:"category_id,omitempty"`
	SellerID       string   `json:"seller_id,omitempty"`
	CartProductIDs []string `json:"cart_product_ids,omitempty"`
	Context        string   `json:"context"`
	Limit          int      `json:"limit,omitempty"`
	Type           string   `json:"type,omitempty"`
}

type resolveTypeResponse struct {
	Request          recommendationRequestResponse `json:"request"`
	Type             string                        `json:"type"`
	StrategyID       string                        `json:"strategy_id"`
	StrategyIDFormat string                        `json:"strategy_id_format"`
	Context          contextDefinitionResponse     `json:"context"`
	Definition       typeDefinitionResponse        `json:"definition"`
	Fallbacks        []fallbackStepResponse        `json:"fallbacks"`
}

type recommendationRequestResponse struct {
	UserID         string   `json:"user_id,omitempty"`
	AnonymousID    string   `json:"anonymous_id,omitempty"`
	SessionID      string   `json:"session_id,omitempty"`
	ProductID      string   `json:"product_id,omitempty"`
	CategoryID     string   `json:"category_id,omitempty"`
	SellerID       string   `json:"seller_id,omitempty"`
	CartProductIDs []string `json:"cart_product_ids,omitempty"`
	Context        string   `json:"context"`
	Limit          int      `json:"limit"`
	Type           string   `json:"type,omitempty"`
}

type typeListResponse struct {
	Types            []typeDefinitionResponse `json:"types"`
	StrategyIDFormat string                   `json:"strategy_id_format"`
}

type contextListResponse struct {
	Contexts []contextDefinitionResponse `json:"contexts"`
}

type typeDefinitionResponse struct {
	Type                   string                        `json:"type"`
	DisplayName            string                        `json:"display_name"`
	Purpose                string                        `json:"purpose"`
	BestContexts           []string                      `json:"best_contexts"`
	RequiredInputs         []inputRequirementResponse    `json:"required_inputs"`
	RequiredAnyOf          []alternativeInputRequirement `json:"required_any_of,omitempty"`
	OptionalInputs         []inputRequirementResponse    `json:"optional_inputs,omitempty"`
	ProductSignals         []signalDefinitionResponse    `json:"product_signals,omitempty"`
	BehaviorSignals        []signalDefinitionResponse    `json:"behavior_signals,omitempty"`
	ScoringFactors         []scoringFactorResponse       `json:"scoring_factors"`
	OutputFields           []outputFieldResponse         `json:"output_fields"`
	DefaultStrategyID      string                        `json:"default_strategy_id"`
	Fallbacks              []fallbackStepResponse        `json:"fallbacks"`
	ColdStartFallbackNotes []string                      `json:"cold_start_fallback_notes,omitempty"`
}

type contextDefinitionResponse struct {
	Context        string   `json:"context"`
	Description    string   `json:"description"`
	PreferredTypes []string `json:"preferred_types"`
	FallbackTypes  []string `json:"fallback_types"`
}

type inputRequirementResponse struct {
	Field       string `json:"field"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type alternativeInputRequirement struct {
	Fields      []string `json:"fields"`
	Description string   `json:"description"`
}

type signalDefinitionResponse struct {
	Source      string `json:"source"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type scoringFactorResponse struct {
	Name        string `json:"name"`
	Weight      string `json:"weight"`
	Description string `json:"description"`
}

type outputFieldResponse struct {
	Field       string `json:"field"`
	Public      bool   `json:"public"`
	Description string `json:"description"`
}

type fallbackStepResponse struct {
	Priority    int    `json:"priority"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	StrategyID  string `json:"strategy_id"`
	Description string `json:"description"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type storagePlanResponse struct {
	DatabaseName     string                        `json:"database_name"`
	CacheKeyPrefix   string                        `json:"cache_key_prefix"`
	CacheKeyPattern  string                        `json:"cache_key_pattern"`
	OwnershipRule    string                        `json:"ownership_rule"`
	MongoCollections []mongoCollectionPlanResponse `json:"mongo_collections"`
	RedisCaches      []redisCachePlanResponse      `json:"redis_caches"`
	TTLPolicy        cacheTTLPolicyResponse        `json:"ttl_policy"`
	FailureBehavior  []failureBehaviorResponse     `json:"failure_behavior"`
}

type mongoCollectionPlanResponse struct {
	Name        string                   `json:"name"`
	Purpose     string                   `json:"purpose"`
	DetailLevel string                   `json:"detail_level"`
	Indexes     []mongoIndexPlanResponse `json:"indexes,omitempty"`
}

type mongoIndexPlanResponse struct {
	Name               string   `json:"name"`
	Keys               []string `json:"keys"`
	Unique             bool     `json:"unique,omitempty"`
	ExpireAfterSeconds int64    `json:"expire_after_seconds,omitempty"`
	Purpose            string   `json:"purpose"`
}

type redisCachePlanResponse struct {
	Name       string `json:"name"`
	KeyPattern string `json:"key_pattern"`
	ValueType  string `json:"value_type"`
	TTLSeconds int64  `json:"ttl_seconds"`
	Purpose    string `json:"purpose"`
}

type cacheTTLPolicyResponse struct {
	DefaultSeconds      int64 `json:"default_seconds"`
	PersonalizedSeconds int64 `json:"personalized_seconds"`
	GuestSeconds        int64 `json:"guest_seconds"`
	RebuildLockSeconds  int64 `json:"rebuild_lock_seconds"`
}

type failureBehaviorResponse struct {
	Failure  string `json:"failure"`
	Behavior string `json:"behavior"`
}

type storageStatusResponse struct {
	Ready     bool                           `json:"ready"`
	CheckedAt string                         `json:"checked_at"`
	MongoDB   storageComponentStatusResponse `json:"mongodb"`
	Redis     storageComponentStatusResponse `json:"redis"`
}

type storageComponentStatusResponse struct {
	Layer      string  `json:"layer"`
	Configured bool    `json:"configured"`
	State      string  `json:"state"`
	LatencyMS  float64 `json:"latency_ms,omitempty"`
	Error      string  `json:"error,omitempty"`
}

func recommendationRequestFromDomain(req domain.RecommendationRequest) recommendationRequestResponse {
	return recommendationRequestResponse{
		UserID:         req.UserID,
		AnonymousID:    req.AnonymousID,
		SessionID:      req.SessionID,
		ProductID:      req.ProductID,
		CategoryID:     req.CategoryID,
		SellerID:       req.SellerID,
		CartProductIDs: append([]string(nil), req.CartProductIDs...),
		Context:        string(req.Context),
		Limit:          req.Limit,
		Type:           string(req.Type),
	}
}

func typeDefinitionFromDomain(definition domain.TypeDefinition) typeDefinitionResponse {
	return typeDefinitionResponse{
		Type:                   string(definition.Type),
		DisplayName:            definition.DisplayName,
		Purpose:                definition.Purpose,
		BestContexts:           contextsFromDomain(definition.BestContexts),
		RequiredInputs:         inputRequirementsFromDomain(definition.RequiredInputs),
		RequiredAnyOf:          alternativeInputRequirementsFromDomain(definition.RequiredAnyOf),
		OptionalInputs:         inputRequirementsFromDomain(definition.OptionalInputs),
		ProductSignals:         signalsFromDomain(definition.ProductSignals),
		BehaviorSignals:        signalsFromDomain(definition.BehaviorSignals),
		ScoringFactors:         scoringFactorsFromDomain(definition.ScoringFactors),
		OutputFields:           outputFieldsFromDomain(definition.OutputFields),
		DefaultStrategyID:      string(definition.DefaultStrategyID),
		Fallbacks:              fallbackStepsFromDomain(definition.Fallbacks),
		ColdStartFallbackNotes: append([]string(nil), definition.ColdStartFallbackNotes...),
	}
}

func contextDefinitionFromDomain(definition domain.ContextDefinition) contextDefinitionResponse {
	return contextDefinitionResponse{
		Context:        string(definition.Context),
		Description:    definition.Description,
		PreferredTypes: typesFromDomain(definition.PreferredTypes),
		FallbackTypes:  typesFromDomain(definition.FallbackTypes),
	}
}

func contextsFromDomain(contexts []domain.RecommendationContext) []string {
	values := make([]string, 0, len(contexts))
	for _, context := range contexts {
		values = append(values, string(context))
	}
	return values
}

func typesFromDomain(types []domain.RecommendationType) []string {
	values := make([]string, 0, len(types))
	for _, typ := range types {
		values = append(values, string(typ))
	}
	return values
}

func inputRequirementsFromDomain(requirements []domain.InputRequirement) []inputRequirementResponse {
	values := make([]inputRequirementResponse, 0, len(requirements))
	for _, requirement := range requirements {
		values = append(values, inputRequirementResponse{
			Field:       requirement.Field,
			Required:    requirement.Required,
			Description: requirement.Description,
		})
	}
	return values
}

func alternativeInputRequirementsFromDomain(requirements []domain.AlternativeInputRequirement) []alternativeInputRequirement {
	values := make([]alternativeInputRequirement, 0, len(requirements))
	for _, requirement := range requirements {
		values = append(values, alternativeInputRequirement{
			Fields:      append([]string(nil), requirement.Fields...),
			Description: requirement.Description,
		})
	}
	return values
}

func signalsFromDomain(signals []domain.SignalDefinition) []signalDefinitionResponse {
	values := make([]signalDefinitionResponse, 0, len(signals))
	for _, signal := range signals {
		values = append(values, signalDefinitionResponse{
			Source:      string(signal.Source),
			Name:        signal.Name,
			Description: signal.Description,
		})
	}
	return values
}

func scoringFactorsFromDomain(factors []domain.ScoringFactor) []scoringFactorResponse {
	values := make([]scoringFactorResponse, 0, len(factors))
	for _, factor := range factors {
		values = append(values, scoringFactorResponse{
			Name:        factor.Name,
			Weight:      factor.Weight,
			Description: factor.Description,
		})
	}
	return values
}

func outputFieldsFromDomain(fields []domain.OutputField) []outputFieldResponse {
	values := make([]outputFieldResponse, 0, len(fields))
	for _, field := range fields {
		values = append(values, outputFieldResponse{
			Field:       field.Field,
			Public:      field.Public,
			Description: field.Description,
		})
	}
	return values
}

func fallbackStepsFromDomain(steps []domain.FallbackStep) []fallbackStepResponse {
	values := make([]fallbackStepResponse, 0, len(steps))
	for _, step := range steps {
		values = append(values, fallbackStepResponse{
			Priority:    step.Priority,
			Name:        step.Name,
			Type:        string(step.Type),
			StrategyID:  string(step.StrategyID),
			Description: step.Description,
		})
	}
	return values
}

func storagePlanFromDomain(plan domain.StoragePlan) storagePlanResponse {
	return storagePlanResponse{
		DatabaseName:     plan.DatabaseName,
		CacheKeyPrefix:   plan.CacheKeyPrefix,
		CacheKeyPattern:  plan.CacheKeyPattern,
		OwnershipRule:    plan.OwnershipRule,
		MongoCollections: mongoCollectionsFromDomain(plan.MongoCollections),
		RedisCaches:      redisCachesFromDomain(plan.RedisCaches),
		TTLPolicy: cacheTTLPolicyResponse{
			DefaultSeconds:      durationSeconds(plan.TTLPolicy.Default),
			PersonalizedSeconds: durationSeconds(plan.TTLPolicy.Personalized),
			GuestSeconds:        durationSeconds(plan.TTLPolicy.Guest),
			RebuildLockSeconds:  durationSeconds(plan.TTLPolicy.RebuildLock),
		},
		FailureBehavior: failureBehaviorsFromDomain(plan.FailureBehavior),
	}
}

func mongoCollectionsFromDomain(collections []domain.MongoCollectionPlan) []mongoCollectionPlanResponse {
	values := make([]mongoCollectionPlanResponse, 0, len(collections))
	for _, collection := range collections {
		values = append(values, mongoCollectionPlanResponse{
			Name:        collection.Name,
			Purpose:     collection.Purpose,
			DetailLevel: collection.DetailLevel,
			Indexes:     mongoIndexesFromDomain(collection.Indexes),
		})
	}
	return values
}

func mongoIndexesFromDomain(indexes []domain.MongoIndexPlan) []mongoIndexPlanResponse {
	values := make([]mongoIndexPlanResponse, 0, len(indexes))
	for _, index := range indexes {
		values = append(values, mongoIndexPlanResponse{
			Name:               index.Name,
			Keys:               append([]string(nil), index.Keys...),
			Unique:             index.Unique,
			ExpireAfterSeconds: index.ExpireAfterSeconds,
			Purpose:            index.Purpose,
		})
	}
	return values
}

func redisCachesFromDomain(caches []domain.RedisCachePlan) []redisCachePlanResponse {
	values := make([]redisCachePlanResponse, 0, len(caches))
	for _, cache := range caches {
		values = append(values, redisCachePlanResponse{
			Name:       cache.Name,
			KeyPattern: cache.KeyPattern,
			ValueType:  string(cache.ValueType),
			TTLSeconds: durationSeconds(cache.TTL),
			Purpose:    cache.Purpose,
		})
	}
	return values
}

func failureBehaviorsFromDomain(behaviors []domain.FailureBehavior) []failureBehaviorResponse {
	values := make([]failureBehaviorResponse, 0, len(behaviors))
	for _, behavior := range behaviors {
		values = append(values, failureBehaviorResponse{
			Failure:  behavior.Failure,
			Behavior: behavior.Behavior,
		})
	}
	return values
}

func storageStatusFromUsecase(status usecase.StorageStatus) storageStatusResponse {
	return storageStatusResponse{
		Ready:     status.Ready(),
		CheckedAt: status.CheckedAt.Format(time.RFC3339Nano),
		MongoDB:   storageComponentFromUsecase(status.MongoDB),
		Redis:     storageComponentFromUsecase(status.Redis),
	}
}

func storageComponentFromUsecase(status usecase.StorageComponentStatus) storageComponentStatusResponse {
	return storageComponentStatusResponse{
		Layer:      string(status.Layer),
		Configured: status.Configured,
		State:      string(status.State),
		LatencyMS:  float64(status.Latency.Microseconds()) / 1000,
		Error:      status.Error,
	}
}

func durationSeconds(duration time.Duration) int64 {
	return int64(duration.Seconds())
}
