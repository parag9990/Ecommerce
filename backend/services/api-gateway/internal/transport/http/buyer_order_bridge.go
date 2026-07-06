package httptransport

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"ecommerce/api-gateway/internal/clients"
	"ecommerce/api-gateway/internal/domain"
	"ecommerce/api-gateway/internal/validation"
	orderv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/order/v1"
	userv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/user/v1"
	"google.golang.org/grpc"
)

const buyerOrderGRPCTimeout = 3 * time.Second

type buyerOrderRPCClient interface {
	CreateOrder(context.Context, *orderv1.CreateOrderRequest, ...grpc.CallOption) (*orderv1.CreateOrderResponse, error)
	GetOrder(context.Context, *orderv1.GetOrderRequest, ...grpc.CallOption) (*orderv1.GetOrderResponse, error)
	ListOrders(context.Context, *orderv1.ListOrdersRequest, ...grpc.CallOption) (*orderv1.ListOrdersResponse, error)
	CancelOrder(context.Context, *orderv1.CancelOrderRequest, ...grpc.CallOption) (*orderv1.CancelOrderResponse, error)
}

type BuyerOrderHandler struct {
	orderClient buyerOrderRPCClient
	userClient  clients.UserClient
	logger      *slog.Logger
	timeout     time.Duration
}

type checkoutRequestBody struct {
	AddressID       string `json:"address_id"`
	CartID          string `json:"cart_id"`
	CouponCode      string `json:"coupon_code"`
	IdempotencyKey  string `json:"idempotency_key"`
	PaymentProvider string `json:"payment_provider"`
}

type checkoutResponseDTO struct {
	Order         buyerOrderDTO     `json:"order"`
	PaymentIntent *paymentIntentDTO `json:"payment_intent,omitempty"`
}

type paymentIntentDTO struct {
	Amount       orderMoneyDTO `json:"amount"`
	ClientSecret string        `json:"client_secret,omitempty"`
	ExpiresAt    string        `json:"expires_at,omitempty"`
	PaymentID    string        `json:"payment_id,omitempty"`
	Provider     string        `json:"provider,omitempty"`
	Status       string        `json:"status,omitempty"`
}

type buyerOrderListResponse struct {
	Orders        []buyerOrderDTO `json:"orders"`
	Total         int             `json:"total"`
	NextPageToken string          `json:"next_page_token,omitempty"`
}

type buyerOrderDTO struct {
	CreatedAt         string                     `json:"created_at,omitempty"`
	FulfillmentStatus string                     `json:"fulfillment_status,omitempty"`
	Items             []buyerOrderItemDTO        `json:"items"`
	OrderID           string                     `json:"order_id"`
	PaymentID         string                     `json:"payment_id,omitempty"`
	PaymentStatus     string                     `json:"payment_status,omitempty"`
	ShippingAddress   *buyerOrderAddressDTO      `json:"shipping_address,omitempty"`
	Status            string                     `json:"status"`
	StatusHistory     []buyerOrderStatusEventDTO `json:"status_history"`
	Total             orderMoneyDTO              `json:"total"`
	UpdatedAt         string                     `json:"updated_at,omitempty"`
	UserID            string                     `json:"user_id"`
}

type buyerOrderItemDTO struct {
	LineTotal   orderMoneyDTO `json:"line_total"`
	OrderItemID string        `json:"order_item_id,omitempty"`
	Price       orderMoneyDTO `json:"price"`
	ProductID   string        `json:"product_id,omitempty"`
	Quantity    int32         `json:"quantity"`
	SellerID    string        `json:"seller_id,omitempty"`
	Title       string        `json:"title,omitempty"`
	UnitPrice   orderMoneyDTO `json:"unit_price"`
	VariantID   string        `json:"variant_id,omitempty"`
}

type buyerOrderAddressDTO struct {
	City       string `json:"city,omitempty"`
	Country    string `json:"country,omitempty"`
	Line1      string `json:"line1,omitempty"`
	Line2      string `json:"line2,omitempty"`
	Name       string `json:"name,omitempty"`
	Phone      string `json:"phone,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	State      string `json:"state,omitempty"`
}

type buyerOrderStatusEventDTO struct {
	CreatedAt string `json:"created_at,omitempty"`
	Status    string `json:"status,omitempty"`
}

type buyerCancelOrderRequest struct {
	Reason string `json:"reason"`
}

func NewBuyerOrderHandler(orderClient buyerOrderRPCClient, userClient clients.UserClient, logger *slog.Logger) *BuyerOrderHandler {
	if orderClient == nil {
		panic("order client is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &BuyerOrderHandler{
		orderClient: orderClient,
		userClient:  userClient,
		logger:      logger,
		timeout:     buyerOrderGRPCTimeout,
	}
}

func buyerOrderRouteEndpoint(route domain.RouteDefinition, handler *BuyerOrderHandler, fallback http.HandlerFunc) http.HandlerFunc {
	if handler == nil {
		return fallback
	}
	switch string(route.Method) + " " + route.Path {
	case "POST /api/v1/orders/checkout":
		return handler.CreateCheckout
	case "GET /api/v1/orders":
		return handler.ListOrders
	case "GET /api/v1/orders/{order_id}":
		return handler.GetOrder
	case "POST /api/v1/orders/{order_id}/cancel":
		return handler.CancelOrder
	default:
		return fallback
	}
}

func (h *BuyerOrderHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	var body checkoutRequestBody
	if !decodeBuyerOrderJSON(w, r, &body) {
		return
	}
	if key, ok := validation.IdempotencyKeyFromContext(r.Context()); ok && body.IdempotencyKey == "" {
		body.IdempotencyKey = key
	}
	body.CartID = strings.TrimSpace(body.CartID)
	body.AddressID = strings.TrimSpace(body.AddressID)
	body.IdempotencyKey = strings.TrimSpace(body.IdempotencyKey)
	body.PaymentProvider = strings.TrimSpace(body.PaymentProvider)
	if body.CartID == "" || body.AddressID == "" || body.IdempotencyKey == "" {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "cart_id, address_id, and idempotency_key are required")
		return
	}

	claims, ok := gatewayauth.ClaimsFromContext(r.Context())
	if !ok || strings.TrimSpace(claims.UserID()) == "" {
		writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	address, resolved := h.resolveAddress(w, r, claims.UserID(), body.AddressID)
	if !resolved {
		return
	}
	countryCode, ok := checkoutCountryCode(address.GetCountry())
	if !ok {
		writeError(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Delivery address country must be a supported ISO country code")
		return
	}

	ctx, cancel := h.grpcContext(r)
	defer cancel()
	response, err := h.orderClient.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		CartId: body.CartID,
		ShippingAddress: &orderv1.AddressSnapshot{
			RecipientName: address.GetName(),
			Phone:         address.GetPhone(),
			Line1:         address.GetLine1(),
			Line2:         address.GetLine2(),
			City:          address.GetCity(),
			State:         address.GetState(),
			PostalCode:    address.GetPostalCode(),
			CountryCode:   countryCode,
		},
		IdempotencyKey: body.IdempotencyKey,
	})
	if err != nil {
		h.logger.WarnContext(r.Context(), "buyer_order_checkout_failed", "error", err, "request_id", RequestIDFromContext(r.Context()))
		writeMappedError(w, r, err)
		return
	}
	if response.GetOrder() == nil {
		writeError(w, r, http.StatusBadGateway, "UPSTREAM_RESPONSE_INVALID", "Order service response did not include an order")
		return
	}
	writeSuccess(w, r, http.StatusOK, checkoutResponseDTO{
		Order:         mapBuyerOrderDTO(response.GetOrder()),
		PaymentIntent: mapPaymentIntentDTO(response.GetPaymentAction(), response.GetOrder(), body.PaymentProvider),
	})
}

func (h *BuyerOrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	pageSize := parseBuyerOrderPositiveInt(r.URL.Query().Get("page_size"), 20)
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	pageToken := strings.TrimSpace(r.URL.Query().Get("page_token"))
	if pageToken == "" {
		pageToken = strings.TrimSpace(r.URL.Query().Get("cursor"))
	}
	statusFilter, ok := buyerOrderStatusFilter(r.URL.Query().Get("status"))
	if !ok {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "status filter is not supported")
		return
	}

	ctx, cancel := h.grpcContext(r)
	defer cancel()
	response, err := h.orderClient.ListOrders(ctx, &orderv1.ListOrdersRequest{
		PageSize:     int32(pageSize),
		PageToken:    pageToken,
		StatusFilter: statusFilter,
	})
	if err != nil {
		h.logger.WarnContext(r.Context(), "buyer_order_list_failed", "error", err, "request_id", RequestIDFromContext(r.Context()))
		writeMappedError(w, r, err)
		return
	}
	orders := make([]buyerOrderDTO, 0, len(response.GetOrders()))
	for _, order := range response.GetOrders() {
		orders = append(orders, mapBuyerOrderDTO(order))
	}
	writeSuccess(w, r, http.StatusOK, buyerOrderListResponse{
		Orders:        orders,
		Total:         len(orders),
		NextPageToken: response.GetNextPageToken(),
	})
}

func (h *BuyerOrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimSpace(r.PathValue("order_id"))
	if orderID == "" {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "order_id is required")
		return
	}
	ctx, cancel := h.grpcContext(r)
	defer cancel()
	response, err := h.orderClient.GetOrder(ctx, &orderv1.GetOrderRequest{OrderId: orderID})
	if err != nil {
		h.logger.WarnContext(r.Context(), "buyer_order_get_failed", "error", err, "request_id", RequestIDFromContext(r.Context()))
		writeMappedError(w, r, err)
		return
	}
	if response.GetOrder() == nil {
		writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Order not found")
		return
	}
	writeSuccess(w, r, http.StatusOK, mapBuyerOrderDTO(response.GetOrder()))
}

func (h *BuyerOrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimSpace(r.PathValue("order_id"))
	if orderID == "" {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "order_id is required")
		return
	}
	var body buyerCancelOrderRequest
	if !decodeBuyerOrderJSON(w, r, &body) {
		return
	}
	reason := strings.TrimSpace(body.Reason)
	if reason == "" {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "reason is required")
		return
	}
	ctx, cancel := h.grpcContext(r)
	defer cancel()
	response, err := h.orderClient.CancelOrder(ctx, &orderv1.CancelOrderRequest{
		OrderId:    orderID,
		ReasonCode: reason,
	})
	if err != nil {
		h.logger.WarnContext(r.Context(), "buyer_order_cancel_failed", "error", err, "request_id", RequestIDFromContext(r.Context()))
		writeMappedError(w, r, err)
		return
	}
	if response.GetOrder() == nil {
		writeError(w, r, http.StatusBadGateway, "UPSTREAM_RESPONSE_INVALID", "Order service response did not include an order")
		return
	}
	writeSuccess(w, r, http.StatusOK, mapBuyerOrderDTO(response.GetOrder()))
}

func (h *BuyerOrderHandler) resolveAddress(w http.ResponseWriter, r *http.Request, userID string, addressID string) (*userv1.Address, bool) {
	if h.userClient == nil {
		writeError(w, r, http.StatusBadGateway, "UPSTREAM_NOT_CONFIGURED", "User address lookup is not configured")
		return nil, false
	}
	ctx, cancel := h.grpcContext(r)
	defer cancel()
	response, err := h.userClient.ListUserAddresses(ctx, &userv1.ListUserAddressesRequest{
		UserId:   userID,
		Page:     1,
		PageSize: 100,
	})
	if err != nil {
		h.logger.WarnContext(r.Context(), "buyer_order_address_lookup_failed", "error", err, "request_id", RequestIDFromContext(r.Context()))
		writeMappedError(w, r, err)
		return nil, false
	}
	for _, address := range response.GetAddresses() {
		if address.GetAddressId() == addressID {
			return address, true
		}
	}
	writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Delivery address not found")
	return nil, false
}

func (h *BuyerOrderHandler) grpcContext(r *http.Request) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	ctx = clients.WithRequestMetadata(ctx, RequestIDFromContext(r.Context()))
	ctx = clients.WithAuthMetadata(ctx)
	return ctx, cancel
}

func decodeBuyerOrderJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if payload, ok := validation.BodyPayloadFromContext(r.Context()); ok {
		body, err := json.Marshal(payload)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Request body is invalid")
			return false
		}
		if err := json.Unmarshal(body, dst); err != nil {
			writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Request body is invalid")
			return false
		}
		return true
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, 64*1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Request body must be valid JSON")
		return false
	}
	return true
}

func parseBuyerOrderPositiveInt(value string, fallback int) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}

func buyerOrderStatusFilter(value string) (orderv1.OrderStatus, bool) {
	if strings.TrimSpace(value) == "" || strings.EqualFold(strings.TrimSpace(value), "all") {
		return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED, true
	}
	status, ok := orderStatusFromDashboard(value)
	return status, ok
}

func checkoutCountryCode(value string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, ".", "")
	normalized = strings.Join(strings.Fields(normalized), " ")
	if len(normalized) == 2 {
		return strings.ToUpper(normalized), true
	}
	switch normalized {
	case "india":
		return "IN", true
	case "united states", "united states of america", "usa":
		return "US", true
	case "united kingdom", "great britain", "uk":
		return "GB", true
	case "canada":
		return "CA", true
	case "australia":
		return "AU", true
	default:
		return "", false
	}
}

func mapBuyerOrderDTO(order *orderv1.Order) buyerOrderDTO {
	if order == nil {
		return buyerOrderDTO{Items: []buyerOrderItemDTO{}, StatusHistory: []buyerOrderStatusEventDTO{}}
	}
	items := make([]buyerOrderItemDTO, 0, len(order.GetItems()))
	for _, item := range order.GetItems() {
		items = append(items, buyerOrderItemDTO{
			LineTotal:   mapOrderMoneyDTO(item.GetLineTotal()),
			OrderItemID: item.GetOrderItemId(),
			Price:       mapOrderMoneyDTO(item.GetLineTotal()),
			ProductID:   item.GetProductId(),
			Quantity:    item.GetQuantity(),
			SellerID:    item.GetSellerId(),
			Title:       item.GetProductName(),
			UnitPrice:   mapOrderMoneyDTO(item.GetUnitPrice()),
			VariantID:   item.GetVariantId(),
		})
	}
	status := orderStatusString(order.GetStatus())
	createdAt := timestampString(order.GetCreatedAt())
	return buyerOrderDTO{
		CreatedAt:         createdAt,
		FulfillmentStatus: fulfillmentStatusForBuyerOrder(order.GetStatus()),
		Items:             items,
		OrderID:           order.GetOrderId(),
		PaymentID:         order.GetPaymentId(),
		PaymentStatus:     paymentStatusForBuyerOrder(order.GetStatus()),
		ShippingAddress:   mapBuyerOrderAddressDTO(order.GetShippingAddress()),
		Status:            status,
		StatusHistory:     []buyerOrderStatusEventDTO{{CreatedAt: createdAt, Status: status}},
		Total:             mapOrderMoneyDTO(order.GetTotal()),
		UpdatedAt:         timestampString(order.GetUpdatedAt()),
		UserID:            order.GetUserId(),
	}
}

func mapBuyerOrderAddressDTO(address *orderv1.AddressSnapshot) *buyerOrderAddressDTO {
	if address == nil {
		return nil
	}
	return &buyerOrderAddressDTO{
		City:       address.GetCity(),
		Country:    address.GetCountryCode(),
		Line1:      address.GetLine1(),
		Line2:      address.GetLine2(),
		Name:       address.GetRecipientName(),
		Phone:      address.GetPhone(),
		PostalCode: address.GetPostalCode(),
		State:      address.GetState(),
	}
}

func mapPaymentIntentDTO(action *orderv1.PaymentAction, order *orderv1.Order, provider string) *paymentIntentDTO {
	if action == nil {
		return nil
	}
	status := "requires_action"
	if strings.TrimSpace(action.GetClientActionToken()) == "" {
		status = "pending"
	}
	return &paymentIntentDTO{
		Amount:       mapOrderMoneyDTO(order.GetTotal()),
		ClientSecret: action.GetClientActionToken(),
		ExpiresAt:    timestampString(action.GetExpiresAt()),
		PaymentID:    action.GetPaymentId(),
		Provider:     strings.TrimSpace(provider),
		Status:       status,
	}
}

func paymentStatusForBuyerOrder(status orderv1.OrderStatus) string {
	switch status {
	case orderv1.OrderStatus_ORDER_STATUS_PENDING_PAYMENT:
		return "pending"
	case orderv1.OrderStatus_ORDER_STATUS_PAID,
		orderv1.OrderStatus_ORDER_STATUS_PACKED,
		orderv1.OrderStatus_ORDER_STATUS_SHIPPED,
		orderv1.OrderStatus_ORDER_STATUS_DELIVERED:
		return "paid"
	case orderv1.OrderStatus_ORDER_STATUS_PAYMENT_FAILED:
		return "failed"
	case orderv1.OrderStatus_ORDER_STATUS_CANCELLED:
		return "cancelled"
	case orderv1.OrderStatus_ORDER_STATUS_REFUNDED:
		return "refunded"
	default:
		return ""
	}
}

func fulfillmentStatusForBuyerOrder(status orderv1.OrderStatus) string {
	switch status {
	case orderv1.OrderStatus_ORDER_STATUS_PACKED:
		return "packed"
	case orderv1.OrderStatus_ORDER_STATUS_SHIPPED:
		return "shipped"
	case orderv1.OrderStatus_ORDER_STATUS_DELIVERED:
		return "delivered"
	case orderv1.OrderStatus_ORDER_STATUS_CANCELLED:
		return "cancelled"
	default:
		return ""
	}
}
