package usecase

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider/settlement"
)

const (
	PaymentReconciliationAlertEventType = "PaymentReconciliationMismatchDetected"
	DefaultReconciliationAlertTopic     = "payment.reconciliation.alerts"
	DefaultReconciliationBatchSize      = 500
)

const (
	reasonAmountMismatch                 = "amount_mismatch"
	reasonCurrencyMismatch               = "currency_mismatch"
	reasonProviderOutcomeMismatch        = "provider_outcome_mismatch"
	reasonLocalStatusNotSettleable       = "local_status_not_settleable"
	reasonProviderPaymentNotFoundLocally = "provider_payment_not_found_locally"
	reasonCapturedPaymentMissingInReport = "captured_payment_missing_in_report"
)

type ReconciliationRepository interface {
	GetPaymentByProviderPaymentID(ctx context.Context, provider string, providerPaymentID string) (domain.Payment, error)
	ListSettlementCandidates(ctx context.Context, provider string, from time.Time, to time.Time, afterPaymentID string, limit int) ([]domain.Payment, error)
	CreateReconciliationIfAbsent(ctx context.Context, result domain.PaymentReconciliation) (bool, error)
	GetReconciliationByID(ctx context.Context, reconciliationID string) (domain.PaymentReconciliation, error)
}

type ReconciliationAlertPublisher interface {
	PublishMismatch(ctx context.Context, topic string, alert domain.ReconciliationAlert) error
}

type ReconcileSettlementsConfig struct {
	AlertTopic string
	BatchSize  int
}

func (c ReconcileSettlementsConfig) normalized() ReconcileSettlementsConfig {
	c.AlertTopic = strings.TrimSpace(c.AlertTopic)
	if c.AlertTopic == "" {
		c.AlertTopic = DefaultReconciliationAlertTopic
	}
	if c.BatchSize == 0 {
		c.BatchSize = DefaultReconciliationBatchSize
	}
	return c
}

func (c ReconcileSettlementsConfig) validate() error {
	c = c.normalized()
	if c.BatchSize <= 0 || c.BatchSize > 10000 {
		return errors.New("reconciliation batch size must be between 1 and 10000")
	}
	return nil
}

type ReconciliationSummary struct {
	Provider             string
	SettlementID         string
	ReportDate           string
	MatchedCount         int
	MismatchCount        int
	MissingLocalCount    int
	MissingProviderCount int
	CreatedCount         int
	AlertedCount         int
	AlertFailureCount    int
}

type ReconcileSettlementsUsecase struct {
	repository ReconciliationRepository
	source     settlement.Source
	alerts     ReconciliationAlertPublisher
	config     ReconcileSettlementsConfig
	clock      func() time.Time
	logger     *slog.Logger
}

type ReconcileSettlementsOption func(*ReconcileSettlementsUsecase)

func WithReconciliationClock(clock func() time.Time) ReconcileSettlementsOption {
	return func(u *ReconcileSettlementsUsecase) {
		if clock != nil {
			u.clock = clock
		}
	}
}

func NewReconcileSettlementsUsecase(
	repository ReconciliationRepository,
	source settlement.Source,
	alerts ReconciliationAlertPublisher,
	cfg ReconcileSettlementsConfig,
	logger *slog.Logger,
	opts ...ReconcileSettlementsOption,
) (*ReconcileSettlementsUsecase, error) {
	if repository == nil {
		return nil, errors.New("reconciliation repository is required")
	}
	if source == nil {
		return nil, errors.New("settlement report source is required")
	}
	if alerts == nil {
		return nil, errors.New("reconciliation alert publisher is required")
	}
	cfg = cfg.normalized()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	u := &ReconcileSettlementsUsecase{
		repository: repository,
		source:     source,
		alerts:     alerts,
		config:     cfg,
		clock:      func() time.Time { return time.Now().UTC() },
		logger:     logger,
	}
	for _, option := range opts {
		option(u)
	}
	return u, nil
}

func (u *ReconcileSettlementsUsecase) Execute(ctx context.Context, providerName string, reportDate time.Time) (ReconciliationSummary, error) {
	if err := ctx.Err(); err != nil {
		return ReconciliationSummary{}, err
	}
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	if providerName == "" || reportDate.IsZero() {
		return ReconciliationSummary{}, fmt.Errorf("%w: provider and report date are required", domain.ErrInvalidReconciliation)
	}
	report, err := u.source.Fetch(ctx, providerName, reportDate)
	if err != nil {
		return ReconciliationSummary{}, fmt.Errorf("fetch settlement report: %w", err)
	}
	if err := report.Validate(providerName, reportDate); err != nil {
		return ReconciliationSummary{}, fmt.Errorf("%w: %v", domain.ErrInvalidReconciliation, err)
	}

	summary := ReconciliationSummary{
		Provider:     report.Provider,
		SettlementID: report.SettlementID,
		ReportDate:   report.ReportDate,
	}
	seen := make(map[string]struct{}, len(report.Rows))
	for _, row := range report.Rows {
		result, alert, err := u.resultForProviderRow(ctx, report, row)
		if err != nil {
			return summary, err
		}
		created, err := u.persistResult(ctx, result)
		if err != nil {
			return summary, err
		}
		summary.add(result.Status, created)
		if created && alert != nil {
			u.publishAlert(ctx, *alert, &summary)
		}
		seen[row.ProviderPaymentID] = struct{}{}
	}

	afterPaymentID := ""
	for {
		candidates, err := u.repository.ListSettlementCandidates(
			ctx,
			report.Provider,
			report.WindowStart,
			report.WindowEnd,
			afterPaymentID,
			u.config.BatchSize,
		)
		if err != nil {
			return summary, fmt.Errorf("list settlement candidates: %w", err)
		}
		for _, candidate := range candidates {
			afterPaymentID = candidate.PaymentID
			if _, exists := seen[candidate.ProviderPaymentID]; exists {
				continue
			}
			result, alert, err := u.missingProviderResult(report, candidate)
			if err != nil {
				return summary, err
			}
			created, err := u.persistResult(ctx, result)
			if err != nil {
				return summary, err
			}
			summary.add(result.Status, created)
			if created {
				u.publishAlert(ctx, alert, &summary)
			}
		}
		if len(candidates) < u.config.BatchSize {
			break
		}
	}
	return summary, nil
}

func (u *ReconcileSettlementsUsecase) resultForProviderRow(
	ctx context.Context,
	report settlement.Report,
	row settlement.Row,
) (domain.PaymentReconciliation, *domain.ReconciliationAlert, error) {
	local, err := u.repository.GetPaymentByProviderPaymentID(ctx, row.Provider, row.ProviderPaymentID)
	if errors.Is(err, domain.ErrPaymentRecordNotFound) {
		reasons := []string{reasonProviderPaymentNotFoundLocally}
		result, err := u.newResult(report, row.ProviderPaymentID, "", domain.ReconciliationStatusMissingLocal, reasons, nil, &row, nil)
		if err != nil {
			return domain.PaymentReconciliation{}, nil, err
		}
		alert := newReconciliationAlert(result, reasons, nil, row.Currency, u.clock())
		return result, &alert, nil
	}
	if err != nil {
		return domain.PaymentReconciliation{}, nil, fmt.Errorf("find local payment for provider payment %s: %w", row.ProviderPaymentID, err)
	}
	comparison := compareCapturedPayment(local, row)
	result, err := u.newResult(
		report,
		row.ProviderPaymentID,
		local.PaymentID,
		comparison.Status,
		comparison.ReasonCodes,
		&local,
		&row,
		comparison.DifferenceAmount,
	)
	if err != nil {
		return domain.PaymentReconciliation{}, nil, err
	}
	if comparison.Status == domain.ReconciliationStatusMatched {
		return result, nil, nil
	}
	alert := newReconciliationAlert(result, comparison.ReasonCodes, comparison.DifferenceAmount, row.Currency, u.clock())
	return result, &alert, nil
}

func (u *ReconcileSettlementsUsecase) missingProviderResult(
	report settlement.Report,
	local domain.Payment,
) (domain.PaymentReconciliation, domain.ReconciliationAlert, error) {
	reasons := []string{reasonCapturedPaymentMissingInReport}
	result, err := u.newResult(
		report,
		local.ProviderPaymentID,
		local.PaymentID,
		domain.ReconciliationStatusMissingProvider,
		reasons,
		&local,
		nil,
		nil,
	)
	if err != nil {
		return domain.PaymentReconciliation{}, domain.ReconciliationAlert{}, err
	}
	return result, newReconciliationAlert(result, reasons, nil, local.Amount.Currency, u.clock()), nil
}

func (u *ReconcileSettlementsUsecase) newResult(
	report settlement.Report,
	providerPaymentID string,
	paymentID string,
	status domain.ReconciliationStatus,
	reasons []string,
	local *domain.Payment,
	external *settlement.Row,
	difference *int64,
) (domain.PaymentReconciliation, error) {
	details := domain.ReconciliationDetails{
		ReportDate:       report.ReportDate,
		ReasonCodes:      append([]string(nil), reasons...),
		DifferenceAmount: difference,
	}
	if local != nil {
		details.Local = &domain.ReconciliationLocalEvidence{
			Status:         local.Status,
			Currency:       local.Amount.Currency,
			CapturedAmount: local.CapturedAmount,
			RefundedAmount: local.RefundedAmount,
		}
	}
	if external != nil {
		details.Provider = &domain.ReconciliationProviderEvidence{
			Outcome:       string(external.Outcome),
			Currency:      external.Currency,
			SettledAmount: external.SettledAmount,
			FeeAmount:     external.FeeAmount,
		}
	}
	rawDetails, err := json.Marshal(details)
	if err != nil {
		return domain.PaymentReconciliation{}, fmt.Errorf("marshal reconciliation details: %w", err)
	}
	idParts := []string{report.Provider, report.SettlementID, providerPaymentID}
	if status == domain.ReconciliationStatusMissingProvider {
		idParts = []string{report.Provider, report.SettlementID, paymentID, string(status)}
	}
	result := domain.PaymentReconciliation{
		ReconciliationID:  reconciliationID(idParts...),
		Provider:          report.Provider,
		SettlementID:      report.SettlementID,
		Status:            status,
		PaymentID:         paymentID,
		ProviderPaymentID: providerPaymentID,
		Details:           rawDetails,
		CreatedAt:         u.clock().UTC(),
	}.Normalized()
	if err := result.Validate(); err != nil {
		return domain.PaymentReconciliation{}, err
	}
	return result, nil
}

func (u *ReconcileSettlementsUsecase) persistResult(ctx context.Context, result domain.PaymentReconciliation) (bool, error) {
	created, err := u.repository.CreateReconciliationIfAbsent(ctx, result)
	if err != nil {
		return false, fmt.Errorf("persist reconciliation %s: %w", result.ReconciliationID, err)
	}
	if created {
		return true, nil
	}
	existing, err := u.repository.GetReconciliationByID(ctx, result.ReconciliationID)
	if err != nil {
		return false, fmt.Errorf("load replayed reconciliation %s: %w", result.ReconciliationID, err)
	}
	if !sameReconciliationEvidence(existing, result) {
		return false, fmt.Errorf("%w: %s", domain.ErrReconciliationConflict, result.ReconciliationID)
	}
	return false, nil
}

func (u *ReconcileSettlementsUsecase) publishAlert(ctx context.Context, alert domain.ReconciliationAlert, summary *ReconciliationSummary) {
	if err := u.alerts.PublishMismatch(ctx, u.config.AlertTopic, alert); err != nil {
		summary.AlertFailureCount++
		u.logger.ErrorContext(ctx, "payment.reconciliation.alert_publish_failed",
			slog.String("reconciliation_id", alert.ReconciliationID),
			slog.String("provider", alert.Provider),
			slog.String("status", string(alert.Status)),
			slog.String("error", err.Error()),
		)
		return
	}
	summary.AlertedCount++
}

func (s *ReconciliationSummary) add(status domain.ReconciliationStatus, created bool) {
	switch status {
	case domain.ReconciliationStatusMatched:
		s.MatchedCount++
	case domain.ReconciliationStatusMismatch:
		s.MismatchCount++
	case domain.ReconciliationStatusMissingLocal:
		s.MissingLocalCount++
	case domain.ReconciliationStatusMissingProvider:
		s.MissingProviderCount++
	}
	if created {
		s.CreatedCount++
	}
}

type reconciliationComparison struct {
	Status           domain.ReconciliationStatus
	ReasonCodes      []string
	DifferenceAmount *int64
}

func compareCapturedPayment(local domain.Payment, external settlement.Row) reconciliationComparison {
	reasons := make([]string, 0, 4)
	if local.Amount.Currency != external.Currency {
		reasons = append(reasons, reasonCurrencyMismatch)
	}
	if local.CapturedAmount != external.SettledAmount {
		reasons = append(reasons, reasonAmountMismatch)
	}
	if !settleableStatus(local.Status) {
		reasons = append(reasons, reasonLocalStatusNotSettleable)
	}
	if external.Outcome != settlement.OutcomeCaptured {
		reasons = append(reasons, reasonProviderOutcomeMismatch)
	}
	if len(reasons) == 0 {
		return reconciliationComparison{Status: domain.ReconciliationStatusMatched}
	}
	difference := external.SettledAmount - local.CapturedAmount
	return reconciliationComparison{
		Status:           domain.ReconciliationStatusMismatch,
		ReasonCodes:      reasons,
		DifferenceAmount: &difference,
	}
}

func settleableStatus(status domain.PaymentStatus) bool {
	switch status {
	case domain.PaymentStatusCaptured, domain.PaymentStatusPartiallyRefunded, domain.PaymentStatusRefunded:
		return true
	default:
		return false
	}
}

func newReconciliationAlert(
	result domain.PaymentReconciliation,
	reasons []string,
	difference *int64,
	currency string,
	now time.Time,
) domain.ReconciliationAlert {
	return domain.ReconciliationAlert{
		EventID:           "ral_" + strings.TrimPrefix(result.ReconciliationID, "rec_"),
		EventType:         PaymentReconciliationAlertEventType,
		ReconciliationID:  result.ReconciliationID,
		Provider:          result.Provider,
		SettlementID:      result.SettlementID,
		Status:            result.Status,
		PaymentID:         result.PaymentID,
		ProviderPaymentID: result.ProviderPaymentID,
		ReasonCodes:       append([]string(nil), reasons...),
		DifferenceAmount:  difference,
		Currency:          currency,
		Severity:          alertSeverity(result.Status, reasons),
		OccurredAt:        now.UTC(),
	}.Normalized()
}

func alertSeverity(status domain.ReconciliationStatus, reasons []string) string {
	if status == domain.ReconciliationStatusMissingLocal {
		return "critical"
	}
	for _, reason := range reasons {
		if reason == reasonAmountMismatch || reason == reasonCurrencyMismatch {
			return "critical"
		}
	}
	return "high"
}

func reconciliationID(parts ...string) string {
	normalized := strings.Join(parts, "|")
	sum := sha256.Sum256([]byte(normalized))
	return "rec_" + hex.EncodeToString(sum[:])[:32]
}

func sameReconciliationEvidence(left domain.PaymentReconciliation, right domain.PaymentReconciliation) bool {
	left = left.Normalized()
	right = right.Normalized()
	if left.ReconciliationID != right.ReconciliationID ||
		left.Provider != right.Provider ||
		left.SettlementID != right.SettlementID ||
		left.Status != right.Status ||
		left.PaymentID != right.PaymentID ||
		left.ProviderPaymentID != right.ProviderPaymentID {
		return false
	}
	var leftDetails any
	var rightDetails any
	leftDecoder := json.NewDecoder(bytes.NewReader(left.Details))
	rightDecoder := json.NewDecoder(bytes.NewReader(right.Details))
	leftDecoder.UseNumber()
	rightDecoder.UseNumber()
	if leftDecoder.Decode(&leftDetails) != nil || rightDecoder.Decode(&rightDetails) != nil {
		return false
	}
	return reflect.DeepEqual(leftDetails, rightDetails)
}
