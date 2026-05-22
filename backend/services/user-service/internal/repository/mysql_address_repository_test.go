package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
)

func TestMySQLAddressRepositorySetDefaultAddress(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo, err := NewMySQLAddressRepository(db, WithLogger(testLogger()))
	if err != nil {
		t.Fatalf("NewMySQLAddressRepository returned error: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT address_id.*FROM user_addresses.*FOR UPDATE`).
		WithArgs("user_123", "addr_123").
		WillReturnRows(sqlmock.NewRows([]string{"address_id"}).AddRow("addr_123"))
	mock.ExpectExec(`(?s)UPDATE user_addresses.*SET is_default = FALSE`).
		WithArgs("user_123").
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(`(?s)UPDATE user_addresses.*SET is_default = TRUE`).
		WithArgs("user_123", "addr_123").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = repo.SetDefaultAddress(context.Background(), "user_123", "addr_123")
	if err != nil {
		t.Fatalf("SetDefaultAddress returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestMySQLAddressRepositorySetDefaultAddressNotFoundRollsBack(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo, err := NewMySQLAddressRepository(db, WithLogger(testLogger()))
	if err != nil {
		t.Fatalf("NewMySQLAddressRepository returned error: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT address_id.*FROM user_addresses.*FOR UPDATE`).
		WithArgs("user_123", "addr_missing").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err = repo.SetDefaultAddress(context.Background(), "user_123", "addr_missing")
	if !errors.Is(err, domain.ErrAddressNotFound) {
		t.Fatalf("expected ErrAddressNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestMySQLAddressRepositoryDeleteAddressNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo, err := NewMySQLAddressRepository(db, WithLogger(testLogger()))
	if err != nil {
		t.Fatalf("NewMySQLAddressRepository returned error: %v", err)
	}

	mock.ExpectExec(`(?s)UPDATE user_addresses.*SET.*deleted_at = CURRENT_TIMESTAMP`).
		WithArgs("user_123", "addr_missing").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.DeleteAddress(context.Background(), "user_123", "addr_missing")
	if !errors.Is(err, domain.ErrAddressNotFound) {
		t.Fatalf("expected ErrAddressNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
