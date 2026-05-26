package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
	"github.com/go-sql-driver/mysql"
)

func TestMySQLIdempotencyRepositoryClaimAcquiresNewKey(t *testing.T) {
	db, mock := newMockDB(t)
	repository, _ := NewMySQLIdempotencyRepository(db)
	expiresAt := time.Date(2026, time.May, 27, 10, 0, 0, 0, time.UTC)
	hash := strings.Repeat("a", 64)
	mock.ExpectExec("INSERT INTO order_idempotency_keys").
		WithArgs("user_1", "key_1", hash, expiresAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	claim, err := repository.Claim(context.Background(), "user_1", "key_1", hash, expiresAt)
	if err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	if claim.Decision != domain.ClaimAcquired || claim.Status != domain.IdempotencyStatusProcessing {
		t.Fatalf("claim = %+v, want acquired processing claim", claim)
	}
	assertMockExpectations(t, mock)
}

func TestMySQLIdempotencyRepositoryDuplicateKeyReplaysOrConflicts(t *testing.T) {
	expiresAt := time.Date(2026, time.May, 27, 10, 0, 0, 0, time.UTC)
	hash := strings.Repeat("a", 64)
	tests := []struct {
		name       string
		storedHash string
		orderID    any
		status     string
		want       domain.ClaimDecision
		wantErr    error
	}{
		{name: "completed request", storedHash: hash, orderID: "ord_1", status: "completed", want: domain.ClaimReplay},
		{name: "active request", storedHash: hash, orderID: nil, status: "processing", want: domain.ClaimInProgress},
		{name: "bound request", storedHash: hash, orderID: "ord_1", status: "processing", want: domain.ClaimResume},
		{name: "terminal failure", storedHash: hash, orderID: nil, status: "failed", want: domain.ClaimFailed},
		{name: "changed payload", storedHash: strings.Repeat("b", 64), orderID: "ord_1", status: "completed", wantErr: domain.ErrIdempotencyConflict},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			repository, _ := NewMySQLIdempotencyRepository(db)
			mock.ExpectExec("INSERT INTO order_idempotency_keys").
				WillReturnError(&mysql.MySQLError{Number: 1062, Message: "duplicate"})
			mock.ExpectQuery("SELECT request_hash, order_id, status, expires_at").
				WithArgs("user_1", "key_1").
				WillReturnRows(sqlmock.NewRows([]string{"request_hash", "order_id", "status", "expires_at"}).
					AddRow(test.storedHash, test.orderID, test.status, expiresAt))

			claim, err := repository.Claim(context.Background(), "user_1", "key_1", hash, expiresAt)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Claim() error = %v, want %v", err, test.wantErr)
			}
			if test.wantErr == nil && claim.Decision != test.want {
				t.Fatalf("claim decision = %q, want %q", claim.Decision, test.want)
			}
			assertMockExpectations(t, mock)
		})
	}
}

func TestMySQLIdempotencyRepositoryCompletesBoundClaim(t *testing.T) {
	db, mock := newMockDB(t)
	repository, _ := NewMySQLIdempotencyRepository(db)
	mock.ExpectExec("UPDATE order_idempotency_keys").
		WithArgs("user_1", "key_1", "ord_1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repository.Complete(context.Background(), "user_1", "key_1", "ord_1"); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	assertMockExpectations(t, mock)
}

func TestMySQLIdempotencyRepositoryMarksUnboundClaimFailed(t *testing.T) {
	db, mock := newMockDB(t)
	repository, _ := NewMySQLIdempotencyRepository(db)
	mock.ExpectExec("UPDATE order_idempotency_keys").
		WithArgs("user_1", "key_1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repository.MarkFailed(context.Background(), "user_1", "key_1"); err != nil {
		t.Fatalf("MarkFailed() error = %v", err)
	}
	assertMockExpectations(t, mock)
}

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func assertMockExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("database expectations were not met: %v", err)
	}
}
