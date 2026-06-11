package domain

import "time"

type InventoryReservationStatus string

const (
	ReservationStatusReserved  InventoryReservationStatus = "reserved"
	ReservationStatusCommitted InventoryReservationStatus = "committed"
	ReservationStatusReleased  InventoryReservationStatus = "released"
	ReservationStatusExpired   InventoryReservationStatus = "expired"
)

type InventoryReservationItem struct {
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id"`
	SKU       string `json:"sku"`
	SellerID  string `json:"seller_id"`
	Quantity  int64  `json:"quantity"`
}

type InventoryReservation struct {
	ID             string                     `json:"reservation_id"`
	OrderID        string                     `json:"order_id"`
	IdempotencyKey string                     `json:"idempotency_key,omitempty"`
	Status         InventoryReservationStatus `json:"status"`
	Items          []InventoryReservationItem `json:"items"`
	ExpiresAt      time.Time                  `json:"expires_at"`
	CreatedAt      time.Time                  `json:"created_at"`
	UpdatedAt      time.Time                  `json:"updated_at"`
	CommittedAt    *time.Time                 `json:"committed_at,omitempty"`
	ReleasedAt     *time.Time                 `json:"released_at,omitempty"`
	ExpiredAt      *time.Time                 `json:"expired_at,omitempty"`
	Reason         string                     `json:"reason,omitempty"`
}

func (s InventoryReservationStatus) Valid() bool {
	switch s {
	case ReservationStatusReserved,
		ReservationStatusCommitted,
		ReservationStatusReleased,
		ReservationStatusExpired:
		return true
	default:
		return false
	}
}

func (s InventoryReservationStatus) Terminal() bool {
	switch s {
	case ReservationStatusCommitted, ReservationStatusReleased, ReservationStatusExpired:
		return true
	default:
		return false
	}
}

func (r InventoryReservation) Active() bool {
	return r.Status == ReservationStatusReserved
}

func (r InventoryReservation) Validate() ValidationReport {
	var report ValidationReport
	if !requiredString(r.ID) {
		report.AddError(CodeReservationIDRequired, "reservation_id", "reservation id is required")
	}
	if !requiredString(r.OrderID) {
		report.AddError(CodeOrderIDRequired, "order_id", "order id is required")
	}
	if !r.Status.Valid() {
		report.AddError(CodeInvalidReservationStatus, "status", "reservation status is not supported")
	}
	if len(r.Items) == 0 {
		report.AddError(CodeReservationItemsRequired, "items", "at least one reservation item is required")
	}
	for index, item := range r.Items {
		report.Merge(item.Validate("items" + indexPath(index)))
	}
	if r.ExpiresAt.IsZero() {
		report.AddError(CodeInvalidReservationTTL, "expires_at", "reservation expiry timestamp is required")
	}
	if r.CreatedAt.IsZero() {
		report.AddError(CodeInvalidReservationTTL, "created_at", "reservation creation timestamp is required")
	}
	if r.UpdatedAt.IsZero() {
		report.AddError(CodeInvalidReservationTTL, "updated_at", "reservation update timestamp is required")
	}
	return report
}

func (i InventoryReservationItem) Validate(field string) ValidationReport {
	var report ValidationReport
	if !requiredString(i.ProductID) {
		report.AddError(CodeProductIDRequired, fieldPath(field, "product_id"), "product id is required")
	}
	if !requiredString(i.VariantID) {
		report.AddError(CodeVariantIDRequired, fieldPath(field, "variant_id"), "variant id is required")
	}
	if !requiredString(i.SKU) {
		report.AddError(CodeSKURequired, fieldPath(field, "sku"), "SKU is required")
	}
	if !requiredString(i.SellerID) {
		report.AddError(CodeSellerRequired, fieldPath(field, "seller_id"), "seller id is required")
	}
	if i.Quantity <= 0 {
		report.AddError(CodeInvalidReservationItem, fieldPath(field, "quantity"), "reservation quantity must be greater than zero")
	}
	return report
}

func indexPath(index int) string {
	return "[" + itoa(index) + "]"
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	position := len(digits)
	for value > 0 {
		position--
		digits[position] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[position:])
}
