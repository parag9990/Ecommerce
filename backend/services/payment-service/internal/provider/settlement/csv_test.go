package settlement

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestParseCSVNormalizesSettlementReport(t *testing.T) {
	input := strings.NewReader(
		"settlement_id,provider_payment_id,outcome,currency,settled_amount,fee_amount,settled_at\n" +
			"stl_20260526,psp_pay_701,settled,inr,249900,7250,2026-05-26T16:30:00Z\n",
	)
	reportDate := time.Date(2026, time.May, 26, 0, 0, 0, 0, time.UTC)

	report, err := ParseCSV(context.Background(), input, "Stripe_Like", reportDate)
	if err != nil {
		t.Fatalf("ParseCSV() error = %v", err)
	}
	if report.Provider != "stripe_like" || report.SettlementID != "stl_20260526" || len(report.Rows) != 1 {
		t.Fatalf("report = %+v, want one normalized stripe report row", report)
	}
	if report.Rows[0].Outcome != OutcomeCaptured || report.Rows[0].Currency != "INR" {
		t.Fatalf("row = %+v, want normalized captured INR outcome", report.Rows[0])
	}
	if !report.WindowEnd.Equal(report.WindowStart.AddDate(0, 0, 1)) {
		t.Fatalf("window = %s to %s, want one day", report.WindowStart, report.WindowEnd)
	}
}

func TestParseCSVRejectsDuplicateProviderPaymentID(t *testing.T) {
	input := strings.NewReader(
		"settlement_id,provider_payment_id,outcome,currency,settled_amount,fee_amount,settled_at\n" +
			"stl_one,psp_duplicate,captured,INR,100,0,2026-05-26T01:00:00Z\n" +
			"stl_one,psp_duplicate,captured,INR,100,0,2026-05-26T02:00:00Z\n",
	)

	_, err := ParseCSV(context.Background(), input, "stripe_like", time.Date(2026, time.May, 26, 0, 0, 0, 0, time.UTC))
	if err == nil || !strings.Contains(err.Error(), "duplicate provider_payment_id") {
		t.Fatalf("ParseCSV() error = %v, want duplicate provider payment failure", err)
	}
}

func TestParseCSVRejectsEmptyReport(t *testing.T) {
	input := strings.NewReader("settlement_id,provider_payment_id,outcome,currency,settled_amount,fee_amount,settled_at\n")

	if _, err := ParseCSV(context.Background(), input, "stripe_like", time.Now().UTC()); err == nil {
		t.Fatal("ParseCSV() error = nil, want empty report rejected")
	}
}

func TestReportValidationRejectsChangedRowMetadata(t *testing.T) {
	reportDate := time.Date(2026, time.May, 26, 0, 0, 0, 0, time.UTC)
	report, err := NewReport("stripe_like", reportDate, []Row{{
		SettlementID:      "stl_one",
		ProviderPaymentID: "psp_one",
		Outcome:           OutcomeCaptured,
		Currency:          "INR",
		SettledAmount:     100,
		SettledAt:         reportDate.Add(time.Hour),
	}})
	if err != nil {
		t.Fatalf("NewReport() error = %v", err)
	}
	report.Rows[0].ReportDate = "2026-05-25"
	if err := report.Validate("stripe_like", reportDate); err == nil {
		t.Fatal("Validate() error = nil, want changed row metadata rejected")
	}
}
