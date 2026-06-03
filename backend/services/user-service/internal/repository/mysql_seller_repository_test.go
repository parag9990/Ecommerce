package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
)

func TestMySQLSellerRepositoryCreateSellerProfileDuplicate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo, err := NewMySQLSellerRepository(db, WithLogger(testLogger()))
	if err != nil {
		t.Fatalf("NewMySQLSellerRepository returned error: %v", err)
	}

	seller := domain.SellerProfile{
		SellerID:  "seller_123",
		UserID:    "user_123",
		StoreName: "Aarav Retail",
		Status:    domain.SellerStatusDraft,
		AuditFields: domain.AuditFields{
			CreatedBy: "user_123",
			UpdatedBy: "user_123",
			CreatedAt: fixedRepositoryTime(),
			UpdatedAt: fixedRepositoryTime(),
		},
		StatusAuditFields: domain.StatusAuditFields{
			StatusChangedBy: stringPtrRepository("user_123"),
			StatusChangedAt: timePtrRepository(fixedRepositoryTime()),
		},
	}

	mock.ExpectExec(`(?s)INSERT INTO seller_profiles`).
		WithArgs(
			seller.SellerID,
			seller.UserID,
			seller.StoreName,
			nil,
			nil,
			nil,
			seller.Status,
			nil,
			nil,
			seller.CreatedBy,
			seller.UpdatedBy,
			"user_123",
			fixedRepositoryTime(),
			nil,
			fixedRepositoryTime(),
			fixedRepositoryTime(),
		).
		WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry"})

	_, err = repo.CreateSellerProfile(context.Background(), seller)
	if !errors.Is(err, domain.ErrDuplicateSeller) {
		t.Fatalf("expected ErrDuplicateSeller, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestMySQLSellerRepositoryAddKYCDocumentMissingSeller(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo, err := NewMySQLSellerRepository(db, WithLogger(testLogger()))
	if err != nil {
		t.Fatalf("NewMySQLSellerRepository returned error: %v", err)
	}

	document := domain.KYCDocument{
		DocumentID:   "doc_123",
		SellerID:     "seller_missing",
		DocumentType: domain.KYCDocumentTypePANCard,
		StorageURL:   "s3://private-kyc/seller_missing/doc_123.pdf",
		Status:       domain.KYCStatusPending,
		AuditFields: domain.AuditFields{
			CreatedBy: "user_123",
			UpdatedBy: "user_123",
			CreatedAt: fixedRepositoryTime(),
			UpdatedAt: fixedRepositoryTime(),
		},
		StatusAuditFields: domain.StatusAuditFields{
			StatusChangedBy: stringPtrRepository("user_123"),
			StatusChangedAt: timePtrRepository(fixedRepositoryTime()),
		},
	}

	mock.ExpectExec(`(?s)INSERT INTO seller_kyc_documents`).
		WithArgs(
			document.DocumentID,
			document.SellerID,
			document.DocumentType,
			document.StorageURL,
			document.Status,
			nil,
			nil,
			nil,
			document.CreatedBy,
			document.UpdatedBy,
			fixedRepositoryTime(),
			fixedRepositoryTime(),
			"user_123",
			fixedRepositoryTime(),
		).
		WillReturnError(&mysql.MySQLError{Number: 1452, Message: "Cannot add or update a child row"})

	_, err = repo.AddKYCDocument(context.Background(), document)
	if !errors.Is(err, domain.ErrSellerNotFound) {
		t.Fatalf("expected ErrSellerNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
