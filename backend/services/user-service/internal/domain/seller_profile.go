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
	AuditFields
	StatusAuditFields
}

type NewSellerProfileParams struct {
	SellerID     string
	UserID       string
	StoreName    string
	DisplayName  *string
	GSTNumber    *string
	SupportEmail *string
	CreatedBy    string
	CreatedAt    time.Time
}

type SellerProfilePatch struct {
	StoreName    *string
	DisplayName  *string
	GSTNumber    *string
	SupportEmail *string
	UpdatedBy    string
	UpdatedAt    time.Time
}

func NewSellerProfile(params NewSellerProfileParams) (SellerProfile, error) {
	createdAt := params.CreatedAt.UTC()
	profile := SellerProfile{
		SellerID:     trim(params.SellerID),
		UserID:       trim(params.UserID),
		StoreName:    cleanText(params.StoreName),
		DisplayName:  cleanOptionalText(params.DisplayName),
		GSTNumber:    cleanOptionalUpper(params.GSTNumber),
		SupportEmail: cleanOptionalEmail(params.SupportEmail),
		Status:       SellerStatusDraft,
		AuditFields: AuditFields{
			CreatedBy: trim(params.CreatedBy),
			UpdatedBy: trim(params.CreatedBy),
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
		StatusAuditFields: StatusAuditFields{
			StatusChangedBy: cleanOptional(&params.CreatedBy),
			StatusChangedAt: &createdAt,
		},
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
	validateRequiredSafeText(&v, "store_name", s.StoreName, 3, maxNameLength)
	if s.DisplayName != nil {
		validateOptionalSafeText(&v, "display_name", *s.DisplayName, 2, maxNameLength)
	}
	validateOptionalGSTNumber(&v, "gst_number", s.GSTNumber)
	validateOptionalEmail(&v, "support_email", s.SupportEmail)
	if !s.Status.Valid() {
		v.add("status", "is not supported")
	}
	validateOptionalID(&v, "approved_by", s.ApprovedBy)
	validateAuditFields(&v, s.AuditFields)
	validateStatusAuditFields(&v, s.StatusAuditFields, s.CreatedAt)
	if s.ApprovedAt != nil && s.ApprovedAt.Before(s.CreatedAt) {
		v.add("approved_at", "cannot be before created_at")
	}
	if s.Status == SellerStatusActive && (s.ApprovedBy == nil || s.ApprovedAt == nil) {
		v.add("approved_by", "is required for active seller")
	}
	if s.Status != SellerStatusActive && s.ApprovedAt != nil && s.ApprovedBy == nil {
		v.add("approved_by", "is required when approved_at is present")
	}
	if (s.Status == SellerStatusRejected || s.Status == SellerStatusSuspended) && s.StatusReason == nil {
		v.add("status_reason", "is required for rejected or suspended seller")
	}

	return v.err()
}

func (s *SellerProfile) ApplyPatch(patch SellerProfilePatch) error {
	if trim(patch.UpdatedBy) == "" {
		return ValidationError{Fields: []FieldError{{Field: "updated_by", Message: "is required"}}}
	}
	if patch.UpdatedAt.IsZero() {
		return ValidationError{Fields: []FieldError{{Field: "updated_at", Message: "is required"}}}
	}

	next := *s
	if patch.StoreName != nil {
		next.StoreName = cleanText(*patch.StoreName)
	}
	if patch.DisplayName != nil {
		next.DisplayName = cleanOptionalText(patch.DisplayName)
	}
	if patch.GSTNumber != nil {
		next.GSTNumber = cleanOptionalUpper(patch.GSTNumber)
	}
	if patch.SupportEmail != nil {
		next.SupportEmail = cleanOptionalEmail(patch.SupportEmail)
	}
	next.UpdatedBy = trim(patch.UpdatedBy)
	next.UpdatedAt = patch.UpdatedAt.UTC()

	if err := next.Validate(); err != nil {
		return err
	}

	*s = next
	return nil
}

func (s *SellerProfile) SubmitForReview(actorID string, at time.Time) error {
	return s.transition(SellerStatusPendingReview, actorID, nil, nil, nil, at)
}

func (s *SellerProfile) Approve(approvedBy string, at time.Time) error {
	reviewer := trim(approvedBy)
	return s.transition(SellerStatusActive, reviewer, &reviewer, &at, nil, at)
}

func (s *SellerProfile) Reject(actorID string, reason string, at time.Time) error {
	statusReason := trim(reason)
	return s.transition(SellerStatusRejected, actorID, nil, nil, &statusReason, at)
}

func (s *SellerProfile) Suspend(actorID string, reason string, at time.Time) error {
	statusReason := trim(reason)
	return s.transition(SellerStatusSuspended, actorID, s.ApprovedBy, s.ApprovedAt, &statusReason, at)
}

func (s *SellerProfile) Reactivate(actorID string, reason string, at time.Time) error {
	statusReason := cleanOptional(&reason)
	return s.transition(SellerStatusActive, actorID, s.ApprovedBy, s.ApprovedAt, statusReason, at)
}

func (s *SellerProfile) transition(to SellerStatus, actorID string, approvedBy *string, approvedAt *time.Time, reason *string, at time.Time) error {
	actorID = trim(actorID)
	if actorID == "" {
		return ValidationError{Fields: []FieldError{{Field: "updated_by", Message: "is required"}}}
	}
	if at.IsZero() {
		return ValidationError{Fields: []FieldError{{Field: "updated_at", Message: "is required"}}}
	}
	if !to.Valid() {
		return ValidationError{Fields: []FieldError{{Field: "status", Message: "is not supported"}}}
	}
	if s.Status == to {
		s.UpdatedBy = actorID
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
	changedAt := at.UTC()
	next.StatusChangedBy = &actorID
	next.StatusChangedAt = &changedAt
	next.StatusReason = cleanOptional(reason)
	next.UpdatedBy = actorID
	next.UpdatedAt = changedAt

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
