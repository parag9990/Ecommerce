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

type UpdateSynonymOptions struct {
	Timeout time.Duration
}

type UpdateSynonymUsecase struct {
	repo    SynonymRepository
	options UpdateSynonymOptions
	logger  *slog.Logger
}

func NewUpdateSynonymUsecase(repo SynonymRepository, options UpdateSynonymOptions, logger *slog.Logger) (*UpdateSynonymUsecase, error) {
	if repo == nil {
		return nil, errors.New("synonym repository is required")
	}
	if options.Timeout <= 0 {
		options.Timeout = defaultSynonymAdminTimeout
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &UpdateSynonymUsecase{repo: repo, options: options, logger: logger}, nil
}

func (u *UpdateSynonymUsecase) Execute(ctx context.Context, synonymID string, input domain.SearchSynonymInput) (domain.SearchSynonym, error) {
	started := time.Now()
	synonymID = strings.TrimSpace(synonymID)
	if synonymID == "" {
		return domain.SearchSynonym{}, domain.ErrInvalidSynonym
	}

	synonym, err := domain.NewSearchSynonym(input)
	if err != nil {
		u.logger.Warn("search.synonym.validation_failed",
			slog.String("request_id", requestID(ctx)),
			slog.String("synonym_id", synonymID),
			slog.String("error", err.Error()),
		)
		return domain.SearchSynonym{}, err
	}

	repoCtx, cancel := context.WithTimeout(ctx, u.options.Timeout)
	defer cancel()
	before, existed, err := u.repo.GetSynonym(repoCtx, synonymID)
	if err != nil {
		u.logger.Error("search.synonym.read_before_update_failed",
			slog.String("request_id", requestID(ctx)),
			slog.String("synonym_id", synonymID),
			slog.String("error", err.Error()),
		)
		return domain.SearchSynonym{}, err
	}
	if !existed {
		return domain.SearchSynonym{}, domain.ErrSearchSynonymNotFound
	}

	saved, err := u.repo.UpsertSynonym(repoCtx, synonym)
	if err != nil {
		u.logger.Error("search.synonym.update_failed",
			slog.String("request_id", requestID(ctx)),
			slog.String("synonym_id", synonymID),
			slog.String("new_synonym_id", synonym.ID),
			slog.String("error", err.Error()),
		)
		return domain.SearchSynonym{}, err
	}
	if saved.ID != synonymID {
		if _, deleted, err := u.repo.DeleteSynonym(repoCtx, synonymID); err != nil {
			u.logger.Error("search.synonym.delete_previous_after_rename_failed",
				slog.String("request_id", requestID(ctx)),
				slog.String("synonym_id", synonymID),
				slog.String("new_synonym_id", saved.ID),
				slog.String("error", err.Error()),
			)
			return domain.SearchSynonym{}, err
		} else if !deleted {
			u.logger.Warn("search.synonym.previous_missing_after_rename",
				slog.String("request_id", requestID(ctx)),
				slog.String("synonym_id", synonymID),
				slog.String("new_synonym_id", saved.ID),
			)
		}
	}

	u.logger.Info("search.synonym.updated",
		slog.String("request_id", requestID(ctx)),
		slog.String("actor_id", requestctx.Analytics(ctx).UserID),
		slog.String("action", "search.synonym.update"),
		slog.String("synonym_id", saved.ID),
		slog.String("previous_synonym_id", synonymID),
		slog.String("reason", strings.TrimSpace(input.Reason)),
		slog.Any("before", before),
		slog.Any("after", saved),
		slog.Int("duration_ms", elapsedMS(started)),
	)
	return saved, nil
}
