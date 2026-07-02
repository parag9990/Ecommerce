package domain

import "fmt"

type VariantStatus string

const (
	VariantStatusActive     VariantStatus = "active"
	VariantStatusInactive   VariantStatus = "inactive"
	VariantStatusOutOfStock VariantStatus = "out_of_stock"
)

type Variant struct {
	ID               string        `json:"variant_id"`
	SKU              string        `json:"sku"`
	Title            string        `json:"title,omitempty"`
	Attributes       Attributes    `json:"attributes"`
	Price            Money         `json:"price"`
	MRP              *Money        `json:"mrp,omitempty"`
	StockQuantity    int64         `json:"stock_quantity"`
	ReservedQuantity int64         `json:"reserved_quantity"`
	SafetyStock      int64         `json:"safety_stock"`
	Status           VariantStatus `json:"status"`
	Barcode          string        `json:"barcode,omitempty"`
}

func (s VariantStatus) Valid() bool {
	switch s {
	case VariantStatusActive, VariantStatusInactive, VariantStatusOutOfStock:
		return true
	default:
		return false
	}
}

func (v Variant) Active() bool {
	return v.Status == VariantStatusActive
}

func (v Variant) InventoryState() InventoryState {
	return InventoryState{
		StockQuantity:    v.StockQuantity,
		ReservedQuantity: v.ReservedQuantity,
		SafetyStock:      v.SafetyStock,
	}
}

func (v Variant) AvailableQuantity() int64 {
	return v.InventoryState().AvailableQuantity()
}

func (v Variant) Validate(field string, options ValidationOptions, schema []AttributeDefinition) ValidationReport {
	var report ValidationReport
	if !requiredString(v.ID) {
		report.AddError(CodeVariantIDRequired, fieldPath(field, "variant_id"), "variant id is required")
	}
	if !requiredString(v.SKU) {
		report.AddError(CodeSKURequired, fieldPath(field, "sku"), "variant SKU is required")
	}
	report.Merge(v.Price.Validate(fieldPath(field, "price")))
	if v.MRP != nil {
		report.Merge(v.MRP.Validate(fieldPath(field, "mrp")))
		if v.MRP.Currency == v.Price.Currency && v.MRP.Amount < v.Price.Amount {
			report.AddError(CodeInvalidMRP, fieldPath(field, "mrp.amount"), "MRP cannot be lower than sale price")
		}
	}
	report.Merge(v.InventoryState().Validate(field))
	if !v.Status.Valid() {
		report.AddError(CodeInvalidVariantStatus, fieldPath(field, "status"), "variant status is not supported")
	}
	if len(schema) > 0 {
		report.Merge(ValidateAttributes(v.Attributes, schema, AttributeScopeVariant, options, fieldPath(field, "attributes")))
	}
	if _, err := AttributeSignature(v.Attributes); err != nil {
		report.AddError(CodeInvalidAttributeValue, fieldPath(field, "attributes"), fmt.Sprintf("variant attributes cannot be normalized: %v", err))
	}
	return report
}
