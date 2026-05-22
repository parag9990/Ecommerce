package domain

import (
	"time"
)

type KYCDocumentType string

const (
	KYCDocumentTypeGSTCertificate       KYCDocumentType = "gst_certificate"
	KYCDocumentTypePANCard              KYCDocumentType = "pan_card"
	KYCDocumentTypeAddressProof         KYCDocumentType = "address_proof"
	KYCDocumentTypeBankProof            KYCDocumentType = "bank_proof"
	KYCDocumentTypeBusinessRegistration KYCDocumentType = "business_registration"
)

type KYCStatus string

const (
	KYCStatusPending  KYCStatus = "pending"
	KYCStatusApproved KYCStatus = "approved"
	KYCStatusRejected KYCStatus = "rejected"
)

type KYCDocument struct {
	DocumentID      string
	SellerID        string
	DocumentType    KYCDocumentType
	StorageURL      string
	Status          KYCStatus
	ReviewedBy      *string
	ReviewedAt      *time.Time
	RejectionReason *string
	CreatedAt       time.Time
}

type NewKYCDocumentParams struct {
	DocumentID   string
	SellerID     string
	DocumentType KYCDocumentType
	StorageURL   string
	CreatedAt    time.Time
}

func NewKYCDocument(params NewKYCDocumentParams) (KYCDocument, error) {
	document := KYCDocument{
		DocumentID:   trim(params.DocumentID),
		SellerID:     trim(params.SellerID),
		DocumentType: KYCDocumentType(trim(string(params.DocumentType))),
		StorageURL:   trim(params.StorageURL),
		Status:       KYCStatusPending,
		CreatedAt:    params.CreatedAt.UTC(),
	}

	if err := document.Validate(); err != nil {
		return KYCDocument{}, err
	}

	return document, nil
}

func (d KYCDocumentType) Valid() bool {
	switch d {
	case KYCDocumentTypeGSTCertificate, KYCDocumentTypePANCard, KYCDocumentTypeAddressProof, KYCDocumentTypeBankProof, KYCDocumentTypeBusinessRegistration:
		return true
	default:
		return false
	}
}

func (s KYCStatus) Valid() bool {
	switch s {
	case KYCStatusPending, KYCStatusApproved, KYCStatusRejected:
		return true
	default:
		return false
	}
}

func (d KYCDocument) Validate() error {
	var v validationCollector

	validateID(&v, "document_id", d.DocumentID)
	validateID(&v, "seller_id", d.SellerID)
	validateRequiredString(&v, "document_type", string(d.DocumentType), maxDocumentTypeLength)
	if !d.DocumentType.Valid() {
		v.add("document_type", "is not supported")
	}
	validateRequiredString(&v, "storage_url", d.StorageURL, maxStorageURLLength)
	if !d.Status.Valid() {
		v.add("status", "is not supported")
	}
	validateOptionalID(&v, "reviewed_by", d.ReviewedBy)
	validateOptionalString(&v, "rejection_reason", d.RejectionReason, maxRejectionReasonLength)
	validateTimestamp(&v, "created_at", d.CreatedAt)
	if d.ReviewedAt != nil && d.ReviewedAt.Before(d.CreatedAt) {
		v.add("reviewed_at", "cannot be before created_at")
	}

	switch d.Status {
	case KYCStatusPending:
		if d.ReviewedBy != nil || d.ReviewedAt != nil || d.RejectionReason != nil {
			v.add("status", "pending document cannot have review metadata")
		}
	case KYCStatusApproved:
		if d.ReviewedBy == nil || d.ReviewedAt == nil {
			v.add("reviewed_by", "is required for approved document")
		}
		if d.RejectionReason != nil {
			v.add("rejection_reason", "must be empty for approved document")
		}
	case KYCStatusRejected:
		if d.ReviewedBy == nil || d.ReviewedAt == nil {
			v.add("reviewed_by", "is required for rejected document")
		}
		if d.RejectionReason == nil {
			v.add("rejection_reason", "is required for rejected document")
		}
	}

	return v.err()
}

func (d *KYCDocument) Approve(reviewedBy string, at time.Time) error {
	return d.review(KYCStatusApproved, reviewedBy, nil, at)
}

func (d *KYCDocument) Reject(reviewedBy string, reason string, at time.Time) error {
	rejectionReason := trim(reason)
	return d.review(KYCStatusRejected, reviewedBy, &rejectionReason, at)
}

func (d *KYCDocument) review(status KYCStatus, reviewedBy string, rejectionReason *string, at time.Time) error {
	if at.IsZero() {
		return ValidationError{Fields: []FieldError{{Field: "reviewed_at", Message: "is required"}}}
	}
	if d.Status != KYCStatusPending {
		return invalidTransition("kyc_document", string(d.Status), string(status))
	}

	reviewer := trim(reviewedBy)
	reviewedAt := at.UTC()
	next := *d
	next.Status = status
	next.ReviewedBy = &reviewer
	next.ReviewedAt = &reviewedAt
	next.RejectionReason = cleanOptional(rejectionReason)

	if err := next.Validate(); err != nil {
		return err
	}

	*d = next
	return nil
}
