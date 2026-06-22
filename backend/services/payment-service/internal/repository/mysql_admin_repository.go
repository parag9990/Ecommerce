package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
)

type AdminPaymentPage struct {
	Payments []domain.Payment
	Total    int64
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
