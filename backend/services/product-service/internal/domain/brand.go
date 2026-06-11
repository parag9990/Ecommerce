package domain

import "time"

type BrandStatus string

const (
	BrandStatusActive   BrandStatus = "active"
	BrandStatusInactive BrandStatus = "inactive"
	BrandStatusArchived BrandStatus = "archived"
)

type Brand struct {
	ID          string      `json:"brand_id"`
	Name        string      `json:"name"`
	Slug        string      `json:"slug"`
	Description string      `json:"description,omitempty"`
	LogoURL     string      `json:"logo_url,omitempty"`
	WebsiteURL  string      `json:"website_url,omitempty"`
	Status      BrandStatus `json:"status"`
	CreatedBy   string      `json:"created_by,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func (s BrandStatus) Valid() bool {
	switch s {
	case BrandStatusActive, BrandStatusInactive, BrandStatusArchived:
		return true
	default:
		return false
	}
}

func (b Brand) Active() bool {
	return b.Status == BrandStatusActive
}

func (b Brand) Validate() ValidationReport {
	var report ValidationReport
	if !requiredString(b.ID) {
		report.AddError(CodeBrandIDRequired, "brand_id", "brand id is required")
	}
	if !requiredString(b.Name) {
		report.AddError(CodeBrandNameRequired, "name", "brand name is required")
	}
	if !requiredString(b.Slug) || !isSlug(b.Slug) {
		report.AddError(CodeInvalidBrandSlug, "slug", "brand slug must be URL friendly")
	}
	if !b.Status.Valid() {
		report.AddError(CodeInvalidBrandStatus, "status", "brand status is not supported")
	}
	if b.LogoURL != "" && !isPublicURL(b.LogoURL) {
		report.AddError(CodeInvalidBrandURL, "logo_url", "brand logo URL must be an absolute http or https URL")
	}
	if b.WebsiteURL != "" && !isPublicURL(b.WebsiteURL) {
		report.AddError(CodeInvalidBrandURL, "website_url", "brand website URL must be an absolute http or https URL")
	}
	if b.CreatedAt.IsZero() {
		report.AddError(CodeInvalidBrandStatus, "created_at", "brand creation timestamp is required")
	}
	if b.UpdatedAt.IsZero() {
		report.AddError(CodeInvalidBrandStatus, "updated_at", "brand update timestamp is required")
	}
	return report
}
