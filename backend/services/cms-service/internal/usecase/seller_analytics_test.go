package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

func TestGetSellerAnalyticsReturnsAggregatesWithDefaults(t *testing.T) {
	repo := &memorySellerAnalyticsRepository{
		summary: domain.SellerAnalyticsSummary{
			RevenueAmount: 1250000,
			PaidOrders:    84,
			PaidItems:     156,
		},
		conversion: domain.SellerConversionSummary{
			ProductViewSessions: 2000,
			PaidOrderSessions:   68,
		},
		products: []domain.TopProductMetric{
			{
				ProductID: "prod_101",
				Title:     "Cotton T-Shirt",
				UnitsSold: 156,
				Orders:    72,
				Revenue:   domain.Money{Amount: 312000, Currency: "INR"},
			},
		},
	}
	uc := newTestSellerAnalyticsUsecase(t, repo)

	result, err := uc.GetSellerAnalytics(context.Background(), GetSellerAnalyticsInput{
		Actor: activeActor(domain.RoleSellerCatalogEditor),
	})
	if err != nil {
		t.Fatalf("GetSellerAnalytics: %v", err)
	}
	if result.Revenue.Amount != 1250000 || result.Revenue.Currency != "INR" {
		t.Fatalf("unexpected revenue: %+v", result.Revenue)
	}
	if result.Orders != 84 {
		t.Fatalf("expected 84 orders, got %d", result.Orders)
	}
	if result.ConversionRate != 3.4 {
		t.Fatalf("expected conversion 3.4, got %f", result.ConversionRate)
	}
	if len(result.TopProducts) != 1 || result.TopProducts[0].ProductID != "prod_101" {
		t.Fatalf("unexpected top products: %+v", result.TopProducts)
	}
	if got := domain.FormatAnalyticsDate(repo.summaryQuery.From); got != "2026-04-25" {
		t.Fatalf("expected default from 2026-04-25, got %s", got)
	}
	if got := domain.FormatAnalyticsDate(repo.summaryQuery.To); got != "2026-05-24" {
		t.Fatalf("expected default to 2026-05-24, got %s", got)
	}
	if repo.summaryQuery.TopProductsLimit != 5 {
		t.Fatalf("expected default top products limit 5, got %d", repo.summaryQuery.TopProductsLimit)
	}
}

func TestGetSellerAnalyticsRejectsMissingSellerContext(t *testing.T) {
	uc := newTestSellerAnalyticsUsecase(t, &memorySellerAnalyticsRepository{})
	actor := activeActor(domain.RoleSeller)
	actor.SellerID = ""

	_, err := uc.GetSellerAnalytics(context.Background(), GetSellerAnalyticsInput{Actor: actor})
	if !errors.Is(err, domain.ErrSellerContextRequired) {
		t.Fatalf("expected seller context error, got %v", err)
	}
}

func TestGetSellerAnalyticsRejectsOversizedDateRange(t *testing.T) {
	uc := newTestSellerAnalyticsUsecase(t, &memorySellerAnalyticsRepository{})
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC)

	_, err := uc.GetSellerAnalytics(context.Background(), GetSellerAnalyticsInput{
		Actor: activeActor(domain.RoleSeller),
		From:  &from,
		To:    &to,
	})
	if !errors.Is(err, domain.ErrAnalyticsRangeTooLarge) {
		t.Fatalf("expected range too large, got %v", err)
	}
}

func TestGetSellerAnalyticsClampsTopProductsLimit(t *testing.T) {
	repo := &memorySellerAnalyticsRepository{}
	uc := newTestSellerAnalyticsUsecase(t, repo)
	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC)

	_, err := uc.GetSellerAnalytics(context.Background(), GetSellerAnalyticsInput{
		Actor:            activeActor(domain.RoleSeller),
		From:             &from,
		To:               &to,
		TopProductsLimit: 500,
	})
	if err != nil {
		t.Fatalf("GetSellerAnalytics: %v", err)
	}
	if repo.summaryQuery.TopProductsLimit != 20 {
		t.Fatalf("expected clamp to 20, got %d", repo.summaryQuery.TopProductsLimit)
	}
}

func TestCalculateSellerConversionRateHandlesZeroAndRounds(t *testing.T) {
	if got := CalculateSellerConversionRate(domain.SellerConversionSummary{}); got != 0 {
		t.Fatalf("expected zero conversion, got %f", got)
	}
	got := CalculateSellerConversionRate(domain.SellerConversionSummary{
		ProductViewSessions: 300,
		PaidOrderSessions:   10,
	})
	if got != 3.33 {
		t.Fatalf("expected rounded conversion 3.33, got %f", got)
	}
}

func newTestSellerAnalyticsUsecase(t *testing.T, repo SellerAnalyticsRepository) *SellerAnalyticsUsecase {
	t.Helper()
	uc, err := NewSellerAnalyticsUsecase(
		newTestAuthorizer(t, nil, nil),
		repo,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		SellerAnalyticsOptions{},
	)
	if err != nil {
		t.Fatalf("NewSellerAnalyticsUsecase: %v", err)
	}
	uc.now = func() time.Time { return time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC) }
	return uc
}

type memorySellerAnalyticsRepository struct {
	summary      domain.SellerAnalyticsSummary
	conversion   domain.SellerConversionSummary
	products     []domain.TopProductMetric
	summaryQuery domain.AnalyticsQuery
}

func (r *memorySellerAnalyticsRepository) GetSellerSummary(_ context.Context, query domain.AnalyticsQuery) (domain.SellerAnalyticsSummary, error) {
	r.summaryQuery = query
	return r.summary, nil
}

func (r *memorySellerAnalyticsRepository) GetSellerConversion(_ context.Context, _ domain.AnalyticsQuery) (domain.SellerConversionSummary, error) {
	return r.conversion, nil
}

func (r *memorySellerAnalyticsRepository) ListTopProducts(_ context.Context, _ domain.AnalyticsQuery) ([]domain.TopProductMetric, error) {
	return append([]domain.TopProductMetric(nil), r.products...), nil
}
