package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/usecase"
)

func TestHTTPCouponValidatorPostsCartContext(t *testing.T) {
	var captured couponValidationRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/v1/coupons/validate" {
			t.Fatalf("path = %q, want /internal/v1/coupons/validate", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("Decode request error = %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"valid":true,"coupon_id":"coupon_save10","code":"SAVE10","discount":{"amount":200,"currency":"INR"},"reason":""}`))
	}))
	defer server.Close()

	validator, err := NewHTTPCouponValidator(server.URL, "/internal/v1/coupons/validate", time.Second)
	if err != nil {
		t.Fatalf("NewHTTPCouponValidator() error = %v", err)
	}

	result, err := validator.ValidateCoupon(context.Background(), usecase.CouponValidationRequest{
		Code:     " save10 ",
		UserID:   "user_123",
		CartID:   "cart_123",
		Currency: domain.CurrencyINR,
		Subtotal: domain.NewMoney(2000, domain.CurrencyINR),
		Items: []usecase.CouponCartItem{
			{
				ProductID:    "prod_123",
				VariantID:    "var_1",
				SellerID:     "seller_456",
				UnitPrice:    domain.NewMoney(1000, domain.CurrencyINR),
				Quantity:     2,
				LineSubtotal: domain.NewMoney(2000, domain.CurrencyINR),
			},
		},
		RequestedAt: time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("ValidateCoupon() error = %v", err)
	}
	if result == nil || !result.Valid || result.CouponID != "coupon_save10" || result.Discount.Amount != 200 {
		t.Fatalf("result = %#v, want valid coupon", result)
	}
	if captured.CouponCode != "SAVE10" || captured.UserID != "user_123" || captured.CartID != "cart_123" || len(captured.Items) != 1 {
		t.Fatalf("captured request = %#v, want cart context", captured)
	}
}

func TestHTTPCouponValidatorDecodesEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"valid":false,"coupon_id":"","discount":{"amount":0,"currency":"INR"},"reason":"COUPON_EXPIRED"}}`))
	}))
	defer server.Close()

	validator, err := NewHTTPCouponValidator(server.URL, "internal/v1/coupons/validate", time.Second)
	if err != nil {
		t.Fatalf("NewHTTPCouponValidator() error = %v", err)
	}

	result, err := validator.ValidateCoupon(context.Background(), usecase.CouponValidationRequest{
		Code:        "SAVE10",
		CartID:      "cart_123",
		Currency:    domain.CurrencyINR,
		Subtotal:    domain.NewMoney(2000, domain.CurrencyINR),
		RequestedAt: time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("ValidateCoupon() error = %v", err)
	}
	if result.Valid || result.Reason != "COUPON_EXPIRED" {
		t.Fatalf("result = %#v, want invalid expired coupon", result)
	}
}
