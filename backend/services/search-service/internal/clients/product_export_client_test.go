package clients

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/requestctx"
)

func TestHTTPProductExportClientListSearchableProducts(t *testing.T) {
	updatedAt := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	price := 2499.0
	popularity := int32(50)
	inStock := true

	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != defaultProductSearchExportPath {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("cursor") != "abc" || r.URL.Query().Get("limit") != "500" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		if r.Header.Get("X-Request-ID") != "req_123" {
			t.Fatalf("request id = %q", r.Header.Get("X-Request-ID"))
		}
		return jsonResponse(t, http.StatusOK, productSearchExportResponse{
			Items: []productSearchExportDTO{{
				ProductID:       " prod_1 ",
				Title:           " Shoe ",
				Brand:           " Nike ",
				CategoryIDs:     []string{"cat_shoes", "cat_shoes"},
				SellerID:        "seller_1",
				Price:           &price,
				PopularityScore: &popularity,
				InStock:         &inStock,
				Status:          " Published ",
				UpdatedAt:       updatedAt,
			}},
			NextCursor: "next",
			HasMore:    true,
			Total:      10,
		}), nil
	})}

	client, err := NewHTTPProductExportClient(HTTPProductExportClientConfig{BaseURL: "http://product-service", Timeout: time.Second}, httpClient)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	ctx := requestctx.WithRequestID(context.Background(), "req_123")
	page, err := client.ListSearchableProducts(ctx, domain.ProductExportRequest{Cursor: "abc", Limit: 500})
	if err != nil {
		t.Fatalf("list searchable products: %v", err)
	}
	if !page.HasMore || page.NextCursor != "next" || page.Total != 10 {
		t.Fatalf("page = %#v", page)
	}
	if len(page.Items) != 1 || page.Items[0].ProductID != "prod_1" || page.Items[0].Status != domain.ProductStatusPublished {
		b, _ := json.Marshal(page.Items)
		t.Fatalf("items = %s", b)
	}
	if len(page.Items[0].CategoryIDs) != 1 {
		t.Fatalf("category ids = %#v", page.Items[0].CategoryIDs)
	}
}

func TestHTTPProductExportClientMapsNonSuccessStatus(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return textResponse(http.StatusServiceUnavailable, "down"), nil
	})}

	client, err := NewHTTPProductExportClient(HTTPProductExportClientConfig{BaseURL: "http://product-service", Timeout: time.Second}, httpClient)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.ListSearchableProducts(context.Background(), domain.ProductExportRequest{Limit: 10})
	if !errors.Is(err, domain.ErrProductCatalogExportUnavailable) {
		t.Fatalf("expected product export unavailable, got %v", err)
	}
}
