package domain

import (
	"fmt"
	"time"
)

type ProductStatus string

const (
	ProductStatusDraft       ProductStatus = "draft"
	ProductStatusSubmitted   ProductStatus = "submitted"
	ProductStatusRejected    ProductStatus = "rejected"
	ProductStatusPublished   ProductStatus = "published"
	ProductStatusUnpublished ProductStatus = "unpublished"
	ProductStatusArchived    ProductStatus = "archived"
)

type Product struct {
	ID            string         `json:"product_id"`
	SellerID      string         `json:"seller_id"`
	Title         string         `json:"title"`
	Slug          string         `json:"slug,omitempty"`
	Description   string         `json:"description,omitempty"`
	Brand         string         `json:"brand,omitempty"`
	CategoryID    string         `json:"category_id"`
	CategoryPath  []string       `json:"category_path,omitempty"`
	Status        ProductStatus  `json:"status"`
	Attributes    Attributes     `json:"attributes,omitempty"`
	Images        []ProductImage `json:"images,omitempty"`
	Variants      []Variant      `json:"variants"`
	RatingSummary RatingSummary  `json:"rating_summary"`
	CreatedBy     string         `json:"created_by,omitempty"`
	UpdatedBy     string         `json:"updated_by,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	PublishedAt   *time.Time     `json:"published_at,omitempty"`
}

type RatingSummary struct {
	Average float64 `json:"average"`
	Count   int64   `json:"count"`
}

func (s ProductStatus) Valid() bool {
	switch s {
	case ProductStatusDraft,
		ProductStatusSubmitted,
		ProductStatusRejected,
		ProductStatusPublished,
		ProductStatusUnpublished,
		ProductStatusArchived:
		return true
	default:
		return false
	}
}

func (s ProductStatus) CanTransitionTo(target ProductStatus) bool {
	switch s {
	case ProductStatusDraft:
		return target == ProductStatusSubmitted ||
			target == ProductStatusPublished ||
			target == ProductStatusArchived
	case ProductStatusSubmitted:
		return target == ProductStatusPublished ||
			target == ProductStatusRejected ||
			target == ProductStatusArchived
	case ProductStatusRejected:
		return target == ProductStatusDraft ||
			target == ProductStatusSubmitted ||
			target == ProductStatusPublished ||
			target == ProductStatusArchived
	case ProductStatusPublished:
		return target == ProductStatusUnpublished || target == ProductStatusArchived
	case ProductStatusUnpublished:
		return target == ProductStatusDraft ||
			target == ProductStatusSubmitted ||
			target == ProductStatusPublished ||
			target == ProductStatusArchived
	default:
		return false
	}
}

func (p *Product) TransitionTo(target ProductStatus, now time.Time) error {
	if !target.Valid() || !p.Status.CanTransitionTo(target) {
		report := ValidationReport{}
		report.AddError(CodeInvalidStatusTransition, "status", "product status transition is not allowed")
		return ValidationError{Report: report}
	}
	p.Status = target
	p.UpdatedAt = now.UTC()
	if target == ProductStatusPublished {
		p.PublishedAt = &p.UpdatedAt
	} else if target == ProductStatusDraft || target == ProductStatusSubmitted {
		p.PublishedAt = nil
	}
	return nil
}

func CanSellerEdit(status ProductStatus) bool {
	switch status {
	case ProductStatusDraft, ProductStatusRejected, ProductStatusUnpublished:
		return true
	default:
		return false
	}
}

func CanSellerPublish(status ProductStatus) bool {
	switch status {
	case ProductStatusDraft, ProductStatusRejected, ProductStatusUnpublished:
		return true
	default:
		return false
	}
}

func (p Product) ActiveVariants() []Variant {
	variants := make([]Variant, 0, len(p.Variants))
	for _, variant := range p.Variants {
		if variant.Active() {
			variants = append(variants, variant)
		}
	}
	return variants
}

func (p Product) PrimaryImage() *ProductImage {
	for index := range p.Images {
		if p.Images[index].IsPrimary && p.Images[index].Active() {
			return &p.Images[index]
		}
	}
	return nil
}

func (p Product) Validate(options ValidationOptions, category *Category) ValidationReport {
	options = options.normalized()
	var report ValidationReport
	if !requiredString(p.ID) {
		report.AddError(CodeProductIDRequired, "product_id", "product id is required")
	}
	if !requiredString(p.SellerID) {
		report.AddError(CodeSellerRequired, "seller_id", "seller id is required")
	}
	if !requiredString(p.Title) {
		report.AddError(CodeTitleRequired, "title", "product title is required")
	}
	if !requiredString(p.CategoryID) {
		report.AddError(CodeCategoryRequired, "category_id", "category id is required")
	}
	if !p.Status.Valid() {
		report.AddError(CodeInvalidProductStatus, "status", "product status is not supported")
	}
	if len(p.Variants) == 0 {
		report.AddError(CodeVariantRequired, "variants", "at least one sellable variant is required")
	}
	if len(p.Variants) > options.MaxVariantsPerProduct {
		report.AddError(CodeLimitExceeded, "variants", "product exceeds maximum variant limit")
	}
	if len(p.Images) > options.MaxImagesPerProduct {
		report.AddError(CodeLimitExceeded, "images", "product exceeds maximum image limit")
	}
	if len(p.CategoryPath) > 0 && p.CategoryPath[len(p.CategoryPath)-1] != p.CategoryID {
		report.AddError(CodeInvalidCategoryPath, "category_path", "category path must end with product category id")
	}

	var schema []AttributeDefinition
	if category != nil {
		report.Merge(category.Validate())
		if category.ID != "" && p.CategoryID != "" && category.ID != p.CategoryID {
			report.AddError(CodeInvalidCategory, "category_id", "product category does not match supplied category")
		}
		if !category.IsActive {
			report.AddError(CodeInactiveCategory, "category_id", "new products cannot use an inactive category")
		}
		schema = category.AttributeSchema
		report.Merge(ValidateAttributes(p.Attributes, schema, AttributeScopeProduct, options, "attributes"))
	}

	report.Merge(p.RatingSummary.Validate())
	report.Merge(p.validateImages())
	report.Merge(p.validateVariants(options, schema))
	return report
}

func (p Product) ValidateForPublish(options ValidationOptions, category *Category) ValidationReport {
	report := p.Validate(options, category)
	if p.Status == ProductStatusArchived {
		report.AddError(CodeInvalidStatusTransition, "status", "archived products cannot be published")
	}
	if len(p.ActiveVariants()) == 0 {
		report.AddError(CodeActiveVariantRequired, "variants", "at least one active variant is required before publish")
	}
	if p.PrimaryImage() == nil {
		if options.RequirePrimaryImageForPublish {
			report.AddError(CodePrimaryImageRequired, "images", "published products require an active primary image")
		} else {
			report.AddWarning(CodePrimaryImageRequired, "images", "published products should have an active primary image")
		}
	}
	return report
}

func (r RatingSummary) Validate() ValidationReport {
	var report ValidationReport
	if r.Average < 0 || r.Average > 5 || r.Count < 0 {
		report.AddError(CodeInvalidRating, "rating_summary", "rating average must be between 0 and 5 and count cannot be negative")
	}
	if r.Count == 0 && r.Average != 0 {
		report.AddError(CodeInvalidRating, "rating_summary.average", "rating average must be zero when there are no ratings")
	}
	return report
}

func (p Product) validateImages() ValidationReport {
	var report ValidationReport
	positions := make(map[int]struct{}, len(p.Images))
	primaryCount := 0
	for index, image := range p.Images {
		field := fmt.Sprintf("images[%d]", index)
		report.Merge(image.Validate(field))
		if image.Position > 0 {
			if _, exists := positions[image.Position]; exists {
				report.AddError(CodeDuplicateImagePosition, fieldPath(field, "position"), "image position must be unique within product")
			}
			positions[image.Position] = struct{}{}
		}
		if image.IsPrimary && image.Active() {
			primaryCount++
		}
	}
	if primaryCount > 1 {
		report.AddError(CodeMultiplePrimaryImages, "images", "only one active primary image is allowed")
	}
	return report
}

func (p Product) validateVariants(options ValidationOptions, schema []AttributeDefinition) ValidationReport {
	var report ValidationReport
	variantIDs := make(map[string]struct{}, len(p.Variants))
	skus := make(map[string]struct{}, len(p.Variants))
	attributeSignatures := make(map[string]struct{}, len(p.Variants))
	for index, variant := range p.Variants {
		field := fmt.Sprintf("variants[%d]", index)
		report.Merge(variant.Validate(field, options, schema))
		if variant.ID != "" {
			if _, exists := variantIDs[variant.ID]; exists {
				report.AddError(CodeVariantIDRequired, fieldPath(field, "variant_id"), "variant id must be unique within product")
			}
			variantIDs[variant.ID] = struct{}{}
		}
		if variant.SKU != "" {
			if _, exists := skus[variant.SKU]; exists {
				report.AddError(CodeDuplicateSKU, fieldPath(field, "sku"), "variant SKU must be unique within product")
			}
			skus[variant.SKU] = struct{}{}
		}
		signature, err := AttributeSignature(variant.Attributes)
		if err == nil && signature != "{}" {
			if _, exists := attributeSignatures[signature]; exists {
				report.AddError(CodeDuplicateVariantAttributes, fieldPath(field, "attributes"), "variant attribute combination must be unique")
			}
			attributeSignatures[signature] = struct{}{}
		}
	}
	return report
}
