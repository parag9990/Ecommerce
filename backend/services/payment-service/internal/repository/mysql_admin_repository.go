package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
)

type AdminPaymentPage struct {
	Payments []domain.Payment
	Total    int64
}

type AdminRefundPage struct {
	Refunds []domain.Refund
	Total   int64
}

type AdminReconciliationPage struct {
	Reconciliations []domain.PaymentReconciliation
	Total           int64
}

const paymentSelectQuery = `SELECT payment_id, order_id, user_id, provider, provider_payment_id,
provider_intent_id, status, currency, amount, captured_amount, refunded_amount,
idempotency_key, retry_of_payment_id, root_payment_id, attempt_no, retry_request_key,
failure_code, failure_message, created_at, updated_at FROM payments`

func (r *MySQLPaymentRepository) ListPaymentsForAdmin(ctx context.Context, status, providerName, orderID string, limit, offset int) (AdminPaymentPage, error) {
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 5)
	if status = strings.TrimSpace(status); status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}
	if providerName = strings.TrimSpace(providerName); providerName != "" {
		conditions = append(conditions, "provider = ?")
		args = append(args, strings.ToLower(providerName))
	}
	if orderID = strings.TrimSpace(orderID); orderID != "" {
		conditions = append(conditions, "order_id = ?")
		args = append(args, orderID)
	}
	where := " WHERE " + strings.Join(conditions, " AND ")
	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM payments"+where, args...).Scan(&total); err != nil {
		return AdminPaymentPage{}, fmt.Errorf("count admin payments: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, paymentSelectQuery+where+" ORDER BY created_at DESC, payment_id DESC LIMIT ? OFFSET ?", append(args, limit, offset)...)
	if err != nil {
		return AdminPaymentPage{}, fmt.Errorf("list admin payments: %w", err)
	}
	defer rows.Close()
	payments := make([]domain.Payment, 0, limit)
	for rows.Next() {
		payment, err := scanPayment(rows)
		if err != nil {
			return AdminPaymentPage{}, fmt.Errorf("scan admin payment: %w", err)
		}
		payments = append(payments, payment)
	}
	if err := rows.Err(); err != nil {
		return AdminPaymentPage{}, fmt.Errorf("iterate admin payments: %w", err)
	}
	return AdminPaymentPage{Payments: payments, Total: total}, nil
}

func (r *MySQLPaymentRepository) GetPaymentForAdmin(ctx context.Context, paymentID string) (domain.Payment, error) {
	payment, err := r.GetPaymentByID(ctx, strings.TrimSpace(paymentID))
	if err != nil {
		return domain.Payment{}, err
	}
	return payment, nil
}

func (r *MySQLPaymentRepository) ListRefundsForAdmin(ctx context.Context, status string, limit, offset int) (AdminRefundPage, error) {
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 2)
	if status = strings.TrimSpace(status); status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}
	where := " WHERE " + strings.Join(conditions, " AND ")
	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM refunds"+where, args...).Scan(&total); err != nil {
		return AdminRefundPage{}, fmt.Errorf("count admin refunds: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, refundSelectQuery+where+" ORDER BY created_at DESC, refund_id DESC LIMIT ? OFFSET ?", append(args, limit, offset)...)
	if err != nil {
		return AdminRefundPage{}, fmt.Errorf("list admin refunds: %w", err)
	}
	defer rows.Close()
	refunds := make([]domain.Refund, 0, limit)
	for rows.Next() {
		refund, err := scanRefund(rows)
		if err != nil {
			return AdminRefundPage{}, fmt.Errorf("scan admin refund: %w", err)
		}
		refunds = append(refunds, refund)
	}
	if err := rows.Err(); err != nil {
		return AdminRefundPage{}, fmt.Errorf("iterate admin refunds: %w", err)
	}
	return AdminRefundPage{Refunds: refunds, Total: total}, nil
}

func (r *MySQLPaymentRepository) ListRefundsForPaymentAdmin(ctx context.Context, paymentID string) ([]domain.Refund, error) {
	rows, err := r.db.QueryContext(ctx, refundSelectQuery+" WHERE payment_id = ? ORDER BY created_at DESC, refund_id DESC", strings.TrimSpace(paymentID))
	if err != nil {
		return nil, fmt.Errorf("list payment refunds: %w", err)
	}
	defer rows.Close()
	refunds := make([]domain.Refund, 0)
	for rows.Next() {
		refund, err := scanRefund(rows)
		if err != nil {
			return nil, fmt.Errorf("scan payment refund: %w", err)
		}
		refunds = append(refunds, refund)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate payment refunds: %w", err)
	}
	return refunds, nil
}

func (r *MySQLPaymentRepository) ListReconciliationsForAdmin(ctx context.Context, status string, limit, offset int) (AdminReconciliationPage, error) {
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 2)
	if status = strings.TrimSpace(status); status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}
	where := " WHERE " + strings.Join(conditions, " AND ")
	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM payment_reconciliations"+where, args...).Scan(&total); err != nil {
		return AdminReconciliationPage{}, fmt.Errorf("count admin reconciliations: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT reconciliation_id, provider, settlement_id, status, payment_id,
		       provider_payment_id, details, created_at
		FROM payment_reconciliations`+where+`
		ORDER BY created_at DESC, reconciliation_id DESC
		LIMIT ? OFFSET ?`, append(args, limit, offset)...)
	if err != nil {
		return AdminReconciliationPage{}, fmt.Errorf("list admin reconciliations: %w", err)
	}
	defer rows.Close()
	reconciliations := make([]domain.PaymentReconciliation, 0, limit)
	for rows.Next() {
		reconciliation, err := scanReconciliation(rows)
		if err != nil {
			return AdminReconciliationPage{}, fmt.Errorf("scan admin reconciliation: %w", err)
		}
		reconciliations = append(reconciliations, reconciliation)
	}
	if err := rows.Err(); err != nil {
		return AdminReconciliationPage{}, fmt.Errorf("iterate admin reconciliations: %w", err)
	}
	return AdminReconciliationPage{Reconciliations: reconciliations, Total: total}, nil
}

func ReconciliationAlertFromDomain(reconciliation domain.PaymentReconciliation) map[string]any {
	var details domain.ReconciliationDetails
	if len(reconciliation.Details) > 0 {
		_ = json.Unmarshal(reconciliation.Details, &details)
	}
	alert := map[string]any{
		"reconciliation_id": reconciliation.ReconciliationID,
		"payment_id":        reconciliation.PaymentID,
		"provider":          reconciliation.Provider,
		"status":            string(reconciliation.Status),
		"settlement_id":     reconciliation.SettlementID,
		"detected_at":       reconciliation.CreatedAt.UTC().Format(time.RFC3339),
	}
	if details.Local != nil {
		alert["local_status"] = string(details.Local.Status)
		alert["local_amount"] = map[string]any{"amount": details.Local.CapturedAmount - details.Local.RefundedAmount, "currency": details.Local.Currency}
	}
	if details.Provider != nil {
		alert["provider_status"] = details.Provider.Outcome
		alert["provider_amount"] = map[string]any{"amount": details.Provider.SettledAmount, "currency": details.Provider.Currency}
	}
	if len(details.ReasonCodes) > 0 {
		alert["note"] = strings.Join(details.ReasonCodes, ", ")
	}
	return alert
}
