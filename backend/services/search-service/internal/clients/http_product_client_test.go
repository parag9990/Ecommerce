package clients

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/requestctx"
)

func TestHTTPProductClientBatchGetProducts(t *testing.T) {
	var received batchGetProductsRequest
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != defaultProductBatchGetPath {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Header.Get("X-Request-ID") != "req_123" {
			t.Fatalf("request id = %q", r.Header.Get("X-Request-ID"))
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		return jsonResponse(t, http.StatusOK, productListResponse{
			Products: []productDTO{{ProductID: "prod_1", SellerID: "seller_1", Title: "Shoe", Brand: "Nike", CategoryID: "cat_shoes", Status: "published"}},
		}), nil
	})}

	client, err := NewHTTPProductClient(HTTPProductClientConfig{BaseURL: "http://product-service", Timeout: time.Second}, httpClient)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	ctx := requestctx.WithRequestID(context.Background(), "req_123")
	products, err := client.BatchGetProducts(ctx, []string{"prod_1", "prod_1", " "})
	if err != nil {
		t.Fatalf("batch get: %v", err)
	}
	if len(received.IDs) != 1 || received.IDs[0] != "prod_1" {
		t.Fatalf("ids = %#v", received.IDs)
	}
	if len(products) != 1 || products[0].ProductID != "prod_1" {
		t.Fatalf("products = %#v", products)
	}
}

func TestHTTPProductClientMapsNonSuccessStatus(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return textResponse(http.StatusServiceUnavailable, "down"), nil
	})}

	client, err := NewHTTPProductClient(HTTPProductClientConfig{BaseURL: "http://product-service", Timeout: time.Second}, httpClient)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.BatchGetProducts(context.Background(), []string{"prod_1"})
	if !errors.Is(err, domain.ErrProductHydrationUnavailable) {
		t.Fatalf("expected product hydration unavailable, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func jsonResponse(t *testing.T, status int, value any) *http.Response {
	t.Helper()
	var body strings.Builder
	if err := json.NewEncoder(&body).Encode(value); err != nil {
		t.Fatalf("encode response: %v", err)
	}
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body.String())),
	}
}

func textResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
