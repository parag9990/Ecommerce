package domain

import (
	"fmt"
	"net/url"
	"strings"
)

type ImageStatus string

const (
	ImageStatusActive     ImageStatus = "active"
	ImageStatusHidden     ImageStatus = "hidden"
	ImageStatusProcessing ImageStatus = "processing"
	ImageStatusFailed     ImageStatus = "failed"
)

type ProductImage struct {
	ID         string      `json:"image_id"`
	URL        string      `json:"url"`
	AltText    string      `json:"alt_text,omitempty"`
	Position   int         `json:"position"`
	IsPrimary  bool        `json:"is_primary"`
	VariantIDs []string    `json:"variant_ids,omitempty"`
	Width      int         `json:"width,omitempty"`
	Height     int         `json:"height,omitempty"`
	Status     ImageStatus `json:"status"`
}

func (s ImageStatus) Valid() bool {
	switch s {
	case ImageStatusActive, ImageStatusHidden, ImageStatusProcessing, ImageStatusFailed:
		return true
	default:
		return false
	}
}

func (i ProductImage) Active() bool {
	return i.Status == ImageStatusActive
}

func (i ProductImage) Validate(field string) ValidationReport {
	var report ValidationReport
	if !requiredString(i.ID) {
		report.AddError(CodeImageIDRequired, fieldPath(field, "image_id"), "image id is required")
	}
	if !requiredString(i.URL) {
		report.AddError(CodeImageURLRequired, fieldPath(field, "url"), "image URL is required")
	} else if !isPublicURL(i.URL) {
		report.AddError(CodeInvalidImageURL, fieldPath(field, "url"), "image URL must be an absolute http or https URL")
	}
	if i.Position <= 0 {
		report.AddError(CodeInvalidImagePosition, fieldPath(field, "position"), "image position must be a positive integer")
	}
	if !i.Status.Valid() {
		report.AddError(CodeInvalidImageStatus, fieldPath(field, "status"), "image status is not supported")
	}
	if i.Width < 0 || i.Height < 0 {
		report.AddError(CodeInvalidImageDimensions, field, "image dimensions cannot be negative")
	}
	for index, variantID := range i.VariantIDs {
		if strings.TrimSpace(variantID) == "" {
			report.AddError(CodeInvalidImageVariantReference, fmt.Sprintf("%s[%d]", fieldPath(field, "variant_ids"), index), "variant id reference cannot be empty")
		}
	}
	return report
}

func isPublicURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return parsed.IsAbs() && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}
