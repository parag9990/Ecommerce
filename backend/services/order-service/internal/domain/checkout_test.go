package domain

import (
	"errors"
	"testing"
	"time"
)

func TestBuildCreatedOrderUsesFreshProductSnapshots(t *testing.T) {
	input := validBuildCreatedOrderInput()
	input.Cart.Currency = " inr "
	input.Products[0].Currency = "INR"
	input.Products[0].UnitAmount = 299950

	order, err := BuildCreatedOrder(input)
	if err != nil {
		t.Fatalf("BuildCreatedOrder() error = %v", err)
	}
	if order.Status != OrderStatusCreated || order.Currency != "INR" {
		t.Fatalf("order status/currency = %q/%q, want created/INR", order.Status, order.Currency)
	}
	if order.TotalAmount != 599900 || order.Items[0].UnitAmount != 299950 {
		t.Fatalf("order total/unit amount = %d/%d, want 599900/299950", order.TotalAmount, order.Items[0].UnitAmount)
	}
	if order.Items[0].TitleSnapshot != "Running Shoes" || order.InitialHistory.ToStatus != OrderStatusCreated {
		t.Fatalf("fresh item snapshot or initial status history was not created: %+v", order)
	}
}

func TestValidateCartForCheckoutRejectsDuplicateRows(t *testing.T) {
	cart := validBuildCreatedOrderInput().Cart
	cart.Items = append(cart.Items, cart.Items[0])

	err := ValidateCartForCheckout(cart, cart.UserID)
	if !errors.Is(err, ErrDuplicateCartItem) {
		t.Fatalf("ValidateCartForCheckout() error = %v, want ErrDuplicateCartItem", err)
	}
}

func TestBuildCreatedOrderRejectsInvalidCheckoutProducts(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*BuildCreatedOrderInput)
		want   error
	}{
		{
			name: "currency mismatch",
			mutate: func(input *BuildCreatedOrderInput) {
				input.Products[0].Currency = "USD"
			},
			want: ErrCurrencyMismatch,
		},
		{
			name: "unpublished product",
			mutate: func(input *BuildCreatedOrderInput) {
				input.Products[0].Published = false
			},
			want: ErrProductUnavailable,
		},
		{
			name: "inactive variant",
			mutate: func(input *BuildCreatedOrderInput) {
				input.Products[0].VariantActive = false
			},
			want: ErrVariantUnavailable,
		},
		{
			name: "stock unavailable",
			mutate: func(input *BuildCreatedOrderInput) {
				input.Products[0].InStock = false
			},
			want: ErrInventoryUnavailable,
		},
		{
			name: "overflow total",
			mutate: func(input *BuildCreatedOrderInput) {
				input.Products[0].UnitAmount = maxInt64
			},
			want: ErrInvalidOrderTotal,
		},
		{
			name: "invalid shipping address",
			mutate: func(input *BuildCreatedOrderInput) {
				input.Address.Line1 = ""
			},
			want: ErrInvalidAddress,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validBuildCreatedOrderInput()
			test.mutate(&input)
			_, err := BuildCreatedOrder(input)
			if !errors.Is(err, test.want) {
				t.Fatalf("BuildCreatedOrder() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestValidateForCreationRejectsTamperedTotal(t *testing.T) {
	order, err := BuildCreatedOrder(validBuildCreatedOrderInput())
	if err != nil {
		t.Fatalf("BuildCreatedOrder() error = %v", err)
	}
	order.TotalAmount++

	err = order.ValidateForCreation()
	if !errors.Is(err, ErrInvalidOrderTotal) {
		t.Fatalf("ValidateForCreation() error = %v, want ErrInvalidOrderTotal", err)
	}
}

func TestValidateForCreationRequiresDurableInventoryReservation(t *testing.T) {
	order, err := BuildCreatedOrder(validBuildCreatedOrderInput())
	if err != nil {
		t.Fatalf("BuildCreatedOrder() error = %v", err)
	}

	if err := order.ValidateForCreation(); !errors.Is(err, ErrInvalidReservation) {
		t.Fatalf("ValidateForCreation() error = %v, want ErrInvalidReservation", err)
	}

	now := time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC)
	if err := order.AttachInventoryReservation("res_1", now.Add(time.Minute), now); err != nil {
		t.Fatalf("AttachInventoryReservation() error = %v", err)
	}
	if err := order.ValidateForCreation(); err != nil {
		t.Fatalf("ValidateForCreation() error = %v after reservation attachment", err)
	}
}

func validBuildCreatedOrderInput() BuildCreatedOrderInput {
	return BuildCreatedOrderInput{
		OrderID:      "ord_1",
		HistoryID:    "osh_1",
		OrderItemIDs: []string{"oi_1"},
		UserID:       "user_1",
		Cart: CartSnapshot{
			CartID:   "cart_1",
			UserID:   "user_1",
			Currency: "INR",
			Items: []CartItemSnapshot{{
				ProductID: "product_1",
				VariantID: "variant_1",
				Quantity:  2,
			}},
		},
		Products: []ProductSnapshot{{
			ProductID:     "product_1",
			VariantID:     "variant_1",
			SellerID:      "seller_1",
			SKU:           "SKU-1",
			Title:         "Running Shoes",
			Currency:      "INR",
			UnitAmount:    1000,
			Published:     true,
			VariantActive: true,
			InStock:       true,
		}},
		Address: AddressSnapshot{
			RecipientName: "Buyer",
			Line1:         "101 Main Street",
			City:          "Delhi",
			PostalCode:    "110001",
			Country:       "IN",
		},
		CreatedAt: time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC),
	}
}
