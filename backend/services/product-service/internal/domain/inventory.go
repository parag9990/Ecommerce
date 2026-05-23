package domain

type InventoryState struct {
	StockQuantity    int64 `json:"stock_quantity"`
	ReservedQuantity int64 `json:"reserved_quantity"`
	SafetyStock      int64 `json:"safety_stock"`
}

func (i InventoryState) AvailableQuantity() int64 {
	return i.StockQuantity - i.ReservedQuantity - i.SafetyStock
}

func (i InventoryState) CanFulfill(quantity int64) bool {
	return quantity > 0 && i.AvailableQuantity() >= quantity
}

func (i InventoryState) Validate(field string) ValidationReport {
	var report ValidationReport
	if i.StockQuantity < 0 {
		report.AddError(CodeInvalidStock, fieldPath(field, "stock_quantity"), "stock quantity cannot be negative")
	}
	if i.ReservedQuantity < 0 {
		report.AddError(CodeInvalidReservedQuantity, fieldPath(field, "reserved_quantity"), "reserved quantity cannot be negative")
	}
	if i.SafetyStock < 0 {
		report.AddError(CodeInvalidSafetyStock, fieldPath(field, "safety_stock"), "safety stock cannot be negative")
	}
	if i.ReservedQuantity > i.StockQuantity {
		report.AddError(CodeInvalidReservedQuantity, fieldPath(field, "reserved_quantity"), "reserved quantity cannot exceed stock quantity")
	}
	if i.AvailableQuantity() < 0 {
		report.AddError(CodeInvalidAvailableQuantity, fieldPath(field, "available_quantity"), "available quantity cannot be negative")
	}
	return report
}
