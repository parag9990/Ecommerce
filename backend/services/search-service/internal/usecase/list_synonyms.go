package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

type ListSynonymsOptions struct {
	Timeout time.Duration
}

type ListSynonymsUsecase struct {
	repo    SynonymRepository
	options ListSynonymsOptions
	logger  *slog.Logger
}

func NewListSynonymsUsecase(repo SynonymRepository, options ListSynonymsOptions, logger *slog.Logger) (*ListSynonymsUsecase, error) {
	if repo == nil {
		return nil, errors.New("synonym repository is required")
	}
	if options.Timeout <= 0 {
		options.Timeout = defaultSynonymAdminTimeout
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &ListSynonymsUsecase{
		repo:    repo,
		options: options,
		logger:  logger,
	}, nil
}

func (u *ListSynonymsUsecase) Execute(ctx context.Context, req domain.SearchSynonymPageRequest) ([]domain.SearchSynonym, error) {
	started := time.Now()
	page := domain.NormalizeSearchSynonymPageRequest(req)

	repoCtx, cancel := context.WithTimeout(ctx, u.options.Timeout)
	defer cancel()

	synonyms, err := u.repo.ListSynonyms(repoCtx, page)
	if err != nil {
		u.logger.Error("search.synonym.list_failed",
			slog.String("request_id", requestID(ctx)),
			slog.Int("page", page.Page),
			slog.Int("page_size", page.PageSize),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	u.logger.Info("search.synonym.listed",
		slog.String("request_id", requestID(ctx)),
		slog.Int("page", page.Page),
		slog.Int("page_size", page.PageSize),
		slog.Int("synonym_count", len(synonyms)),
		slog.Int("duration_ms", elapsedMS(started)),
	)
	return synonyms, nil
}
