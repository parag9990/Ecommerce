package clients

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
	"ecommerce/backend/services/wishlist-service/internal/usecase"
)

func TestHTTPProductValidatorBuildsSnapshotFromProductService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/products/prod_123" {
			t.Fatalf("path = %q, want /api/v1/products/prod_123", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": {
				"product_id": "prod_123",
				"status": "published",
				"variants": [
					{
						"variant_id": "var_1",
						"price": {"amount": 299900, "currency": "inr"},
						"stock_quantity": 10,
						"reserved_quantity": 2
					}
				]
			},
			"error": null
		}`))
	}))
	defer server.Close()
	validator := mustProductValidator(t, server.URL)

	snapshot, err := validator.ValidateWishlistProduct(context.Background(), "prod_123", "var_1")
	if err != nil {
		t.Fatalf("ValidateWishlistProduct returned error: %v", err)
	}
	if snapshot.ProductID != "prod_123" {
		t.Fatalf("ProductID = %q, want prod_123", snapshot.ProductID)
	}
	if snapshot.VariantID != "var_1" {
		t.Fatalf("VariantID = %q, want var_1", snapshot.VariantID)
	}
	if snapshot.LastKnownPrice == nil || snapshot.LastKnownPrice.Currency != "INR" {
		t.Fatalf("LastKnownPrice = %#v, want INR price", snapshot.LastKnownPrice)
	}
	if snapshot.Availability != domain.AvailabilityInStock {
		t.Fatalf("Availability = %q, want in_stock", snapshot.Availability)
	}
}

func TestHTTPProductValidatorRejectsUnpublishedProduct(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"product_id":"prod_123","status":"draft"}`))
	}))
	defer server.Close()
	validator := mustProductValidator(t, server.URL)

	_, err := validator.ValidateWishlistProduct(context.Background(), "prod_123", "")
	if !errors.Is(err, usecase.ErrProductNotAvailable) {
		t.Fatalf("ValidateWishlistProduct error = %v, want ErrProductNotAvailable", err)
	}
}

func TestHTTPProductValidatorRejectsMissingVariant(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"product_id":"prod_123","status":"published","variants":[{"variant_id":"var_1"}]}`))
	}))
	defer server.Close()
	validator := mustProductValidator(t, server.URL)

	_, err := validator.ValidateWishlistProduct(context.Background(), "prod_123", "var_missing")
	if !errors.Is(err, usecase.ErrVariantNotFound) {
		t.Fatalf("ValidateWishlistProduct error = %v, want ErrVariantNotFound", err)
	}
}

func TestHTTPProductValidatorMapsNotFound(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	validator := mustProductValidator(t, server.URL)

	_, err := validator.ValidateWishlistProduct(context.Background(), "prod_missing", "")
	if !errors.Is(err, usecase.ErrProductNotFound) {
		t.Fatalf("ValidateWishlistProduct error = %v, want ErrProductNotFound", err)
	}
}

func mustProductValidator(t *testing.T, baseURL string) *HTTPProductValidator {
	t.Helper()
	validator, err := NewHTTPProductValidator(baseURL, time.Second, nil)
	if err != nil {
		t.Fatalf("NewHTTPProductValidator returned error: %v", err)
	}
	return validator
}
