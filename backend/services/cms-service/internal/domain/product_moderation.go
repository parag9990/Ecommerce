package domain

import (
	"fmt"
	"strings"
	"time"
)

const (
	maxProductTitleLength       = 180
	minProductDescriptionLength = 20
	maxModerationReasonLength   = 512
)

type ProductStatus string

const (
	ProductStatusDraft       ProductStatus = "draft"
	ProductStatusSubmitted   ProductStatus = "submitted"
	ProductStatusApproved    ProductStatus = "approved"
	ProductStatusRejected    ProductStatus = "rejected"
	ProductStatusPublished   ProductStatus = "published"
	ProductStatusUnpublished ProductStatus = "unpublished"
)

type ProductReviewStatus string

const (
	ProductReviewStatusSubmitted ProductReviewStatus = "submitted"
	ProductReviewStatusApproved  ProductReviewStatus = "approved"
	ProductReviewStatusRejected  ProductReviewStatus = "rejected"
	ProductReviewStatusCancelled ProductReviewStatus = "cancelled"
)

type ModerationDecision string

const (
	ModerationDecisionApprove ModerationDecision = "approved"
	ModerationDecisionReject  ModerationDecision = "rejected"
)

type ProductEditRisk string

const (
	ProductEditRiskMinor ProductEditRisk = "minor"
	ProductEditRiskMajor ProductEditRisk = "major"
)

type Product struct {
	ID           string
	SellerID     string
	Title        string
	Description  string
	CategoryID   string
	Brand        string
	GenericBrand bool
	Status       ProductStatus
	Images       []ProductImage
	Variants     []ProductVariant
	UpdatedAt    time.Time
}

type ProductImage struct {
	URL string
}

type ProductVariant struct {
	SKU           string
	Price         Money
	MRP           Money
	StockQuantity int64
}

type Money struct {
	Amount   int64
	Currency string
}

type ProductStatusUpdate struct {
	ProductID      string
	SellerID       string
	ActorUserID    string
	FromStatus     ProductStatus
	ToStatus       ProductStatus
	ReviewID       string
	Reason         string
	RequestID      string
	ForceUnpublish bool
}

type ProductModerationReview struct {
	ReviewID        string
	ProductID       string
	SellerID        string
	Status          ProductReviewStatus
	SubmittedBy     string
	ReviewedBy      string
	RejectionReason string
	SubmittedAt     time.Time
	ReviewedAt      time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ProductModerationResult struct {
	ProductID string
	SellerID  string
	ReviewID  string
	Status    ProductStatus
	Message   string
}

type FieldViolation struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type ValidationError struct {
	Message string
	Fields  []FieldViolation
}

func (e ValidationError) Error() string {
	if strings.TrimSpace(e.Message) != "" {
		return e.Message
	}
	return ErrValidationFailed.Error()
}

func (e ValidationError) Unwrap() error {
	return ErrValidationFailed
}

func NewValidationError(message string, fields []FieldViolation) ValidationError {
	return ValidationError{
		Message: strings.TrimSpace(message),
		Fields:  append([]FieldViolation(nil), fields...),
	}
}

func (s ProductStatus) Normalized() ProductStatus {
	return ProductStatus(strings.ToLower(strings.TrimSpace(string(s))))
}

func (s ProductStatus) String() string {
	return string(s)
}

func (s ProductStatus) Valid() bool {
	switch s.Normalized() {
	case ProductStatusDraft, ProductStatusSubmitted, ProductStatusApproved, ProductStatusRejected, ProductStatusPublished, ProductStatusUnpublished:
		return true
	default:
		return false
	}
}

func (s ProductStatus) PubliclyVisible() bool {
	return s.Normalized() == ProductStatusPublished
}

func (s ProductReviewStatus) Normalized() ProductReviewStatus {
	return ProductReviewStatus(strings.ToLower(strings.TrimSpace(string(s))))
}

func (s ProductReviewStatus) Terminal() bool {
	switch s.Normalized() {
	case ProductReviewStatusApproved, ProductReviewStatusRejected, ProductReviewStatusCancelled:
		return true
	default:
		return false
	}
}

func (d ModerationDecision) Normalized() ModerationDecision {
	return ModerationDecision(strings.ToLower(strings.TrimSpace(string(d))))
}

func (d ModerationDecision) Valid() bool {
	switch d.Normalized() {
	case ModerationDecisionApprove, ModerationDecisionReject:
		return true
	default:
		return false
	}
}

func (d ModerationDecision) ProductStatus() ProductStatus {
	switch d.Normalized() {
	case ModerationDecisionApprove:
		return ProductStatusApproved
	case ModerationDecisionReject:
		return ProductStatusRejected
	default:
		return ""
	}
}

func (d ModerationDecision) ReviewStatus() ProductReviewStatus {
	switch d.Normalized() {
	case ModerationDecisionApprove:
		return ProductReviewStatusApproved
	case ModerationDecisionReject:
		return ProductReviewStatusRejected
	default:
		return ""
	}
}

var allowedProductTransitions = map[ProductStatus]map[ProductStatus]struct{}{
	ProductStatusDraft: {
		ProductStatusSubmitted: {},
	},
	ProductStatusSubmitted: {
		ProductStatusApproved: {},
		ProductStatusRejected: {},
	},
	ProductStatusRejected: {
		ProductStatusDraft:     {},
		ProductStatusSubmitted: {},
	},
	ProductStatusApproved: {
		ProductStatusPublished: {},
	},
	ProductStatusPublished: {
		ProductStatusUnpublished: {},
		ProductStatusSubmitted:   {},
	},
	ProductStatusUnpublished: {
		ProductStatusPublished: {},
		ProductStatusDraft:     {},
	},
}

func CanTransitionProduct(from, to ProductStatus) bool {
	from = from.Normalized()
	to = to.Normalized()
	if !from.Valid() || !to.Valid() {
		return false
	}
	targets, ok := allowedProductTransitions[from]
	if !ok {
		return false
	}
	_, ok = targets[to]
	return ok
}

func ValidateProductTransition(from, to ProductStatus) error {
	from = from.Normalized()
	to = to.Normalized()
	if !from.Valid() || !to.Valid() {
		return fmt.Errorf("%w: %s to %s", ErrInvalidProductStatus, from, to)
	}
	if CanTransitionProduct(from, to) {
		return nil
	}
	return fmt.Errorf("%w: cannot move product from %s to %s", ErrInvalidProductTransition, from, to)
}

func ValidateProductReadyForReview(product Product) error {
	product = product.Normalized()
	fields := make([]FieldViolation, 0)

	if product.ID == "" {
		fields = append(fields, FieldViolation{Field: "product_id", Reason: "Product id is required."})
	}
	if product.SellerID == "" {
		fields = append(fields, FieldViolation{Field: "seller_id", Reason: "Seller id is required."})
	}
	if product.Title == "" {
		fields = append(fields, FieldViolation{Field: "title", Reason: "Title is required."})
	} else if len(product.Title) > maxProductTitleLength {
		fields = append(fields, FieldViolation{Field: "title", Reason: fmt.Sprintf("Title must be %d characters or fewer.", maxProductTitleLength)})
	}
	if len(product.Description) < minProductDescriptionLength {
		fields = append(fields, FieldViolation{Field: "description", Reason: "Description must contain meaningful product details."})
	}
	if product.CategoryID == "" {
		fields = append(fields, FieldViolation{Field: "category_id", Reason: "Category is required."})
	}
	if product.Brand == "" && !product.GenericBrand {
		fields = append(fields, FieldViolation{Field: "brand", Reason: "Brand is required unless the product is marked generic."})
	}
	if len(product.Images) == 0 {
		fields = append(fields, FieldViolation{Field: "images", Reason: "At least one product image is required."})
	} else {
		for i, image := range product.Images {
			if strings.TrimSpace(image.URL) == "" {
				fields = append(fields, FieldViolation{Field: fmt.Sprintf("images[%d].url", i), Reason: "Image URL is required."})
			}
		}
	}
	if len(product.Variants) == 0 {
		fields = append(fields, FieldViolation{Field: "variants", Reason: "At least one product variant is required."})
	}

	seenSKU := make(map[string]struct{}, len(product.Variants))
	for i, variant := range product.Variants {
		sku := strings.TrimSpace(variant.SKU)
		if sku == "" {
			fields = append(fields, FieldViolation{Field: fmt.Sprintf("variants[%d].sku", i), Reason: "SKU is required."})
		} else if _, exists := seenSKU[strings.ToLower(sku)]; exists {
			fields = append(fields, FieldViolation{Field: fmt.Sprintf("variants[%d].sku", i), Reason: "SKU must be unique within the product."})
		} else {
			seenSKU[strings.ToLower(sku)] = struct{}{}
		}
		if variant.Price.Amount <= 0 {
			fields = append(fields, FieldViolation{Field: fmt.Sprintf("variants[%d].price.amount", i), Reason: "Price must be greater than zero."})
		}
		if !validCurrency(variant.Price.CurrencyOrDefault()) {
			fields = append(fields, FieldViolation{Field: fmt.Sprintf("variants[%d].price.currency", i), Reason: "Price currency is invalid."})
		}
		if variant.MRP.Amount < 0 {
			fields = append(fields, FieldViolation{Field: fmt.Sprintf("variants[%d].mrp.amount", i), Reason: "MRP cannot be negative."})
		}
		if variant.MRP.Amount > 0 && !validCurrency(variant.MRP.CurrencyOrDefault()) {
			fields = append(fields, FieldViolation{Field: fmt.Sprintf("variants[%d].mrp.currency", i), Reason: "MRP currency is invalid."})
		}
	}

	if len(fields) > 0 {
		return NewValidationError("Product is not ready for review.", fields)
	}
	return nil
}

func NormalizeModerationReason(reason string) string {
	reason = strings.Join(strings.Fields(strings.TrimSpace(reason)), " ")
	runes := []rune(reason)
	if len(runes) > maxModerationReasonLength {
		return string(runes[:maxModerationReasonLength])
	}
	return reason
}

func ProductEditRiskForField(field string) ProductEditRisk {
	normalized := strings.ToLower(strings.TrimSpace(field))
	switch normalized {
	case "title", "description", "category_id", "brand", "images", "attributes", "variants.sku":
		return ProductEditRiskMajor
	default:
		if strings.HasPrefix(normalized, "images.") || strings.Contains(normalized, ".images.") {
			return ProductEditRiskMajor
		}
		if strings.HasPrefix(normalized, "attributes.") || strings.Contains(normalized, ".attributes.") {
			return ProductEditRiskMajor
		}
		if strings.HasPrefix(normalized, "variants[") && strings.Contains(normalized, "].sku") {
			return ProductEditRiskMajor
		}
		return ProductEditRiskMinor
	}
}

func MajorEditRequiresReview(fields []string) bool {
	for _, field := range fields {
		if ProductEditRiskForField(field) == ProductEditRiskMajor {
			return true
		}
	}
	return false
}

func (p Product) Normalized() Product {
	p.ID = strings.TrimSpace(p.ID)
	p.SellerID = strings.TrimSpace(p.SellerID)
	p.Title = strings.TrimSpace(p.Title)
	p.Description = strings.TrimSpace(p.Description)
	p.CategoryID = strings.TrimSpace(p.CategoryID)
	p.Brand = strings.TrimSpace(p.Brand)
	p.Status = p.Status.Normalized()
	for i := range p.Images {
		p.Images[i].URL = strings.TrimSpace(p.Images[i].URL)
	}
	for i := range p.Variants {
		p.Variants[i].SKU = strings.TrimSpace(p.Variants[i].SKU)
		p.Variants[i].Price.Currency = strings.TrimSpace(strings.ToUpper(p.Variants[i].Price.Currency))
		p.Variants[i].MRP.Currency = strings.TrimSpace(strings.ToUpper(p.Variants[i].MRP.Currency))
	}
	return p
}

func (m Money) CurrencyOrDefault() string {
	currency := strings.ToUpper(strings.TrimSpace(m.Currency))
	if currency == "" {
		return "INR"
	}
	return currency
}

func validCurrency(currency string) bool {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "INR", "USD", "EUR", "GBP", "AED", "SGD":
		return true
	default:
		return false
	}
}
