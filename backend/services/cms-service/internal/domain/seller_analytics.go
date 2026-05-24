package domain

import (
	"strings"
	"time"
)

const (
	AnalyticsDefaultCurrency         = "INR"
	AnalyticsDefaultRangeDays        = 30
	AnalyticsMaxRangeDays            = 366
	AnalyticsDefaultTopProductsLimit = 5
	AnalyticsMaxTopProductsLimit     = 20
	analyticsDateLayout              = "2006-01-02"
)

type AnalyticsQuery struct {
	SellerID         string
	From             time.Time
	To               time.Time
	Currency         string
	TopProductsLimit int
}

type SellerAnalytics struct {
	Revenue        Money
	Orders         int64
	ConversionRate float64
	TopProducts    []TopProductMetric
}

type TopProductMetric struct {
	ProductID string
	Title     string
	UnitsSold int64
	Orders    int64
	Revenue   Money
}

type SellerAnalyticsSummary struct {
	Currency      string
	RevenueAmount int64
	PaidOrders    int64
	PaidItems     int64
}

type SellerConversionSummary struct {
	ProductViewSessions     int64
	AddToCartSessions       int64
	CheckoutStartedSessions int64
	PaidOrderSessions       int64
}

func (q AnalyticsQuery) Normalized() AnalyticsQuery {
	q.SellerID = strings.TrimSpace(q.SellerID)
	q.From = AnalyticsDate(q.From)
	q.To = AnalyticsDate(q.To)
	q.Currency = NormalizeAnalyticsCurrency(q.Currency)
	return q
}

func (q AnalyticsQuery) InclusiveDays() int {
	q = q.Normalized()
	if q.From.IsZero() || q.To.IsZero() || q.From.After(q.To) {
		return 0
	}
	return int(q.To.Sub(q.From).Hours()/24) + 1
}

func AnalyticsDate(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func FormatAnalyticsDate(value time.Time) string {
	return AnalyticsDate(value).Format(analyticsDateLayout)
}

func NormalizeAnalyticsCurrency(currency string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return AnalyticsDefaultCurrency
	}
	return currency
}

func ValidAnalyticsCurrency(currency string) bool {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if len(currency) != 3 {
		return false
	}
	for _, ch := range currency {
		if ch < 'A' || ch > 'Z' {
			return false
		}
	}
	return true
}

func (m TopProductMetric) Normalized(defaultCurrency string) TopProductMetric {
	m.ProductID = strings.TrimSpace(m.ProductID)
	m.Title = strings.TrimSpace(m.Title)
	m.Revenue.Currency = NormalizeAnalyticsCurrency(firstAnalyticsCurrency(m.Revenue.Currency, defaultCurrency))
	return m
}

func firstAnalyticsCurrency(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return AnalyticsDefaultCurrency
}
