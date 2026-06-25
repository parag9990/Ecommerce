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

type DeleteSynonymOptions struct {
	Timeout time.Duration
}

type DeleteSynonymUsecase struct {
	repo    SynonymRepository
	options DeleteSynonymOptions
	logger  *slog.Logger
}

func NewDeleteSynonymUsecase(repo SynonymRepository, options DeleteSynonymOptions, logger *slog.Logger) (*DeleteSynonymUsecase, error) {
	if repo == nil {
		return nil, errors.New("synonym repository is required")
	}
	if options.Timeout <= 0 {
		options.Timeout = defaultSynonymAdminTimeout
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &DeleteSynonymUsecase{repo: repo, options: options, logger: logger}, nil
}

func (u *DeleteSynonymUsecase) Execute(ctx context.Context, synonymID string, reason string) (domain.SearchSynonym, error) {
	started := time.Now()
	synonymID = strings.TrimSpace(synonymID)
	if synonymID == "" {
		return domain.SearchSynonym{}, domain.ErrInvalidSynonym
	}
	if len([]rune(strings.TrimSpace(reason))) > domain.MaxSynonymReasonLength {
		return domain.SearchSynonym{}, domain.ErrInvalidSynonym
	}

	repoCtx, cancel := context.WithTimeout(ctx, u.options.Timeout)
	defer cancel()
	before, existed, err := u.repo.GetSynonym(repoCtx, synonymID)
	if err != nil {
		u.logger.Error("search.synonym.read_before_delete_failed",
			slog.String("request_id", requestID(ctx)),
			slog.String("synonym_id", synonymID),
			slog.String("error", err.Error()),
		)
		return domain.SearchSynonym{}, err
	}
	if !existed {
		return domain.SearchSynonym{}, domain.ErrSearchSynonymNotFound
	}

	deleted, ok, err := u.repo.DeleteSynonym(repoCtx, synonymID)
	if err != nil {
		u.logger.Error("search.synonym.delete_failed",
			slog.String("request_id", requestID(ctx)),
			slog.String("synonym_id", synonymID),
			slog.String("error", err.Error()),
		)
		return domain.SearchSynonym{}, err
	}
	if !ok {
		return domain.SearchSynonym{}, domain.ErrSearchSynonymNotFound
	}
	if strings.TrimSpace(deleted.Root) == "" {
		deleted = before
	}

	u.logger.Info("search.synonym.deleted",
		slog.String("request_id", requestID(ctx)),
		slog.String("actor_id", requestctx.Analytics(ctx).UserID),
		slog.String("action", "search.synonym.delete"),
		slog.String("synonym_id", synonymID),
		slog.String("reason", strings.TrimSpace(reason)),
		slog.Any("before", before),
		slog.Int("duration_ms", elapsedMS(started)),
	)
	return deleted, nil
}
