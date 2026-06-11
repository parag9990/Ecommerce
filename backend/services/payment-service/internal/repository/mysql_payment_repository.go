package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
)

const mysqlDuplicateEntryCode uint16 = 1062

type MySQLPaymentRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func OpenMySQL(dsn string) (*sql.DB, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("mysql dsn is required")
	}
	return sql.Open("mysql", dsn)
}

func NewMySQLPaymentRepository(db *sql.DB, logger *slog.Logger) (*MySQLPaymentRepository, error) {
	if db == nil {
		return nil, errors.New("mysql db is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &MySQLPaymentRepository{db: db, logger: logger}, nil
}

func (r *MySQLPaymentRepository) VerifyPaymentSchema(ctx context.Context, definition domain.PaymentSchemaDefinition) (domain.PaymentSchemaVerification, error) {
	if err := ctx.Err(); err != nil {
		return domain.PaymentSchemaVerification{}, err
	}
	if err := r.db.PingContext(ctx); err != nil {
		return domain.PaymentSchemaVerification{}, fmt.Errorf("%w: %v", domain.ErrPaymentSchemaUnavailable, err)
	}

	verification := domain.PaymentSchemaVerification{DatabaseName: definition.DatabaseName}
	for _, table := range definition.Tables {
		exists, err := r.exists(ctx, `
			SELECT COUNT(1)
			FROM information_schema.tables
			WHERE table_schema = ? AND table_name = ?
		`, definition.DatabaseName, table.Name)
		if err != nil {
			return domain.PaymentSchemaVerification{}, err
		}
		if !exists {
			verification.MissingTables = append(verification.MissingTables, table.Name)
			continue
		}

		expectedIndexes := append(append([]string(nil), table.UniqueKeys...), table.Indexes...)
		for _, indexName := range expectedIndexes {
			exists, err := r.exists(ctx, `
				SELECT COUNT(1)
				FROM information_schema.statistics
				WHERE table_schema = ? AND table_name = ? AND index_name = ?
			`, definition.DatabaseName, table.Name, indexName)
			if err != nil {
				return domain.PaymentSchemaVerification{}, err
			}
			if !exists {
				verification.MissingIndexes = append(verification.MissingIndexes, table.Name+"."+indexName)
			}
		}

		for _, constraintName := range table.ForeignKeys {
			exists, err := r.exists(ctx, `
				SELECT COUNT(1)
				FROM information_schema.table_constraints
				WHERE table_schema = ?
				  AND table_name = ?
				  AND constraint_name = ?
				  AND constraint_type = 'FOREIGN KEY'
			`, definition.DatabaseName, table.Name, constraintName)
			if err != nil {
				return domain.PaymentSchemaVerification{}, err
			}
			if !exists {
				verification.MissingForeignKeys = append(verification.MissingForeignKeys, table.Name+"."+constraintName)
			}
		}
	}
	return verification.WithReadyFlag(), nil
}

func (r *MySQLPaymentRepository) CreatePayment(ctx context.Context, payment domain.Payment) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	payment = payment.Normalized()
	if err := payment.ValidateForPersistence(); err != nil {
		return err
	}
	payment.CreatedAt = timeOrNow(payment.CreatedAt)
	payment.UpdatedAt = timeOr(payment.UpdatedAt, payment.CreatedAt)
	return insertPayment(ctx, r.db, payment)
}

func (r *MySQLPaymentRepository) GetPaymentByID(ctx context.Context, paymentID string) (domain.Payment, error) {
	if err := ctx.Err(); err != nil {
		return domain.Payment{}, err
	}
	paymentID = strings.TrimSpace(paymentID)
	if paymentID == "" {
		return domain.Payment{}, fmt.Errorf("%w: payment_id is required", domain.ErrInvalidPayment)
	}

	const query = `
		SELECT
			payment_id,
			order_id,
			user_id,
			provider,
			provider_payment_id,
			provider_intent_id,
			status,
			currency,
			amount,
			captured_amount,
			refunded_amount,
			idempotency_key,
			retry_of_payment_id,
			root_payment_id,
			attempt_no,
			retry_request_key,
			failure_code,
			failure_message,
			created_at,
			updated_at
		FROM payments
		WHERE payment_id = ?
	`
	payment, err := scanPayment(r.db.QueryRowContext(ctx, query, paymentID))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Payment{}, fmt.Errorf("%w: %s", domain.ErrPaymentRecordNotFound, paymentID)
	}
	return payment, err
}

func (r *MySQLPaymentRepository) FindPaymentByProviderIdempotencyKey(ctx context.Context, providerName string, idempotencyKey string) (domain.Payment, error) {
	if err := ctx.Err(); err != nil {
		return domain.Payment{}, err
	}
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if providerName == "" {
		return domain.Payment{}, fmt.Errorf("%w: provider is required", domain.ErrInvalidPayment)
	}
	if idempotencyKey == "" {
		return domain.Payment{}, fmt.Errorf("%w: idempotency_key is required", domain.ErrInvalidPayment)
	}

	const query = `
		SELECT
			payment_id,
			order_id,
			user_id,
			provider,
			provider_payment_id,
			provider_intent_id,
			status,
			currency,
			amount,
			captured_amount,
			refunded_amount,
			idempotency_key,
			retry_of_payment_id,
			root_payment_id,
			attempt_no,
			retry_request_key,
			failure_code,
			failure_message,
			created_at,
			updated_at
		FROM payments
		WHERE provider = ? AND idempotency_key = ?
	`
	payment, err := scanPayment(r.db.QueryRowContext(ctx, query, providerName, idempotencyKey))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Payment{}, fmt.Errorf("%w: provider=%s idempotency_key=%s", domain.ErrPaymentRecordNotFound, providerName, idempotencyKey)
	}
	return payment, err
}

func (r *MySQLPaymentRepository) GetPaymentByProviderPaymentID(ctx context.Context, provider string, providerPaymentID string) (domain.Payment, error) {
	if err := ctx.Err(); err != nil {
		return domain.Payment{}, err
	}
	provider = strings.ToLower(strings.TrimSpace(provider))
	providerPaymentID = strings.TrimSpace(providerPaymentID)
	if provider == "" {
		return domain.Payment{}, fmt.Errorf("%w: provider is required", domain.ErrInvalidPayment)
	}
	if providerPaymentID == "" {
		return domain.Payment{}, fmt.Errorf("%w: provider_payment_id is required", domain.ErrInvalidPayment)
	}

	const query = `
		SELECT
			payment_id,
			order_id,
			user_id,
			provider,
			provider_payment_id,
			provider_intent_id,
			status,
			currency,
			amount,
			captured_amount,
			refunded_amount,
			idempotency_key,
			retry_of_payment_id,
			root_payment_id,
			attempt_no,
			retry_request_key,
			failure_code,
			failure_message,
			created_at,
			updated_at
		FROM payments
		WHERE provider = ? AND provider_payment_id = ?
		ORDER BY id DESC
		LIMIT 2
	`
	rows, err := r.db.QueryContext(ctx, query, provider, providerPaymentID)
	if err != nil {
		return domain.Payment{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.Payment{}, fmt.Errorf("%w: provider=%s provider_payment_id=%s", domain.ErrPaymentRecordNotFound, provider, providerPaymentID)
	}
	payment, err := scanPayment(rows)
	if err != nil {
		return domain.Payment{}, err
	}
	if rows.Next() {
		return domain.Payment{}, fmt.Errorf("%w: provider=%s provider_payment_id=%s", domain.ErrProviderPaymentAmbiguous, provider, providerPaymentID)
	}
	return payment, rows.Err()
}

func (r *MySQLPaymentRepository) CreatePaymentAttempt(ctx context.Context, attempt domain.PaymentAttempt) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	attempt = attempt.Normalized()
	if err := attempt.Validate(); err != nil {
		return err
	}
	attempt.CreatedAt = timeOrNow(attempt.CreatedAt)
	return insertPaymentAttempt(ctx, r.db, attempt)
}

func (r *MySQLPaymentRepository) CreatePaymentWithAttempt(ctx context.Context, payment domain.Payment, attempt domain.PaymentAttempt) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	payment = payment.Normalized()
	attempt = attempt.Normalized()
	if err := payment.ValidateForPersistence(); err != nil {
		return err
	}
	if err := attempt.Validate(); err != nil {
		return err
	}
	if attempt.PaymentID != payment.PaymentID {
		return fmt.Errorf("%w: attempt payment_id does not match payment", domain.ErrInvalidPayment)
	}
	payment.CreatedAt = timeOrNow(payment.CreatedAt)
	payment.UpdatedAt = timeOr(payment.UpdatedAt, payment.CreatedAt)
	attempt.CreatedAt = timeOr(attempt.CreatedAt, payment.CreatedAt)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := insertPayment(ctx, tx, payment); err != nil {
		return err
	}
	if err := insertPaymentAttempt(ctx, tx, attempt); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *MySQLPaymentRepository) UpdatePaymentIntent(ctx context.Context, payment domain.Payment, attempt domain.PaymentAttempt) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	payment = payment.Normalized()
	attempt = attempt.Normalized()
	if err := payment.ValidateForPersistence(); err != nil {
		return err
	}
	if err := attempt.Validate(); err != nil {
		return err
	}
	if attempt.PaymentID != payment.PaymentID {
		return fmt.Errorf("%w: attempt payment_id does not match payment", domain.ErrInvalidPayment)
	}
	payment.UpdatedAt = timeOrNow(payment.UpdatedAt)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, `
		UPDATE payments
		SET provider_intent_id = ?,
		    provider_payment_id = ?,
		    status = ?,
		    failure_code = NULL,
		    failure_message = NULL,
		    updated_at = ?
		WHERE payment_id = ? AND status = 'initiated'
	`,
		nullableString(payment.ProviderIntentID),
		nullableString(payment.ProviderPaymentID),
		string(payment.Status),
		payment.UpdatedAt,
		payment.PaymentID,
	)
	if err != nil {
		return err
	}
	if err := ensureUpdated(result, payment.PaymentID); err != nil {
		return err
	}

	result, err = tx.ExecContext(ctx, `
		UPDATE payment_attempts
		SET provider_attempt_id = ?,
		    status = ?,
		    failure_code = NULL,
		    failure_message = NULL,
		    raw_provider_response = ?
		WHERE attempt_id = ? AND payment_id = ? AND status = 'initiated'
	`,
		nullableString(attempt.ProviderAttemptID),
		string(attempt.Status),
		nullableBytes(attempt.RawProviderResponse),
		attempt.AttemptID,
		attempt.PaymentID,
	)
	if err != nil {
		return err
	}
	if err := ensureUpdated(result, attempt.AttemptID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *MySQLPaymentRepository) FailPaymentIntent(ctx context.Context, paymentID string, attemptID string, failureCode string, failureMessage string, at time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	paymentID = strings.TrimSpace(paymentID)
	attemptID = strings.TrimSpace(attemptID)
	failureCode = strings.TrimSpace(failureCode)
	failureMessage = strings.TrimSpace(failureMessage)
	if paymentID == "" || attemptID == "" || failureCode == "" {
		return fmt.Errorf("%w: payment_id, attempt_id, and failure_code are required", domain.ErrInvalidPayment)
	}
	if len(failureCode) > 128 || len(failureMessage) > 512 {
		return fmt.Errorf("%w: failure fields exceed schema limits", domain.ErrInvalidPayment)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, `
		UPDATE payments
		SET status = 'failed',
		    failure_code = ?,
		    failure_message = ?,
		    updated_at = ?
		WHERE payment_id = ? AND status = 'initiated'
	`, failureCode, failureMessage, timeOrNow(at), paymentID)
	if err != nil {
		return err
	}
	if err := ensureUpdated(result, paymentID); err != nil {
		return err
	}
	result, err = tx.ExecContext(ctx, `
		UPDATE payment_attempts
		SET status = 'failed',
		    failure_code = ?,
		    failure_message = ?
		WHERE attempt_id = ? AND payment_id = ? AND status = 'initiated'
	`, failureCode, failureMessage, attemptID, paymentID)
	if err != nil {
		return err
	}
	if err := ensureUpdated(result, attemptID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *MySQLPaymentRepository) ReserveRetryPayment(ctx context.Context, input domain.ReserveRetryPaymentInput) (domain.ReservedRetryPayment, error) {
	if err := ctx.Err(); err != nil {
		return domain.ReservedRetryPayment{}, err
	}
	input = input.Normalized()
	if err := input.Validate(); err != nil {
		return domain.ReservedRetryPayment{}, err
	}
	input.Now = timeOrNow(input.Now)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ReservedRetryPayment{}, err
	}
	defer func() { _ = tx.Rollback() }()

	observedParent, err := findPaymentByIDInTransaction(ctx, tx, input.FailedPaymentID)
	if err != nil {
		return domain.ReservedRetryPayment{}, err
	}
	if observedParent.UserID != input.BuyerUserID {
		return domain.ReservedRetryPayment{}, domain.ErrPaymentRecordNotFound
	}

	// Lock all existing payment rows for an order so distinct retry keys and
	// distinct failed attempts cannot allocate concurrently for the same cart.
	if err := lockPaymentsForOrder(ctx, tx, observedParent.OrderID); err != nil {
		return domain.ReservedRetryPayment{}, err
	}
	parent, err := findPaymentByIDForUpdate(ctx, tx, input.FailedPaymentID)
	if err != nil {
		return domain.ReservedRetryPayment{}, err
	}
	if parent.UserID != input.BuyerUserID {
		return domain.ReservedRetryPayment{}, domain.ErrPaymentRecordNotFound
	}
	if parent.Status != domain.PaymentStatusFailed || parent.CapturedAmount > 0 {
		return domain.ReservedRetryPayment{}, domain.ErrPaymentNotRetryable
	}

	existing, found, err := findRetryPaymentByRequestKey(ctx, tx, parent.PaymentID, input.RetryRequestKey)
	if err != nil {
		return domain.ReservedRetryPayment{}, err
	}
	if found {
		attempt, err := findRetryAttempt(ctx, tx, existing.PaymentID)
		if err != nil {
			return domain.ReservedRetryPayment{}, err
		}
		if err := tx.Commit(); err != nil {
			return domain.ReservedRetryPayment{}, err
		}
		return domain.ReservedRetryPayment{Payment: existing, Attempt: attempt, Replayed: true}, nil
	}

	summary, err := summarizeOrderPayments(ctx, tx, parent.OrderID)
	if err != nil {
		return domain.ReservedRetryPayment{}, err
	}
	if summary.Captured > 0 {
		return domain.ReservedRetryPayment{}, domain.ErrPaymentAlreadyCaptured
	}
	if summary.Active > 0 {
		return domain.ReservedRetryPayment{}, domain.ErrPaymentRetryInProgress
	}
	nextAttemptNo := summary.MaxAttemptNo + 1
	if nextAttemptNo > input.Policy.MaxAttempts {
		return domain.ReservedRetryPayment{}, domain.ErrPaymentRetryLimitReached
	}
	if input.Policy.Cooldown > 0 && !summary.LatestCreatedAt.IsZero() &&
		input.Now.Before(summary.LatestCreatedAt.Add(input.Policy.Cooldown)) {
		return domain.ReservedRetryPayment{}, domain.ErrPaymentRetryCooldown
	}

	child, err := domain.NewRetryPayment(domain.NewRetryPaymentInput{
		Parent:                 parent,
		PaymentID:              input.PaymentID,
		RootPaymentID:          parent.RetryRootPaymentID(),
		AttemptNo:              nextAttemptNo,
		RetryRequestKey:        input.RetryRequestKey,
		ProviderIdempotencyKey: provider.BuildPaymentIntentIdempotencyKey(parent.OrderID, int(nextAttemptNo)),
		Now:                    input.Now,
	})
	if err != nil {
		return domain.ReservedRetryPayment{}, err
	}
	attempt := domain.PaymentAttempt{
		AttemptID: input.AttemptID,
		PaymentID: child.PaymentID,
		Status:    domain.PaymentAttemptStatusInitiated,
		CreatedAt: input.Now,
	}
	if err := attempt.Validate(); err != nil {
		return domain.ReservedRetryPayment{}, err
	}
	if err := insertPayment(ctx, tx, child); err != nil {
		return domain.ReservedRetryPayment{}, err
	}
	if err := insertPaymentAttempt(ctx, tx, attempt); err != nil {
		return domain.ReservedRetryPayment{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.ReservedRetryPayment{}, err
	}
	return domain.ReservedRetryPayment{Payment: child, Attempt: attempt}, nil
}

func (r *MySQLPaymentRepository) CreateRefund(ctx context.Context, refund domain.Refund) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	refund = refund.Normalized()
	if err := refund.Validate(); err != nil {
		return err
	}
	refund.CreatedAt = timeOrNow(refund.CreatedAt)
	refund.UpdatedAt = timeOr(refund.UpdatedAt, refund.CreatedAt)

	const query = `
		INSERT INTO refunds (
			refund_id,
			payment_id,
			provider_refund_id,
			status,
			currency,
			amount,
			reason,
			requested_by,
			reviewed_by,
			review_reason,
			reviewed_at,
			idempotency_key,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		refund.RefundID,
		refund.PaymentID,
		nullableString(refund.ProviderRefundID),
		string(refund.Status),
		refund.Amount.Currency,
		refund.Amount.Amount,
		refund.Reason,
		refund.RequestedBy,
		nullableString(refund.ReviewedBy),
		nullableString(refund.ReviewReason),
		nullableTime(refund.ReviewedAt),
		refund.IdempotencyKey,
		refund.CreatedAt,
		refund.UpdatedAt,
	)
	if err != nil {
		return mapRefundWriteError(err)
	}
	return nil
}

func (r *MySQLPaymentRepository) GetRefundByID(ctx context.Context, refundID string) (domain.Refund, error) {
	if err := ctx.Err(); err != nil {
		return domain.Refund{}, err
	}
	refundID = strings.TrimSpace(refundID)
	if refundID == "" {
		return domain.Refund{}, fmt.Errorf("%w: refund_id is required", domain.ErrInvalidRefund)
	}
	refund, err := scanRefund(r.db.QueryRowContext(ctx, refundSelectQuery+" WHERE refund_id = ?", refundID))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Refund{}, fmt.Errorf("%w: %s", domain.ErrRefundRecordNotFound, refundID)
	}
	return refund, err
}

func (r *MySQLPaymentRepository) FindRefundByIdempotencyKey(ctx context.Context, paymentID string, idempotencyKey string) (domain.Refund, error) {
	if err := ctx.Err(); err != nil {
		return domain.Refund{}, err
	}
	paymentID = strings.TrimSpace(paymentID)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if paymentID == "" || idempotencyKey == "" {
		return domain.Refund{}, fmt.Errorf("%w: payment_id and idempotency_key are required", domain.ErrInvalidRefund)
	}
	refund, err := scanRefund(r.db.QueryRowContext(ctx, refundSelectQuery+" WHERE payment_id = ? AND idempotency_key = ?", paymentID, idempotencyKey))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Refund{}, fmt.Errorf("%w: payment_id=%s", domain.ErrRefundRecordNotFound, paymentID)
	}
	return refund, err
}

func (r *MySQLPaymentRepository) ReserveRefund(ctx context.Context, refund domain.Refund) (domain.Payment, error) {
	if err := ctx.Err(); err != nil {
		return domain.Payment{}, err
	}
	refund = refund.Normalized()
	if err := refund.Validate(); err != nil {
		return domain.Payment{}, err
	}
	refund.CreatedAt = timeOrNow(refund.CreatedAt)
	refund.UpdatedAt = timeOr(refund.UpdatedAt, refund.CreatedAt)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Payment{}, err
	}
	defer func() { _ = tx.Rollback() }()

	payment, err := findPaymentByIDForUpdate(ctx, tx, refund.PaymentID)
	if err != nil {
		return domain.Payment{}, err
	}
	var reservedOpenAmount int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM refunds
		WHERE payment_id = ? AND status IN ('requested', 'approved', 'processing')
	`, refund.PaymentID).Scan(&reservedOpenAmount); err != nil {
		return domain.Payment{}, err
	}
	if err := domain.ValidateRefundReservation(payment, refund, reservedOpenAmount); err != nil {
		return domain.Payment{}, err
	}
	if err := insertRefund(ctx, tx, refund); err != nil {
		return domain.Payment{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Payment{}, err
	}
	return payment, nil
}

func (r *MySQLPaymentRepository) ReviewRefund(
	ctx context.Context,
	refundID string,
	approved bool,
	reviewedBy string,
	reviewReason string,
	at time.Time,
) (domain.Refund, domain.Payment, error) {
	if err := ctx.Err(); err != nil {
		return domain.Refund{}, domain.Payment{}, err
	}
	refundID = strings.TrimSpace(refundID)
	reviewedBy = strings.TrimSpace(reviewedBy)
	reviewReason = strings.TrimSpace(reviewReason)
	if refundID == "" || reviewedBy == "" || reviewReason == "" || len(reviewedBy) > 64 || len(reviewReason) > 512 {
		return domain.Refund{}, domain.Payment{}, fmt.Errorf("%w: valid refund_id, reviewed_by, and review_reason are required", domain.ErrInvalidRefund)
	}
	existing, err := r.GetRefundByID(ctx, refundID)
	if err != nil {
		return domain.Refund{}, domain.Payment{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Refund{}, domain.Payment{}, err
	}
	defer func() { _ = tx.Rollback() }()

	payment, err := findPaymentByIDForUpdate(ctx, tx, existing.PaymentID)
	if err != nil {
		return domain.Refund{}, domain.Payment{}, err
	}
	refund, err := findRefundByIDForUpdate(ctx, tx, refundID)
	if err != nil {
		return domain.Refund{}, domain.Payment{}, err
	}
	if refund.Status != domain.RefundStatusRequested {
		return domain.Refund{}, domain.Payment{}, fmt.Errorf("%w: %s cannot be reviewed", domain.ErrInvalidRefundTransition, refund.Status)
	}
	status := domain.RefundStatusRejected
	if approved {
		status = domain.RefundStatusApproved
	}
	if !refund.CanTransitionTo(status) {
		return domain.Refund{}, domain.Payment{}, fmt.Errorf("%w: %s -> %s", domain.ErrInvalidRefundTransition, refund.Status, status)
	}
	refund.Status = status
	refund.ReviewedBy = reviewedBy
	refund.ReviewReason = reviewReason
	reviewedAt := timeOrNow(at)
	refund.ReviewedAt = &reviewedAt
	refund.UpdatedAt = reviewedAt
	if err := updateRefundReview(ctx, tx, refund); err != nil {
		return domain.Refund{}, domain.Payment{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Refund{}, domain.Payment{}, err
	}
	return refund, payment, nil
}

func (r *MySQLPaymentRepository) MarkRefundProcessing(ctx context.Context, refundID string, providerRefundID string, at time.Time) (domain.Refund, error) {
	if err := ctx.Err(); err != nil {
		return domain.Refund{}, err
	}
	refundID = strings.TrimSpace(refundID)
	providerRefundID = strings.TrimSpace(providerRefundID)
	if refundID == "" || providerRefundID == "" || len(providerRefundID) > 128 {
		return domain.Refund{}, fmt.Errorf("%w: valid refund_id and provider_refund_id are required", domain.ErrInvalidRefund)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Refund{}, err
	}
	defer func() { _ = tx.Rollback() }()
	refund, err := findRefundByIDForUpdate(ctx, tx, refundID)
	if err != nil {
		return domain.Refund{}, err
	}
	if refund.Status == domain.RefundStatusProcessing && refund.ProviderRefundID == providerRefundID {
		if err := tx.Commit(); err != nil {
			return domain.Refund{}, err
		}
		return refund, nil
	}
	if !refund.CanTransitionTo(domain.RefundStatusProcessing) {
		return domain.Refund{}, fmt.Errorf("%w: %s -> processing", domain.ErrInvalidRefundTransition, refund.Status)
	}
	refund.Status = domain.RefundStatusProcessing
	refund.ProviderRefundID = providerRefundID
	refund.UpdatedAt = timeOrNow(at)
	if err := updateRefundProcessing(ctx, tx, refund); err != nil {
		return domain.Refund{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Refund{}, err
	}
	return refund, nil
}

func (r *MySQLPaymentRepository) CompleteRefund(ctx context.Context, refundID string, providerRefundID string, outcome domain.RefundStatus, at time.Time) (domain.Refund, domain.Payment, error) {
	refund, err := r.GetRefundByID(ctx, refundID)
	if err != nil {
		return domain.Refund{}, domain.Payment{}, err
	}
	event := domain.VerifiedRefundWebhook{
		Provider:         "",
		RefundID:         refund.RefundID,
		ProviderRefundID: strings.TrimSpace(providerRefundID),
		Amount:           refund.Amount,
		Outcome:          outcome,
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Refund{}, domain.Payment{}, err
	}
	defer func() { _ = tx.Rollback() }()
	payment, err := findPaymentByIDForUpdate(ctx, tx, refund.PaymentID)
	if err != nil {
		return domain.Refund{}, domain.Payment{}, err
	}
	event.Provider = payment.Provider
	locked, err := findRefundByIDForUpdate(ctx, tx, refundID)
	if err != nil {
		return domain.Refund{}, domain.Payment{}, err
	}
	updatedPayment, updatedRefund, reason := domain.ApplyVerifiedRefundOutcome(payment, locked, event, at)
	if reason != "" {
		return domain.Refund{}, domain.Payment{}, fmt.Errorf("%w: %s", domain.ErrInvalidRefundTransition, reason)
	}
	if err := persistRefundOutcome(ctx, tx, updatedRefund, updatedPayment, outcome == domain.RefundStatusSucceeded); err != nil {
		return domain.Refund{}, domain.Payment{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Refund{}, domain.Payment{}, err
	}
	return updatedRefund, updatedPayment, nil
}

func (r *MySQLPaymentRepository) StoreWebhookEvent(ctx context.Context, event domain.PaymentWebhookEvent) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	event = event.Normalized()
	if err := event.Validate(); err != nil {
		return false, err
	}
	event.ReceivedAt = timeOrNow(event.ReceivedAt)

	const query = `
		INSERT INTO payment_webhook_events (
			webhook_event_id,
			provider,
			provider_event_id,
			event_type,
			processed,
			payload,
			received_at,
			processed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		event.WebhookEventID,
		event.Provider,
		event.ProviderEventID,
		event.EventType,
		event.Processed,
		event.Payload,
		event.ReceivedAt,
		nullableTime(event.ProcessedAt),
	)
	if isDuplicateEntry(err) {
		r.logger.InfoContext(ctx, "payment.webhook_event.duplicate",
			slog.String("provider", event.Provider),
			slog.String("provider_event_id", event.ProviderEventID),
		)
		return false, nil
	}
	if err != nil {
		return false, mapWriteError("payment_webhook_events", err)
	}
	return true, nil
}

func (r *MySQLPaymentRepository) ProcessWebhook(ctx context.Context, event domain.VerifiedPaymentWebhook) (domain.WebhookProcessingResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.WebhookProcessingResult{}, err
	}
	event = event.Normalized()
	event.ReceivedAt = timeOrNow(event.ReceivedAt)
	if err := event.Validate(); err != nil {
		return domain.WebhookProcessingResult{}, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.WebhookProcessingResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	storedEvent := event.StoredEvent()
	inserted, err := insertWebhookEvent(ctx, tx, storedEvent)
	if err != nil {
		return domain.WebhookProcessingResult{}, err
	}
	if !inserted {
		storedEvent, err = findWebhookEventForUpdate(ctx, tx, event.Provider, event.ProviderEventID)
		if err != nil {
			return domain.WebhookProcessingResult{}, err
		}
		if storedEvent.Processed {
			if err := tx.Commit(); err != nil {
				return domain.WebhookProcessingResult{}, err
			}
			return domain.WebhookProcessingResult{
				WebhookEventID:   storedEvent.WebhookEventID,
				ProviderEventID:  storedEvent.ProviderEventID,
				ProcessingStatus: domain.WebhookProcessingStatusDuplicate,
			}, nil
		}
		event.WebhookEventID = storedEvent.WebhookEventID
	}

	result := domain.WebhookProcessingResult{
		WebhookEventID:  event.WebhookEventID,
		ProviderEventID: event.ProviderEventID,
	}
	payment, err := findPaymentForWebhook(ctx, tx, event)
	if errors.Is(err, domain.ErrPaymentRecordNotFound) || errors.Is(err, domain.ErrWebhookPaymentAmbiguous) {
		result.ProcessingStatus = domain.WebhookProcessingStatusIgnored
		if errors.Is(err, domain.ErrWebhookPaymentAmbiguous) {
			result.IgnoreReason = "webhook lookup matched multiple payments"
		} else {
			result.IgnoreReason = "payment was not found"
		}
		if err := markWebhookProcessed(ctx, tx, event.WebhookEventID, event.ReceivedAt); err != nil {
			return domain.WebhookProcessingResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return domain.WebhookProcessingResult{}, err
		}
		r.logger.WarnContext(ctx, "payment.webhook.ignored",
			slog.String("provider", event.Provider),
			slog.String("provider_event_id", event.ProviderEventID),
			slog.String("reason", result.IgnoreReason),
		)
		return result, nil
	}
	if err != nil {
		return domain.WebhookProcessingResult{}, err
	}

	result.Payment = payment
	result.PreviousStatus = payment.Status
	updated, decision, ignoreReason := domain.ApplyVerifiedPaymentWebhook(payment, event, event.ReceivedAt)
	if ignoreReason != "" {
		result.ProcessingStatus = domain.WebhookProcessingStatusIgnored
		result.IgnoreReason = ignoreReason
		if err := markWebhookProcessed(ctx, tx, event.WebhookEventID, event.ReceivedAt); err != nil {
			return domain.WebhookProcessingResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return domain.WebhookProcessingResult{}, err
		}
		r.logger.WarnContext(ctx, "payment.webhook.ignored",
			slog.String("provider", event.Provider),
			slog.String("provider_event_id", event.ProviderEventID),
			slog.String("payment_id", payment.PaymentID),
			slog.String("reason", ignoreReason),
		)
		return result, nil
	}

	result.Payment = updated
	result.StatusChanged = !decision.Noop
	result.ProcessingStatus = domain.WebhookProcessingStatusProcessed
	result.LateCapture = result.PreviousStatus == domain.PaymentStatusFailed &&
		updated.Status == domain.PaymentStatusCaptured
	if result.StatusChanged {
		if err := updatePaymentForWebhook(ctx, tx, updated); err != nil {
			return domain.WebhookProcessingResult{}, err
		}
		if err := updateAttemptForFinalWebhook(ctx, tx, updated); err != nil {
			return domain.WebhookProcessingResult{}, err
		}
		if updated.Status == domain.PaymentStatusCaptured {
			result.DuplicateCapture, err = hasCapturedSibling(ctx, tx, updated)
			if err != nil {
				return domain.WebhookProcessingResult{}, err
			}
		}
	}
	if err := markWebhookProcessed(ctx, tx, event.WebhookEventID, event.ReceivedAt); err != nil {
		return domain.WebhookProcessingResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.WebhookProcessingResult{}, err
	}
	if result.LateCapture || result.DuplicateCapture {
		r.logger.ErrorContext(ctx, "payment.duplicate_capture_risk_detected",
			slog.String("payment_id", updated.PaymentID),
			slog.String("order_id", updated.OrderID),
			slog.String("root_payment_id", updated.RetryRootPaymentID()),
			slog.Bool("late_capture", result.LateCapture),
			slog.Bool("duplicate_capture", result.DuplicateCapture),
		)
	}
	return result, nil
}

func (r *MySQLPaymentRepository) ProcessRefundWebhook(ctx context.Context, event domain.VerifiedRefundWebhook) (domain.RefundWebhookProcessingResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.RefundWebhookProcessingResult{}, err
	}
	event = event.Normalized()
	event.ReceivedAt = timeOrNow(event.ReceivedAt)
	if err := event.Validate(); err != nil {
		return domain.RefundWebhookProcessingResult{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.RefundWebhookProcessingResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	inserted, err := insertWebhookEvent(ctx, tx, event.StoredEvent())
	if err != nil {
		return domain.RefundWebhookProcessingResult{}, err
	}
	if !inserted {
		stored, err := findWebhookEventForUpdate(ctx, tx, event.Provider, event.ProviderEventID)
		if err != nil {
			return domain.RefundWebhookProcessingResult{}, err
		}
		if stored.Processed {
			if err := tx.Commit(); err != nil {
				return domain.RefundWebhookProcessingResult{}, err
			}
			return domain.RefundWebhookProcessingResult{
				WebhookEventID:   stored.WebhookEventID,
				ProviderEventID:  stored.ProviderEventID,
				ProcessingStatus: domain.WebhookProcessingStatusDuplicate,
			}, nil
		}
		event.WebhookEventID = stored.WebhookEventID
	}

	result := domain.RefundWebhookProcessingResult{
		WebhookEventID:  event.WebhookEventID,
		ProviderEventID: event.ProviderEventID,
	}
	candidate, err := findRefundForWebhook(ctx, tx, event)
	if errors.Is(err, domain.ErrRefundRecordNotFound) {
		result.ProcessingStatus = domain.WebhookProcessingStatusIgnored
		result.IgnoreReason = "refund was not found"
		if err := markWebhookProcessed(ctx, tx, event.WebhookEventID, event.ReceivedAt); err != nil {
			return domain.RefundWebhookProcessingResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return domain.RefundWebhookProcessingResult{}, err
		}
		r.logger.WarnContext(ctx, "payment.refund_webhook.ignored",
			slog.String("provider", event.Provider),
			slog.String("provider_event_id", event.ProviderEventID),
			slog.String("reason", result.IgnoreReason),
		)
		return result, nil
	}
	if err != nil {
		return domain.RefundWebhookProcessingResult{}, err
	}
	payment, err := findPaymentByIDForUpdate(ctx, tx, candidate.PaymentID)
	if err != nil {
		return domain.RefundWebhookProcessingResult{}, err
	}
	refund, err := findRefundByIDForUpdate(ctx, tx, candidate.RefundID)
	if err != nil {
		return domain.RefundWebhookProcessingResult{}, err
	}
	result.Refund = refund
	result.Payment = payment
	result.PreviousStatus = refund.Status
	updatedPayment, updatedRefund, ignoreReason := domain.ApplyVerifiedRefundOutcome(payment, refund, event, event.ReceivedAt)
	if ignoreReason != "" {
		result.ProcessingStatus = domain.WebhookProcessingStatusIgnored
		result.IgnoreReason = ignoreReason
		if err := markWebhookProcessed(ctx, tx, event.WebhookEventID, event.ReceivedAt); err != nil {
			return domain.RefundWebhookProcessingResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return domain.RefundWebhookProcessingResult{}, err
		}
		r.logger.WarnContext(ctx, "payment.refund_webhook.ignored",
			slog.String("provider", event.Provider),
			slog.String("provider_event_id", event.ProviderEventID),
			slog.String("refund_id", refund.RefundID),
			slog.String("reason", ignoreReason),
		)
		return result, nil
	}
	result.Payment = updatedPayment
	result.Refund = updatedRefund
	result.StatusChanged = true
	result.ProcessingStatus = domain.WebhookProcessingStatusProcessed
	if err := persistRefundOutcome(ctx, tx, updatedRefund, updatedPayment, event.Outcome == domain.RefundStatusSucceeded); err != nil {
		return domain.RefundWebhookProcessingResult{}, err
	}
	if err := markWebhookProcessed(ctx, tx, event.WebhookEventID, event.ReceivedAt); err != nil {
		return domain.RefundWebhookProcessingResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.RefundWebhookProcessingResult{}, err
	}
	return result, nil
}

func (r *MySQLPaymentRepository) CreateReconciliation(ctx context.Context, reconciliation domain.PaymentReconciliation) error {
	created, err := r.createReconciliation(ctx, reconciliation)
	if err != nil {
		return err
	}
	if !created {
		return fmt.Errorf("%w: payment_reconciliations", domain.ErrDuplicatePaymentRecord)
	}
	return nil
}

func (r *MySQLPaymentRepository) CreateReconciliationIfAbsent(ctx context.Context, reconciliation domain.PaymentReconciliation) (bool, error) {
	return r.createReconciliation(ctx, reconciliation)
}

func (r *MySQLPaymentRepository) createReconciliation(ctx context.Context, reconciliation domain.PaymentReconciliation) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	reconciliation = reconciliation.Normalized()
	if err := reconciliation.Validate(); err != nil {
		return false, err
	}
	reconciliation.CreatedAt = timeOrNow(reconciliation.CreatedAt)

	const query = `
		INSERT INTO payment_reconciliations (
			reconciliation_id,
			provider,
			settlement_id,
			status,
			payment_id,
			provider_payment_id,
			details,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		reconciliation.ReconciliationID,
		reconciliation.Provider,
		nullableString(reconciliation.SettlementID),
		string(reconciliation.Status),
		nullableString(reconciliation.PaymentID),
		nullableString(reconciliation.ProviderPaymentID),
		nullableBytes(reconciliation.Details),
		reconciliation.CreatedAt,
	)
	if err != nil {
		if isDuplicateEntry(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *MySQLPaymentRepository) GetReconciliationByID(ctx context.Context, reconciliationID string) (domain.PaymentReconciliation, error) {
	if err := ctx.Err(); err != nil {
		return domain.PaymentReconciliation{}, err
	}
	reconciliationID = strings.TrimSpace(reconciliationID)
	if reconciliationID == "" {
		return domain.PaymentReconciliation{}, fmt.Errorf("%w: reconciliation_id is required", domain.ErrInvalidReconciliation)
	}
	reconciliation, err := scanReconciliation(r.db.QueryRowContext(ctx, `
		SELECT reconciliation_id, provider, settlement_id, status, payment_id,
		       provider_payment_id, details, created_at
		FROM payment_reconciliations
		WHERE reconciliation_id = ?
	`, reconciliationID))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PaymentReconciliation{}, fmt.Errorf("%w: %s", domain.ErrReconciliationNotFound, reconciliationID)
	}
	return reconciliation, err
}

func (r *MySQLPaymentRepository) ListSettlementCandidates(
	ctx context.Context,
	providerName string,
	from time.Time,
	to time.Time,
	afterPaymentID string,
	limit int,
) ([]domain.Payment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	afterPaymentID = strings.TrimSpace(afterPaymentID)
	if providerName == "" || from.IsZero() || to.IsZero() || !from.Before(to) {
		return nil, fmt.Errorf("%w: provider and a valid settlement window are required", domain.ErrInvalidReconciliation)
	}
	if limit <= 0 || limit > 10000 {
		return nil, fmt.Errorf("%w: candidate query limit must be between 1 and 10000", domain.ErrInvalidReconciliation)
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			payment_id, order_id, user_id, provider, provider_payment_id, provider_intent_id,
			status, currency, amount, captured_amount, refunded_amount, idempotency_key,
			retry_of_payment_id, root_payment_id, attempt_no, retry_request_key,
			failure_code, failure_message, created_at, updated_at
		FROM payments
		WHERE provider = ?
		  AND status IN ('captured', 'partially_refunded', 'refunded')
		  AND provider_payment_id IS NOT NULL
		  AND provider_payment_id <> ''
		  AND updated_at >= ?
		  AND updated_at < ?
		  AND payment_id > ?
		ORDER BY payment_id
		LIMIT ?
	`, providerName, from.UTC(), to.UTC(), afterPaymentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	candidates := make([]domain.Payment, 0, limit)
	for rows.Next() {
		payment, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, payment)
	}
	return candidates, rows.Err()
}

func (r *MySQLPaymentRepository) exists(ctx context.Context, query string, args ...any) (bool, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

type contextExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

const refundSelectQuery = `
	SELECT
		refund_id, payment_id, provider_refund_id, status, currency, amount, reason,
		requested_by, reviewed_by, review_reason, reviewed_at, idempotency_key,
		created_at, updated_at
	FROM refunds
`

func insertPayment(ctx context.Context, executor contextExecutor, payment domain.Payment) error {
	const query = `
		INSERT INTO payments (
			payment_id,
			order_id,
			user_id,
			provider,
			provider_payment_id,
			provider_intent_id,
			status,
			currency,
			amount,
			captured_amount,
			refunded_amount,
			idempotency_key,
			retry_of_payment_id,
			root_payment_id,
			attempt_no,
			retry_request_key,
			failure_code,
			failure_message,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := executor.ExecContext(ctx, query,
		payment.PaymentID,
		payment.OrderID,
		payment.UserID,
		payment.Provider,
		nullableString(payment.ProviderPaymentID),
		nullableString(payment.ProviderIntentID),
		string(payment.Status),
		payment.Amount.Currency,
		payment.Amount.Amount,
		payment.CapturedAmount,
		payment.RefundedAmount,
		payment.IdempotencyKey,
		nullableString(payment.RetryOfPaymentID),
		nullableString(payment.RootPaymentID),
		payment.AttemptNo,
		nullableString(payment.RetryRequestKey),
		nullableString(payment.FailureCode),
		nullableString(payment.FailureMessage),
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	return mapWriteError("payments", err)
}

func insertPaymentAttempt(ctx context.Context, executor contextExecutor, attempt domain.PaymentAttempt) error {
	const query = `
		INSERT INTO payment_attempts (
			attempt_id,
			payment_id,
			provider_attempt_id,
			status,
			failure_code,
			failure_message,
			raw_provider_response,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := executor.ExecContext(ctx, query,
		attempt.AttemptID,
		attempt.PaymentID,
		nullableString(attempt.ProviderAttemptID),
		string(attempt.Status),
		nullableString(attempt.FailureCode),
		nullableString(attempt.FailureMessage),
		nullableBytes(attempt.RawProviderResponse),
		attempt.CreatedAt,
	)
	return mapWriteError("payment_attempts", err)
}

func insertRefund(ctx context.Context, executor contextExecutor, refund domain.Refund) error {
	_, err := executor.ExecContext(ctx, `
		INSERT INTO refunds (
			refund_id, payment_id, provider_refund_id, status, currency, amount, reason,
			requested_by, reviewed_by, review_reason, reviewed_at, idempotency_key,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		refund.RefundID,
		refund.PaymentID,
		nullableString(refund.ProviderRefundID),
		string(refund.Status),
		refund.Amount.Currency,
		refund.Amount.Amount,
		refund.Reason,
		refund.RequestedBy,
		nullableString(refund.ReviewedBy),
		nullableString(refund.ReviewReason),
		nullableTime(refund.ReviewedAt),
		refund.IdempotencyKey,
		refund.CreatedAt,
		refund.UpdatedAt,
	)
	return mapRefundWriteError(err)
}

func insertWebhookEvent(ctx context.Context, executor contextExecutor, event domain.PaymentWebhookEvent) (bool, error) {
	const query = `
		INSERT INTO payment_webhook_events (
			webhook_event_id,
			provider,
			provider_event_id,
			event_type,
			processed,
			payload,
			received_at,
			processed_at
		) VALUES (?, ?, ?, ?, FALSE, ?, ?, NULL)
	`
	_, err := executor.ExecContext(ctx, query,
		event.WebhookEventID,
		event.Provider,
		event.ProviderEventID,
		event.EventType,
		event.Payload,
		event.ReceivedAt,
	)
	if isDuplicateEntry(err) {
		return false, nil
	}
	if err != nil {
		return false, mapWriteError("payment_webhook_events", err)
	}
	return true, nil
}

func findWebhookEventForUpdate(ctx context.Context, tx *sql.Tx, providerName string, providerEventID string) (domain.PaymentWebhookEvent, error) {
	var event domain.PaymentWebhookEvent
	var processedAt sql.NullTime
	err := tx.QueryRowContext(ctx, `
		SELECT webhook_event_id, provider, provider_event_id, event_type, processed, payload, received_at, processed_at
		FROM payment_webhook_events
		WHERE provider = ? AND provider_event_id = ?
		FOR UPDATE
	`, providerName, providerEventID).Scan(
		&event.WebhookEventID,
		&event.Provider,
		&event.ProviderEventID,
		&event.EventType,
		&event.Processed,
		&event.Payload,
		&event.ReceivedAt,
		&processedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PaymentWebhookEvent{}, fmt.Errorf("%w: provider=%s event=%s", domain.ErrPaymentRecordNotFound, providerName, providerEventID)
	}
	if err != nil {
		return domain.PaymentWebhookEvent{}, err
	}
	if processedAt.Valid {
		processed := processedAt.Time.UTC()
		event.ProcessedAt = &processed
	}
	return event.Normalized(), nil
}

func findPaymentByIDForUpdate(ctx context.Context, tx *sql.Tx, paymentID string) (domain.Payment, error) {
	payment, err := scanPayment(tx.QueryRowContext(ctx, `
		SELECT
			payment_id, order_id, user_id, provider, provider_payment_id, provider_intent_id,
			status, currency, amount, captured_amount, refunded_amount, idempotency_key,
			retry_of_payment_id, root_payment_id, attempt_no, retry_request_key,
			failure_code, failure_message, created_at, updated_at
		FROM payments
		WHERE payment_id = ?
		FOR UPDATE
	`, paymentID))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Payment{}, fmt.Errorf("%w: %s", domain.ErrPaymentRecordNotFound, paymentID)
	}
	return payment, err
}

func findPaymentByIDInTransaction(ctx context.Context, tx *sql.Tx, paymentID string) (domain.Payment, error) {
	payment, err := scanPayment(tx.QueryRowContext(ctx, `
		SELECT
			payment_id, order_id, user_id, provider, provider_payment_id, provider_intent_id,
			status, currency, amount, captured_amount, refunded_amount, idempotency_key,
			retry_of_payment_id, root_payment_id, attempt_no, retry_request_key,
			failure_code, failure_message, created_at, updated_at
		FROM payments
		WHERE payment_id = ?
	`, paymentID))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Payment{}, domain.ErrPaymentRecordNotFound
	}
	return payment, err
}

func lockPaymentsForOrder(ctx context.Context, tx *sql.Tx, orderID string) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT payment_id
		FROM payments
		WHERE order_id = ?
		ORDER BY id
		FOR UPDATE
	`, orderID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var paymentID string
	for rows.Next() {
		if err := rows.Scan(&paymentID); err != nil {
			return err
		}
	}
	return rows.Err()
}

func findRetryPaymentByRequestKey(ctx context.Context, tx *sql.Tx, parentPaymentID string, requestKey string) (domain.Payment, bool, error) {
	payment, err := scanPayment(tx.QueryRowContext(ctx, `
		SELECT
			payment_id, order_id, user_id, provider, provider_payment_id, provider_intent_id,
			status, currency, amount, captured_amount, refunded_amount, idempotency_key,
			retry_of_payment_id, root_payment_id, attempt_no, retry_request_key,
			failure_code, failure_message, created_at, updated_at
		FROM payments
		WHERE retry_of_payment_id = ? AND retry_request_key = ?
	`, parentPaymentID, requestKey))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Payment{}, false, nil
	}
	return payment, err == nil, err
}

func findRetryAttempt(ctx context.Context, tx *sql.Tx, paymentID string) (domain.PaymentAttempt, error) {
	var attempt domain.PaymentAttempt
	var providerAttemptID sql.NullString
	var failureCode sql.NullString
	var failureMessage sql.NullString
	var rawProviderResponse []byte
	var status string
	err := tx.QueryRowContext(ctx, `
		SELECT attempt_id, payment_id, provider_attempt_id, status, failure_code,
		       failure_message, raw_provider_response, created_at
		FROM payment_attempts
		WHERE payment_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, paymentID).Scan(
		&attempt.AttemptID,
		&attempt.PaymentID,
		&providerAttemptID,
		&status,
		&failureCode,
		&failureMessage,
		&rawProviderResponse,
		&attempt.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PaymentAttempt{}, domain.ErrPaymentRecordNotFound
	}
	if err != nil {
		return domain.PaymentAttempt{}, err
	}
	attempt.ProviderAttemptID = providerAttemptID.String
	attempt.Status = domain.PaymentAttemptStatus(status)
	attempt.FailureCode = failureCode.String
	attempt.FailureMessage = failureMessage.String
	attempt.RawProviderResponse = append(attempt.RawProviderResponse, rawProviderResponse...)
	return attempt.Normalized(), nil
}

type orderPaymentSummary struct {
	MaxAttemptNo    uint32
	Active          int
	Captured        int
	LatestCreatedAt time.Time
}

func summarizeOrderPayments(ctx context.Context, tx *sql.Tx, orderID string) (orderPaymentSummary, error) {
	var summary orderPaymentSummary
	var latest sql.NullTime
	err := tx.QueryRowContext(ctx, `
		SELECT
			COALESCE(MAX(attempt_no), 1),
			MAX(created_at),
			COALESCE(SUM(status IN ('initiated', 'requires_action', 'authorized')), 0),
			COALESCE(SUM(status IN ('captured', 'partially_refunded', 'refunded')), 0)
		FROM payments
		WHERE order_id = ?
	`, orderID).Scan(&summary.MaxAttemptNo, &latest, &summary.Active, &summary.Captured)
	if err != nil {
		return orderPaymentSummary{}, err
	}
	if latest.Valid {
		summary.LatestCreatedAt = latest.Time.UTC()
	}
	return summary, nil
}

func findRefundByIDForUpdate(ctx context.Context, tx *sql.Tx, refundID string) (domain.Refund, error) {
	refund, err := scanRefund(tx.QueryRowContext(ctx, refundSelectQuery+" WHERE refund_id = ? FOR UPDATE", refundID))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Refund{}, fmt.Errorf("%w: %s", domain.ErrRefundRecordNotFound, refundID)
	}
	return refund, err
}

func findRefundForWebhook(ctx context.Context, tx *sql.Tx, event domain.VerifiedRefundWebhook) (domain.Refund, error) {
	if event.RefundID != "" {
		refund, err := scanRefund(tx.QueryRowContext(ctx, refundSelectQuery+" WHERE refund_id = ?", event.RefundID))
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Refund{}, domain.ErrRefundRecordNotFound
		}
		return refund, err
	}
	rows, err := tx.QueryContext(ctx, refundSelectQuery+`
		WHERE provider_refund_id = ?
		  AND payment_id IN (SELECT payment_id FROM payments WHERE provider = ?)
		ORDER BY id DESC
		LIMIT 2
	`, event.ProviderRefundID, event.Provider)
	if err != nil {
		return domain.Refund{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.Refund{}, domain.ErrRefundRecordNotFound
	}
	refund, err := scanRefund(rows)
	if err != nil {
		return domain.Refund{}, err
	}
	if rows.Next() {
		return domain.Refund{}, fmt.Errorf("%w: provider refund id is ambiguous", domain.ErrInvalidWebhookEvent)
	}
	return refund, rows.Err()
}

func findPaymentForWebhook(ctx context.Context, tx *sql.Tx, event domain.VerifiedPaymentWebhook) (domain.Payment, error) {
	const fields = `
		SELECT
			payment_id, order_id, user_id, provider, provider_payment_id, provider_intent_id,
			status, currency, amount, captured_amount, refunded_amount, idempotency_key,
			retry_of_payment_id, root_payment_id, attempt_no, retry_request_key,
			failure_code, failure_message, created_at, updated_at
		FROM payments
	`
	var row *sql.Row
	switch {
	case event.PaymentID != "":
		row = tx.QueryRowContext(ctx, fields+" WHERE payment_id = ? FOR UPDATE", event.PaymentID)
	case event.ProviderIntentID != "":
		return findSingleWebhookPaymentForUpdate(ctx, tx, fields+" WHERE provider = ? AND provider_intent_id = ? ORDER BY id DESC LIMIT 2 FOR UPDATE", event.Provider, event.ProviderIntentID)
	case event.ProviderPaymentID != "":
		return findSingleWebhookPaymentForUpdate(ctx, tx, fields+" WHERE provider = ? AND provider_payment_id = ? ORDER BY id DESC LIMIT 2 FOR UPDATE", event.Provider, event.ProviderPaymentID)
	default:
		return findSingleWebhookPaymentForUpdate(ctx, tx, fields+" WHERE provider = ? AND order_id = ? ORDER BY id DESC LIMIT 2 FOR UPDATE", event.Provider, event.OrderID)
	}
	payment, err := scanPayment(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Payment{}, domain.ErrPaymentRecordNotFound
	}
	return payment, err
}

func findSingleWebhookPaymentForUpdate(ctx context.Context, tx *sql.Tx, query string, args ...any) (domain.Payment, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return domain.Payment{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.Payment{}, domain.ErrPaymentRecordNotFound
	}
	payment, err := scanPayment(rows)
	if err != nil {
		return domain.Payment{}, err
	}
	if rows.Next() {
		return domain.Payment{}, domain.ErrWebhookPaymentAmbiguous
	}
	return payment, rows.Err()
}

func updatePaymentForWebhook(ctx context.Context, tx *sql.Tx, payment domain.Payment) error {
	if err := payment.ValidateForPersistence(); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE payments
		SET provider_intent_id = ?,
		    provider_payment_id = ?,
		    status = ?,
		    captured_amount = ?,
		    failure_code = ?,
		    failure_message = ?,
		    updated_at = ?
		WHERE payment_id = ?
	`,
		nullableString(payment.ProviderIntentID),
		nullableString(payment.ProviderPaymentID),
		string(payment.Status),
		payment.CapturedAmount,
		nullableString(payment.FailureCode),
		nullableString(payment.FailureMessage),
		payment.UpdatedAt,
		payment.PaymentID,
	)
	if err != nil {
		return err
	}
	return ensureUpdated(result, payment.PaymentID)
}

func updateAttemptForFinalWebhook(ctx context.Context, tx *sql.Tx, payment domain.Payment) error {
	var status domain.PaymentAttemptStatus
	switch payment.Status {
	case domain.PaymentStatusCaptured:
		status = domain.PaymentAttemptStatusSucceeded
	case domain.PaymentStatusFailed:
		status = domain.PaymentAttemptStatusFailed
	default:
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		UPDATE payment_attempts
		SET status = ?,
		    failure_code = ?,
		    failure_message = ?
		WHERE payment_id = ?
		ORDER BY id DESC
		LIMIT 1
	`,
		string(status),
		nullableString(payment.FailureCode),
		nullableString(payment.FailureMessage),
		payment.PaymentID,
	)
	return err
}

func hasCapturedSibling(ctx context.Context, tx *sql.Tx, payment domain.Payment) (bool, error) {
	var count int
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM payments
		WHERE order_id = ?
		  AND payment_id <> ?
		  AND status IN ('captured', 'partially_refunded', 'refunded')
	`, payment.OrderID, payment.PaymentID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func updateRefundReview(ctx context.Context, tx *sql.Tx, refund domain.Refund) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE refunds
		SET status = ?, reviewed_by = ?, review_reason = ?, reviewed_at = ?, updated_at = ?
		WHERE refund_id = ? AND status = 'requested'
	`, string(refund.Status), refund.ReviewedBy, refund.ReviewReason, nullableTime(refund.ReviewedAt), refund.UpdatedAt, refund.RefundID)
	if err != nil {
		return err
	}
	return ensureUpdated(result, refund.RefundID)
}

func updateRefundProcessing(ctx context.Context, tx *sql.Tx, refund domain.Refund) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE refunds
		SET status = 'processing', provider_refund_id = ?, updated_at = ?
		WHERE refund_id = ? AND status = 'approved'
	`, refund.ProviderRefundID, refund.UpdatedAt, refund.RefundID)
	if err != nil {
		return err
	}
	return ensureUpdated(result, refund.RefundID)
}

func persistRefundOutcome(ctx context.Context, tx *sql.Tx, refund domain.Refund, payment domain.Payment, succeeded bool) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE refunds
		SET status = ?, provider_refund_id = ?, updated_at = ?
		WHERE refund_id = ? AND status IN ('approved', 'processing')
	`, string(refund.Status), nullableString(refund.ProviderRefundID), refund.UpdatedAt, refund.RefundID)
	if err != nil {
		return err
	}
	if err := ensureUpdated(result, refund.RefundID); err != nil {
		return err
	}
	if !succeeded {
		return nil
	}
	result, err = tx.ExecContext(ctx, `
		UPDATE payments
		SET refunded_amount = ?, status = ?, updated_at = ?
		WHERE payment_id = ? AND refunded_amount <= captured_amount AND ? <= captured_amount
	`, payment.RefundedAmount, string(payment.Status), payment.UpdatedAt, payment.PaymentID, payment.RefundedAmount)
	if err != nil {
		return err
	}
	return ensureUpdated(result, payment.PaymentID)
}

func markWebhookProcessed(ctx context.Context, tx *sql.Tx, webhookEventID string, processedAt time.Time) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE payment_webhook_events
		SET processed = TRUE, processed_at = ?
		WHERE webhook_event_id = ?
	`, timeOrNow(processedAt), webhookEventID)
	if err != nil {
		return err
	}
	return ensureUpdated(result, webhookEventID)
}

func ensureUpdated(result sql.Result, identifier string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("%w: %s", domain.ErrPaymentRecordNotFound, identifier)
	}
	return nil
}

func scanPayment(row rowScanner) (domain.Payment, error) {
	var providerPaymentID sql.NullString
	var providerIntentID sql.NullString
	var retryOfPaymentID sql.NullString
	var rootPaymentID sql.NullString
	var retryRequestKey sql.NullString
	var failureCode sql.NullString
	var failureMessage sql.NullString
	var status string
	var currency string
	var payment domain.Payment

	err := row.Scan(
		&payment.PaymentID,
		&payment.OrderID,
		&payment.UserID,
		&payment.Provider,
		&providerPaymentID,
		&providerIntentID,
		&status,
		&currency,
		&payment.Amount.Amount,
		&payment.CapturedAmount,
		&payment.RefundedAmount,
		&payment.IdempotencyKey,
		&retryOfPaymentID,
		&rootPaymentID,
		&payment.AttemptNo,
		&retryRequestKey,
		&failureCode,
		&failureMessage,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		return domain.Payment{}, err
	}
	payment.ProviderPaymentID = providerPaymentID.String
	payment.ProviderIntentID = providerIntentID.String
	payment.RetryOfPaymentID = retryOfPaymentID.String
	payment.RootPaymentID = rootPaymentID.String
	payment.RetryRequestKey = retryRequestKey.String
	payment.Status = domain.PaymentStatus(status)
	payment.Amount.Currency = currency
	payment.FailureCode = failureCode.String
	payment.FailureMessage = failureMessage.String
	return payment.Normalized(), nil
}

func scanRefund(row rowScanner) (domain.Refund, error) {
	var providerRefundID sql.NullString
	var reviewedBy sql.NullString
	var reviewReason sql.NullString
	var reviewedAt sql.NullTime
	var status string
	var currency string
	var refund domain.Refund
	err := row.Scan(
		&refund.RefundID,
		&refund.PaymentID,
		&providerRefundID,
		&status,
		&currency,
		&refund.Amount.Amount,
		&refund.Reason,
		&refund.RequestedBy,
		&reviewedBy,
		&reviewReason,
		&reviewedAt,
		&refund.IdempotencyKey,
		&refund.CreatedAt,
		&refund.UpdatedAt,
	)
	if err != nil {
		return domain.Refund{}, err
	}
	refund.ProviderRefundID = providerRefundID.String
	refund.Status = domain.RefundStatus(status)
	refund.Amount.Currency = currency
	refund.ReviewedBy = reviewedBy.String
	refund.ReviewReason = reviewReason.String
	if reviewedAt.Valid {
		value := reviewedAt.Time.UTC()
		refund.ReviewedAt = &value
	}
	return refund.Normalized(), nil
}

func scanReconciliation(row rowScanner) (domain.PaymentReconciliation, error) {
	var settlementID sql.NullString
	var paymentID sql.NullString
	var providerPaymentID sql.NullString
	var details []byte
	var status string
	var reconciliation domain.PaymentReconciliation
	err := row.Scan(
		&reconciliation.ReconciliationID,
		&reconciliation.Provider,
		&settlementID,
		&status,
		&paymentID,
		&providerPaymentID,
		&details,
		&reconciliation.CreatedAt,
	)
	if err != nil {
		return domain.PaymentReconciliation{}, err
	}
	reconciliation.SettlementID = settlementID.String
	reconciliation.Status = domain.ReconciliationStatus(status)
	reconciliation.PaymentID = paymentID.String
	reconciliation.ProviderPaymentID = providerPaymentID.String
	reconciliation.Details = append(reconciliation.Details, details...)
	return reconciliation.Normalized(), nil
}

func nullableString(value string) sql.NullString {
	value = strings.TrimSpace(value)
	return sql.NullString{String: value, Valid: value != ""}
}

func nullableBytes(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return value
}

func nullableTime(value *time.Time) sql.NullTime {
	if value == nil || value.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: value.UTC(), Valid: true}
}

func timeOrNow(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}

func timeOr(value time.Time, fallback time.Time) time.Time {
	if value.IsZero() {
		return fallback.UTC()
	}
	return value.UTC()
}

func mapWriteError(table string, err error) error {
	if err == nil {
		return nil
	}
	if isDuplicateEntry(err) {
		return fmt.Errorf("%w: %s", domain.ErrDuplicatePaymentRecord, table)
	}
	return err
}

func mapRefundWriteError(err error) error {
	if err == nil {
		return nil
	}
	if isDuplicateEntry(err) {
		return fmt.Errorf("%w: refunds", domain.ErrDuplicateRefundRecord)
	}
	return err
}

func isDuplicateEntry(err error) bool {
	if err == nil {
		return false
	}
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicateEntryCode
}
