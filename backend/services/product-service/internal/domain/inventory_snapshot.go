package domain

import "time"

type InventorySnapshotType string

const (
	InventorySnapshotTypePeriodic         InventorySnapshotType = "periodic"
	InventorySnapshotTypeManualAdjustment InventorySnapshotType = "manual_adjustment"
	InventorySnapshotTypeReservation      InventorySnapshotType = "reservation"
	InventorySnapshotTypeRelease          InventorySnapshotType = "release"
	InventorySnapshotTypeOrderCommit      InventorySnapshotType = "order_commit"
	InventorySnapshotTypeCorrection       InventorySnapshotType = "correction"
)

type InventoryReference struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	OrderID string `json:"order_id,omitempty"`
}

type InventorySnapshot struct {
	ID                string                `json:"inventory_snapshot_id"`
	ProductID         string                `json:"product_id"`
	VariantID         string                `json:"variant_id"`
	SKU               string                `json:"sku"`
	SellerID          string                `json:"seller_id"`
	SnapshotType      InventorySnapshotType `json:"snapshot_type"`
	StockQuantity     int64                 `json:"stock_quantity"`
	ReservedQuantity  int64                 `json:"reserved_quantity"`
	SafetyStock       int64                 `json:"safety_stock,omitempty"`
	AvailableQuantity int64                 `json:"available_quantity"`
	Reason            string                `json:"reason,omitempty"`
	Reference         *InventoryReference   `json:"reference,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
}

func NewInventorySnapshot(
	id string,
	productID string,
	variant Variant,
	sellerID string,
	snapshotType InventorySnapshotType,
	reason string,
	reference *InventoryReference,
	now time.Time,
) InventorySnapshot {
	state := variant.InventoryState()
	return InventorySnapshot{
		ID:                id,
		ProductID:         productID,
		VariantID:         variant.ID,
		SKU:               variant.SKU,
		SellerID:          sellerID,
		SnapshotType:      snapshotType,
		StockQuantity:     state.StockQuantity,
		ReservedQuantity:  state.ReservedQuantity,
		SafetyStock:       state.SafetyStock,
		AvailableQuantity: state.AvailableQuantity(),
		Reason:            reason,
		Reference:         reference,
		CreatedAt:         now.UTC(),
	}
}

func (s InventorySnapshotType) Valid() bool {
	switch s {
	case InventorySnapshotTypePeriodic,
		InventorySnapshotTypeManualAdjustment,
		InventorySnapshotTypeReservation,
		InventorySnapshotTypeRelease,
		InventorySnapshotTypeOrderCommit,
		InventorySnapshotTypeCorrection:
		return true
	default:
		return false
	}
}

func (s InventorySnapshot) Validate() ValidationReport {
	var report ValidationReport
	if !requiredString(s.ID) {
		report.AddError(CodeInventorySnapshotIDRequired, "inventory_snapshot_id", "inventory snapshot id is required")
	}
	if !requiredString(s.ProductID) {
		report.AddError(CodeProductIDRequired, "product_id", "product id is required")
	}
	if !requiredString(s.VariantID) {
		report.AddError(CodeVariantIDRequired, "variant_id", "variant id is required")
	}
	if !requiredString(s.SKU) {
		report.AddError(CodeSKURequired, "sku", "SKU is required")
	}
	if !requiredString(s.SellerID) {
		report.AddError(CodeSellerRequired, "seller_id", "seller id is required")
	}
	if !s.SnapshotType.Valid() {
		report.AddError(CodeInvalidInventorySnapshotType, "snapshot_type", "inventory snapshot type is not supported")
	}
	state := InventoryState{
		StockQuantity:    s.StockQuantity,
		ReservedQuantity: s.ReservedQuantity,
		SafetyStock:      s.SafetyStock,
	}
	report.Merge(state.Validate(""))
	if s.AvailableQuantity != state.AvailableQuantity() {
		report.AddError(CodeInvalidAvailableQuantity, "available_quantity", "available quantity must equal stock minus reserved minus safety stock")
	}
	if s.Reference != nil {
		if !requiredString(s.Reference.Type) {
			report.AddError(CodeInvalidInventorySnapshot, "reference.type", "inventory snapshot reference type is required when reference is present")
		}
		if !requiredString(s.Reference.ID) {
			report.AddError(CodeInvalidInventorySnapshot, "reference.id", "inventory snapshot reference id is required when reference is present")
		}
	}
	if s.CreatedAt.IsZero() {
		report.AddError(CodeInvalidInventorySnapshot, "created_at", "inventory snapshot creation timestamp is required")
	}
	return report
}
