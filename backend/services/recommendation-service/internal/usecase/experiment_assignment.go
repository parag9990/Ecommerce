package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

type StrategyAssignmentInput struct {
	RequestID          string
	Request            domain.RecommendationRequest
	DefaultStrategyID  domain.StrategyID
	AssignmentSnapshot time.Time
}

type ExperimentAssignmentRepository interface {
	GetExperimentAssignment(ctx context.Context, experimentID string, assignmentKey string, now time.Time) (domain.ExperimentAssignment, error)
	UpsertExperimentAssignment(ctx context.Context, assignment domain.ExperimentAssignment) error
}

type ABTestingMetrics interface {
	RecordABAssignment(experimentID string, variantID string, strategyID string)
	RecordABAssignmentError(reason string)
}

type ExperimentAssignmentConfig struct {
	Enabled             bool
	Experiments         []domain.ExperimentDefinition
	AssignmentTTL       time.Duration
	DefaultSalt         string
	MaxIdentifierLength int
}

type ExperimentAssignmentService struct {
	repo        ExperimentAssignmentRepository
	metrics     ABTestingMetrics
	logger      *slog.Logger
	enabled     bool
	ttl         time.Duration
	maxIDLen    int
	experiments []domain.ExperimentDefinition
}

func NewExperimentAssignmentService(
	repo ExperimentAssignmentRepository,
	cfg ExperimentAssignmentConfig,
	metrics ABTestingMetrics,
	logger *slog.Logger,
) (*ExperimentAssignmentService, error) {
	if cfg.AssignmentTTL <= 0 {
		cfg.AssignmentTTL = domain.DefaultABAssignmentTTL
	}
	cfg.DefaultSalt = strings.TrimSpace(cfg.DefaultSalt)
	if cfg.DefaultSalt == "" {
		cfg.DefaultSalt = domain.DefaultABTestSalt
	}
	if cfg.MaxIdentifierLength <= 0 {
		cfg.MaxIdentifierLength = 128
	}
	if logger == nil {
		logger = slog.Default()
	}

	experiments := make([]domain.ExperimentDefinition, 0, len(cfg.Experiments))
	seen := make(map[string]struct{}, len(cfg.Experiments))
	for _, experiment := range cfg.Experiments {
		experiment = experiment.Normalize(cfg.DefaultSalt)
		if err := experiment.Validate(); err != nil {
			return nil, err
		}
		if _, ok := seen[experiment.ExperimentID]; ok {
			return nil, fmt.Errorf("%w: duplicate experiment_id %q", domain.ErrInvalidExperiment, experiment.ExperimentID)
		}
		seen[experiment.ExperimentID] = struct{}{}
		experiments = append(experiments, experiment)
	}
	sort.SliceStable(experiments, func(i int, j int) bool {
		return experiments[i].ExperimentID < experiments[j].ExperimentID
	})

	return &ExperimentAssignmentService{
		repo:        repo,
		metrics:     metrics,
		logger:      logger,
		enabled:     cfg.Enabled,
		ttl:         cfg.AssignmentTTL,
		maxIDLen:    cfg.MaxIdentifierLength,
		experiments: experiments,
	}, nil
}

func (s *ExperimentAssignmentService) Assign(ctx context.Context, input StrategyAssignmentInput) (domain.StrategyAssignment, error) {
	if err := ctx.Err(); err != nil {
		return domain.StrategyAssignment{}, err
	}
	defaultAssignment := domain.StrategyAssignment{StrategyID: input.DefaultStrategyID}
	if s == nil || !s.enabled || len(s.experiments) == 0 {
		return defaultAssignment, nil
	}
	req := input.Request.Normalize()
	if input.DefaultStrategyID == "" {
		defaultStrategy, err := domain.DefaultStrategyForType(req.Type)
		if err != nil {
			return domain.StrategyAssignment{}, err
		}
		defaultAssignment.StrategyID = defaultStrategy
		input.DefaultStrategyID = defaultStrategy
	}
	if err := domain.ValidateStrategyID(input.DefaultStrategyID); err != nil {
		return domain.StrategyAssignment{}, err
	}
	now := input.AssignmentSnapshot
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	experiment, ok := s.matchingExperiment(req, now)
	if !ok {
		return defaultAssignment, nil
	}
	identity, ok := domain.NewAssignmentIdentity(req.UserID, req.AnonymousID, req.SessionID)
	if !ok {
		s.logger.DebugContext(ctx, "recommendation.ab.assignment_skipped",
			slog.String("request_id", input.RequestID),
			slog.String("experiment_id", experiment.ExperimentID),
			slog.String("reason", "identity_missing"),
		)
		return defaultAssignment, nil
	}
	if err := s.validateIdentity(identity); err != nil {
		s.recordError("identity_invalid")
		return domain.StrategyAssignment{}, err
	}
	if s.repo == nil {
		s.recordError("repository_unavailable")
		return domain.StrategyAssignment{}, fmt.Errorf("%w: assignment repository is not configured", domain.ErrRecommendationStorage)
	}

	existing, err := s.repo.GetExperimentAssignment(ctx, experiment.ExperimentID, identity.AssignmentKey, now)
	switch {
	case err == nil:
		return s.assignmentFromStored(existing), nil
	case errors.Is(err, domain.ErrExperimentAssignmentNotFound):
	default:
		s.recordError("repository_read")
		return domain.StrategyAssignment{}, err
	}

	bucket := domain.BucketPercent(experiment.ExperimentID, identity.AssignmentKey, experiment.Salt)
	variant, err := domain.ChooseExperimentVariant(bucket, experiment.Variants)
	if err != nil {
		s.recordError("variant_selection")
		return domain.StrategyAssignment{}, err
	}
	assignment := domain.NewExperimentAssignment(experiment, identity, variant, now, s.ttl)
	if err := s.repo.UpsertExperimentAssignment(ctx, assignment); err != nil {
		s.recordError("repository_write")
		return domain.StrategyAssignment{}, err
	}

	result := domain.StrategyAssignment{
		ExperimentID: experiment.ExperimentID,
		VariantID:    variant.VariantID,
		StrategyID:   variant.StrategyID,
		Assigned:     true,
	}
	s.recordAssignment(result)
	s.logger.InfoContext(ctx, "recommendation.ab.assigned",
		slog.String("request_id", input.RequestID),
		slog.String("experiment_id", result.ExperimentID),
		slog.String("variant_id", result.VariantID),
		slog.String("strategy_id", string(result.StrategyID)),
		slog.String("context", string(req.Context)),
	)
	return result, nil
}

func (s *ExperimentAssignmentService) matchingExperiment(req domain.RecommendationRequest, now time.Time) (domain.ExperimentDefinition, bool) {
	for _, experiment := range s.experiments {
		if experiment.ActiveAt(now) && experiment.Matches(req) {
			return experiment, true
		}
	}
	return domain.ExperimentDefinition{}, false
}

func (s *ExperimentAssignmentService) assignmentFromStored(assignment domain.ExperimentAssignment) domain.StrategyAssignment {
	result := domain.StrategyAssignment{
		ExperimentID: assignment.ExperimentID,
		VariantID:    assignment.VariantID,
		StrategyID:   assignment.StrategyID,
		Assigned:     true,
	}
	s.recordAssignment(result)
	return result
}

func (s *ExperimentAssignmentService) validateIdentity(identity domain.AssignmentIdentity) error {
	for field, value := range map[string]string{
		"assignment_key": identity.AssignmentKey,
		"user_id":        identity.UserID,
		"anonymous_id":   identity.AnonymousID,
		"session_id":     identity.SessionID,
	} {
		if len(value) > s.maxIDLen {
			return fmt.Errorf("%w: %s is too long", domain.ErrInvalidRecommendationRequest, field)
		}
		if strings.ContainsAny(value, "\x00\r\n\t") {
			return fmt.Errorf("%w: %s contains unsupported control characters", domain.ErrInvalidRecommendationRequest, field)
		}
	}
	return nil
}

func (s *ExperimentAssignmentService) recordAssignment(assignment domain.StrategyAssignment) {
	if s.metrics != nil {
		s.metrics.RecordABAssignment(assignment.ExperimentID, assignment.VariantID, string(assignment.StrategyID))
	}
}

func (s *ExperimentAssignmentService) recordError(reason string) {
	if s.metrics != nil {
		s.metrics.RecordABAssignmentError(reason)
	}
}
