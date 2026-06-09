package domain

import "time"

const maxAuditActorIDLength = 64

type AuditFields struct {
	CreatedBy string
	UpdatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SoftDeleteFields struct {
	DeletedBy *string
	DeletedAt *time.Time
}

type StatusAuditFields struct {
	StatusChangedBy *string
	StatusChangedAt *time.Time
	StatusReason    *string
}

type MutationAudit struct {
	ActorID string
	At      time.Time
}

func NewMutationAudit(actorID string, at time.Time) MutationAudit {
	return MutationAudit{ActorID: trim(actorID), At: at.UTC()}
}

func validateAuditFields(v *validationCollector, audit AuditFields) {
	validateRequiredAuditID(v, "created_by", audit.CreatedBy)
	validateRequiredAuditID(v, "updated_by", audit.UpdatedBy)
	validateTimestamp(v, "created_at", audit.CreatedAt)
	validateTimestamp(v, "updated_at", audit.UpdatedAt)
	if !audit.CreatedAt.IsZero() && !audit.UpdatedAt.IsZero() && audit.UpdatedAt.Before(audit.CreatedAt) {
		v.add("updated_at", "cannot be before created_at")
	}
}

func validateStatusAuditFields(v *validationCollector, audit StatusAuditFields, createdAt time.Time) {
	validateOptionalAuditID(v, "status_changed_by", audit.StatusChangedBy)
	validateOptionalString(v, "status_reason", audit.StatusReason, maxRejectionReasonLength)
	if audit.StatusChangedAt != nil && audit.StatusChangedAt.Before(createdAt) {
		v.add("status_changed_at", "cannot be before created_at")
	}
}

func validateSoftDeleteFields(v *validationCollector, audit SoftDeleteFields, createdAt time.Time) {
	validateOptionalAuditID(v, "deleted_by", audit.DeletedBy)
	if audit.DeletedAt != nil && audit.DeletedAt.Before(createdAt) {
		v.add("deleted_at", "cannot be before created_at")
	}
	if audit.DeletedAt != nil && audit.DeletedBy == nil {
		v.add("deleted_by", "is required when deleted_at is present")
	}
	if audit.DeletedBy != nil && audit.DeletedAt == nil {
		v.add("deleted_at", "is required when deleted_by is present")
	}
}

func validateMutationAudit(v *validationCollector, audit MutationAudit) {
	validateRequiredAuditID(v, "actor_id", audit.ActorID)
	validateTimestamp(v, "updated_at", audit.At)
}

func validateRequiredAuditID(v *validationCollector, field string, value string) {
	value = trim(value)
	if value == "" {
		v.add(field, "is required")
		return
	}
	validateMaxRunes(v, field, value, maxAuditActorIDLength)
	validateNoControlChars(v, field, value)
}

func validateOptionalAuditID(v *validationCollector, field string, value *string) {
	if value == nil {
		return
	}
	validateRequiredAuditID(v, field, *value)
}
