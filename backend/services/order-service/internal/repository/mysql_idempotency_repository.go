package repository

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
	"github.com/go-sql-driver/mysql"
)

type MySQLIdempotencyRepository struct {
	db *sql.DB
}

func NewMySQLIdempotencyRepository(db *sql.DB) (*MySQLIdempotencyRepository, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	return &MySQLIdempotencyRepository{db: db}, nil
}

func (r *MySQLIdempotencyRepository) Claim(
	ctx context.Context,
	userID string,
	key string,
	requestHash string,
	expiresAt time.Time,
) (domain.IdempotencyClaim, error) {
	userID = strings.TrimSpace(userID)
	key = strings.TrimSpace(key)
	if err := validateClaimInput(userID, key, requestHash, expiresAt); err != nil {
		return domain.IdempotencyClaim{}, err
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO order_idempotency_keys
			(user_id, idempotency_key, request_hash, status, expires_at)
		VALUES (?, ?, ?, 'processing', ?)
	`, userID, key, requestHash, expiresAt.UTC())
	if err == nil {
		return domain.IdempotencyClaim{
			UserID:      userID,
			Key:         key,
			RequestHash: requestHash,
			Status:      domain.IdempotencyStatusProcessing,
			Decision:    domain.ClaimAcquired,
			ExpiresAt:   expiresAt.UTC(),
		}, nil
	}
	if !isDuplicateKey(err) {
		return domain.IdempotencyClaim{}, databaseUnavailable("claim checkout idempotency key", err)
	}

	var (
		stored       domain.IdempotencyClaim
		orderID      sql.NullString
		storedStatus string
	)
	err = r.db.QueryRowContext(ctx, `
		SELECT request_hash, order_id, status, expires_at
		FROM order_idempotency_keys
		WHERE user_id = ? AND idempotency_key = ?
	`, userID, key).Scan(&stored.RequestHash, &orderID, &storedStatus, &stored.ExpiresAt)
	if err != nil {
		return domain.IdempotencyClaim{}, databaseUnavailable("load checkout idempotency claim", err)
	}
	stored.UserID = userID
	stored.Key = key
	if orderID.Valid {
		stored.OrderID = orderID.String
	}
	stored.Status, err = domain.ParseIdempotencyStatus(storedStatus)
	if err != nil {
		return domain.IdempotencyClaim{}, err
	}
	if stored.RequestHash != requestHash {
		return domain.IdempotencyClaim{}, domain.ErrIdempotencyConflict
	}
	switch {
	case stored.Status == domain.IdempotencyStatusCompleted && stored.OrderID != "":
		stored.Decision = domain.ClaimReplay
	case stored.Status == domain.IdempotencyStatusProcessing && stored.OrderID != "":
		stored.Decision = domain.ClaimResume
	case stored.Status == domain.IdempotencyStatusProcessing:
		stored.Decision = domain.ClaimInProgress
	case stored.Status == domain.IdempotencyStatusFailed && stored.OrderID == "":
		stored.Decision = domain.ClaimFailed
	default:
		return domain.IdempotencyClaim{}, fmt.Errorf("%w: inconsistent idempotency claim state", domain.ErrTemporarilyUnavailable)
	}
	return stored, nil
}

func (r *MySQLIdempotencyRepository) Complete(ctx context.Context, userID string, key string, orderID string) error {
	userID = strings.TrimSpace(userID)
	key = strings.TrimSpace(key)
	orderID = strings.TrimSpace(orderID)
	if userID == "" || orderID == "" || domain.ValidateIdempotencyKey(key) != nil {
		return domain.ErrInvalidCheckoutCommand
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE order_idempotency_keys
		SET status = 'completed'
		WHERE user_id = ?
		  AND idempotency_key = ?
		  AND order_id = ?
		  AND status = 'processing'
	`, userID, key, orderID)
	if err != nil {
		return databaseUnavailable("complete checkout idempotency claim", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return databaseUnavailable("read completed idempotency claim result", err)
	}
	if affected == 1 {
		return nil
	}

	var (
		storedOrderID sql.NullString
		status        string
	)
	err = r.db.QueryRowContext(ctx, `
		SELECT order_id, status
		FROM order_idempotency_keys
		WHERE user_id = ? AND idempotency_key = ?
	`, userID, key).Scan(&storedOrderID, &status)
	if err != nil {
		return databaseUnavailable("reload completed idempotency claim", err)
	}
	if storedOrderID.Valid && storedOrderID.String == orderID && status == string(domain.IdempotencyStatusCompleted) {
		return nil
	}
	return domain.ErrCheckoutInProgress
}

func (r *MySQLIdempotencyRepository) MarkFailed(ctx context.Context, userID string, key string) error {
	userID = strings.TrimSpace(userID)
	key = strings.TrimSpace(key)
	if userID == "" || domain.ValidateIdempotencyKey(key) != nil {
		return domain.ErrInvalidCheckoutCommand
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE order_idempotency_keys
		SET status = 'failed'
		WHERE user_id = ?
		  AND idempotency_key = ?
		  AND status = 'processing'
		  AND order_id IS NULL
	`, userID, key)
	if err != nil {
		return databaseUnavailable("fail checkout idempotency claim", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return databaseUnavailable("read failed idempotency claim result", err)
	}
	if affected == 1 {
		return nil
	}
	var (
		orderID sql.NullString
		status  string
	)
	err = r.db.QueryRowContext(ctx, `
		SELECT order_id, status
		FROM order_idempotency_keys
		WHERE user_id = ? AND idempotency_key = ?
	`, userID, key).Scan(&orderID, &status)
	if err != nil {
		return databaseUnavailable("reload failed idempotency claim", err)
	}
	if !orderID.Valid && status == string(domain.IdempotencyStatusFailed) {
		return nil
	}
	return domain.ErrCheckoutInProgress
}

func validateClaimInput(userID string, key string, requestHash string, expiresAt time.Time) error {
	if userID == "" || len(userID) > 64 || expiresAt.IsZero() {
		return domain.ErrInvalidCheckoutCommand
	}
	if err := domain.ValidateIdempotencyKey(key); err != nil {
		return err
	}
	decoded, err := hex.DecodeString(requestHash)
	if err != nil || len(decoded) != 32 {
		return domain.ErrInvalidCheckoutCommand
	}
	return nil
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
