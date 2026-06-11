package settlement

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const ReportDateLayout = "2006-01-02"

type Outcome string

const (
	OutcomeCaptured Outcome = "captured"
	OutcomeRefunded Outcome = "refunded"
)

type Row struct {
	Provider          string
	SettlementID      string
	ProviderPaymentID string
	Outcome           Outcome
	Currency          string
	SettledAmount     int64
	FeeAmount         int64
	SettledAt         time.Time
	ReportDate        string
}

type Report struct {
	Provider     string
	SettlementID string
	ReportDate   string
	WindowStart  time.Time
	WindowEnd    time.Time
	Rows         []Row
}

type Source interface {
	Fetch(ctx context.Context, provider string, reportDate time.Time) (Report, error)
}

func (r Row) Normalized() Row {
	r.Provider = strings.ToLower(strings.TrimSpace(r.Provider))
	r.SettlementID = strings.TrimSpace(r.SettlementID)
	r.ProviderPaymentID = strings.TrimSpace(r.ProviderPaymentID)
	r.Outcome = Outcome(strings.ToLower(strings.TrimSpace(string(r.Outcome))))
	r.Currency = strings.ToUpper(strings.TrimSpace(r.Currency))
	r.SettledAt = r.SettledAt.UTC()
	r.ReportDate = strings.TrimSpace(r.ReportDate)
	return r
}

func (r Row) Validate() error {
	r = r.Normalized()
	switch {
	case r.Provider == "" || len(r.Provider) > 64:
		return errors.New("settlement row provider is required and cannot exceed 64 characters")
	case r.SettlementID == "" || len(r.SettlementID) > 128:
		return errors.New("settlement row settlement_id is required and cannot exceed 128 characters")
	case r.ProviderPaymentID == "" || len(r.ProviderPaymentID) > 128:
		return errors.New("settlement row provider_payment_id is required and cannot exceed 128 characters")
	case r.Outcome != OutcomeCaptured && r.Outcome != OutcomeRefunded:
		return fmt.Errorf("settlement row outcome %q is unsupported", r.Outcome)
	case !isCurrency(r.Currency):
		return errors.New("settlement row currency must be a 3-letter ISO code")
	case r.SettledAmount <= 0:
		return errors.New("settlement row settled_amount must be greater than zero")
	case r.FeeAmount < 0:
		return errors.New("settlement row fee_amount cannot be negative")
	case r.SettledAt.IsZero():
		return errors.New("settlement row settled_at is required")
	}
	if _, err := time.Parse(ReportDateLayout, r.ReportDate); err != nil {
		return fmt.Errorf("settlement row report_date must be YYYY-MM-DD: %w", err)
	}
	return nil
}

func NewReport(provider string, reportDate time.Time, rows []Row) (Report, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" || len(provider) > 64 {
		return Report{}, errors.New("settlement report provider is required and cannot exceed 64 characters")
	}
	if reportDate.IsZero() {
		return Report{}, errors.New("settlement report date is required")
	}
	reportDate = reportDate.UTC()
	reportDateText := reportDate.Format(ReportDateLayout)
	start, err := time.Parse(ReportDateLayout, reportDateText)
	if err != nil {
		return Report{}, fmt.Errorf("normalize settlement report date: %w", err)
	}
	if len(rows) == 0 {
		return Report{}, errors.New("settlement report contains no rows")
	}

	normalizedRows := make([]Row, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	settlementID := ""
	for index, row := range rows {
		row.Provider = provider
		row.ReportDate = reportDateText
		row = row.Normalized()
		if err := row.Validate(); err != nil {
			return Report{}, fmt.Errorf("settlement row %d: %w", index+1, err)
		}
		if settlementID == "" {
			settlementID = row.SettlementID
		}
		if row.SettlementID != settlementID {
			return Report{}, errors.New("settlement report contains multiple settlement_id values")
		}
		if _, exists := seen[row.ProviderPaymentID]; exists {
			return Report{}, fmt.Errorf("settlement report contains duplicate provider_payment_id %q", row.ProviderPaymentID)
		}
		seen[row.ProviderPaymentID] = struct{}{}
		normalizedRows = append(normalizedRows, row)
	}
	return Report{
		Provider:     provider,
		SettlementID: settlementID,
		ReportDate:   reportDateText,
		WindowStart:  start.UTC(),
		WindowEnd:    start.AddDate(0, 0, 1).UTC(),
		Rows:         normalizedRows,
	}, nil
}

func (r Report) Validate(expectedProvider string, expectedDate time.Time) error {
	expectedProvider = strings.ToLower(strings.TrimSpace(expectedProvider))
	if expectedProvider == "" || expectedDate.IsZero() {
		return errors.New("expected settlement provider and report date are required")
	}
	for _, row := range r.Rows {
		row = row.Normalized()
		if row.Provider != strings.ToLower(strings.TrimSpace(r.Provider)) || row.ReportDate != strings.TrimSpace(r.ReportDate) {
			return errors.New("settlement report row metadata does not match report metadata")
		}
	}
	normalized, err := NewReport(r.Provider, expectedDate, r.Rows)
	if err != nil {
		return err
	}
	if normalized.Provider != expectedProvider {
		return errors.New("settlement report provider does not match requested provider")
	}
	if normalized.ReportDate != r.ReportDate ||
		normalized.SettlementID != strings.TrimSpace(r.SettlementID) ||
		!normalized.WindowStart.Equal(r.WindowStart.UTC()) ||
		!normalized.WindowEnd.Equal(r.WindowEnd.UTC()) {
		return errors.New("settlement report metadata does not match its reporting window or rows")
	}
	return nil
}

func NormalizeOutcome(raw string) (Outcome, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "captured", "settled":
		return OutcomeCaptured, nil
	case "refunded":
		return OutcomeRefunded, nil
	default:
		return "", fmt.Errorf("unsupported settlement outcome %q", raw)
	}
}

func isCurrency(currency string) bool {
	if len(currency) != 3 {
		return false
	}
	for _, char := range currency {
		if char < 'A' || char > 'Z' {
			return false
		}
	}
	return true
}
