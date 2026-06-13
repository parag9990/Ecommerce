package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

type SellerAnalyticsRepository interface {
	GetSellerSummary(ctx context.Context, query domain.AnalyticsQuery) (domain.SellerAnalyticsSummary, error)
	GetSellerConversion(ctx context.Context, query domain.AnalyticsQuery) (domain.SellerConversionSummary, error)
	ListTopProducts(ctx context.Context, query domain.AnalyticsQuery) ([]domain.TopProductMetric, error)
}

type SellerAnalyticsUsecase struct {
	authorizer              *Authorizer
	repo                    SellerAnalyticsRepository
	logger                  *slog.Logger
	now                     func() time.Time
	defaultCurrency         string
	defaultRangeDays        int
	maxRangeDays            int
	defaultTopProductsLimit int
	maxTopProductsLimit     int
}

type SellerAnalyticsOptions struct {
	DefaultCurrency         string
	DefaultRangeDays        int
	MaxRangeDays            int
	DefaultTopProductsLimit int
	MaxTopProductsLimit     int
}

type GetSellerAnalyticsInput struct {
	Actor            domain.ActorContext
	From             *time.Time
	To               *time.Time
	Currency         string
	TopProductsLimit int
	RequestID        string
}

func NewSellerAnalyticsUsecase(authorizer *Authorizer, repo SellerAnalyticsRepository, logger *slog.Logger, opts SellerAnalyticsOptions) (*SellerAnalyticsUsecase, error) {
	if authorizer == nil {
		return nil, domain.ErrInvalidAuthorizationIn
	}
	if repo == nil {
		return nil, domain.ErrAnalyticsRepositoryRequired
	}
	if logger == nil {
		logger = slog.Default()
	}
	opts = normalizeSellerAnalyticsOptions(opts)
	return &SellerAnalyticsUsecase{
		authorizer:              authorizer,
		repo:                    repo,
		logger:                  logger,
		now:                     func() time.Time { return time.Now().UTC() },
		defaultCurrency:         opts.DefaultCurrency,
		defaultRangeDays:        opts.DefaultRangeDays,
		maxRangeDays:            opts.MaxRangeDays,
		defaultTopProductsLimit: opts.DefaultTopProductsLimit,
		maxTopProductsLimit:     opts.MaxTopProductsLimit,
	}, nil
}

func (uc *SellerAnalyticsUsecase) GetSellerAnalytics(ctx context.Context, input GetSellerAnalyticsInput) (domain.SellerAnalytics, error) {
	input = input.normalized()
	actor := input.Actor.Normalized()
	if !actor.Authenticated() {
		return domain.SellerAnalytics{}, domain.ErrUnauthenticated
	}
	if actor.SellerID == "" {
		return domain.SellerAnalytics{}, domain.ErrSellerContextRequired
	}
	if err := uc.authorizeAnalyticsRead(ctx, actor, input.requestID()); err != nil {
		return domain.SellerAnalytics{}, err
	}

	query, err := uc.analyticsQuery(input)
	if err != nil {
		return domain.SellerAnalytics{}, err
	}

	startedAt := uc.now()
	summary, err := uc.repo.GetSellerSummary(ctx, query)
	if err != nil {
		uc.logAnalyticsFailure(ctx, input, query, startedAt, "summary_query_failed", err)
		return domain.SellerAnalytics{}, fmt.Errorf("%w: %v", domain.ErrAnalyticsUnavailable, err)
	}
	conversion, err := uc.repo.GetSellerConversion(ctx, query)
	if err != nil {
		uc.logAnalyticsFailure(ctx, input, query, startedAt, "conversion_query_failed", err)
		return domain.SellerAnalytics{}, fmt.Errorf("%w: %v", domain.ErrAnalyticsUnavailable, err)
	}
	topProducts, err := uc.repo.ListTopProducts(ctx, query)
	if err != nil {
		uc.logAnalyticsFailure(ctx, input, query, startedAt, "top_products_query_failed", err)
		return domain.SellerAnalytics{}, fmt.Errorf("%w: %v", domain.ErrAnalyticsUnavailable, err)
	}
	topProducts = normalizeTopProducts(topProducts, query.Currency)

	result := domain.SellerAnalytics{
		Revenue: domain.Money{
			Amount:   summary.RevenueAmount,
			Currency: query.Currency,
		},
		Orders:         summary.PaidOrders,
		ConversionRate: CalculateSellerConversionRate(conversion),
		TopProducts:    topProducts,
	}

	uc.logger.InfoContext(ctx, "cms.seller_analytics.requested",
		slog.String("seller_id_hash", hashLogValue(query.SellerID)),
		slog.String("from", domain.FormatAnalyticsDate(query.From)),
		slog.String("to", domain.FormatAnalyticsDate(query.To)),
		slog.String("currency", query.Currency),
		slog.Int("top_products_limit", query.TopProductsLimit),
		slog.String("request_id", input.requestID()),
		slog.Int64("latency_ms", uc.now().Sub(startedAt).Milliseconds()),
	)
	return result, nil
}

func CalculateSellerConversionRate(summary domain.SellerConversionSummary) float64 {
	if summary.ProductViewSessions <= 0 || summary.PaidOrderSessions <= 0 {
		return 0
	}
	rate := float64(summary.PaidOrderSessions) * 100 / float64(summary.ProductViewSessions)
	return math.Round(rate*100) / 100
}

func (uc *SellerAnalyticsUsecase) analyticsQuery(input GetSellerAnalyticsInput) (domain.AnalyticsQuery, error) {
	now := domain.AnalyticsDate(uc.now())
	to := now
	if input.To != nil {
		to = domain.AnalyticsDate(*input.To)
	}
	from := to.AddDate(0, 0, -(uc.defaultRangeDays - 1))
	if input.From != nil {
		from = domain.AnalyticsDate(*input.From)
	}
	currency := input.Currency
	if strings.TrimSpace(currency) == "" {
		currency = uc.defaultCurrency
	}
	limit := input.TopProductsLimit
	if limit <= 0 {
		limit = uc.defaultTopProductsLimit
	}
	if limit > uc.maxTopProductsLimit {
		limit = uc.maxTopProductsLimit
	}

	query := domain.AnalyticsQuery{
		SellerID:         input.Actor.SellerID,
		From:             from,
		To:               to,
		Currency:         currency,
		TopProductsLimit: limit,
	}.Normalized()
	if err := uc.validateQuery(query, now); err != nil {
		return domain.AnalyticsQuery{}, err
	}
	return query, nil
}

func (uc *SellerAnalyticsUsecase) validateQuery(query domain.AnalyticsQuery, today time.Time) error {
	fields := make([]domain.FieldViolation, 0)
	if query.SellerID == "" {
		return domain.ErrSellerContextRequired
	}
	if query.From.IsZero() {
		fields = append(fields, domain.FieldViolation{Field: "from", Reason: "From date is required."})
	}
	if query.To.IsZero() {
		fields = append(fields, domain.FieldViolation{Field: "to", Reason: "To date is required."})
	}
	if !domain.ValidAnalyticsCurrency(query.Currency) {
		fields = append(fields, domain.FieldViolation{Field: "currency", Reason: "Currency must be a 3-letter ISO-style code."})
	}
	if len(fields) > 0 {
		return fmt.Errorf("%w: %v", domain.ErrInvalidAnalyticsDateRange, domain.NewValidationError("Invalid analytics query.", fields))
	}
	if query.From.After(query.To) {
		return fmt.Errorf("%w: from date must be before or equal to to date", domain.ErrInvalidAnalyticsDateRange)
	}
	if query.To.After(today) {
		return fmt.Errorf("%w: to date cannot be in the future", domain.ErrInvalidAnalyticsDateRange)
	}
	if query.InclusiveDays() > uc.maxRangeDays {
		return fmt.Errorf("%w: requested %d days, max %d", domain.ErrAnalyticsRangeTooLarge, query.InclusiveDays(), uc.maxRangeDays)
	}
	return nil
}

func (uc *SellerAnalyticsUsecase) authorizeAnalyticsRead(ctx context.Context, actor domain.ActorContext, requestID string) error {
	_, err := uc.authorizer.Authorize(ctx, AuthorizeInput{
		Actor:              actor,
		ResourceSellerID:   actor.SellerID,
		RequiredPermission: domain.PermissionAnalyticsRead,
		ResourceType:       "analytics",
		ResourceID:         "seller_dashboard_summary",
		RequestID:          requestID,
	})
	return err
}

func (uc *SellerAnalyticsUsecase) logAnalyticsFailure(ctx context.Context, input GetSellerAnalyticsInput, query domain.AnalyticsQuery, startedAt time.Time, reason string, err error) {
	uc.logger.ErrorContext(ctx, "cms.seller_analytics.request_failed",
		slog.String("reason", reason),
		slog.String("seller_id_hash", hashLogValue(query.SellerID)),
		slog.String("from", domain.FormatAnalyticsDate(query.From)),
		slog.String("to", domain.FormatAnalyticsDate(query.To)),
		slog.String("currency", query.Currency),
		slog.String("request_id", input.requestID()),
		slog.Int64("latency_ms", uc.now().Sub(startedAt).Milliseconds()),
		slog.String("error", err.Error()),
	)
}

func normalizeSellerAnalyticsOptions(opts SellerAnalyticsOptions) SellerAnalyticsOptions {
	opts.DefaultCurrency = domain.NormalizeAnalyticsCurrency(opts.DefaultCurrency)
	if opts.DefaultRangeDays <= 0 {
		opts.DefaultRangeDays = domain.AnalyticsDefaultRangeDays
	}
	if opts.MaxRangeDays <= 0 || opts.MaxRangeDays > domain.AnalyticsMaxRangeDays {
		opts.MaxRangeDays = domain.AnalyticsMaxRangeDays
	}
	if opts.DefaultRangeDays > opts.MaxRangeDays {
		opts.DefaultRangeDays = opts.MaxRangeDays
	}
	if opts.DefaultTopProductsLimit <= 0 {
		opts.DefaultTopProductsLimit = domain.AnalyticsDefaultTopProductsLimit
	}
	if opts.MaxTopProductsLimit <= 0 || opts.MaxTopProductsLimit > domain.AnalyticsMaxTopProductsLimit {
		opts.MaxTopProductsLimit = domain.AnalyticsMaxTopProductsLimit
	}
	if opts.DefaultTopProductsLimit > opts.MaxTopProductsLimit {
		opts.DefaultTopProductsLimit = opts.MaxTopProductsLimit
	}
	return opts
}

func normalizeTopProducts(products []domain.TopProductMetric, currency string) []domain.TopProductMetric {
	out := make([]domain.TopProductMetric, 0, len(products))
	for _, product := range products {
		product = product.Normalized(currency)
		if product.ProductID == "" {
			continue
		}
		out = append(out, product)
	}
	return out
}

func hashLogValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:8])
}

func (i GetSellerAnalyticsInput) normalized() GetSellerAnalyticsInput {
	i.Actor = i.Actor.Normalized()
	if i.From != nil {
		from := i.From.UTC()
		i.From = &from
	}
	if i.To != nil {
		to := i.To.UTC()
		i.To = &to
	}
	i.Currency = strings.ToUpper(strings.TrimSpace(i.Currency))
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func (i GetSellerAnalyticsInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}
