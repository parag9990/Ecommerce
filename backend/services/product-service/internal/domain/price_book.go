package domain

import (
	"fmt"
	"time"
)

type PriceBookStatus string

const (
	PriceBookStatusDraft    PriceBookStatus = "draft"
	PriceBookStatusActive   PriceBookStatus = "active"
	PriceBookStatusExpired  PriceBookStatus = "expired"
	PriceBookStatusArchived PriceBookStatus = "archived"
)

type PriceBook struct {
	ID        string           `json:"price_book_id"`
	SellerID  string           `json:"seller_id"`
	Name      string           `json:"name"`
	Currency  string           `json:"currency"`
	Status    PriceBookStatus  `json:"status"`
	Priority  int              `json:"priority"`
	StartsAt  *time.Time       `json:"starts_at,omitempty"`
	EndsAt    *time.Time       `json:"ends_at,omitempty"`
	Entries   []PriceBookEntry `json:"entries"`
	CreatedBy string           `json:"created_by,omitempty"`
	UpdatedBy string           `json:"updated_by,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type PriceBookEntry struct {
	ID          string `json:"entry_id"`
	ProductID   string `json:"product_id"`
	VariantID   string `json:"variant_id"`
	SKU         string `json:"sku"`
	Price       Money  `json:"price"`
	MRP         *Money `json:"mrp,omitempty"`
	MinQuantity int64  `json:"min_quantity,omitempty"`
}

func (s PriceBookStatus) Valid() bool {
	switch s {
	case PriceBookStatusDraft, PriceBookStatusActive, PriceBookStatusExpired, PriceBookStatusArchived:
		return true
	default:
		return false
	}
}

func (p PriceBook) ActiveAt(now time.Time) bool {
	if p.Status != PriceBookStatusActive {
		return false
	}
	if p.StartsAt != nil && now.Before(*p.StartsAt) {
		return false
	}
	return p.EndsAt == nil || !now.After(*p.EndsAt)
}

func (p PriceBook) Validate() ValidationReport {
	var report ValidationReport
	if !requiredString(p.ID) {
		report.AddError(CodePriceBookIDRequired, "price_book_id", "price book id is required")
	}
	if !requiredString(p.SellerID) {
		report.AddError(CodeSellerRequired, "seller_id", "seller id is required")
	}
	if !requiredString(p.Name) {
		report.AddError(CodeTitleRequired, "name", "price book name is required")
	}
	if !isCurrencyCode(p.Currency) {
		report.AddError(CodeInvalidCurrency, "currency", "price book currency must be a three-letter ISO-style code")
	}
	if !p.Status.Valid() {
		report.AddError(CodeInvalidPriceBookStatus, "status", "price book status is not supported")
	}
	if p.StartsAt != nil && p.EndsAt != nil && !p.StartsAt.Before(*p.EndsAt) {
		report.AddError(CodeInvalidPriceBookWindow, "ends_at", "price book end time must be after start time")
	}
	if len(p.Entries) == 0 {
		report.AddError(CodePriceBookEntryRequired, "entries", "price book requires at least one price entry")
	}
	entryIDs := make(map[string]struct{}, len(p.Entries))
	targets := make(map[string]struct{}, len(p.Entries))
	for index, entry := range p.Entries {
		field := indexField("entries", index)
		report.Merge(entry.Validate(field, p.Currency))
		if entry.ID != "" {
			if _, exists := entryIDs[entry.ID]; exists {
				report.AddError(CodePriceBookEntryRequired, fieldPath(field, "entry_id"), "price book entry id must be unique")
			}
			entryIDs[entry.ID] = struct{}{}
		}
		target := entry.ProductID + ":" + entry.VariantID + ":" + entry.SKU
		if target != "::" {
			if _, exists := targets[target]; exists {
				report.AddError(CodeDuplicatePriceBookEntry, field, "price book cannot contain duplicate variant targets")
			}
			targets[target] = struct{}{}
		}
	}
	if p.CreatedAt.IsZero() {
		report.AddError(CodeInvalidPriceBookWindow, "created_at", "price book creation timestamp is required")
	}
	if p.UpdatedAt.IsZero() {
		report.AddError(CodeInvalidPriceBookWindow, "updated_at", "price book update timestamp is required")
	}
	return report
}

func (e PriceBookEntry) Validate(field string, priceBookCurrency string) ValidationReport {
	var report ValidationReport
	if !requiredString(e.ID) {
		report.AddError(CodePriceBookEntryRequired, fieldPath(field, "entry_id"), "price book entry id is required")
	}
	if !requiredString(e.ProductID) {
		report.AddError(CodeProductIDRequired, fieldPath(field, "product_id"), "product id is required")
	}
	if !requiredString(e.VariantID) {
		report.AddError(CodeVariantIDRequired, fieldPath(field, "variant_id"), "variant id is required")
	}
	if !requiredString(e.SKU) {
		report.AddError(CodeSKURequired, fieldPath(field, "sku"), "SKU is required")
	}
	report.Merge(e.Price.Validate(fieldPath(field, "price")))
	if isCurrencyCode(priceBookCurrency) && e.Price.Currency != normalizeCurrency(priceBookCurrency) {
		report.AddError(CodeInvalidCurrency, fieldPath(field, "price.currency"), "entry price currency must match price book currency")
	}
	if e.MRP != nil {
		report.Merge(e.MRP.Validate(fieldPath(field, "mrp")))
		if e.MRP.Currency == e.Price.Currency && e.MRP.Amount < e.Price.Amount {
			report.AddError(CodeInvalidMRP, fieldPath(field, "mrp.amount"), "MRP cannot be lower than sale price")
		}
	}
	if e.MinQuantity < 0 {
		report.AddError(CodeInvalidPrice, fieldPath(field, "min_quantity"), "minimum quantity cannot be negative")
	}
	return report
}

func indexField(field string, index int) string {
	return fmt.Sprintf("%s[%d]", field, index)
}
