package domain

import (
	"testing"
	"time"
)

func TestAnalyticsQueryNormalizationUsesUTCDateAndCurrency(t *testing.T) {
	query := AnalyticsQuery{
		SellerID: " seller_1 ",
		From:     time.Date(2026, 5, 24, 22, 30, 0, 0, time.FixedZone("IST", 5*60*60+30*60)),
		To:       time.Date(2026, 5, 25, 1, 30, 0, 0, time.FixedZone("IST", 5*60*60+30*60)),
		Currency: " inr ",
	}.Normalized()

	if query.SellerID != "seller_1" {
		t.Fatalf("unexpected seller id %q", query.SellerID)
	}
	if got := FormatAnalyticsDate(query.From); got != "2026-05-24" {
		t.Fatalf("expected UTC from date 2026-05-24, got %s", got)
	}
	if got := FormatAnalyticsDate(query.To); got != "2026-05-24" {
		t.Fatalf("expected UTC to date 2026-05-24, got %s", got)
	}
	if query.Currency != "INR" {
		t.Fatalf("expected INR, got %q", query.Currency)
	}
}

func TestAnalyticsQueryInclusiveDays(t *testing.T) {
	query := AnalyticsQuery{
		From: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC),
	}
	if got := query.InclusiveDays(); got != 24 {
		t.Fatalf("expected 24 inclusive days, got %d", got)
	}
}

func TestValidAnalyticsCurrency(t *testing.T) {
	for _, currency := range []string{"INR", "USD", "inr", " INR "} {
		if !ValidAnalyticsCurrency(currency) {
			t.Fatalf("expected %s to be valid", currency)
		}
	}
	for _, currency := range []string{"IN", "12R"} {
		if ValidAnalyticsCurrency(currency) {
			t.Fatalf("expected %s to be invalid", currency)
		}
	}
}
