package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

const (
	defaultAuditLogRangeDays = 30
	defaultAuditLogPageSize  = 20
	maxAuditLogPageSize      = 100
)

type AuditLogRepository interface {
	ListAuditLogs(ctx context.Context, filter domain.AuditLogFilter) (domain.AuditLogPage, error)
}

type AuditLogUsecase struct {
	authorizer *Authorizer
	repo       AuditLogRepository
	logger     *slog.Logger
	now        func() time.Time
	options    AuditLogOptions
}

type AuditLogOptions struct {
	DefaultRangeDays int
	DefaultPageSize  int
	MaxPageSize      int
}

type ListAuditLogsInput struct {
	Actor        domain.ActorContext
	ActorUserID  string
	Action       domain.Permission
	ResourceType string
	ResourceID   string
	From         *time.Time
	To           *time.Time
	PageSize     int
	Cursor       string
	RequestID    string
}

func NewAuditLogUsecase(authorizer *Authorizer, repo AuditLogRepository, logger *slog.Logger, opts AuditLogOptions) (*AuditLogUsecase, error) {
	if authorizer == nil {
		return nil, domain.ErrInvalidAuthorizationIn
	}
	if repo == nil {
		return nil, domain.ErrAuditLogRepositoryRequired
	}
	if logger == nil {
		logger = slog.Default()
	}
	opts = opts.normalized()
	return &AuditLogUsecase{
		authorizer: authorizer,
		repo:       repo,
		logger:     logger,
		now:        func() time.Time { return time.Now().UTC() },
		options:    opts,
	}, nil
}

func (uc *AuditLogUsecase) ListAuditLogs(ctx context.Context, input ListAuditLogsInput) (domain.AuditLogPage, error) {
	input = input.normalized()
	if !input.Actor.Authenticated() {
		return domain.AuditLogPage{}, domain.ErrUnauthenticated
	}
	sellerID := input.Actor.SellerID
	if sellerID == "" {
		return domain.AuditLogPage{}, domain.ErrSellerContextRequired
	}
	if err := uc.authorizeAuditLogRead(ctx, input.Actor, sellerID, input.requestID()); err != nil {
		return domain.AuditLogPage{}, err
	}

	from, to, err := uc.timeRange(input.From, input.To)
	if err != nil {
		return domain.AuditLogPage{}, err
	}
	pageSize, err := uc.pageSize(input.PageSize)
	if err != nil {
		return domain.AuditLogPage{}, err
	}

	page, err := uc.repo.ListAuditLogs(ctx, domain.AuditLogFilter{
		SellerID:     sellerID,
		ActorUserID:  input.ActorUserID,
		Action:       input.Action,
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		From:         from,
		To:           to,
		PageSize:     pageSize,
		Cursor:       input.Cursor,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAuditLogFilter) {
			return domain.AuditLogPage{}, err
		}
		uc.logger.ErrorContext(ctx, "cms.audit_logs.list_failed",
			slog.String("seller_id", sellerID),
			slog.String("request_id", input.requestID()),
			slog.String("error", err.Error()),
		)
		return domain.AuditLogPage{}, fmt.Errorf("%w: %v", domain.ErrAuditLogsUnavailable, err)
	}

	uc.logger.InfoContext(ctx, "cms.audit_logs.listed",
		slog.String("seller_id", sellerID),
		slog.Int("count", len(page.Entries)),
		slog.Bool("has_next", page.NextCursor != ""),
		slog.String("request_id", input.requestID()),
	)
	return page, nil
}

func (uc *AuditLogUsecase) authorizeAuditLogRead(ctx context.Context, actor domain.ActorContext, sellerID string, requestID string) error {
	_, err := uc.authorizer.Authorize(ctx, AuthorizeInput{
		Actor:              actor,
		ResourceSellerID:   sellerID,
		RequiredPermission: domain.PermissionAuditLogsRead,
		ResourceType:       "audit_logs",
		ResourceID:         sellerID,
		RequestID:          requestID,
	})
	return err
}

func (uc *AuditLogUsecase) timeRange(fromInput *time.Time, toInput *time.Time) (time.Time, time.Time, error) {
	now := uc.now().UTC()
	to := now
	if toInput != nil {
		to = toInput.UTC()
	}
	from := to.AddDate(0, 0, -uc.options.DefaultRangeDays)
	if fromInput != nil {
		from = fromInput.UTC()
	}
	if from.IsZero() || to.IsZero() || !from.Before(to) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: from must be before to", domain.ErrInvalidAuditLogFilter)
	}
	return from, to, nil
}

func (uc *AuditLogUsecase) pageSize(value int) (int, error) {
	if value < 0 {
		return 0, fmt.Errorf("%w: page_size cannot be negative", domain.ErrInvalidAuditLogFilter)
	}
	if value == 0 {
		return uc.options.DefaultPageSize, nil
	}
	if value > uc.options.MaxPageSize {
		return uc.options.MaxPageSize, nil
	}
	return value, nil
}

func (o AuditLogOptions) normalized() AuditLogOptions {
	if o.DefaultRangeDays <= 0 {
		o.DefaultRangeDays = defaultAuditLogRangeDays
	}
	if o.DefaultPageSize <= 0 {
		o.DefaultPageSize = defaultAuditLogPageSize
	}
	if o.MaxPageSize <= 0 {
		o.MaxPageSize = maxAuditLogPageSize
	}
	if o.MaxPageSize > maxAuditLogPageSize {
		o.MaxPageSize = maxAuditLogPageSize
	}
	if o.DefaultPageSize > o.MaxPageSize {
		o.DefaultPageSize = o.MaxPageSize
	}
	return o
}

func (i ListAuditLogsInput) normalized() ListAuditLogsInput {
	i.Actor = i.Actor.Normalized()
	i.ActorUserID = strings.TrimSpace(i.ActorUserID)
	i.Action = domain.Permission(strings.TrimSpace(i.Action.String()))
	i.ResourceType = strings.TrimSpace(i.ResourceType)
	i.ResourceID = strings.TrimSpace(i.ResourceID)
	i.Cursor = strings.TrimSpace(i.Cursor)
	i.RequestID = strings.TrimSpace(i.RequestID)
	if i.From != nil {
		value := i.From.UTC()
		i.From = &value
	}
	if i.To != nil {
		value := i.To.UTC()
		i.To = &value
	}
	return i
}

func (i ListAuditLogsInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}
