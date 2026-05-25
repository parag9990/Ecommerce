package httptransport

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/usecase"
)

type SchemaUsecase interface {
	Spec(ctx context.Context) (domain.CollectionSpec, error)
	EnsureCollections(ctx context.Context) (domain.CollectionReport, error)
	Ready(ctx context.Context) error
}

type CartUsecase interface {
	AddItem(ctx context.Context, cmd usecase.AddItemCommand) (*domain.Cart, error)
	RemoveItem(ctx context.Context, cmd usecase.RemoveItemCommand) (*domain.Cart, error)
	ApplyCouponPreview(ctx context.Context, cmd usecase.ApplyCouponPreviewCommand) (*usecase.CouponPreviewResult, error)
	MergeGuestCart(ctx context.Context, cmd usecase.MergeGuestCartCommand) (*domain.Cart, error)
}

type Handler struct {
	schemaUsecase SchemaUsecase
	cartUsecase   CartUsecase
	logger        *slog.Logger
}

func NewHandler(schemaUsecase SchemaUsecase, cartUsecase CartUsecase, logger *slog.Logger) (*Handler, error) {
	if schemaUsecase == nil {
		return nil, errors.New("schema usecase is required")
	}
	if cartUsecase == nil {
		return nil, errors.New("cart usecase is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{schemaUsecase: schemaUsecase, cartUsecase: cartUsecase, logger: logger}, nil
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/cart/items", h.handleAddItem)
	mux.HandleFunc("/api/v1/cart/items/", h.handleRemoveItem)
	mux.HandleFunc("/api/v1/cart/coupons/preview", h.handleApplyCouponPreview)
	mux.HandleFunc("/api/v1/cart/merge", h.handleMergeGuestCart)
	mux.HandleFunc("/internal/v1/cart/schema", h.handleSchema)
	mux.HandleFunc("/internal/v1/cart/schema/bootstrap", h.handleBootstrapSchema)
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("/readyz", h.handleReady)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func (h *Handler) handleReady(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.schemaUsecase.Ready(ctx); err != nil {
		h.logger.Error("cart.ready.failed", slog.String("error", err.Error()))
		writeAPIError(w, http.StatusServiceUnavailable, "NOT_READY", "cart service dependencies are not ready")
		return
	}
	writeJSON(w, http.StatusOK, healthResponse{Status: "ready"})
}

func (h *Handler) handleSchema(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	spec, err := h.schemaUsecase.Spec(r.Context())
	if err != nil {
		h.logger.Error("cart.schema.spec_failed", slog.String("error", err.Error()))
		writeAPIError(w, http.StatusInternalServerError, "SCHEMA_SPEC_FAILED", "could not load cart schema specification")
		return
	}
	writeJSON(w, http.StatusOK, specResponse(spec))
}

func (h *Handler) handleBootstrapSchema(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	report, err := h.schemaUsecase.EnsureCollections(r.Context())
	if err != nil {
		h.logger.Error("cart.schema.bootstrap_failed", slog.String("error", err.Error()))
		writeAPIError(w, http.StatusInternalServerError, "SCHEMA_BOOTSTRAP_FAILED", "could not ensure cart collection schema")
		return
	}
	writeJSON(w, http.StatusOK, reportResponse(report))
}

func (h *Handler) handleAddItem(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	defer r.Body.Close()

	var req addItemRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", "request body must be valid JSON")
		return
	}

	cart, err := h.cartUsecase.AddItem(r.Context(), usecase.AddItemCommand{
		UserID:         ownerHeader(r, "X-User-ID"),
		GuestSessionID: firstHeader(r, "X-Guest-Session-ID", "X-Guest-Session-Id", "X-Session-ID"),
		ProductID:      req.ProductID,
		VariantID:      req.VariantID,
		Quantity:       req.Quantity,
	})
	if err != nil {
		h.writeMergeCartError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newCartResponse(cart))
}

func (h *Handler) handleRemoveItem(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodDelete) {
		return
	}

	itemID := strings.TrimPrefix(r.URL.Path, "/api/v1/cart/items/")
	if strings.Contains(itemID, "/") {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "cart item route was not found")
		return
	}

	cart, err := h.cartUsecase.RemoveItem(r.Context(), usecase.RemoveItemCommand{
		UserID:         ownerHeader(r, "X-User-ID"),
		GuestSessionID: firstHeader(r, "X-Guest-Session-ID", "X-Guest-Session-Id", "X-Session-ID"),
		ItemID:         itemID,
	})
	if err != nil {
		h.writeCartError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newCartResponse(cart))
}

func (h *Handler) handleApplyCouponPreview(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	defer r.Body.Close()

	var req couponPreviewRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", "request body must be valid JSON")
		return
	}

	preview, err := h.cartUsecase.ApplyCouponPreview(r.Context(), usecase.ApplyCouponPreviewCommand{
		UserID:         ownerHeader(r, "X-User-ID"),
		GuestSessionID: firstHeader(r, "X-Guest-Session-ID", "X-Guest-Session-Id", "X-Session-ID"),
		CartID:         req.CartID,
		CouponCode:     req.CouponCode,
	})
	if err != nil {
		h.writeCartError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newCouponPreviewResponse(preview))
}

func (h *Handler) handleMergeGuestCart(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	defer r.Body.Close()

	var req mergeCartRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", "request body must be valid JSON")
		return
	}

	cart, err := h.cartUsecase.MergeGuestCart(r.Context(), usecase.MergeGuestCartCommand{
		UserID:         ownerHeader(r, "X-User-ID"),
		GuestSessionID: firstHeader(r, "X-Guest-Session-ID", "X-Guest-Session-Id", "X-Session-ID"),
		GuestCartID:    req.GuestCartID,
	})
	if err != nil {
		h.writeCartError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newCartResponse(cart))
}

func specResponse(spec domain.CollectionSpec) collectionSpecResponse {
	statuses := make([]string, 0, len(spec.Statuses))
	for _, status := range spec.Statuses {
		statuses = append(statuses, string(status))
	}
	indexes := make([]indexResponse, 0, len(spec.Indexes))
	for _, index := range spec.Indexes {
		indexes = append(indexes, indexResponse{
			Name:    index.Name,
			Purpose: index.Purpose,
			Unique:  index.Unique,
			TTL:     index.TTL,
		})
	}
	return collectionSpecResponse{
		DatabaseName:   spec.DatabaseName,
		CollectionName: spec.CollectionName,
		MaxItems:       spec.MaxItems,
		Statuses:       statuses,
		Indexes:        indexes,
	}
}

func (h *Handler) writeCartError(w http.ResponseWriter, err error) {
	var stockErr domain.StockError
	if errors.As(err, &stockErr) {
		writeAPIErrorDetails(w, http.StatusConflict, "INSUFFICIENT_STOCK", "requested quantity is not available for this variant", map[string]any{
			"requested_quantity": stockErr.Requested,
			"available_quantity": stockErr.Available,
		})
		return
	}
	switch {
	case errors.Is(err, domain.ErrCartOwnerMissing), errors.Is(err, domain.ErrInvalidCartOwner):
		writeAPIError(w, http.StatusUnauthorized, "CART_OWNER_REQUIRED", "user or guest session is required")
	case errors.Is(err, domain.ErrUserIDRequired):
		writeAPIError(w, http.StatusUnauthorized, "USER_AUTH_REQUIRED", "authenticated user is required")
	case errors.Is(err, domain.ErrGuestCartIDRequired):
		writeAPIError(w, http.StatusBadRequest, "GUEST_CART_ID_REQUIRED", "guest_cart_id is required")
	case errors.Is(err, domain.ErrGuestSessionIDRequired), errors.Is(err, domain.ErrGuestCartAccessDenied):
		writeAPIError(w, http.StatusForbidden, "GUEST_CART_ACCESS_DENIED", "guest cart does not belong to this session")
	case errors.Is(err, domain.ErrProductIDRequired):
		writeAPIError(w, http.StatusBadRequest, "PRODUCT_ID_REQUIRED", "product_id is required")
	case errors.Is(err, domain.ErrVariantIDRequired):
		writeAPIError(w, http.StatusBadRequest, "VARIANT_ID_REQUIRED", "variant_id is required")
	case errors.Is(err, domain.ErrItemIDRequired):
		writeAPIError(w, http.StatusBadRequest, "ITEM_ID_REQUIRED", "item_id is required")
	case errors.Is(err, domain.ErrCouponCodeRequired):
		writeAPIError(w, http.StatusBadRequest, "COUPON_CODE_REQUIRED", "coupon_code is required")
	case errors.Is(err, domain.ErrCouponCodeTooLong):
		writeAPIError(w, http.StatusBadRequest, "COUPON_CODE_TOO_LONG", "coupon_code must not exceed 64 characters")
	case errors.Is(err, domain.ErrQuantityTooSmall):
		writeAPIError(w, http.StatusBadRequest, "QUANTITY_TOO_SMALL", "quantity must be at least 1")
	case errors.Is(err, domain.ErrQuantityTooLarge):
		writeAPIError(w, http.StatusBadRequest, "QUANTITY_TOO_LARGE", "quantity must not exceed 10")
	case errors.Is(err, domain.ErrProductNotFound):
		writeAPIError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "product was not found")
	case errors.Is(err, domain.ErrVariantNotFound):
		writeAPIError(w, http.StatusNotFound, "VARIANT_NOT_FOUND", "variant was not found")
	case errors.Is(err, domain.ErrProductNotSellable):
		writeAPIError(w, http.StatusConflict, "PRODUCT_NOT_SELLABLE", "product is not available for sale")
	case errors.Is(err, domain.ErrVariantNotSellable):
		writeAPIError(w, http.StatusConflict, "VARIANT_NOT_SELLABLE", "variant is not available for sale")
	case errors.Is(err, domain.ErrProductPriceInvalid):
		writeAPIError(w, http.StatusConflict, "PRODUCT_PRICE_INVALID", "product price is not available")
	case errors.Is(err, domain.ErrCartItemLimitReached):
		writeAPIError(w, http.StatusConflict, "CART_ITEM_LIMIT_REACHED", "cart item limit reached")
	case errors.Is(err, domain.ErrCartNotFound):
		writeAPIError(w, http.StatusNotFound, "CART_NOT_FOUND", "active cart was not found for this shopper")
	case errors.Is(err, domain.ErrCartNotActive):
		writeAPIError(w, http.StatusConflict, "CART_NOT_ACTIVE", "cart is not editable")
	case errors.Is(err, domain.ErrCartExpired):
		writeAPIError(w, http.StatusConflict, "CART_EXPIRED", "cart expired due to inactivity")
	case errors.Is(err, domain.ErrEmptyCart):
		writeAPIError(w, http.StatusConflict, "EMPTY_CART", "add items before applying a coupon")
	case errors.Is(err, domain.ErrCartOwnershipMismatch):
		writeAPIError(w, http.StatusForbidden, "CART_OWNERSHIP_MISMATCH", "cart does not belong to this shopper")
	case errors.Is(err, domain.ErrMixedCurrencyCart), errors.Is(err, domain.ErrInvalidCartTotals):
		writeAPIError(w, http.StatusConflict, "CART_TOTALS_INVALID", "cart totals could not be recalculated")
	case errors.Is(err, domain.ErrCouponCurrencyMismatch), errors.Is(err, domain.ErrCouponDiscountTooLarge), errors.Is(err, domain.ErrInvalidCouponPreview):
		writeAPIError(w, http.StatusBadGateway, "COUPON_PREVIEW_INVALID", "coupon preview result was invalid")
	case errors.Is(err, domain.ErrCartVersionConflict):
		writeAPIError(w, http.StatusConflict, "CART_VERSION_CONFLICT", "cart was updated concurrently; retry the request")
	case errors.Is(err, domain.ErrCouponPreviewUnavailable):
		writeAPIError(w, http.StatusServiceUnavailable, "COUPON_PREVIEW_UNAVAILABLE", "coupon preview is temporarily unavailable")
	case errors.Is(err, domain.ErrProductUnavailable), errors.Is(err, domain.ErrCartStoreUnavailable):
		writeAPIError(w, http.StatusServiceUnavailable, "CART_TEMPORARILY_UNAVAILABLE", "cart is temporarily unavailable")
	default:
		h.logger.Error("cart.mutation.failed", slog.String("error", err.Error()))
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not update cart")
	}
}

func (h *Handler) writeMergeCartError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrUserIDRequired):
		writeAPIError(w, http.StatusUnauthorized, "USER_AUTH_REQUIRED", "authenticated user is required")
	case errors.Is(err, domain.ErrGuestCartIDRequired):
		writeAPIError(w, http.StatusBadRequest, "GUEST_CART_ID_REQUIRED", "guest_cart_id is required")
	case errors.Is(err, domain.ErrGuestSessionIDRequired), errors.Is(err, domain.ErrGuestCartAccessDenied), errors.Is(err, domain.ErrCartOwnershipMismatch):
		writeAPIError(w, http.StatusForbidden, "GUEST_CART_ACCESS_DENIED", "guest cart does not belong to this session")
	case errors.Is(err, domain.ErrCartNotFound):
		writeAPIError(w, http.StatusNotFound, "GUEST_CART_NOT_FOUND", "guest cart was not found")
	case errors.Is(err, domain.ErrCartNotActive):
		writeAPIError(w, http.StatusConflict, "GUEST_CART_NOT_ACTIVE", "guest cart is not editable")
	case errors.Is(err, domain.ErrCartExpired):
		writeAPIError(w, http.StatusConflict, "CART_EXPIRED", "cart expired due to inactivity")
	case errors.Is(err, domain.ErrMixedCurrencyCart):
		writeAPIError(w, http.StatusConflict, "MIXED_CURRENCY_CART", "guest cart currency does not match user cart currency")
	default:
		h.writeCartError(w, err)
	}
}

func ownerHeader(r *http.Request, key string) string {
	return r.Header.Get(key)
}

func firstHeader(r *http.Request, keys ...string) string {
	for _, key := range keys {
		if value := r.Header.Get(key); value != "" {
			return value
		}
	}
	return ""
}

func reportResponse(report domain.CollectionReport) collectionReportResponse {
	return collectionReportResponse{
		DatabaseName:      report.DatabaseName,
		CollectionName:    report.CollectionName,
		CollectionCreated: report.CollectionCreated,
		ValidatorUpdated:  report.ValidatorUpdated,
		IndexesEnsured:    report.IndexesEnsured,
	}
}
