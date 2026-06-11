package domain

import (
	"fmt"
	"strings"
	"time"
)

const (
	maxPublicIDLength     = 64
	maxSKULength          = 128
	maxTitleLength        = 512
	maxImageURLLength     = 1024
	maxCouponCodeLength   = 64
	maxStatusReasonLength = 512
	maxRecipientLength    = 256
	maxAddressLineLength  = 512
	maxAddressCityLength  = 128
	maxAddressStateLength = 128
	maxPostalCodeLength   = 32
	maxPhoneLength        = 32
	maxInt64              = int64(1<<63 - 1)
)

type FulfillmentStatus string

const (
	FulfillmentStatusPending   FulfillmentStatus = "pending"
	FulfillmentStatusPacked    FulfillmentStatus = "packed"
	FulfillmentStatusShipped   FulfillmentStatus = "shipped"
	FulfillmentStatusDelivered FulfillmentStatus = "delivered"
	FulfillmentStatusCancelled FulfillmentStatus = "cancelled"
	FulfillmentStatusReturned  FulfillmentStatus = "returned"
)

// AddressSnapshot is copied into an order so subsequent profile edits never
// alter the shipping destination associated with the purchase.
type AddressSnapshot struct {
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone,omitempty"`
	Line1         string `json:"line1"`
	Line2         string `json:"line2,omitempty"`
	City          string `json:"city"`
	State         string `json:"state,omitempty"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
}

type CartSnapshot struct {
	CartID   string
	UserID   string
	Currency string
	Items    []CartItemSnapshot
}

type CartItemSnapshot struct {
	ProductID string
	VariantID string
	Quantity  int32
}

// ProductSnapshot contains only checkout-authoritative fields returned by the
// Product Service. UnitAmount is stored in minor currency units.
type ProductSnapshot struct {
	ProductID     string
	VariantID     string
	SellerID      string
	SKU           string
	Title         string
	ImageURL      string
	Currency      string
	UnitAmount    int64
	Published     bool
	VariantActive bool
	InStock       bool
}

type OrderItem struct {
	OrderItemID       string
	OrderID           string
	SellerID          string
	ProductID         string
	VariantID         string
	SKU               string
	TitleSnapshot     string
	ImageURLSnapshot  string
	Quantity          int32
	Currency          string
	UnitAmount        int64
	DiscountAmount    int64
	TaxAmount         int64
	TotalAmount       int64
	FulfillmentStatus FulfillmentStatus
}

type Order struct {
	OrderID                string
	UserID                 string
	CartID                 string
	Status                 OrderStatus
	Currency               string
	SubtotalAmount         int64
	DiscountAmount         int64
	ShippingAmount         int64
	TaxAmount              int64
	TotalAmount            int64
	AddressSnapshot        AddressSnapshot
	CouponCode             string
	PaymentID              string
	InventoryReservationID string
	InventoryReservedUntil time.Time
	Items                  []OrderItem
	InitialHistory         OrderStatusHistoryEntry
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type BuildCreatedOrderInput struct {
	OrderID      string
	HistoryID    string
	OrderItemIDs []string
	UserID       string
	Cart         CartSnapshot
	Products     []ProductSnapshot
	Address      AddressSnapshot
	CouponCode   string
	CreatedAt    time.Time
}

func ValidateCartForCheckout(cart CartSnapshot, userID string) error {
	userID = strings.TrimSpace(userID)
	if strings.TrimSpace(cart.CartID) == "" {
		return ErrCartNotFound
	}
	if strings.TrimSpace(cart.UserID) != userID {
		return ErrCartForbidden
	}
	if len(cart.Items) == 0 {
		return ErrCartEmpty
	}
	if _, err := normalizeCurrency(cart.Currency); err != nil {
		return err
	}

	seen := make(map[string]struct{}, len(cart.Items))
	for _, item := range cart.Items {
		if err := validateCartItem(item); err != nil {
			return err
		}
		key := productKey(item.ProductID, item.VariantID)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("%w: product_id=%q variant_id=%q", ErrDuplicateCartItem, item.ProductID, item.VariantID)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func BuildCreatedOrder(input BuildCreatedOrderInput) (Order, error) {
	if err := ValidateCartForCheckout(input.Cart, input.UserID); err != nil {
		return Order{}, err
	}
	if err := validateRequiredID("order_id", input.OrderID); err != nil {
		return Order{}, err
	}
	if err := validateRequiredID("history_id", input.HistoryID); err != nil {
		return Order{}, err
	}
	if len(input.OrderItemIDs) != len(input.Cart.Items) {
		return Order{}, fmt.Errorf("%w: order item id count does not match cart item count", ErrInvalidOrder)
	}

	currency, _ := normalizeCurrency(input.Cart.Currency)
	address := normalizeAddress(input.Address)
	if err := address.Validate(); err != nil {
		return Order{}, err
	}
	couponCode := strings.TrimSpace(input.CouponCode)
	if len(couponCode) > maxCouponCodeLength {
		return Order{}, fmt.Errorf("%w: coupon_code exceeds %d characters", ErrInvalidOrder, maxCouponCodeLength)
	}

	productIndex, err := indexProducts(input.Products)
	if err != nil {
		return Order{}, err
	}
	order := Order{
		OrderID:         strings.TrimSpace(input.OrderID),
		UserID:          strings.TrimSpace(input.UserID),
		CartID:          strings.TrimSpace(input.Cart.CartID),
		Status:          OrderStatusCreated,
		Currency:        currency,
		AddressSnapshot: address,
		CouponCode:      couponCode,
		Items:           make([]OrderItem, 0, len(input.Cart.Items)),
	}

	for index, cartItem := range input.Cart.Items {
		product, exists := productIndex[productKey(cartItem.ProductID, cartItem.VariantID)]
		if !exists {
			return Order{}, fmt.Errorf("%w: product_id=%q variant_id=%q", ErrProductUnavailable, cartItem.ProductID, cartItem.VariantID)
		}
		if err := validateProductForCheckout(product, currency); err != nil {
			return Order{}, err
		}
		if err := validateRequiredID("order_item_id", input.OrderItemIDs[index]); err != nil {
			return Order{}, err
		}

		lineAmount, err := checkedMultiply(product.UnitAmount, int64(cartItem.Quantity))
		if err != nil {
			return Order{}, err
		}
		orderItem := OrderItem{
			OrderItemID:       strings.TrimSpace(input.OrderItemIDs[index]),
			OrderID:           order.OrderID,
			SellerID:          strings.TrimSpace(product.SellerID),
			ProductID:         strings.TrimSpace(product.ProductID),
			VariantID:         strings.TrimSpace(product.VariantID),
			SKU:               strings.TrimSpace(product.SKU),
			TitleSnapshot:     strings.TrimSpace(product.Title),
			ImageURLSnapshot:  strings.TrimSpace(product.ImageURL),
			Quantity:          cartItem.Quantity,
			Currency:          currency,
			UnitAmount:        product.UnitAmount,
			TotalAmount:       lineAmount,
			FulfillmentStatus: FulfillmentStatusPending,
		}
		order.Items = append(order.Items, orderItem)
		order.SubtotalAmount, err = checkedAdd(order.SubtotalAmount, lineAmount)
		if err != nil {
			return Order{}, err
		}
	}
	order.TotalAmount = order.SubtotalAmount
	order.InitialHistory = OrderStatusHistoryEntry{
		ID:        strings.TrimSpace(input.HistoryID),
		OrderID:   order.OrderID,
		ToStatus:  OrderStatusCreated,
		Reason:    "cart_to_order_created",
		ActorType: OrderStatusActorBuyer,
		ActorID:   order.UserID,
		CreatedAt: input.CreatedAt.UTC(),
	}
	if err := order.validateCreatedSnapshot(); err != nil {
		return Order{}, err
	}
	return order, nil
}

func (o Order) ValidateForCreation() error {
	if err := o.validateCreatedSnapshot(); err != nil {
		return err
	}
	if strings.TrimSpace(o.PaymentID) != "" {
		return fmt.Errorf("%w: a created order cannot already have a payment", ErrInvalidOrder)
	}
	if strings.TrimSpace(o.InventoryReservationID) == "" ||
		len(strings.TrimSpace(o.InventoryReservationID)) > maxPublicIDLength ||
		o.InventoryReservedUntil.IsZero() ||
		!o.InventoryReservedUntil.After(o.InitialHistory.CreatedAt) {
		return ErrInvalidReservation
	}
	return nil
}

func (o *Order) AttachInventoryReservation(reservationID string, reservedUntil time.Time, now time.Time) error {
	reservationID = strings.TrimSpace(reservationID)
	if reservationID == "" || len(reservationID) > maxPublicIDLength ||
		reservedUntil.IsZero() || !reservedUntil.After(now) {
		return ErrInvalidReservation
	}
	o.InventoryReservationID = reservationID
	o.InventoryReservedUntil = reservedUntil.UTC()
	return nil
}

func (o Order) validateCreatedSnapshot() error {
	if err := validateRequiredID("order_id", o.OrderID); err != nil {
		return err
	}
	if err := validateRequiredID("user_id", o.UserID); err != nil {
		return err
	}
	if strings.TrimSpace(o.CartID) == "" || len(strings.TrimSpace(o.CartID)) > maxPublicIDLength {
		return fmt.Errorf("%w: invalid cart_id", ErrInvalidOrder)
	}
	currency, err := normalizeCurrency(o.Currency)
	if err != nil || currency != o.Currency {
		return fmt.Errorf("%w: invalid order currency", ErrInvalidOrder)
	}
	if o.Status != OrderStatusCreated || len(o.Items) == 0 {
		return fmt.Errorf("%w: created order status and items are required", ErrInvalidOrder)
	}
	if err := o.AddressSnapshot.Validate(); err != nil {
		return err
	}

	var subtotal int64
	for _, item := range o.Items {
		if err := validateCreatedOrderItem(item, o.OrderID, o.Currency); err != nil {
			return err
		}
		lineAmount, lineErr := checkedMultiply(item.UnitAmount, int64(item.Quantity))
		if lineErr != nil {
			return lineErr
		}
		subtotal, err = checkedAdd(subtotal, lineAmount)
		if err != nil {
			return err
		}
	}
	if o.SubtotalAmount != subtotal || o.DiscountAmount != 0 || o.ShippingAmount != 0 || o.TaxAmount != 0 || o.TotalAmount != subtotal {
		return fmt.Errorf("%w: Task 3 order totals do not match fresh product snapshots", ErrInvalidOrderTotal)
	}
	if err := validateInitialHistory(o.InitialHistory, o); err != nil {
		return err
	}
	return nil
}

func (a AddressSnapshot) Validate() error {
	if strings.TrimSpace(a.RecipientName) == "" ||
		strings.TrimSpace(a.Line1) == "" ||
		strings.TrimSpace(a.City) == "" ||
		strings.TrimSpace(a.PostalCode) == "" ||
		strings.TrimSpace(a.Country) == "" {
		return fmt.Errorf("%w: recipient_name, line1, city, postal_code, and country are required", ErrInvalidAddress)
	}
	if len(strings.TrimSpace(a.RecipientName)) > maxRecipientLength ||
		len(strings.TrimSpace(a.Phone)) > maxPhoneLength ||
		len(strings.TrimSpace(a.Line1)) > maxAddressLineLength ||
		len(strings.TrimSpace(a.Line2)) > maxAddressLineLength ||
		len(strings.TrimSpace(a.City)) > maxAddressCityLength ||
		len(strings.TrimSpace(a.State)) > maxAddressStateLength ||
		len(strings.TrimSpace(a.PostalCode)) > maxPostalCodeLength ||
		len(strings.TrimSpace(a.Country)) != 2 {
		return fmt.Errorf("%w: shipping address exceeds allowed field limits or country is not an ISO alpha-2 code", ErrInvalidAddress)
	}
	return nil
}

func validateCartItem(item CartItemSnapshot) error {
	if strings.TrimSpace(item.ProductID) == "" || strings.TrimSpace(item.VariantID) == "" {
		return ErrInvalidCartItem
	}
	if item.Quantity <= 0 {
		return ErrInvalidQuantity
	}
	return nil
}

func indexProducts(products []ProductSnapshot) (map[string]ProductSnapshot, error) {
	index := make(map[string]ProductSnapshot, len(products))
	for _, product := range products {
		key := productKey(product.ProductID, product.VariantID)
		if strings.TrimSpace(product.ProductID) == "" || strings.TrimSpace(product.VariantID) == "" {
			return nil, ErrProductUnavailable
		}
		if _, exists := index[key]; exists {
			return nil, fmt.Errorf("%w: duplicate Product Service snapshot", ErrProductUnavailable)
		}
		index[key] = product
	}
	return index, nil
}

func validateProductForCheckout(product ProductSnapshot, currency string) error {
	if !product.Published {
		return ErrProductUnavailable
	}
	if !product.VariantActive {
		return ErrVariantUnavailable
	}
	if !product.InStock {
		return ErrInventoryUnavailable
	}
	if strings.TrimSpace(product.SellerID) == "" {
		return ErrMissingSeller
	}
	productCurrency, err := normalizeCurrency(product.Currency)
	if err != nil || productCurrency != currency {
		return ErrCurrencyMismatch
	}
	if product.UnitAmount < 0 {
		return ErrInvalidPrice
	}
	if strings.TrimSpace(product.SKU) == "" || len(strings.TrimSpace(product.SKU)) > maxSKULength ||
		strings.TrimSpace(product.Title) == "" || len(strings.TrimSpace(product.Title)) > maxTitleLength ||
		len(strings.TrimSpace(product.ImageURL)) > maxImageURLLength {
		return ErrProductUnavailable
	}
	return nil
}

func validateCreatedOrderItem(item OrderItem, orderID string, currency string) error {
	if err := validateRequiredID("order_item_id", item.OrderItemID); err != nil {
		return err
	}
	if item.OrderID != orderID || item.Currency != currency || item.Quantity <= 0 ||
		item.UnitAmount < 0 || item.DiscountAmount != 0 || item.TaxAmount != 0 ||
		item.FulfillmentStatus != FulfillmentStatusPending {
		return fmt.Errorf("%w: invalid created order item", ErrInvalidOrder)
	}
	if strings.TrimSpace(item.SellerID) == "" || len(strings.TrimSpace(item.SellerID)) > maxPublicIDLength ||
		strings.TrimSpace(item.ProductID) == "" || len(strings.TrimSpace(item.ProductID)) > maxPublicIDLength ||
		strings.TrimSpace(item.VariantID) == "" || len(strings.TrimSpace(item.VariantID)) > maxPublicIDLength ||
		strings.TrimSpace(item.SKU) == "" || len(strings.TrimSpace(item.SKU)) > maxSKULength ||
		strings.TrimSpace(item.TitleSnapshot) == "" || len(strings.TrimSpace(item.TitleSnapshot)) > maxTitleLength ||
		len(strings.TrimSpace(item.ImageURLSnapshot)) > maxImageURLLength {
		return fmt.Errorf("%w: invalid order item snapshot fields", ErrInvalidOrder)
	}
	lineAmount, err := checkedMultiply(item.UnitAmount, int64(item.Quantity))
	if err != nil || item.TotalAmount != lineAmount {
		return ErrInvalidOrderTotal
	}
	return nil
}

func validateInitialHistory(history OrderStatusHistoryEntry, order Order) error {
	if err := validateRequiredID("history_id", history.ID); err != nil {
		return err
	}
	if history.OrderID != order.OrderID || history.FromStatus != nil ||
		history.ToStatus != OrderStatusCreated || history.ActorType != OrderStatusActorBuyer ||
		history.ActorID != order.UserID || strings.TrimSpace(history.Reason) == "" ||
		len(history.Reason) > maxStatusReasonLength || history.CreatedAt.IsZero() {
		return fmt.Errorf("%w: invalid initial status history", ErrInvalidOrder)
	}
	return nil
}

func validateRequiredID(name string, value string) error {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxPublicIDLength {
		return fmt.Errorf("%w: invalid %s", ErrInvalidOrder, name)
	}
	return nil
}

func normalizeAddress(address AddressSnapshot) AddressSnapshot {
	address.RecipientName = strings.TrimSpace(address.RecipientName)
	address.Phone = strings.TrimSpace(address.Phone)
	address.Line1 = strings.TrimSpace(address.Line1)
	address.Line2 = strings.TrimSpace(address.Line2)
	address.City = strings.TrimSpace(address.City)
	address.State = strings.TrimSpace(address.State)
	address.PostalCode = strings.TrimSpace(address.PostalCode)
	address.Country = strings.ToUpper(strings.TrimSpace(address.Country))
	return address
}

func normalizeCurrency(currency string) (string, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if len(currency) != 3 {
		return "", ErrInvalidCurrency
	}
	for _, char := range currency {
		if char < 'A' || char > 'Z' {
			return "", ErrInvalidCurrency
		}
	}
	return currency, nil
}

func productKey(productID string, variantID string) string {
	return strings.TrimSpace(productID) + ":" + strings.TrimSpace(variantID)
}

func checkedMultiply(amount int64, quantity int64) (int64, error) {
	if amount < 0 || quantity <= 0 || (amount != 0 && quantity > maxInt64/amount) {
		return 0, ErrInvalidOrderTotal
	}
	return amount * quantity, nil
}

func checkedAdd(left int64, right int64) (int64, error) {
	if left < 0 || right < 0 || left > maxInt64-right {
		return 0, ErrInvalidOrderTotal
	}
	return left + right, nil
}
