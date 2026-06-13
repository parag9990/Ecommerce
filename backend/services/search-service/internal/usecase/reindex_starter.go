package usecase

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

type ReindexRunner interface {
	Execute(ctx context.Context, req domain.ReindexRequest) (domain.ReindexResult, error)
}

type AsyncReindexStarterOptions struct {
	DefaultMode      string
	DefaultBatchSize int
	MaxBatchSize     int
	JobTimeout       time.Duration
	Clock            func() time.Time
}

type AsyncReindexStarter struct {
	rootCtx context.Context
	runner  ReindexRunner
	options AsyncReindexStarterOptions
	logger  *slog.Logger

	mu     sync.Mutex
	active bool
}

func NewAsyncReindexStarter(rootCtx context.Context, runner ReindexRunner, options AsyncReindexStarterOptions, logger *slog.Logger) (*AsyncReindexStarter, error) {
	if rootCtx == nil {
		rootCtx = context.Background()
	}
	if runner == nil {
		return nil, errors.New("reindex runner is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	if options.DefaultMode == "" {
		options.DefaultMode = domain.ReindexModeAlias
	}
	if options.DefaultBatchSize <= 0 {
		options.DefaultBatchSize = 500
	}
	if options.MaxBatchSize <= 0 {
		options.MaxBatchSize = defaultReindexMaxBatchSize
	}
	if options.JobTimeout <= 0 {
		options.JobTimeout = 2 * time.Hour
	}
	return &AsyncReindexStarter{
		rootCtx: rootCtx,
		runner:  runner,
		options: options,
		logger:  logger,
	}, nil
}

func (s *AsyncReindexStarter) Start(ctx context.Context, req domain.ReindexRequest) (domain.ReindexAccepted, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return domain.ReindexAccepted{}, err
		}
	}
	now := s.now()
	req = domain.NormalizeReindexRequest(req, s.options.DefaultMode, s.options.DefaultBatchSize, now)
	if err := req.Validate(s.options.MaxBatchSize); err != nil {
		return domain.ReindexAccepted{}, err
	}

	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return domain.ReindexAccepted{}, domain.ErrReindexAlreadyRunning
	}
	s.active = true
	s.mu.Unlock()

	go s.run(req)
	return domain.ReindexAccepted{JobID: req.JobID, Status: "accepted"}, nil
}

func (s *AsyncReindexStarter) run(req domain.ReindexRequest) {
	defer func() {
		s.mu.Lock()
		s.active = false
		s.mu.Unlock()
	}()

	ctx, cancel := context.WithTimeout(s.rootCtx, s.options.JobTimeout)
	defer cancel()

	result, err := s.runner.Execute(ctx, req)
	if err != nil {
		s.logger.ErrorContext(ctx, "search.reindex.async_failed",
			slog.String("job_id", req.JobID),
			slog.String("mode", req.Mode),
			slog.String("reason", req.Reason),
			slog.String("actor_id", req.ActorID),
			slog.String("error", err.Error()),
		)
		return
	}
	s.logger.InfoContext(ctx, "search.reindex.async_completed",
		slog.String("job_id", result.JobID),
		slog.String("mode", result.Mode),
		slog.String("target_collection", result.TargetCollection),
		slog.Int("products_indexed", result.ProductsIndexed),
		slog.Bool("alias_swapped", result.AliasSwapped),
	)
}

func (s *AsyncReindexStarter) now() time.Time {
	if s.options.Clock != nil {
		return s.options.Clock()
	}
	return time.Now().UTC()
}
