package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/usecase"
)

func TestCartHTTPClientMapsCheckoutSnapshot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/v1/carts/cart_1" || r.Header.Get("X-User-ID") != "user_1" {
			t.Fatalf("unexpected cart request: %s user=%q", r.URL.Path, r.Header.Get("X-User-ID"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"cart_id": "cart_1", "user_id": "user_1",
			"items":  []map[string]any{{"product_id": "prod_1", "variant_id": "var_1", "quantity": 2}},
			"totals": map[string]any{"currency": "INR"},
		})
	}))
	defer server.Close()
	client, err := NewCartHTTPClient(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	cart, err := client.GetCart(context.Background(), usecase.GetCartRequest{UserID: "user_1", CartID: "cart_1"})
	if err != nil {
		t.Fatal(err)
	}
	if cart.CartID != "cart_1" || cart.UserID != "user_1" || cart.Currency != "INR" || len(cart.Items) != 1 || cart.Items[0].Quantity != 2 {
		t.Fatalf("unexpected cart snapshot: %+v", cart)
	}
}

func TestProductHTTPClientUsesOrderIDAndMapsSnapshots(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/internal/v1/products/batch":
			_ = json.NewEncoder(w).Encode(map[string]any{"products": []map[string]any{{
				"product_id": "prod_1", "seller_id": "seller_1", "title": "Shirt", "status": "published",
				"images":   []map[string]any{{"url": "https://cdn.test/shirt.jpg", "is_primary": true}},
				"variants": []map[string]any{{"variant_id": "var_1", "sku": "SKU-1", "status": "active", "available_quantity": 4, "price": map[string]any{"amount": 1200, "currency": "INR"}}},
			}}})
		case "/internal/v1/inventory/reservations":
			var body struct {
				OrderID string `json:"order_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.OrderID != "ord_1" {
				t.Fatalf("reservation order_id = %q, err=%v", body.OrderID, err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"reservation_id": "res_1", "expires_at": time.Now().Add(time.Minute)})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, err := NewProductHTTPClient(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	products, err := client.BatchGetProducts(context.Background(), usecase.BatchGetProductsRequest{Items: []usecase.ProductLookupItem{{ProductID: "prod_1", VariantID: "var_1"}}})
	if err != nil || len(products.Items) != 1 || products.Items[0].UnitAmount != 1200 || !products.Items[0].InStock {
		t.Fatalf("products=%+v err=%v", products, err)
	}
	reservation, err := client.ReserveInventory(context.Background(), usecase.ReserveInventoryRequest{OrderID: "ord_1", TTL: time.Minute})
	if err != nil || reservation.ReservationID != "res_1" {
		t.Fatalf("reservation=%+v err=%v", reservation, err)
	}
}

func TestPaymentHTTPClientAuthenticatesAndMapsClientAction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer internal-token" || r.Header.Get("Idempotency-Key") != "pay-key" {
			t.Fatalf("missing payment authorization or idempotency headers")
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"payment_id": "pay_1", "provider": "stripe_like", "provider_intent_id": "pi_1", "status": "requires_action",
			"client_payload": map[string]any{"client_secret": "secret_1"},
		})
	}))
	defer server.Close()
	client, err := NewPaymentHTTPClient(server.URL, "internal-token", time.Minute, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.CreatePaymentIntent(context.Background(), usecase.CreatePaymentIntentRequest{OrderID: "ord_1", IdempotencyKey: "pay-key"})
	if err != nil || result.PaymentID != "pay_1" || result.ClientActionToken != "secret_1" || result.ExpiresAt.Before(time.Now()) {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
