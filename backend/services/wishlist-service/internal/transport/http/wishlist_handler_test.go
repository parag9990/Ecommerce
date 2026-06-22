package httptransport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
	"ecommerce/backend/services/wishlist-service/internal/usecase"
)

func TestWishlistHandlerAddItemUsesAuthenticatedHeaderUser(t *testing.T) {
	fakeUsecase := &fakeWishlistUsecase{
		wishlist: mustTransportWishlist(t),
	}
	handler := mustWishlistHandler(t, fakeUsecase)

	request := httptest.NewRequest(http.MethodPost, wishlistItemsPath, bytes.NewBufferString(`{"product_id":"prod_123","variant_id":"var_1"}`))
	request.Header.Set("X-User-ID", "user_123")
	request.Header.Set("X-User-Roles", "buyer")
	request.Header.Set("X-Request-ID", "req_test")
	response := httptest.NewRecorder()

	handler.handleItems(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 body=%s", response.Code, response.Body.String())
	}
	if fakeUsecase.addInput.UserID != "user_123" {
		t.Fatalf("add input user_id = %q, want user_123", fakeUsecase.addInput.UserID)
	}
	if fakeUsecase.addInput.ProductID != "prod_123" {
		t.Fatalf("add input product_id = %q, want prod_123", fakeUsecase.addInput.ProductID)
	}
	if fakeUsecase.addInput.TraceID != "req_test" {
		t.Fatalf("add input trace_id = %q, want req_test", fakeUsecase.addInput.TraceID)
	}

	var envelope apiEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.RequestID != "req_test" {
		t.Fatalf("request_id = %q, want req_test", envelope.RequestID)
	}
	if envelope.Error != nil {
		t.Fatalf("error = %#v, want nil", envelope.Error)
	}
}

func TestWishlistHandlerGetWishlistUsesAuthenticatedBuyer(t *testing.T) {
	fakeUsecase := &fakeWishlistUsecase{wishlist: mustTransportWishlist(t)}
	handler := mustWishlistHandler(t, fakeUsecase)
	request := httptest.NewRequest(http.MethodGet, wishlistPath, nil)
	request.Header.Set("X-User-ID", "user_123")
	request.Header.Set("X-User-Roles", "buyer")
	response := httptest.NewRecorder()

	handler.handleWishlist(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 body=%s", response.Code, response.Body.String())
	}
	if fakeUsecase.getInput.UserID != "user_123" {
		t.Fatalf("get input user_id = %q, want user_123", fakeUsecase.getInput.UserID)
	}
}

func TestWishlistHandlerRejectsMissingBuyerRole(t *testing.T) {
	handler := mustWishlistHandler(t, &fakeWishlistUsecase{wishlist: mustTransportWishlist(t)})
	request := httptest.NewRequest(http.MethodPost, wishlistItemsPath, bytes.NewBufferString(`{"product_id":"prod_123"}`))
	request.Header.Set("X-User-ID", "user_123")
	response := httptest.NewRecorder()

	handler.handleItems(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}

func TestWishlistHandlerMapsDuplicateAdd(t *testing.T) {
	handler := mustWishlistHandler(t, &fakeWishlistUsecase{addErr: domain.ErrDuplicateProduct})
	request := httptest.NewRequest(http.MethodPost, wishlistItemsPath, bytes.NewBufferString(`{"product_id":"prod_123"}`))
	request.Header.Set("X-User-ID", "user_123")
	request.Header.Set("X-User-Roles", "buyer")
	response := httptest.NewRecorder()

	handler.handleItems(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", response.Code)
	}
	var envelope apiEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Error == nil || envelope.Error.Code != "WISHLIST_ITEM_ALREADY_EXISTS" {
		t.Fatalf("error = %#v, want WISHLIST_ITEM_ALREADY_EXISTS", envelope.Error)
	}
}

func TestWishlistHandlerRemoveItem(t *testing.T) {
	fakeUsecase := &fakeWishlistUsecase{wishlist: mustTransportWishlist(t)}
	handler := mustWishlistHandler(t, fakeUsecase)
	request := httptest.NewRequest(http.MethodDelete, wishlistItemsPathPrefix+"prod_123", nil)
	request.Header.Set("X-User-ID", "user_123")
	request.Header.Set("X-User-Roles", "buyer")
	response := httptest.NewRecorder()

	handler.handleItemByProduct(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if fakeUsecase.removeInput.UserID != "user_123" {
		t.Fatalf("remove input user_id = %q, want user_123", fakeUsecase.removeInput.UserID)
	}
	if fakeUsecase.removeInput.ProductID != "prod_123" {
		t.Fatalf("remove input product_id = %q, want prod_123", fakeUsecase.removeInput.ProductID)
	}
	if fakeUsecase.removeInput.TraceID == "" {
		t.Fatal("remove input trace_id is empty")
	}
}

func TestWishlistHandlerMoveToCartReturnsCart(t *testing.T) {
	fakeUsecase := &fakeWishlistUsecase{
		cart: &usecase.Cart{
			CartID: "cart_123",
			UserID: "user_123",
			Items:  []usecase.CartItem{{ItemID: "item_1", ProductID: "prod_123", VariantID: "var_1", Quantity: 1}},
			Total:  &domain.Money{Amount: 299900, Currency: "INR"},
		},
	}
	handler := mustWishlistHandler(t, fakeUsecase)
	request := httptest.NewRequest(http.MethodPost, wishlistItemsPathPrefix+"prod_123/move-to-cart", nil)
	request.Header.Set("X-User-ID", "user_123")
	request.Header.Set("X-User-Roles", "buyer")
	request.Header.Set("X-Request-ID", "req_move")
	request.Header.Set("X-Idempotency-Key", "idem_move")
	response := httptest.NewRecorder()

	handler.handleItemByProduct(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 body=%s", response.Code, response.Body.String())
	}
	if fakeUsecase.moveInput.UserID != "user_123" {
		t.Fatalf("move input user_id = %q, want user_123", fakeUsecase.moveInput.UserID)
	}
	if fakeUsecase.moveInput.ProductID != "prod_123" {
		t.Fatalf("move input product_id = %q, want prod_123", fakeUsecase.moveInput.ProductID)
	}
	if fakeUsecase.moveInput.RequestID != "req_move" || fakeUsecase.moveInput.IdempotencyKey != "idem_move" {
		t.Fatalf("move request metadata = %#v, want request/idempotency headers", fakeUsecase.moveInput)
	}

	var envelope apiEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := envelope.Data.(map[string]any)
	if !ok {
		t.Fatalf("data = %#v, want object", envelope.Data)
	}
	if data["cart_id"] != "cart_123" {
		t.Fatalf("cart_id = %#v, want cart_123", data["cart_id"])
	}
}

func TestWishlistHandlerMoveToCartMapsErrors(t *testing.T) {
	handler := mustWishlistHandler(t, &fakeWishlistUsecase{moveErr: usecase.ErrVariantRequiredForCart})
	request := httptest.NewRequest(http.MethodPost, wishlistItemsPathPrefix+"prod_123/move-to-cart", nil)
	request.Header.Set("X-User-ID", "user_123")
	request.Header.Set("X-User-Roles", "buyer")
	response := httptest.NewRecorder()

	handler.handleItemByProduct(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
	var envelope apiEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Error == nil || envelope.Error.Code != "VARIANT_REQUIRED_FOR_CART" {
		t.Fatalf("error = %#v, want VARIANT_REQUIRED_FOR_CART", envelope.Error)
	}
}

func TestWishlistHandlerMoveToCartRejectsWrongMethod(t *testing.T) {
	handler := mustWishlistHandler(t, &fakeWishlistUsecase{})
	request := httptest.NewRequest(http.MethodDelete, wishlistItemsPathPrefix+"prod_123/move-to-cart", nil)
	request.Header.Set("X-User-ID", "user_123")
	request.Header.Set("X-User-Roles", "buyer")
	response := httptest.NewRecorder()

	handler.handleItemByProduct(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", response.Code)
	}
	if response.Header().Get("Allow") != http.MethodPost {
		t.Fatalf("Allow = %q, want POST", response.Header().Get("Allow"))
	}
}

func mustWishlistHandler(t *testing.T, wishlistUsecase WishlistUsecase) *WishlistHandler {
	t.Helper()
	handler, err := NewWishlistHandler(wishlistUsecase, WishlistHandlerConfig{
		UserIDHeader:     "X-User-ID",
		RolesHeader:      "X-User-Roles",
		RequestIDHeader:  "X-Request-ID",
		RequireBuyerRole: true,
	}, nil)
	if err != nil {
		t.Fatalf("NewWishlistHandler returned error: %v", err)
	}
	return handler
}

func mustTransportWishlist(t *testing.T) *domain.Wishlist {
	t.Helper()
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	wishlist, err := domain.NewWishlist("wish_123", "user_123", now)
	if err != nil {
		t.Fatalf("NewWishlist returned error: %v", err)
	}
	if err := wishlist.AddItem(domain.AddWishlistItemInput{
		ProductID:    "prod_123",
		VariantID:    "var_1",
		Availability: domain.AvailabilityInStock,
	}, now); err != nil {
		t.Fatalf("AddItem returned error: %v", err)
	}
	return wishlist
}

type fakeWishlistUsecase struct {
	wishlist    *domain.Wishlist
	cart        *usecase.Cart
	getInput    usecase.GetWishlistInput
	addInput    usecase.AddWishlistItemInput
	removeInput usecase.RemoveWishlistItemInput
	moveInput   usecase.MoveToCartInput
	addErr      error
	removeErr   error
	moveErr     error
}

func (f *fakeWishlistUsecase) GetWishlist(ctx context.Context, input usecase.GetWishlistInput) (*domain.Wishlist, error) {
	f.getInput = input
	if f.wishlist == nil {
		return nil, errors.New("missing wishlist")
	}
	return f.wishlist, nil
}

func (f *fakeWishlistUsecase) AddItem(ctx context.Context, input usecase.AddWishlistItemInput) (*domain.Wishlist, error) {
	f.addInput = input
	if f.addErr != nil {
		return nil, f.addErr
	}
	return f.wishlist, nil
}

func (f *fakeWishlistUsecase) MoveToCart(ctx context.Context, input usecase.MoveToCartInput) (*usecase.Cart, error) {
	f.moveInput = input
	if f.moveErr != nil {
		return nil, f.moveErr
	}
	if f.cart == nil {
		return nil, errors.New("missing cart")
	}
	return f.cart, nil
}

func (f *fakeWishlistUsecase) RemoveItem(ctx context.Context, input usecase.RemoveWishlistItemInput) (*domain.Wishlist, error) {
	f.removeInput = input
	if f.removeErr != nil {
		return nil, f.removeErr
	}
	if f.wishlist == nil {
		return nil, errors.New("missing wishlist")
	}
	return f.wishlist, nil
}
