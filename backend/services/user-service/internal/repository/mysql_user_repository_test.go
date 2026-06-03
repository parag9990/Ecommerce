package repository

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func fixedRepositoryTime() time.Time {
	return time.Date(2026, 5, 21, 10, 30, 0, 0, time.UTC)
}

func stringPtrRepository(value string) *string {
	return &value
}

func timePtrRepository(value time.Time) *time.Time {
	return &value
}

func TestMySQLUserRepositoryFindUserByIDNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo, err := NewMySQLUserRepository(db, WithLogger(testLogger()))
	if err != nil {
		t.Fatalf("NewMySQLUserRepository returned error: %v", err)
	}

	mock.ExpectQuery(`(?s)SELECT.*FROM users.*WHERE user_id = \?.*LIMIT 1`).
		WithArgs("user_missing").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.FindUserByID(context.Background(), "user_missing")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestMySQLUserRepositoryCreateUserDuplicate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo, err := NewMySQLUserRepository(db, WithLogger(testLogger()))
	if err != nil {
		t.Fatalf("NewMySQLUserRepository returned error: %v", err)
	}

	user := domain.User{
		UserID:        "user_123",
		AuthAccountID: "auth_123",
		Email:         "buyer@example.com",
		FullName:      "Aarav Sharma",
		Status:        domain.UserStatusActive,
		AuditFields: domain.AuditFields{
			CreatedBy: "service:auth-service",
			UpdatedBy: "service:auth-service",
			CreatedAt: fixedRepositoryTime(),
			UpdatedAt: fixedRepositoryTime(),
		},
		StatusAuditFields: domain.StatusAuditFields{
			StatusChangedBy: stringPtrRepository("service:auth-service"),
			StatusChangedAt: timePtrRepository(fixedRepositoryTime()),
		},
	}

	mock.ExpectExec(`(?s)INSERT INTO users`).
		WithArgs(
			user.UserID,
			user.AuthAccountID,
			user.Email,
			nil,
			user.FullName,
			nil,
			user.Status,
			user.CreatedBy,
			user.UpdatedBy,
			"service:auth-service",
			fixedRepositoryTime(),
			fixedRepositoryTime(),
			fixedRepositoryTime(),
		).
		WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry"})

	_, err = repo.CreateUser(context.Background(), user)
	if !errors.Is(err, domain.ErrDuplicateUser) {
		t.Fatalf("expected ErrDuplicateUser, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestMySQLUserRepositoryBatchFindUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo, err := NewMySQLUserRepository(db, WithLogger(testLogger()))
	if err != nil {
		t.Fatalf("NewMySQLUserRepository returned error: %v", err)
	}

	now := fixedRepositoryTime()
	rows := sqlmock.NewRows([]string{
		"user_id",
		"auth_account_id",
		"email",
		"phone",
		"full_name",
		"avatar_url",
		"status",
		"created_by",
		"updated_by",
		"status_changed_by",
		"status_changed_at",
		"deleted_by",
		"deleted_at",
		"created_at",
		"updated_at",
	}).AddRow("user_1", "auth_1", "one@example.com", nil, "One", nil, "active", "system:backfill", "system:backfill", "system:backfill", now, nil, nil, now, now).
		AddRow("user_2", "auth_2", "two@example.com", "+919999999999", "Two", nil, "active", "system:backfill", "system:backfill", "system:backfill", now, nil, nil, now, now)

	mock.ExpectQuery(`(?s)SELECT.*FROM users.*WHERE user_id IN \(\?,\?\).*ORDER BY user_id`).
		WithArgs("user_1", "user_2").
		WillReturnRows(rows)

	users, err := repo.BatchFindUsers(context.Background(), []string{"user_1", "user_2"})
	if err != nil {
		t.Fatalf("BatchFindUsers returned error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[1].Phone == nil || *users[1].Phone != "+919999999999" {
		t.Fatalf("expected nullable phone to scan, got %#v", users[1].Phone)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
