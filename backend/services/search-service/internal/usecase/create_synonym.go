package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/requestctx"
)

const defaultSynonymAdminTimeout = 500 * time.Millisecond

type CreateSynonymOptions struct {
	Timeout time.Duration
}

type CreateSynonymUsecase struct {
	repo    SynonymRepository
	options CreateSynonymOptions
	logger  *slog.Logger
}

func NewCreateSynonymUsecase(repo SynonymRepository, options CreateSynonymOptions, logger *slog.Logger) (*CreateSynonymUsecase, error) {
	if repo == nil {
		return nil, errors.New("synonym repository is required")
	}
	if options.Timeout <= 0 {
		options.Timeout = defaultSynonymAdminTimeout
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &CreateSynonymUsecase{
		repo:    repo,
		options: options,
		logger:  logger,
	}, nil
}

func (u *CreateSynonymUsecase) Execute(ctx context.Context, input domain.SearchSynonymInput) (domain.SearchSynonym, error) {
	started := time.Now()
	synonym, err := domain.NewSearchSynonym(input)
	if err != nil {
		u.logger.Warn("search.synonym.validation_failed",
			slog.String("request_id", requestID(ctx)),
			slog.String("error", err.Error()),
		)
		return domain.SearchSynonym{}, err
	}

	repoCtx, cancel := context.WithTimeout(ctx, u.options.Timeout)
	defer cancel()
	before, existed, err := u.repo.GetSynonym(repoCtx, synonym.ID)
	if err != nil {
		u.logger.Error("search.synonym.read_before_upsert_failed",
			slog.String("request_id", requestID(ctx)),
			slog.String("synonym_id", synonym.ID),
			slog.String("error", err.Error()),
		)
		return domain.SearchSynonym{}, err
	}

	saved, err := u.repo.UpsertSynonym(repoCtx, synonym)
	if err != nil {
		u.logger.Error("search.synonym.upsert_failed",
			slog.String("request_id", requestID(ctx)),
			slog.String("synonym_id", synonym.ID),
			slog.String("root", synonym.Root),
			slog.Int("synonym_count", len(synonym.Synonyms)),
			slog.String("error", err.Error()),
		)
		return domain.SearchSynonym{}, err
	}

	u.logger.Info("search.synonym.upserted",
		slog.String("request_id", requestID(ctx)),
		slog.String("actor_id", requestctx.Analytics(ctx).UserID),
		slog.String("action", "search.synonym.upsert"),
		slog.String("synonym_id", saved.ID),
		slog.String("reason", strings.TrimSpace(input.Reason)),
		slog.Bool("resource_existed", existed),
		slog.Any("before", before),
		slog.Any("after", saved),
		slog.Int("duration_ms", elapsedMS(started)),
	)
	return saved, nil
}
