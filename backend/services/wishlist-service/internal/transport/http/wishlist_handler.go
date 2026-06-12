package httptransport

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
	"ecommerce/backend/services/wishlist-service/internal/usecase"
)

const (
	wishlistItemsPath       = "/api/v1/wishlist/items"
	wishlistItemsPathPrefix = "/api/v1/wishlist/items/"
	maxWishlistRequestBytes = 1 << 20
)

type WishlistUsecase interface {
	AddItem(ctx context.Context, input usecase.AddWishlistItemInput) (*domain.Wishlist, error)
	RemoveItem(ctx context.Context, input usecase.RemoveWishlistItemInput) (*domain.Wishlist, error)
	MoveToCart(ctx context.Context, input usecase.MoveToCartInput) (*usecase.Cart, error)
}

type WishlistHandlerConfig struct {
	UserIDHeader     string
	RolesHeader      string
	RequestIDHeader  string
	RequireBuyerRole bool
}

type WishlistHandler struct {
	usecase WishlistUsecase
	config  WishlistHandlerConfig
	logger  *slog.Logger
}

type addWishlistItemRequest struct {
	ProductID string  `json:"product_id"`
	VariantID *string `json:"variant_id,omitempty"`
}

type apiEnvelope struct {
	Data      any       `json:"data"`
	RequestID string    `json:"request_id"`
	Error     *apiError `json:"error"`
}

type apiError struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []errorDetail `json:"details,omitempty"`
}

type errorDetail struct {
	Field string `json:"field,omitempty"`
	Value string `json:"value,omitempty"`
}

type wishlistResponse struct {
	WishlistID string                 `json:"wishlist_id"`
	UserID     string                 `json:"user_id"`
	Items      []wishlistItemResponse `json:"items"`
}

type wishlistItemResponse struct {
	ProductID      string         `json:"product_id"`
	VariantID      string         `json:"variant_id,omitempty"`
	AddedAt        time.Time      `json:"added_at"`
	LastKnownPrice *moneyResponse `json:"last_known_price,omitempty"`
	Availability   string         `json:"availability,omitempty"`
}

type moneyResponse struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type cartResponse struct {
	CartID   string             `json:"cart_id"`
	UserID   string             `json:"user_id"`
	Items    []cartItemResponse `json:"items"`
	Subtotal *moneyResponse     `json:"subtotal,omitempty"`
	Discount *moneyResponse     `json:"discount,omitempty"`
	Total    *moneyResponse     `json:"total,omitempty"`
}

type cartItemResponse struct {
	ItemID    string `json:"item_id,omitempty"`
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id"`
	Quantity  int    `json:"quantity"`
}

func NewWishlistHandler(usecase WishlistUsecase, config WishlistHandlerConfig, logger *slog.Logger) (*WishlistHandler, error) {
	if usecase == nil {
		return nil, errors.New("wishlist usecase is required")
	}
	if config.UserIDHeader == "" {
		config.UserIDHeader = "X-User-ID"
	}
	if config.RolesHeader == "" {
		config.RolesHeader = "X-User-Roles"
	}
	if config.RequestIDHeader == "" {
		config.RequestIDHeader = "X-Request-ID"
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &WishlistHandler{
		usecase: usecase,
		config:  config,
		logger:  logger,
	}, nil
}

func (h *WishlistHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc(wishlistItemsPath, h.handleItems)
	mux.HandleFunc(wishlistItemsPathPrefix, h.handleItemByProduct)
}

func (h *WishlistHandler) handleItems(w http.ResponseWriter, r *http.Request) {
	requestID := h.requestID(r)
	if r.URL.Path != wishlistItemsPath {
		h.writeError(w, r, requestID, http.StatusNotFound, "NOT_FOUND", "Resource not found", nil)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		h.writeError(w, r, requestID, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", nil)
		return
	}

	userID, err := h.userIDFromRequest(r)
	if err != nil {
		h.writeMappedError(w, r, requestID, err)
		return
	}

	var request addWishlistItemRequest
	if err := decodeJSONRequest(w, r, &request); err != nil {
		h.writeMappedError(w, r, requestID, err)
		return
	}
	variantID := ""
	if request.VariantID != nil {
		if strings.TrimSpace(*request.VariantID) == "" {
			h.writeMappedError(w, r, requestID, usecase.ValidationError{Field: "variant_id", Message: "must not be empty when provided"})
			return
		}
		variantID = *request.VariantID
	}

	wishlist, err := h.usecase.AddItem(r.Context(), usecase.AddWishlistItemInput{
		UserID:    userID,
		ProductID: request.ProductID,
		VariantID: variantID,
		TraceID:   requestID,
	})
	if err != nil {
		h.writeMappedError(w, r, requestID, err)
		return
	}
	h.writeResponse(w, http.StatusOK, requestID, mapWishlist(wishlist))
}

func (h *WishlistHandler) handleItemByProduct(w http.ResponseWriter, r *http.Request) {
	requestID := h.requestID(r)
	if strings.HasSuffix(r.URL.Path, "/move-to-cart") {
		h.handleMoveToCart(w, r, requestID)
		return
	}

	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodDelete)
		h.writeError(w, r, requestID, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", nil)
		return
	}

	encodedProductID := strings.TrimPrefix(r.URL.Path, wishlistItemsPathPrefix)
	if encodedProductID == "" || strings.Contains(encodedProductID, "/") {
		h.writeError(w, r, requestID, http.StatusNotFound, "NOT_FOUND", "Resource not found", nil)
		return
	}
	productID, err := url.PathUnescape(encodedProductID)
	if err != nil {
		h.writeMappedError(w, r, requestID, usecase.ValidationError{Field: "product_id", Message: "is invalid"})
		return
	}
	userID, err := h.userIDFromRequest(r)
	if err != nil {
		h.writeMappedError(w, r, requestID, err)
		return
	}

	wishlist, err := h.usecase.RemoveItem(r.Context(), usecase.RemoveWishlistItemInput{
		UserID:    userID,
		ProductID: productID,
		TraceID:   requestID,
	})
	if err != nil {
		h.writeMappedError(w, r, requestID, err)
		return
	}
	h.writeResponse(w, http.StatusOK, requestID, mapWishlist(wishlist))
}

func (h *WishlistHandler) handleMoveToCart(w http.ResponseWriter, r *http.Request, requestID string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		h.writeError(w, r, requestID, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", nil)
		return
	}

	productID, err := productIDFromMoveToCartPath(r.URL.Path)
	if err != nil {
		h.writeMappedError(w, r, requestID, err)
		return
	}
	userID, err := h.userIDFromRequest(r)
	if err != nil {
		h.writeMappedError(w, r, requestID, err)
		return
	}

	cart, err := h.usecase.MoveToCart(r.Context(), usecase.MoveToCartInput{
		UserID:         userID,
		ProductID:      productID,
		RequestID:      requestID,
		IdempotencyKey: strings.TrimSpace(r.Header.Get("X-Idempotency-Key")),
	})
	if err != nil {
		h.writeMappedError(w, r, requestID, err)
		return
	}
	h.writeResponse(w, http.StatusOK, requestID, mapCart(cart))
}

func productIDFromMoveToCartPath(path string) (string, error) {
	raw := strings.TrimPrefix(path, wishlistItemsPathPrefix)
	raw = strings.TrimSuffix(raw, "/move-to-cart")
	raw = strings.Trim(raw, "/")
	if raw == "" || strings.Contains(raw, "/") {
		return "", usecase.ValidationError{Field: "product_id", Message: "is invalid"}
	}
	productID, err := url.PathUnescape(raw)
	if err != nil || strings.TrimSpace(productID) == "" {
		return "", usecase.ValidationError{Field: "product_id", Message: "is invalid"}
	}
	return productID, nil
}

func decodeJSONRequest(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxWishlistRequestBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return usecase.ValidationError{Field: "body", Message: "must be valid JSON"}
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return usecase.ValidationError{Field: "body", Message: "must contain a single JSON object"}
	}
	return nil
}

func (h *WishlistHandler) userIDFromRequest(r *http.Request) (string, error) {
	userID := strings.TrimSpace(r.Header.Get(h.config.UserIDHeader))
	if userID == "" {
		return "", usecase.ErrUnauthenticated
	}
	if h.config.RequireBuyerRole && !requestHasBuyerRole(r, h.config.RolesHeader) {
		return "", usecase.ErrPermissionDenied
	}
	return userID, nil
}

func requestHasBuyerRole(r *http.Request, rolesHeader string) bool {
	values := append([]string{}, r.Header.Values(rolesHeader)...)
	values = append(values, r.Header.Values("X-User-Role")...)
	for _, value := range values {
		roles := strings.FieldsFunc(value, func(r rune) bool {
			return r == ',' || r == ' ' || r == ';'
		})
		for _, role := range roles {
			if strings.EqualFold(strings.TrimSpace(role), "buyer") {
				return true
			}
		}
	}
	return false
}

func (h *WishlistHandler) requestID(r *http.Request) string {
	if requestID := strings.TrimSpace(r.Header.Get(h.config.RequestIDHeader)); requestID != "" {
		return requestID
	}
	return generateRequestID()
}

func generateRequestID() string {
	var randomBytes [8]byte
	if _, err := rand.Read(randomBytes[:]); err == nil {
		return "req_" + hex.EncodeToString(randomBytes[:])
	}
	return fmt.Sprintf("req_%d", time.Now().UTC().UnixNano())
}

func (h *WishlistHandler) writeMappedError(w http.ResponseWriter, r *http.Request, requestID string, err error) {
	status, code, message, details := mapError(err)
	h.logger.Warn("wishlist request failed",
		"request_id", requestID,
		"method", r.Method,
		"path", r.URL.Path,
		"status", status,
		"code", code,
		"error", err,
	)
	h.writeError(w, r, requestID, status, code, message, details)
}

func (h *WishlistHandler) writeError(w http.ResponseWriter, _ *http.Request, requestID string, status int, code, message string, details []errorDetail) {
	writeAPIJSON(w, status, apiEnvelope{
		Data:      nil,
		RequestID: requestID,
		Error: &apiError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func (h *WishlistHandler) writeResponse(w http.ResponseWriter, status int, requestID string, data any) {
	writeAPIJSON(w, status, apiEnvelope{
		Data:      data,
		RequestID: requestID,
		Error:     nil,
	})
}

func writeAPIJSON(w http.ResponseWriter, status int, response apiEnvelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

func mapError(err error) (int, string, string, []errorDetail) {
	var validationErr usecase.ValidationError
	var fieldErr domain.FieldError
	switch {
	case errors.Is(err, usecase.ErrUnauthenticated):
		return http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required", nil
	case errors.Is(err, usecase.ErrPermissionDenied):
		return http.StatusForbidden, "FORBIDDEN", "Buyer role is required", nil
	case errors.Is(err, usecase.ErrVariantRequiredForCart):
		return http.StatusBadRequest, "VARIANT_REQUIRED_FOR_CART", "Select a variant before moving to cart", []errorDetail{{Field: "variant_id"}}
	case errors.As(err, &validationErr):
		return http.StatusBadRequest, "VALIDATION_ERROR", validationErr.Error(), []errorDetail{{Field: validationErr.Field}}
	case errors.As(err, &fieldErr):
		return http.StatusBadRequest, "VALIDATION_ERROR", fieldErr.Error(), []errorDetail{{Field: fieldErr.Field}}
	case errors.Is(err, domain.ErrInvalidWishlist), errors.Is(err, usecase.ErrInvalidInput):
		return http.StatusBadRequest, "VALIDATION_ERROR", "Invalid wishlist request", nil
	case errors.Is(err, domain.ErrWishlistItemNotFound):
		return http.StatusNotFound, "WISHLIST_ITEM_NOT_FOUND", "Product is not in wishlist", nil
	case errors.Is(err, usecase.ErrProductNotFound):
		return http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", nil
	case errors.Is(err, usecase.ErrProductNotAvailable):
		return http.StatusConflict, "PRODUCT_NOT_AVAILABLE", "Product is not available", nil
	case errors.Is(err, usecase.ErrVariantNotFound):
		return http.StatusBadRequest, "VARIANT_NOT_FOUND", "Product variant not found", nil
	case errors.Is(err, domain.ErrDuplicateProduct):
		return http.StatusConflict, "WISHLIST_ITEM_ALREADY_EXISTS", "Product already exists in wishlist", nil
	case errors.Is(err, usecase.ErrCartValidation):
		return http.StatusBadRequest, "CART_VALIDATION_ERROR", "Cart rejected the item", nil
	case errors.Is(err, usecase.ErrCartUnauthenticated):
		return http.StatusUnauthorized, "CART_UNAUTHENTICATED", "Cart service authentication failed", nil
	case errors.Is(err, usecase.ErrCartForbidden):
		return http.StatusForbidden, "CART_FORBIDDEN", "Cart service authorization failed", nil
	case errors.Is(err, usecase.ErrCartProductNotFound):
		return http.StatusNotFound, "CART_PRODUCT_NOT_FOUND", "Product or variant was not found for cart", nil
	case errors.Is(err, usecase.ErrCartItemUnavailable):
		return http.StatusConflict, "CART_ITEM_UNAVAILABLE", "Product cannot be added to cart", nil
	case errors.Is(err, usecase.ErrCartServiceUnavailable):
		return http.StatusServiceUnavailable, "CART_SERVICE_UNAVAILABLE", "Cart service is unavailable", nil
	case errors.Is(err, usecase.ErrMoveToCartPartialFailure):
		return http.StatusInternalServerError, "MOVE_TO_CART_PARTIAL_FAILURE", "Cart updated but wishlist cleanup failed", nil
	case errors.Is(err, usecase.ErrProductServiceUnavailable), errors.Is(err, context.DeadlineExceeded):
		return http.StatusServiceUnavailable, "PRODUCT_SERVICE_UNAVAILABLE", "Product service is unavailable", nil
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil
	}
}

func mapWishlist(wishlist *domain.Wishlist) wishlistResponse {
	if wishlist == nil {
		return wishlistResponse{Items: []wishlistItemResponse{}}
	}
	snapshot := wishlist.Snapshot()
	items := make([]wishlistItemResponse, 0, len(snapshot.Items))
	for _, item := range snapshot.Items {
		items = append(items, wishlistItemResponse{
			ProductID:      item.ProductID,
			VariantID:      item.VariantID,
			AddedAt:        item.AddedAt,
			LastKnownPrice: mapMoney(item.LastKnownPrice),
			Availability:   string(item.Availability.Normalized()),
		})
	}
	return wishlistResponse{
		WishlistID: snapshot.ID,
		UserID:     snapshot.UserID,
		Items:      items,
	}
}

func mapCart(cart *usecase.Cart) cartResponse {
	if cart == nil {
		return cartResponse{Items: []cartItemResponse{}}
	}
	items := make([]cartItemResponse, 0, len(cart.Items))
	for _, item := range cart.Items {
		items = append(items, cartItemResponse{
			ItemID:    item.ItemID,
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
		})
	}
	return cartResponse{
		CartID:   cart.CartID,
		UserID:   cart.UserID,
		Items:    items,
		Subtotal: mapMoney(cart.Subtotal),
		Discount: mapMoney(cart.Discount),
		Total:    mapMoney(cart.Total),
	}
}

func mapMoney(money *domain.Money) *moneyResponse {
	if money == nil {
		return nil
	}
	return &moneyResponse{
		Amount:   money.Amount,
		Currency: money.Currency,
	}
}
