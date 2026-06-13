package domain

import "errors"

var (
	ErrInvalidRecommendationRequest  = errors.New("invalid recommendation request")
	ErrUnsupportedRecommendationType = errors.New("unsupported recommendation type")
	ErrUnsupportedContext            = errors.New("unsupported recommendation context")
	ErrUnsupportedStrategy           = errors.New("unsupported recommendation strategy")
	ErrDefinitionNotFound            = errors.New("recommendation definition not found")
	ErrInvalidLimit                  = errors.New("invalid recommendation limit")
	ErrInvalidCacheKey               = errors.New("invalid recommendation cache key")
	ErrRecommendationCacheMiss       = errors.New("recommendation cache miss")
	ErrRecommendationCacheCorrupt    = errors.New("recommendation cache payload is corrupt")
	ErrRecommendationSetNotFound     = errors.New("recommendation set not found")
	ErrInvalidRecommendationSet      = errors.New("invalid recommendation set")
	ErrRecommendationStorage         = errors.New("recommendation storage error")
	ErrInvalidExperiment             = errors.New("invalid recommendation experiment")
	ErrExperimentAssignmentNotFound  = errors.New("recommendation experiment assignment not found")
	ErrFeatureProfileNotFound        = errors.New("recommendation feature profile not found")
	ErrInvalidEventEnvelope          = errors.New("invalid event envelope")
	ErrUnsupportedEventType          = errors.New("unsupported recommendation event type")
	ErrInvalidInteraction            = errors.New("invalid user interaction")
	ErrInvalidFeature                = errors.New("invalid recommendation feature")
)
