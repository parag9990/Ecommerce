package domain

import (
	"time"
)

type SellerStatus string

const (
	SellerStatusDraft         SellerStatus = "draft"
	SellerStatusPendingReview SellerStatus = "pending_review"
	SellerStatusActive        SellerStatus = "active"
	SellerStatusSuspended     SellerStatus = "suspended"
	SellerStatusRejected      SellerStatus = "rejected"
)

type SellerProfile struct {
	SellerID     string
	UserID       string
	StoreName    string
	DisplayName  *string
	GSTNumber    *string
	SupportEmail *string
	Status       SellerStatus
	ApprovedBy   *string
	ApprovedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type NewSellerProfileParams struct {
	SellerID     string
	UserID       string
	StoreName    string
	DisplayName  *string
	GSTNumber    *string
	SupportEmail *string
	CreatedAt    time.Time
}

type SellerProfilePatch struct {
	StoreName    *string
	DisplayName  *string
	GSTNumber    *string
	SupportEmail *string
	UpdatedAt    time.Time
}

func NewSellerProfile(params NewSellerProfileParams) (SellerProfile, error) {
	createdAt := params.CreatedAt.UTC()
	profile := SellerProfile{
		SellerID:     trim(params.SellerID),
		UserID:       trim(params.UserID),
		StoreName:    trim(params.StoreName),
		DisplayName:  cleanOptional(params.DisplayName),
		GSTNumber:    cleanOptionalUpper(params.GSTNumber),
		SupportEmail: cleanOptional(params.SupportEmail),
		Status:       SellerStatusDraft,
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
	}

	if err := profile.Validate(); err != nil {
		return SellerProfile{}, err
	}

	return profile, nil
}

func (s SellerStatus) Valid() bool {
	switch s {
	case SellerStatusDraft, SellerStatusPendingReview, SellerStatusActive, SellerStatusSuspended, SellerStatusRejected:
		return true
	default:
		return false
	}
}

func (s SellerProfile) Validate() error {
	var v validationCollector

	validateID(&v, "seller_id", s.SellerID)
	validateID(&v, "user_id", s.UserID)
	validateRequiredString(&v, "store_name", s.StoreName, maxNameLength)
	validateOptionalString(&v, "display_name", s.DisplayName, maxNameLength)
	validateOptionalGSTNumber(&v, "gst_number", s.GSTNumber)
	validateOptionalEmail(&v, "support_email", s.SupportEmail)
	if !s.Status.Valid() {
		v.add("status", "is not supported")
	}
	validateOptionalID(&v, "approved_by", s.ApprovedBy)
	validateTimestamp(&v, "created_at", s.CreatedAt)
	validateTimestamp(&v, "updated_at", s.UpdatedAt)
	if !s.CreatedAt.IsZero() && !s.UpdatedAt.IsZero() && s.UpdatedAt.Before(s.CreatedAt) {
		v.add("updated_at", "cannot be before created_at")
	}
	if s.ApprovedAt != nil && s.ApprovedAt.Before(s.CreatedAt) {
		v.add("approved_at", "cannot be before created_at")
	}
	if s.Status == SellerStatusActive && (s.ApprovedBy == nil || s.ApprovedAt == nil) {
		v.add("approved_by", "is required for active seller")
	}
	if s.Status != SellerStatusActive && s.ApprovedAt != nil && s.ApprovedBy == nil {
		v.add("approved_by", "is required when approved_at is present")
	}

	return v.err()
}

func (s *SellerProfile) ApplyPatch(patch SellerProfilePatch) error {
	if patch.UpdatedAt.IsZero() {
		return ValidationError{Fields: []FieldError{{Field: "updated_at", Message: "is required"}}}
	}

	next := *s
	if patch.StoreName != nil {
		next.StoreName = trim(*patch.StoreName)
	}
	if patch.DisplayName != nil {
		next.DisplayName = cleanOptional(patch.DisplayName)
	}
	if patch.GSTNumber != nil {
		next.GSTNumber = cleanOptionalUpper(patch.GSTNumber)
	}
	if patch.SupportEmail != nil {
		next.SupportEmail = cleanOptional(patch.SupportEmail)
	}
	next.UpdatedAt = patch.UpdatedAt.UTC()

	if err := next.Validate(); err != nil {
		return err
	}

	*s = next
	return nil
}

func (s *SellerProfile) SubmitForReview(at time.Time) error {
	return s.transition(SellerStatusPendingReview, nil, nil, at)
}

func (s *SellerProfile) Approve(approvedBy string, at time.Time) error {
	reviewer := trim(approvedBy)
	return s.transition(SellerStatusActive, &reviewer, &at, at)
}

func (s *SellerProfile) Reject(at time.Time) error {
	return s.transition(SellerStatusRejected, nil, nil, at)
}

func (s *SellerProfile) Suspend(at time.Time) error {
	return s.transition(SellerStatusSuspended, s.ApprovedBy, s.ApprovedAt, at)
}

func (s *SellerProfile) Reactivate(at time.Time) error {
	return s.transition(SellerStatusActive, s.ApprovedBy, s.ApprovedAt, at)
}

func (s *SellerProfile) transition(to SellerStatus, approvedBy *string, approvedAt *time.Time, at time.Time) error {
	if at.IsZero() {
		return ValidationError{Fields: []FieldError{{Field: "updated_at", Message: "is required"}}}
	}
	if !to.Valid() {
		return ValidationError{Fields: []FieldError{{Field: "status", Message: "is not supported"}}}
	}
	if s.Status == to {
		s.UpdatedAt = at.UTC()
		return s.Validate()
	}
	if !sellerStatusCanMove(s.Status, to) {
		return invalidTransition("seller_profile", string(s.Status), string(to))
	}

	next := *s
	next.Status = to
	next.ApprovedBy = cleanOptional(approvedBy)
	if approvedAt != nil {
		approvedAtUTC := approvedAt.UTC()
		next.ApprovedAt = &approvedAtUTC
	} else if to != SellerStatusActive {
		next.ApprovedAt = nil
	}
	next.UpdatedAt = at.UTC()

	if err := next.Validate(); err != nil {
		return err
	}

	*s = next
	return nil
}

func sellerStatusCanMove(from SellerStatus, to SellerStatus) bool {
	switch from {
	case SellerStatusDraft:
		return to == SellerStatusPendingReview
	case SellerStatusPendingReview:
		return to == SellerStatusActive || to == SellerStatusRejected
	case SellerStatusRejected:
		return to == SellerStatusDraft || to == SellerStatusPendingReview
	case SellerStatusActive:
		return to == SellerStatusSuspended
	case SellerStatusSuspended:
		return to == SellerStatusActive || to == SellerStatusRejected
	default:
		return false
	}
}
