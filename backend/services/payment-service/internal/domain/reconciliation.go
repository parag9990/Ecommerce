package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ReconciliationStatus string

const (
	ReconciliationStatusUnspecified     ReconciliationStatus = ""
	ReconciliationStatusMatched         ReconciliationStatus = "matched"
	ReconciliationStatusMismatch        ReconciliationStatus = "mismatch"
	ReconciliationStatusMissingLocal    ReconciliationStatus = "missing_local"
	ReconciliationStatusMissingProvider ReconciliationStatus = "missing_provider"
)

type PaymentReconciliation struct {
	ReconciliationID  string
	Provider          string
	SettlementID      string
	Status            ReconciliationStatus
	PaymentID         string
	ProviderPaymentID string
	Details           json.RawMessage
	CreatedAt         time.Time
}

type ReconciliationDetails struct {
	ReportDate       string                          `json:"report_date"`
	ReasonCodes      []string                        `json:"reason_codes,omitempty"`
	Local            *ReconciliationLocalEvidence    `json:"local,omitempty"`
	Provider         *ReconciliationProviderEvidence `json:"provider,omitempty"`
	DifferenceAmount *int64                          `json:"difference_amount,omitempty"`
}

type ReconciliationLocalEvidence struct {
	Status         PaymentStatus `json:"status"`
	Currency       string        `json:"currency"`
	CapturedAmount int64         `json:"captured_amount"`
	RefundedAmount int64         `json:"refunded_amount"`
}

type ReconciliationProviderEvidence struct {
	Outcome       string `json:"outcome"`
	Currency      string `json:"currency"`
	SettledAmount int64  `json:"settled_amount"`
	FeeAmount     int64  `json:"fee_amount"`
}

type ReconciliationAlert struct {
	EventID           string               `json:"event_id"`
	EventType         string               `json:"event_type"`
	ReconciliationID  string               `json:"reconciliation_id"`
	Provider          string               `json:"provider"`
	SettlementID      string               `json:"settlement_id"`
	Status            ReconciliationStatus `json:"status"`
	PaymentID         string               `json:"payment_id,omitempty"`
	ProviderPaymentID string               `json:"provider_payment_id,omitempty"`
	ReasonCodes       []string             `json:"reason_codes"`
	DifferenceAmount  *int64               `json:"difference_amount,omitempty"`
	Currency          string               `json:"currency,omitempty"`
	Severity          string               `json:"severity"`
	OccurredAt        time.Time            `json:"occurred_at"`
}

func (r PaymentReconciliation) Normalized() PaymentReconciliation {
	r.ReconciliationID = strings.TrimSpace(r.ReconciliationID)
	r.Provider = strings.ToLower(strings.TrimSpace(r.Provider))
	r.SettlementID = strings.TrimSpace(r.SettlementID)
	r.PaymentID = strings.TrimSpace(r.PaymentID)
	r.ProviderPaymentID = strings.TrimSpace(r.ProviderPaymentID)
	r.CreatedAt = r.CreatedAt.UTC()
	return r
}

func (r PaymentReconciliation) Validate() error {
	r = r.Normalized()
	if r.ReconciliationID == "" {
		return fmt.Errorf("%w: reconciliation_id is required", ErrInvalidReconciliation)
	}
	if len(r.ReconciliationID) > 64 {
		return fmt.Errorf("%w: reconciliation_id exceeds 64 characters", ErrInvalidReconciliation)
	}
	if r.Provider == "" {
		return fmt.Errorf("%w: provider is required", ErrInvalidReconciliation)
	}
	if len(r.Provider) > 64 {
		return fmt.Errorf("%w: provider exceeds 64 characters", ErrInvalidReconciliation)
	}
	if !IsKnownReconciliationStatus(r.Status) {
		return fmt.Errorf("%w: invalid reconciliation status %s", ErrInvalidReconciliation, r.Status)
	}
	if len(r.SettlementID) > 128 {
		return fmt.Errorf("%w: settlement_id exceeds 128 characters", ErrInvalidReconciliation)
	}
	if len(r.PaymentID) > 64 {
		return fmt.Errorf("%w: payment_id exceeds 64 characters", ErrInvalidReconciliation)
	}
	if len(r.ProviderPaymentID) > 128 {
		return fmt.Errorf("%w: provider_payment_id exceeds 128 characters", ErrInvalidReconciliation)
	}
	if r.SettlementID == "" {
		return fmt.Errorf("%w: settlement_id is required", ErrInvalidReconciliation)
	}
	if r.Status != ReconciliationStatusMissingProvider && r.ProviderPaymentID == "" {
		return fmt.Errorf("%w: provider_payment_id is required for provider report results", ErrInvalidReconciliation)
	}
	if r.Status == ReconciliationStatusMissingProvider && r.PaymentID == "" {
		return fmt.Errorf("%w: payment_id is required for a missing provider result", ErrInvalidReconciliation)
	}
	if err := validateSanitizedJSONPayload(r.Details, false, "details"); err != nil {
		return err
	}
	return nil
}

func IsKnownReconciliationStatus(status ReconciliationStatus) bool {
	switch status {
	case ReconciliationStatusMatched, ReconciliationStatusMismatch, ReconciliationStatusMissingLocal, ReconciliationStatusMissingProvider:
		return true
	default:
		return false
	}
}

func (a ReconciliationAlert) Normalized() ReconciliationAlert {
	a.EventID = strings.TrimSpace(a.EventID)
	a.EventType = strings.TrimSpace(a.EventType)
	a.ReconciliationID = strings.TrimSpace(a.ReconciliationID)
	a.Provider = strings.ToLower(strings.TrimSpace(a.Provider))
	a.SettlementID = strings.TrimSpace(a.SettlementID)
	a.PaymentID = strings.TrimSpace(a.PaymentID)
	a.ProviderPaymentID = strings.TrimSpace(a.ProviderPaymentID)
	a.Currency = strings.ToUpper(strings.TrimSpace(a.Currency))
	a.Severity = strings.ToLower(strings.TrimSpace(a.Severity))
	a.OccurredAt = a.OccurredAt.UTC()
	for i := range a.ReasonCodes {
		a.ReasonCodes[i] = strings.TrimSpace(a.ReasonCodes[i])
	}
	return a
}

func (a ReconciliationAlert) Validate() error {
	a = a.Normalized()
	if a.EventID == "" || a.EventType == "" || a.ReconciliationID == "" {
		return fmt.Errorf("%w: event_id, event_type, and reconciliation_id are required", ErrInvalidReconciliation)
	}
	if a.Provider == "" || a.SettlementID == "" {
		return fmt.Errorf("%w: provider and settlement_id are required", ErrInvalidReconciliation)
	}
	if a.Status == ReconciliationStatusMatched || !IsKnownReconciliationStatus(a.Status) {
		return fmt.Errorf("%w: alerts require a non-matched reconciliation status", ErrInvalidReconciliation)
	}
	if len(a.ReasonCodes) == 0 {
		return fmt.Errorf("%w: alerts require at least one reason code", ErrInvalidReconciliation)
	}
	if a.Severity != "critical" && a.Severity != "high" {
		return fmt.Errorf("%w: severity must be critical or high", ErrInvalidReconciliation)
	}
	if a.OccurredAt.IsZero() {
		return fmt.Errorf("%w: occurred_at is required", ErrInvalidReconciliation)
	}
	return nil
}
