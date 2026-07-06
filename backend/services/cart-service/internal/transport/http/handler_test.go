package httptransport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/usecase"
)

func TestHandleSchemaReturnsCollectionSpec(t *testing.T) {
	handler, err := NewHandler(&fakeSchemaUsecase{}, &fakeCartUsecase{}, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	resp, err := http.Get(server.URL + "/internal/v1/cart/schema")
	if err != nil {
		t.Fatalf("GET schema error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body collectionSpecResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body.DatabaseName != domain.CartDatabaseName {
		t.Fatalf("database_name = %q, want %q", body.DatabaseName, domain.CartDatabaseName)
	}
	if body.CollectionName != domain.CartCollectionName {
		t.Fatalf("collection_name = %q, want %q", body.CollectionName, domain.CartCollectionName)
	}
}

func TestHandleBootstrapSchemaRequiresPost(t *testing.T) {
	handler, err := NewHandler(&fakeSchemaUsecase{}, &fakeCartUsecase{}, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	resp, err := http.Get(server.URL + "/internal/v1/cart/schema/bootstrap")
	if err != nil {
		t.Fatalf("GET bootstrap error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestHandleAddItemReturnsUpdatedCart(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	handler, err := NewHandler(&fakeSchemaUsecase{}, &fakeCartUsecase{
		cart: &domain.Cart{
			ID:        "cart_123",
			UserID:    &userID,
			Status:    domain.CartStatusActive,
			Items:     []domain.CartItem{validHTTPItem(now)},
			Totals:    validHTTPTotals(),
			Version:   2,
			CreatedAt: now,
			UpdatedAt: now,
			ExpiresAt: now.Add(90 * 24 * time.Hour),
		},
	}, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/cart/items", strings.NewReader(`{"product_id":"prod_123","variant_id":"var_1","quantity":1}`))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST add item error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body cartResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body.CartID != "cart_123" || body.Subtotal.Amount != 1000 {
		t.Fatalf("unexpected cart response: %#v", body)
	}
}

func TestHandleGetCartReturnsActiveCart(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	handler, err := NewHandler(&fakeSchemaUsecase{}, &fakeCartUsecase{}, nil, WithCartReader(&fakeCartReader{
		cart: &domain.Cart{
			ID:        "cart_123",
			UserID:    &userID,
			Status:    domain.CartStatusActive,
			Items:     []domain.CartItem{validHTTPItem(now)},
			Totals:    validHTTPTotals(),
			Version:   2,
			CreatedAt: now,
			UpdatedAt: now,
			ExpiresAt: now.Add(90 * 24 * time.Hour),
		},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/cart", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("X-User-ID", userID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET cart error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body cartResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body.CartID != "cart_123" || len(body.Items) != 1 || body.UserID == nil || *body.UserID != userID {
		t.Fatalf("unexpected cart response: %#v", body)
	}
}

func TestHandleGetCartReturnsEmptyCartWhenMissing(t *testing.T) {
	handler, err := NewHandler(&fakeSchemaUsecase{}, &fakeCartUsecase{}, nil, WithCartReader(&fakeCartReader{
		err: domain.ErrCartNotFound,
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/cart", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("X-Guest-Session-ID", "sess_123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET cart error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body cartResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body.CartID != "" || len(body.Items) != 0 || body.GuestSessionID == nil || *body.GuestSessionID != "sess_123" {
		t.Fatalf("unexpected empty cart response: %#v", body)
	}
}

func TestHandleAddItemRequiresOwner(t *testing.T) {
	handler, err := NewHandler(&fakeSchemaUsecase{}, &fakeCartUsecase{err: domain.ErrCartOwnerMissing}, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	resp, err := http.Post(server.URL+"/api/v1/cart/items", "application/json", strings.NewReader(`{"product_id":"prod_123","variant_id":"var_1","quantity":1}`))
	if err != nil {
		t.Fatalf("POST add item error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestHandleUpdateItemReturnsUpdatedCart(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	handler, err := NewHandler(&fakeSchemaUsecase{}, &fakeCartUsecase{
		cart: &domain.Cart{
			ID:        "cart_123",
			UserID:    &userID,
			Status:    domain.CartStatusActive,
			Items:     []domain.CartItem{validHTTPItem(now)},
			Totals:    validHTTPTotals(),
			Version:   3,
			CreatedAt: now,
			UpdatedAt: now,
			ExpiresAt: now.Add(90 * 24 * time.Hour),
		},
	}, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPatch, server.URL+"/api/v1/cart/items/item_1", strings.NewReader(`{"quantity":2}`))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH update item error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body cartResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body.CartID != "cart_123" || len(body.Items) != 1 {
		t.Fatalf("unexpected cart response: %#v", body)
	}
}

func TestHandleUpdateItemMapsItemNotFound(t *testing.T) {
	handler, err := NewHandler(&fakeSchemaUsecase{}, &fakeCartUsecase{err: domain.ErrCartItemNotFound}, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPatch, server.URL+"/api/v1/cart/items/item_404", strings.NewReader(`{"quantity":2}`))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user_123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH update item error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestHandleRemoveItemReturnsUpdatedCart(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	handler, err := NewHandler(&fakeSchemaUsecase{}, &fakeCartUsecase{
		cart: &domain.Cart{
			ID:        "cart_123",
			UserID:    &userID,
			Status:    domain.CartStatusActive,
			Items:     []domain.CartItem{},
			Totals:    emptyHTTPTotals(),
			Version:   3,
			CreatedAt: now,
			UpdatedAt: now,
			ExpiresAt: now.Add(90 * 24 * time.Hour),
		},
	}, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	req, err := http.NewRequest(http.MethodDelete, server.URL+"/api/v1/cart/items/item_1", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("X-User-ID", userID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE remove item error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body cartResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body.CartID != "cart_123" || len(body.Items) != 0 || body.Total.Amount != 0 {
		t.Fatalf("unexpected cart response: %#v", body)
	}
}

func TestHandleRemoveItemRequiresItemID(t *testing.T) {
	handler, err := NewHandler(&fakeSchemaUsecase{}, &fakeCartUsecase{err: domain.ErrItemIDRequired}, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	req, err := http.NewRequest(http.MethodDelete, server.URL+"/api/v1/cart/items/", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("X-User-ID", "user_123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE remove item error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleRemoveItemMapsCartNotFound(t *testing.T) {
	handler, err := NewHandler(&fakeSchemaUsecase{}, &fakeCartUsecase{err: domain.ErrCartNotFound}, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	req, err := http.NewRequest(http.MethodDelete, server.URL+"/api/v1/cart/items/item_1", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("X-User-ID", "user_123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE remove item error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestHandleRemoveItemMapsCartExpired(t *testing.T) {
	handler, err := NewHandler(&fakeSchemaUsecase{}, &fakeCartUsecase{err: domain.ErrCartExpired}, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	req, err := http.NewRequest(http.MethodDelete, server.URL+"/api/v1/cart/items/item_1", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("X-User-ID", "user_123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE remove item error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusConflict)
	}
	var body errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body.Error.Code != "CART_EXPIRED" {
		t.Fatalf("error code = %q, want CART_EXPIRED", body.Error.Code)
	}
}

func TestHandleApplyCouponPreviewReturnsPreview(t *testing.T) {
	handler, err := NewHandler(&fakeSchemaUsecase{}, &fakeCartUsecase{
		preview: &usecase.CouponPreviewResult{
			Valid:    true,
			CouponID: "coupon_123",
			Discount: domain.NewMoney(250, domain.CurrencyINR),
		},
	}, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/cart/coupons/preview", strings.NewReader(`{"coupon_code":" save10 "}`))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user_123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST coupon preview error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body couponPreviewResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if !body.Valid || body.CouponID != "coupon_123" || body.Discount.Amount != 250 {
		t.Fatalf("unexpected coupon preview response: %#v", body)
	}
}

func TestHandleApplyCouponPreviewMapsCouponCodeRequired(t *testing.T) {
	handler, err := NewHandler(&fakeSchemaUsecase{}, &fakeCartUsecase{err: domain.ErrCouponCodeRequired}, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/cart/coupons/preview", strings.NewReader(`{"coupon_code":""}`))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user_123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST coupon preview error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleMergeGuestCartReturnsMergedCart(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	handler, err := NewHandler(&fakeSchemaUsecase{}, &fakeCartUsecase{
		cart: &domain.Cart{
			ID:        "cart_user_123",
			UserID:    &userID,
			Status:    domain.CartStatusActive,
			Items:     []domain.CartItem{validHTTPItem(now)},
			Totals:    validHTTPTotals(),
			Version:   3,
			CreatedAt: now,
			UpdatedAt: now,
			ExpiresAt: now.Add(90 * 24 * time.Hour),
		},
	}, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/cart/merge", strings.NewReader(`{"guest_cart_id":"cart_guest_123"}`))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)
	req.Header.Set("X-Guest-Session-ID", "sess_123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST merge cart error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body cartResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body.CartID != "cart_user_123" || body.UserID == nil || *body.UserID != userID {
		t.Fatalf("unexpected merge response: %#v", body)
	}
}

func TestHandleMergeGuestCartMapsGuestCartIDRequired(t *testing.T) {
	handler, err := NewHandler(&fakeSchemaUsecase{}, &fakeCartUsecase{err: domain.ErrGuestCartIDRequired}, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/cart/merge", strings.NewReader(`{"guest_cart_id":""}`))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user_123")
	req.Header.Set("X-Guest-Session-ID", "sess_123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST merge cart error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

type fakeSchemaUsecase struct{}

func (fakeSchemaUsecase) Spec(ctx context.Context) (domain.CollectionSpec, error) {
	return domain.CartCollectionSpec(), nil
}

func (fakeSchemaUsecase) EnsureCollections(ctx context.Context) (domain.CollectionReport, error) {
	return domain.CollectionReport{
		DatabaseName:      domain.CartDatabaseName,
		CollectionName:    domain.CartCollectionName,
		CollectionCreated: true,
		IndexesEnsured:    []string{"idx_carts_user_status"},
	}, nil
}

func (fakeSchemaUsecase) Ready(ctx context.Context) error {
	return nil
}

type fakeCartUsecase struct {
	cart    *domain.Cart
	preview *usecase.CouponPreviewResult
	err     error
}

func (f *fakeCartUsecase) AddItem(ctx context.Context, cmd usecase.AddItemCommand) (*domain.Cart, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.cart, nil
}

func (f *fakeCartUsecase) UpdateItem(ctx context.Context, cmd usecase.UpdateItemCommand) (*domain.Cart, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.cart, nil
}

func (f *fakeCartUsecase) RemoveItem(ctx context.Context, cmd usecase.RemoveItemCommand) (*domain.Cart, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.cart, nil
}

func (f *fakeCartUsecase) ApplyCouponPreview(ctx context.Context, cmd usecase.ApplyCouponPreviewCommand) (*usecase.CouponPreviewResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.preview, nil
}

func (f *fakeCartUsecase) MergeGuestCart(ctx context.Context, cmd usecase.MergeGuestCartCommand) (*domain.Cart, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.cart, nil
}

type fakeCartReader struct {
	cart *domain.Cart
	err  error
}

func (f *fakeCartReader) FindActiveByOwner(ctx context.Context, owner domain.CartOwner) (*domain.Cart, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.cart, nil
}

func validHTTPItem(now time.Time) domain.CartItem {
	return domain.CartItem{
		ItemID:          "item_1",
		ProductID:       "prod_123",
		VariantID:       "var_1",
		SellerID:        "seller_456",
		TitleSnapshot:   "Running Shoes",
		UnitPrice:       domain.NewMoney(1000, domain.CurrencyINR),
		Quantity:        1,
		LineSubtotal:    domain.NewMoney(1000, domain.CurrencyINR),
		PriceSnapshotAt: now,
		AddedAt:         now,
		UpdatedAt:       now,
	}
}

func validHTTPTotals() domain.CartTotals {
	return domain.CartTotals{
		Subtotal:        domain.NewMoney(1000, domain.CurrencyINR),
		Discount:        domain.NewMoney(0, domain.CurrencyINR),
		Total:           domain.NewMoney(1000, domain.CurrencyINR),
		Currency:        domain.CurrencyINR,
		ItemCount:       1,
		UniqueItemCount: 1,
	}
}

func emptyHTTPTotals() domain.CartTotals {
	return domain.CartTotals{
		Subtotal:        domain.NewMoney(0, domain.CurrencyINR),
		Discount:        domain.NewMoney(0, domain.CurrencyINR),
		Total:           domain.NewMoney(0, domain.CurrencyINR),
		Currency:        domain.CurrencyINR,
		ItemCount:       0,
		UniqueItemCount: 0,
	}
}
