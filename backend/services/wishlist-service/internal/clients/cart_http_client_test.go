package clients

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/usecase"
)

func TestHTTPCartClientAddItemCallsCartServiceAndDecodesCart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/api/v1/cart/items" {
			t.Fatalf("path = %q, want /api/v1/cart/items", r.URL.Path)
		}
		if r.Header.Get("X-User-ID") != "user_123" {
			t.Fatalf("X-User-ID = %q, want user_123", r.Header.Get("X-User-ID"))
		}
		if r.Header.Get("X-User-Roles") != "buyer" {
			t.Fatalf("X-User-Roles = %q, want buyer", r.Header.Get("X-User-Roles"))
		}
		if r.Header.Get("X-Request-ID") != "req_test" {
			t.Fatalf("X-Request-ID = %q, want req_test", r.Header.Get("X-Request-ID"))
		}
		if r.Header.Get("X-Idempotency-Key") != "idem_test" {
			t.Fatalf("X-Idempotency-Key = %q, want idem_test", r.Header.Get("X-Idempotency-Key"))
		}

		var request cartAddItemRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.ProductID != "prod_123" || request.VariantID != "var_1" || request.Quantity != 1 {
			t.Fatalf("request = %#v, want prod_123 var_1 quantity 1", request)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": {
				"cart_id": "cart_123",
				"user_id": "user_123",
				"items": [
					{"item_id": "item_1", "product_id": "prod_123", "variant_id": "var_1", "quantity": 1}
				],
				"subtotal": {"amount": 299900, "currency": "inr"},
				"discount": {"amount": 0, "currency": "INR"},
				"total": {"amount": 299900, "currency": "INR"}
			},
			"error": null
		}`))
	}))
	defer server.Close()

	client := mustCartClient(t, server.URL)
	cart, err := client.AddItem(context.Background(), usecase.CartAddItemInput{
		UserID:         " user_123 ",
		ProductID:      " prod_123 ",
		VariantID:      " var_1 ",
		Quantity:       1,
		RequestID:      " req_test ",
		IdempotencyKey: " idem_test ",
	})
	if err != nil {
		t.Fatalf("AddItem returned error: %v", err)
	}
	if cart.CartID != "cart_123" {
		t.Fatalf("CartID = %q, want cart_123", cart.CartID)
	}
	if len(cart.Items) != 1 || cart.Items[0].ItemID != "item_1" {
		t.Fatalf("Items = %#v, want item_1", cart.Items)
	}
	if cart.Total == nil || cart.Total.Currency != "INR" {
		t.Fatalf("Total = %#v, want INR", cart.Total)
	}
}

func TestHTTPCartClientMapsCartErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       error
	}{
		{name: "bad request", statusCode: http.StatusBadRequest, want: usecase.ErrCartValidation},
		{name: "unauthorized", statusCode: http.StatusUnauthorized, want: usecase.ErrCartUnauthenticated},
		{name: "forbidden", statusCode: http.StatusForbidden, want: usecase.ErrCartForbidden},
		{name: "not found", statusCode: http.StatusNotFound, want: usecase.ErrCartProductNotFound},
		{name: "conflict", statusCode: http.StatusConflict, want: usecase.ErrCartItemUnavailable},
		{name: "server error", statusCode: http.StatusInternalServerError, want: usecase.ErrCartServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := mustCartClient(t, server.URL)
			_, err := client.AddItem(context.Background(), usecase.CartAddItemInput{
				UserID:    "user_123",
				ProductID: "prod_123",
				VariantID: "var_1",
				Quantity:  1,
			})
			if !errors.Is(err, tt.want) {
				t.Fatalf("AddItem error = %v, want %v", err, tt.want)
			}
		})
	}
}

func mustCartClient(t *testing.T, baseURL string) *HTTPCartClient {
	t.Helper()
	client, err := NewHTTPCartClient(baseURL, time.Second, nil)
	if err != nil {
		t.Fatalf("NewHTTPCartClient returned error: %v", err)
	}
	return client
}
