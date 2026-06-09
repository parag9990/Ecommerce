package domain

import (
	"errors"
	"testing"
	"time"
)

func TestKYCDocumentApprove(t *testing.T) {
	document, err := NewKYCDocument(NewKYCDocumentParams{
		DocumentID:   "doc_123",
		SellerID:     "seller_123",
		DocumentType: KYCDocumentTypeGSTCertificate,
		StorageURL:   "s3://private-kyc/seller_123/doc_123.pdf",
		CreatedBy:    "user_123",
		CreatedAt:    fixedTime(),
	})
	if err != nil {
		t.Fatalf("NewKYCDocument returned error: %v", err)
	}

	if err := document.Approve("admin_123", fixedTime().Add(time.Minute)); err != nil {
		t.Fatalf("Approve returned error: %v", err)
	}

	if document.Status != KYCStatusApproved {
		t.Fatalf("expected approved status, got %q", document.Status)
	}
	if document.ReviewedBy == nil || *document.ReviewedBy != "admin_123" {
		t.Fatalf("expected reviewer metadata, got %#v", document.ReviewedBy)
	}
}

func TestKYCDocumentRejectRequiresReason(t *testing.T) {
	document, err := NewKYCDocument(NewKYCDocumentParams{
		DocumentID:   "doc_123",
		SellerID:     "seller_123",
		DocumentType: KYCDocumentTypePANCard,
		StorageURL:   "s3://private-kyc/seller_123/doc_123.pdf",
		CreatedBy:    "user_123",
		CreatedAt:    fixedTime(),
	})
	if err != nil {
		t.Fatalf("NewKYCDocument returned error: %v", err)
	}

	err = document.Reject("admin_123", " ", fixedTime().Add(time.Minute))
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestKYCDocumentReviewIsTerminal(t *testing.T) {
	document, err := NewKYCDocument(NewKYCDocumentParams{
		DocumentID:   "doc_123",
		SellerID:     "seller_123",
		DocumentType: KYCDocumentTypeAddressProof,
		StorageURL:   "s3://private-kyc/seller_123/doc_123.pdf",
		CreatedBy:    "user_123",
		CreatedAt:    fixedTime(),
	})
	if err != nil {
		t.Fatalf("NewKYCDocument returned error: %v", err)
	}

	if err := document.Approve("admin_123", fixedTime().Add(time.Minute)); err != nil {
		t.Fatalf("Approve returned error: %v", err)
	}

	err = document.Reject("admin_123", "unclear document", fixedTime().Add(2*time.Minute))
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("expected invalid transition error, got %v", err)
	}
}
