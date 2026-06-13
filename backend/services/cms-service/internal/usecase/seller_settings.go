package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

type SellerSettingsRepository interface {
	GetSellerSettings(ctx context.Context, sellerID string) (domain.SellerSettings, error)
}

type SellerSettingsUsecase struct {
	authorizer *Authorizer
	repo       SellerSettingsRepository
	logger     *slog.Logger
	now        func() time.Time
}

type GetSellerSettingsInput struct {
	Actor          domain.ActorContext
	SellerID       string
	InternalCaller string
	RequestID      string
}

func NewSellerSettingsUsecase(authorizer *Authorizer, repo SellerSettingsRepository, logger *slog.Logger) (*SellerSettingsUsecase, error) {
	if authorizer == nil {
		return nil, domain.ErrInvalidAuthorizationIn
	}
	if repo == nil {
		return nil, domain.ErrSellerSettingsRepositoryRequired
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &SellerSettingsUsecase{
		authorizer: authorizer,
		repo:       repo,
		logger:     logger,
		now:        func() time.Time { return time.Now().UTC() },
	}, nil
}

func (uc *SellerSettingsUsecase) GetSellerSettings(ctx context.Context, input GetSellerSettingsInput) (domain.SellerSettings, error) {
	input = input.normalized()
	sellerID := input.SellerID
	if sellerID == "" {
		sellerID = input.Actor.SellerID
	}
	if sellerID == "" {
		return domain.SellerSettings{}, domain.ErrSellerContextRequired
	}

	if input.InternalCaller == "" {
		if err := uc.authorizeSettingsRead(ctx, input.Actor, sellerID, input.requestID()); err != nil {
			return domain.SellerSettings{}, err
		}
	}

	settings, err := uc.repo.GetSellerSettings(ctx, sellerID)
	if err != nil {
		if errors.Is(err, domain.ErrSellerSettingsNotFound) {
			settings = domain.DefaultSellerSettings(sellerID)
		} else {
			return domain.SellerSettings{}, err
		}
	}
	settings = settings.Normalized()
	if settings.UpdatedAt.IsZero() {
		settings.UpdatedAt = uc.now()
	}

	uc.logger.InfoContext(ctx, "cms.seller_settings.read",
		slog.String("seller_id", sellerID),
		slog.String("internal_caller", input.InternalCaller),
		slog.String("request_id", input.requestID()),
	)
	return settings, nil
}

func (uc *SellerSettingsUsecase) authorizeSettingsRead(ctx context.Context, actor domain.ActorContext, sellerID string, requestID string) error {
	_, err := uc.authorizer.Authorize(ctx, AuthorizeInput{
		Actor:              actor,
		ResourceSellerID:   sellerID,
		RequiredPermission: domain.PermissionSettingsRead,
		ResourceType:       "settings",
		ResourceID:         sellerID,
		RequestID:          requestID,
	})
	return err
}

func (i GetSellerSettingsInput) normalized() GetSellerSettingsInput {
	i.Actor = i.Actor.Normalized()
	i.SellerID = strings.TrimSpace(i.SellerID)
	i.InternalCaller = strings.TrimSpace(i.InternalCaller)
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func (i GetSellerSettingsInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}
