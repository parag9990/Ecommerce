package domain

import (
	"fmt"
	"strings"
	"time"
)

// TemplateStatus controls whether a stored template can be selected for use.
type TemplateStatus string

const (
	TemplateStatusActive   TemplateStatus = "active"
	TemplateStatusInactive TemplateStatus = "inactive"
)

type Template struct {
	ID          string         `bson:"_id" json:"id"`
	TemplateKey string         `bson:"template_key" json:"template_key"`
	Channel     Channel        `bson:"channel" json:"channel"`
	Subject     string         `bson:"subject,omitempty" json:"subject,omitempty"`
	Body        string         `bson:"body" json:"body"`
	Status      TemplateStatus `bson:"status" json:"status"`
	Version     int            `bson:"version" json:"version"`
	CreatedAt   time.Time      `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `bson:"updated_at" json:"updated_at"`
}

func (t Template) Validate() error {
	if blank(t.ID) {
		return fmt.Errorf("%w: id is required", ErrInvalidTemplate)
	}
	if blank(t.TemplateKey) {
		return fmt.Errorf("%w: template key is required", ErrInvalidTemplate)
	}
	if !t.Channel.IsSupported() {
		return fmt.Errorf("%w: %w: %q", ErrInvalidTemplate, ErrUnsupportedChannel, t.Channel)
	}
	if t.Channel == ChannelEmail && blank(t.Subject) {
		return fmt.Errorf("%w: email subject is required", ErrInvalidTemplate)
	}
	if blank(t.Body) {
		return fmt.Errorf("%w: body is required", ErrInvalidTemplate)
	}
	if !t.Status.IsSupported() {
		return fmt.Errorf("%w: unsupported status %q", ErrInvalidTemplate, t.Status)
	}
	if t.Version < 1 {
		return fmt.Errorf("%w: version must be at least 1", ErrInvalidTemplate)
	}
	if err := validateRecordTimestamps(t.CreatedAt, t.UpdatedAt); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidTemplate, err)
	}
	return nil
}

func (s TemplateStatus) IsSupported() bool {
	switch s {
	case TemplateStatusActive, TemplateStatusInactive:
		return true
	default:
		return false
	}
}

func validateRecordTimestamps(createdAt, updatedAt time.Time) error {
	if createdAt.IsZero() || updatedAt.IsZero() {
		return fmt.Errorf("created_at and updated_at are required")
	}
	if updatedAt.Before(createdAt) {
		return fmt.Errorf("updated_at must not be before created_at")
	}
	return nil
}

func requiredText(value, field string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}
	return nil
}
