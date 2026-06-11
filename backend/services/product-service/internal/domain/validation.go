package domain

import (
	"fmt"
	"strings"
)

type IssueSeverity string

const (
	IssueSeverityError   IssueSeverity = "error"
	IssueSeverityWarning IssueSeverity = "warning"
)

const (
	CodeProductIDRequired            = "PRODUCT_ID_REQUIRED"
	CodeSellerRequired               = "SELLER_REQUIRED"
	CodeTitleRequired                = "TITLE_REQUIRED"
	CodeCategoryRequired             = "CATEGORY_REQUIRED"
	CodeCategoryNotFound             = "CATEGORY_NOT_FOUND"
	CodeCategoryNotVerified          = "CATEGORY_NOT_VERIFIED"
	CodeInactiveCategory             = "INACTIVE_CATEGORY"
	CodeInvalidCategoryPath          = "INVALID_CATEGORY_PATH"
	CodeInvalidProductStatus         = "INVALID_PRODUCT_STATUS"
	CodeInvalidStatusTransition      = "INVALID_STATUS_TRANSITION"
	CodeVariantRequired              = "VARIANT_REQUIRED"
	CodeActiveVariantRequired        = "ACTIVE_VARIANT_REQUIRED"
	CodeVariantIDRequired            = "VARIANT_ID_REQUIRED"
	CodeSKURequired                  = "SKU_REQUIRED"
	CodeDuplicateSKU                 = "DUPLICATE_SKU"
	CodeSKUUniquenessNotVerified     = "SKU_UNIQUENESS_NOT_VERIFIED"
	CodeDuplicateVariantAttributes   = "DUPLICATE_VARIANT_ATTRIBUTES"
	CodeInvalidPrice                 = "INVALID_PRICE"
	CodeInvalidMRP                   = "INVALID_MRP"
	CodeInvalidCurrency              = "INVALID_CURRENCY"
	CodeInvalidStock                 = "INVALID_STOCK"
	CodeInvalidReservedQuantity      = "INVALID_RESERVED_QUANTITY"
	CodeInvalidSafetyStock           = "INVALID_SAFETY_STOCK"
	CodeInvalidVariantStatus         = "INVALID_VARIANT_STATUS"
	CodeInvalidAvailableQuantity     = "INVALID_AVAILABLE_QUANTITY"
	CodeImageIDRequired              = "IMAGE_ID_REQUIRED"
	CodeImageURLRequired             = "IMAGE_URL_REQUIRED"
	CodeInvalidImageURL              = "INVALID_IMAGE_URL"
	CodeInvalidImagePosition         = "INVALID_IMAGE_POSITION"
	CodeDuplicateImagePosition       = "DUPLICATE_IMAGE_POSITION"
	CodeMultiplePrimaryImages        = "MULTIPLE_PRIMARY_IMAGES"
	CodePrimaryImageRequired         = "PRIMARY_IMAGE_REQUIRED"
	CodeInvalidImageStatus           = "INVALID_IMAGE_STATUS"
	CodeInvalidImageDimensions       = "INVALID_IMAGE_DIMENSIONS"
	CodeInvalidImageVariantReference = "INVALID_IMAGE_VARIANT_REFERENCE"
	CodeAttributeKeyRequired         = "ATTRIBUTE_KEY_REQUIRED"
	CodeAttributeLabelRequired       = "ATTRIBUTE_LABEL_REQUIRED"
	CodeInvalidAttributeType         = "INVALID_ATTRIBUTE_TYPE"
	CodeInvalidAttributeScope        = "INVALID_ATTRIBUTE_SCOPE"
	CodeRequiredAttributeMissing     = "REQUIRED_ATTRIBUTE_MISSING"
	CodeUnknownAttributeKey          = "UNKNOWN_ATTRIBUTE_KEY"
	CodeInvalidAttributeValue        = "INVALID_ATTRIBUTE_VALUE"
	CodeDuplicateAttributeDefinition = "DUPLICATE_ATTRIBUTE_DEFINITION"
	CodeInvalidCategory              = "INVALID_CATEGORY"
	CodeInvalidCategorySlug          = "INVALID_CATEGORY_SLUG"
	CodeInvalidCategoryLevel         = "INVALID_CATEGORY_LEVEL"
	CodeInvalidRating                = "INVALID_RATING"
	CodeLimitExceeded                = "LIMIT_EXCEEDED"
	CodeBrandIDRequired              = "BRAND_ID_REQUIRED"
	CodeBrandNameRequired            = "BRAND_NAME_REQUIRED"
	CodeInvalidBrandSlug             = "INVALID_BRAND_SLUG"
	CodeInvalidBrandStatus           = "INVALID_BRAND_STATUS"
	CodeInvalidBrandURL              = "INVALID_BRAND_URL"
	CodeInventorySnapshotIDRequired  = "INVENTORY_SNAPSHOT_ID_REQUIRED"
	CodeInvalidInventorySnapshot     = "INVALID_INVENTORY_SNAPSHOT"
	CodeInvalidInventorySnapshotType = "INVALID_INVENTORY_SNAPSHOT_TYPE"
	CodeReservationIDRequired        = "RESERVATION_ID_REQUIRED"
	CodeOrderIDRequired              = "ORDER_ID_REQUIRED"
	CodeReservationItemsRequired     = "RESERVATION_ITEMS_REQUIRED"
	CodeInvalidReservationStatus     = "INVALID_RESERVATION_STATUS"
	CodeInvalidReservationItem       = "INVALID_RESERVATION_ITEM"
	CodeInvalidReservationTTL        = "INVALID_RESERVATION_TTL"
	CodePriceBookIDRequired          = "PRICE_BOOK_ID_REQUIRED"
	CodePriceBookEntryRequired       = "PRICE_BOOK_ENTRY_REQUIRED"
	CodeDuplicatePriceBookEntry      = "DUPLICATE_PRICE_BOOK_ENTRY"
	CodeInvalidPriceBookStatus       = "INVALID_PRICE_BOOK_STATUS"
	CodeInvalidPriceBookWindow       = "INVALID_PRICE_BOOK_WINDOW"
	CodeEventIDRequired              = "EVENT_ID_REQUIRED"
	CodeInvalidEventType             = "INVALID_EVENT_TYPE"
	CodeInvalidEventEnvelope         = "INVALID_EVENT_ENVELOPE"
	CodeInvalidEventPayload          = "INVALID_EVENT_PAYLOAD"
	CodeInvalidSearchAction          = "INVALID_SEARCH_ACTION"
	CodeInvalidOutboxStatus          = "INVALID_OUTBOX_STATUS"
)

type ValidationIssue struct {
	Code     string        `json:"code"`
	Field    string        `json:"field,omitempty"`
	Message  string        `json:"message"`
	Severity IssueSeverity `json:"severity"`
}

type ValidationReport struct {
	Issues []ValidationIssue `json:"issues"`
}

func (r ValidationReport) IsValid() bool {
	return !r.HasErrors()
}

func (r ValidationReport) HasErrors() bool {
	for _, issue := range r.Issues {
		if issue.Severity == IssueSeverityError {
			return true
		}
	}
	return false
}

func (r ValidationReport) ErrorIssues() []ValidationIssue {
	issues := make([]ValidationIssue, 0, len(r.Issues))
	for _, issue := range r.Issues {
		if issue.Severity == IssueSeverityError {
			issues = append(issues, issue)
		}
	}
	return issues
}

func (r *ValidationReport) Merge(other ValidationReport) {
	if len(other.Issues) == 0 {
		return
	}
	r.Issues = append(r.Issues, other.Issues...)
}

func (r *ValidationReport) AddError(code, field, message string) {
	r.add(IssueSeverityError, code, field, message)
}

func (r *ValidationReport) AddWarning(code, field, message string) {
	r.add(IssueSeverityWarning, code, field, message)
}

func (r *ValidationReport) add(severity IssueSeverity, code, field, message string) {
	r.Issues = append(r.Issues, ValidationIssue{
		Code:     code,
		Field:    field,
		Message:  message,
		Severity: severity,
	})
}

type ValidationError struct {
	Report ValidationReport
}

func (e ValidationError) Error() string {
	if len(e.Report.Issues) == 0 {
		return "validation failed"
	}
	codes := make([]string, 0, len(e.Report.Issues))
	for _, issue := range e.Report.ErrorIssues() {
		codes = append(codes, issue.Code)
	}
	if len(codes) == 0 {
		return "validation completed with warnings"
	}
	return fmt.Sprintf("validation failed: %s", strings.Join(codes, ", "))
}

func fieldPath(parent, child string) string {
	if parent == "" {
		return child
	}
	if child == "" {
		return parent
	}
	return parent + "." + child
}

func requiredString(value string) bool {
	return strings.TrimSpace(value) != ""
}
