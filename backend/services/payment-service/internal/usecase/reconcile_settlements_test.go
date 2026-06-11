package usecase

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sort"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider/settlement"
)

func TestReconcileSettlementsClassifiesRowsAndDoesNotAlertAgainOnRerun(t *testing.T) {
	reportDate := time.Date(2026, time.May, 26, 0, 0, 0, 0, time.UTC)
	source := &fakeSettlementSource{report: reconciliationReport(t, reportDate, 899)}
	repo := &fakeReconciliationRepository{
		payments: map[string]domain.Payment{
			"stripe_like/psp_match":    capturedPayment("pay_match", "psp_match", 1000),
			"stripe_like/psp_mismatch": capturedPayment("pay_mismatch", "psp_mismatch", 900),
		},
		candidates: []domain.Payment{
			capturedPayment("pay_match", "psp_match", 1000),
			capturedPayment("pay_mismatch", "psp_mismatch", 900),
			capturedPayment("pay_missing_provider", "psp_missing_provider", 700),
		},
		stored: make(map[string]domain.PaymentReconciliation),
	}
	alerts := &fakeReconciliationAlerts{}
	uc := newReconciliationUsecaseForTest(t, repo, source, alerts)

	summary, err := uc.Execute(context.Background(), "stripe_like", reportDate)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if summary.MatchedCount != 1 || summary.MismatchCount != 1 || summary.MissingLocalCount != 1 || summary.MissingProviderCount != 1 {
		t.Fatalf("summary = %+v, want one result of every classification", summary)
	}
	if summary.CreatedCount != 4 || summary.AlertedCount != 3 || len(alerts.alerts) != 3 {
		t.Fatalf("summary/alerts = %+v/%d, want four inserts and three alerts", summary, len(alerts.alerts))
	}
	if alerts.alerts[0].Severity != "critical" {
		t.Fatalf("first mismatch alert severity = %q, want critical", alerts.alerts[0].Severity)
	}

	replay, err := uc.Execute(context.Background(), "stripe_like", reportDate)
	if err != nil {
		t.Fatalf("replayed Execute() error = %v", err)
	}
	if replay.CreatedCount != 0 || replay.AlertedCount != 0 || len(alerts.alerts) != 3 {
		t.Fatalf("replay summary/alerts = %+v/%d, want no duplicate output", replay, len(alerts.alerts))
	}
}

func TestReconcileSettlementsRejectsChangedEvidenceForSameSettlementLine(t *testing.T) {
	reportDate := time.Date(2026, time.May, 26, 0, 0, 0, 0, time.UTC)
	source := &fakeSettlementSource{report: reconciliationReport(t, reportDate, 899)}
	repo := &fakeReconciliationRepository{
		payments: map[string]domain.Payment{
			"stripe_like/psp_match":    capturedPayment("pay_match", "psp_match", 1000),
			"stripe_like/psp_mismatch": capturedPayment("pay_mismatch", "psp_mismatch", 900),
		},
		stored: make(map[string]domain.PaymentReconciliation),
	}
	uc := newReconciliationUsecaseForTest(t, repo, source, &fakeReconciliationAlerts{})
	if _, err := uc.Execute(context.Background(), "stripe_like", reportDate); err != nil {
		t.Fatalf("initial Execute() error = %v", err)
	}

	source.report = reconciliationReport(t, reportDate, 898)
	_, err := uc.Execute(context.Background(), "stripe_like", reportDate)
	if !errors.Is(err, domain.ErrReconciliationConflict) {
		t.Fatalf("changed-report Execute() error = %v, want reconciliation conflict", err)
	}
}

func TestReconcileSettlementsRetainsResultsWhenAlertPublishingFails(t *testing.T) {
	reportDate := time.Date(2026, time.May, 26, 0, 0, 0, 0, time.UTC)
	source := &fakeSettlementSource{report: reconciliationReport(t, reportDate, 899)}
	repo := &fakeReconciliationRepository{
		payments: map[string]domain.Payment{
			"stripe_like/psp_match":    capturedPayment("pay_match", "psp_match", 1000),
			"stripe_like/psp_mismatch": capturedPayment("pay_mismatch", "psp_mismatch", 900),
		},
		stored: make(map[string]domain.PaymentReconciliation),
	}
	alerts := &fakeReconciliationAlerts{err: errors.New("alert endpoint unavailable")}
	uc := newReconciliationUsecaseForTest(t, repo, source, alerts)

	summary, err := uc.Execute(context.Background(), "stripe_like", reportDate)
	if err != nil {
		t.Fatalf("Execute() error = %v, want persisted reconciliation despite alert failure", err)
	}
	if summary.CreatedCount != 3 || summary.AlertFailureCount != 2 || len(repo.stored) != 3 {
		t.Fatalf("summary/stored = %+v/%d, want stored rows and observable alert failures", summary, len(repo.stored))
	}
}

func TestSameReconciliationEvidenceDoesNotRoundMinorUnits(t *testing.T) {
	left := domain.PaymentReconciliation{
		ReconciliationID: "rec_high_value",
		Provider:         "stripe_like",
		SettlementID:     "stl_high_value",
		Status:           domain.ReconciliationStatusMismatch,
		Details:          []byte(`{"provider":{"settled_amount":9007199254740992}}`),
	}
	right := left
	right.Details = []byte(`{"provider":{"settled_amount":9007199254740993}}`)

	if sameReconciliationEvidence(left, right) {
		t.Fatal("sameReconciliationEvidence() = true, want exact integer minor-unit comparison")
	}
}

func reconciliationReport(t *testing.T, reportDate time.Time, mismatchAmount int64) settlement.Report {
	t.Helper()
	report, err := settlement.NewReport("stripe_like", reportDate, []settlement.Row{
		{
			SettlementID:      "stl_20260526",
			ProviderPaymentID: "psp_match",
			Outcome:           settlement.OutcomeCaptured,
			Currency:          "INR",
			SettledAmount:     1000,
			SettledAt:         reportDate.Add(time.Hour),
		},
		{
			SettlementID:      "stl_20260526",
			ProviderPaymentID: "psp_mismatch",
			Outcome:           settlement.OutcomeCaptured,
			Currency:          "INR",
			SettledAmount:     mismatchAmount,
			SettledAt:         reportDate.Add(2 * time.Hour),
		},
		{
			SettlementID:      "stl_20260526",
			ProviderPaymentID: "psp_missing_local",
			Outcome:           settlement.OutcomeCaptured,
			Currency:          "INR",
			SettledAmount:     500,
			SettledAt:         reportDate.Add(3 * time.Hour),
		},
	})
	if err != nil {
		t.Fatalf("NewReport() error = %v", err)
	}
	return report
}

func capturedPayment(paymentID string, providerPaymentID string, amount int64) domain.Payment {
	return domain.Payment{
		PaymentID:         paymentID,
		Provider:          "stripe_like",
		ProviderPaymentID: providerPaymentID,
		Amount:            domain.Money{Amount: amount, Currency: "INR"},
		Status:            domain.PaymentStatusCaptured,
		CapturedAmount:    amount,
	}
}

func newReconciliationUsecaseForTest(
	t *testing.T,
	repo ReconciliationRepository,
	source settlement.Source,
	alerts ReconciliationAlertPublisher,
) *ReconcileSettlementsUsecase {
	t.Helper()
	uc, err := NewReconcileSettlementsUsecase(
		repo,
		source,
		alerts,
		ReconcileSettlementsConfig{BatchSize: 2},
		slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
		WithReconciliationClock(func() time.Time {
			return time.Date(2026, time.May, 27, 2, 0, 0, 0, time.UTC)
		}),
	)
	if err != nil {
		t.Fatalf("NewReconcileSettlementsUsecase() error = %v", err)
	}
	return uc
}

type fakeSettlementSource struct {
	report settlement.Report
	err    error
}

func (s *fakeSettlementSource) Fetch(_ context.Context, _ string, _ time.Time) (settlement.Report, error) {
	return s.report, s.err
}

type fakeReconciliationRepository struct {
	payments   map[string]domain.Payment
	candidates []domain.Payment
	stored     map[string]domain.PaymentReconciliation
}

func (r *fakeReconciliationRepository) GetPaymentByProviderPaymentID(_ context.Context, providerName string, providerPaymentID string) (domain.Payment, error) {
	payment, exists := r.payments[providerName+"/"+providerPaymentID]
	if !exists {
		return domain.Payment{}, domain.ErrPaymentRecordNotFound
	}
	return payment, nil
}

func (r *fakeReconciliationRepository) ListSettlementCandidates(_ context.Context, _ string, _ time.Time, _ time.Time, afterPaymentID string, limit int) ([]domain.Payment, error) {
	candidates := append([]domain.Payment(nil), r.candidates...)
	sort.Slice(candidates, func(i int, j int) bool { return candidates[i].PaymentID < candidates[j].PaymentID })
	page := make([]domain.Payment, 0, limit)
	for _, candidate := range candidates {
		if candidate.PaymentID > afterPaymentID && len(page) < limit {
			page = append(page, candidate)
		}
	}
	return page, nil
}

func (r *fakeReconciliationRepository) CreateReconciliationIfAbsent(_ context.Context, result domain.PaymentReconciliation) (bool, error) {
	if _, exists := r.stored[result.ReconciliationID]; exists {
		return false, nil
	}
	r.stored[result.ReconciliationID] = result
	return true, nil
}

func (r *fakeReconciliationRepository) GetReconciliationByID(_ context.Context, reconciliationID string) (domain.PaymentReconciliation, error) {
	result, exists := r.stored[reconciliationID]
	if !exists {
		return domain.PaymentReconciliation{}, domain.ErrReconciliationNotFound
	}
	return result, nil
}

type fakeReconciliationAlerts struct {
	alerts []domain.ReconciliationAlert
	err    error
}

func (p *fakeReconciliationAlerts) PublishMismatch(_ context.Context, _ string, alert domain.ReconciliationAlert) error {
	p.alerts = append(p.alerts, alert)
	return p.err
}
