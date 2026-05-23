package domain

import "strings"

type Category struct {
	ID              string                `json:"category_id"`
	Name            string                `json:"name"`
	Slug            string                `json:"slug"`
	ParentID        *string               `json:"parent_id,omitempty"`
	Path            []string              `json:"path"`
	Level           int                   `json:"level"`
	SortOrder       int                   `json:"sort_order"`
	IsActive        bool                  `json:"is_active"`
	AttributeSchema []AttributeDefinition `json:"attribute_schema,omitempty"`
}

func (c Category) Validate() ValidationReport {
	var report ValidationReport
	if !requiredString(c.ID) {
		report.AddError(CodeCategoryRequired, "category_id", "category id is required")
	}
	if !requiredString(c.Name) {
		report.AddError(CodeInvalidCategory, "name", "category name is required")
	}
	if !requiredString(c.Slug) || !isSlug(c.Slug) {
		report.AddError(CodeInvalidCategorySlug, "slug", "category slug must be URL friendly")
	}
	if c.Level < 0 {
		report.AddError(CodeInvalidCategoryLevel, "level", "category level cannot be negative")
	}
	if len(c.Path) == 0 {
		report.AddError(CodeInvalidCategoryPath, "path", "category path must include root through current category")
	} else if c.ID != "" && c.Path[len(c.Path)-1] != c.ID {
		report.AddError(CodeInvalidCategoryPath, "path", "category path must end with category id")
	}
	report.Merge(ValidateAttributeDefinitions(c.AttributeSchema, "attribute_schema"))
	return report
}

func isSlug(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "-") || strings.HasSuffix(value, "-") {
		return false
	}
	if strings.ToLower(value) != value {
		return false
	}
	for _, ch := range value {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			continue
		}
		return false
	}
	return true
}
