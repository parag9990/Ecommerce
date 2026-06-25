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

	"ecommerce/api-gateway/internal/clients"
	"ecommerce/api-gateway/internal/domain"
	"ecommerce/api-gateway/internal/validation"
	orderv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const sellerOrderGRPCTimeout = 2 * time.Second

type sellerOrderRPCClient interface {
	ListSellerOrders(context.Context, *orderv1.ListSellerOrdersRequest, ...grpc.CallOption) (*orderv1.ListSellerOrdersResponse, error)
	UpdateFulfillment(context.Context, *orderv1.UpdateFulfillmentRequest, ...grpc.CallOption) (*orderv1.UpdateFulfillmentResponse, error)
}

type SellerOrderHandler struct {
	client  sellerOrderRPCClient
	logger  *slog.Logger
	timeout time.Duration
}

type sellerOrderListResponse struct {
	Orders        []sellerOrderDTO `json:"orders"`
	Total         int              `json:"total"`
	NextPageToken string           `json:"next_page_token,omitempty"`
}

type sellerOrderDTO struct {
	OrderID         string               `json:"order_id"`
	Status          string               `json:"status"`
	Items           []sellerOrderItemDTO `json:"items"`
	Total           orderMoneyDTO        `json:"total"`
	CreatedAt       string               `json:"created_at,omitempty"`
	Shipments       []sellerShipmentDTO  `json:"shipments,omitempty"`
	StatusHistory   []any                `json:"status_history"`
	SellerOrderMeta sellerOrderMetaDTO   `json:"seller_order,omitempty"`
}

type sellerOrderMetaDTO struct {
	ParentOrderStatus       string `json:"parent_order_status,omitempty"`
	SellerFulfillmentStatus string `json:"seller_fulfillment_status,omitempty"`
}

type sellerOrderItemDTO struct {
	OrderItemID       string        `json:"order_item_id,omitempty"`
	ProductID         string        `json:"product_id,omitempty"`
	VariantID         string        `json:"variant_id,omitempty"`
	SKU               string        `json:"sku,omitempty"`
	Title             string        `json:"title,omitempty"`
	ImageURL          string        `json:"image_url,omitempty"`
	Quantity          int32         `json:"quantity"`
	UnitPrice         orderMoneyDTO `json:"unit_price"`
	Total             orderMoneyDTO `json:"total"`
	FulfillmentStatus string        `json:"fulfillment_status,omitempty"`
}

type sellerShipmentDTO struct {
	ShipmentID     string `json:"shipment_id,omitempty"`
	Status         string `json:"status,omitempty"`
	Carrier        string `json:"carrier,omitempty"`
	TrackingNumber string `json:"tracking_number,omitempty"`
	ShippedAt      string `json:"shipped_at,omitempty"`
	DeliveredAt    string `json:"delivered_at,omitempty"`
}

type orderMoneyDTO struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type fulfillmentUpdateRequest struct {
	OrderID        string `json:"order_id"`
	Status         string `json:"status"`
	TrackingNumber string `json:"tracking_number"`
	Carrier        string `json:"carrier"`
	Note           string `json:"note"`
}

func NewSellerOrderHandler(client sellerOrderRPCClient, logger *slog.Logger) *SellerOrderHandler {
	if client == nil {
		panic("order client is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &SellerOrderHandler{client: client, logger: logger, timeout: sellerOrderGRPCTimeout}
}

func sellerOrderRouteEndpoint(route domain.RouteDefinition, handler *SellerOrderHandler, fallback http.HandlerFunc) http.HandlerFunc {
	if handler == nil {
		return fallback
	}
	switch string(route.Method) + " " + route.Path {
	case "GET /api/v1/seller/orders":
		return handler.ListSellerOrders
	case "PATCH /api/v1/seller/orders/{order_id}/fulfillment":
		return handler.UpdateFulfillment
	default:
		return fallback
	}
}

func (h *SellerOrderHandler) ListSellerOrders(w http.ResponseWriter, r *http.Request) {
	request, ok := sellerOrderListRequestFromQuery(w, r)
	if !ok {
		return
	}
	ctx, cancel := h.grpcContext(r)
	defer cancel()

	response, err := h.client.ListSellerOrders(ctx, request)
	if err != nil {
		h.logger.WarnContext(r.Context(), "seller_order_list_failed", "error", err, "request_id", RequestIDFromContext(r.Context()))
		writeMappedError(w, r, err)
		return
	}
	orders := make([]sellerOrderDTO, 0, len(response.GetOrders()))
	for _, order := range response.GetOrders() {
		orders = append(orders, mapSellerOrderDTO(order))
	}
	writeSuccess(w, r, http.StatusOK, sellerOrderListResponse{
		Orders:        orders,
		Total:         len(orders),
		NextPageToken: response.GetNextPageToken(),
	})
}

func (h *SellerOrderHandler) UpdateFulfillment(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimSpace(r.PathValue("order_id"))
	if orderID == "" {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "order_id is required")
		return
	}
	var body fulfillmentUpdateRequest
	if !decodeSellerOrderJSON(w, r, &body) {
		return
	}
	if body.OrderID != "" && body.OrderID != orderID {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "order_id does not match path")
		return
	}
	status, ok := orderStatusFromDashboard(body.Status)
	if !ok || !isSellerFulfillmentTarget(status) {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "status must be one of packed, shipped, or delivered")
		return
	}

	ctx, cancel := h.grpcContext(r)
	defer cancel()
	response, err := h.client.UpdateFulfillment(ctx, &orderv1.UpdateFulfillmentRequest{
		OrderId:        orderID,
		TargetStatus:   status,
		TrackingNumber: strings.TrimSpace(body.TrackingNumber),
		Carrier:        strings.TrimSpace(body.Carrier),
		Note:           strings.TrimSpace(body.Note),
	})
	if err != nil {
		h.logger.WarnContext(r.Context(), "seller_order_fulfillment_update_failed", "error", err, "request_id", RequestIDFromContext(r.Context()))
		writeMappedError(w, r, err)
		return
	}
	if response.GetSellerOrder() == nil {
		writeError(w, r, http.StatusBadGateway, "UPSTREAM_RESPONSE_INVALID", "Order service response did not include a seller order")
		return
	}
	writeSuccess(w, r, http.StatusOK, mapSellerOrderDTO(response.GetSellerOrder()))
}

func (h *SellerOrderHandler) grpcContext(r *http.Request) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	ctx = clients.WithRequestMetadata(ctx, RequestIDFromContext(r.Context()))
	ctx = clients.WithAuthMetadata(ctx)
	return ctx, cancel
}

func sellerOrderListRequestFromQuery(w http.ResponseWriter, r *http.Request) (*orderv1.ListSellerOrdersRequest, bool) {
	pageSize := parseSellerOrderPositiveInt(r.URL.Query().Get("page_size"), 20)
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	filter, ok := sellerFulfillmentFilterFromDashboard(r.URL.Query().Get("status"))
	if !ok {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "status filter is not supported")
		return nil, false
	}
	pageToken := strings.TrimSpace(r.URL.Query().Get("cursor"))
	if pageToken == "" {
		pageToken = strings.TrimSpace(r.URL.Query().Get("page_token"))
	}
	return &orderv1.ListSellerOrdersRequest{
		PageSize:          int32(pageSize),
		PageToken:         pageToken,
		FulfillmentFilter: filter,
	}, true
}

func decodeSellerOrderJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
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

func parseSellerOrderPositiveInt(value string, fallback int) int {
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

func sellerFulfillmentFilterFromDashboard(value string) (orderv1.SellerFulfillmentStatus, bool) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", "all":
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_UNSPECIFIED, true
	case "created", "pending", "pending_payment", "paid":
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PENDING, true
	case "partially_packed":
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PARTIALLY_PACKED, true
	case "packed":
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PACKED, true
	case "partially_shipped":
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PARTIALLY_SHIPPED, true
	case "shipped":
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_SHIPPED, true
	case "partially_delivered":
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PARTIALLY_DELIVERED, true
	case "delivered":
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_DELIVERED, true
	case "cancelled", "refunded":
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_UNSPECIFIED, true
	default:
		return orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_UNSPECIFIED, false
	}
}

func orderStatusFromDashboard(value string) (orderv1.OrderStatus, bool) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "created":
		return orderv1.OrderStatus_ORDER_STATUS_CREATED, true
	case "pending_payment":
		return orderv1.OrderStatus_ORDER_STATUS_PENDING_PAYMENT, true
	case "paid":
		return orderv1.OrderStatus_ORDER_STATUS_PAID, true
	case "packed":
		return orderv1.OrderStatus_ORDER_STATUS_PACKED, true
	case "shipped":
		return orderv1.OrderStatus_ORDER_STATUS_SHIPPED, true
	case "delivered":
		return orderv1.OrderStatus_ORDER_STATUS_DELIVERED, true
	case "cancelled":
		return orderv1.OrderStatus_ORDER_STATUS_CANCELLED, true
	case "refunded":
		return orderv1.OrderStatus_ORDER_STATUS_REFUNDED, true
	default:
		return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED, false
	}
}

func isSellerFulfillmentTarget(status orderv1.OrderStatus) bool {
	return status == orderv1.OrderStatus_ORDER_STATUS_PACKED ||
		status == orderv1.OrderStatus_ORDER_STATUS_SHIPPED ||
		status == orderv1.OrderStatus_ORDER_STATUS_DELIVERED
}

func mapSellerOrderDTO(order *orderv1.SellerOrderView) sellerOrderDTO {
	if order == nil {
		return sellerOrderDTO{Items: []sellerOrderItemDTO{}, StatusHistory: []any{}}
	}
	items := make([]sellerOrderItemDTO, 0, len(order.GetItems()))
	for _, item := range order.GetItems() {
		items = append(items, sellerOrderItemDTO{
			OrderItemID:       item.GetOrderItemId(),
			ProductID:         item.GetProductId(),
			VariantID:         item.GetVariantId(),
			SKU:               item.GetSku(),
			Title:             item.GetTitleSnapshot(),
			ImageURL:          item.GetImageUrlSnapshot(),
			Quantity:          item.GetQuantity(),
			UnitPrice:         mapOrderMoneyDTO(item.GetUnitPrice()),
			Total:             mapOrderMoneyDTO(item.GetLineTotal()),
			FulfillmentStatus: itemFulfillmentStatusString(item.GetFulfillmentStatus()),
		})
	}
	shipments := make([]sellerShipmentDTO, 0, len(order.GetShipments()))
	for _, shipment := range order.GetShipments() {
		shipments = append(shipments, sellerShipmentDTO{
			ShipmentID:     shipment.GetShipmentId(),
			Status:         shipmentStatusString(shipment.GetStatus()),
			Carrier:        shipment.GetCarrier(),
			TrackingNumber: shipment.GetTrackingNumber(),
			ShippedAt:      timestampString(shipment.GetShippedAt()),
			DeliveredAt:    timestampString(shipment.GetDeliveredAt()),
		})
	}
	return sellerOrderDTO{
		OrderID:       order.GetOrderId(),
		Status:        dashboardStatusFromSellerOrder(order),
		Items:         items,
		Total:         mapOrderMoneyDTO(order.GetSellerItemsTotal()),
		CreatedAt:     timestampString(order.GetCreatedAt()),
		Shipments:     shipments,
		StatusHistory: []any{},
		SellerOrderMeta: sellerOrderMetaDTO{
			ParentOrderStatus:       orderStatusString(order.GetParentOrderStatus()),
			SellerFulfillmentStatus: sellerFulfillmentStatusString(order.GetSellerFulfillmentStatus()),
		},
	}
}

func dashboardStatusFromSellerOrder(order *orderv1.SellerOrderView) string {
	switch order.GetSellerFulfillmentStatus() {
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PENDING:
		return "paid"
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PARTIALLY_PACKED,
		orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PACKED:
		return "packed"
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PARTIALLY_SHIPPED,
		orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_SHIPPED:
		return "shipped"
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PARTIALLY_DELIVERED,
		orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_DELIVERED:
		return "delivered"
	default:
		return orderStatusString(order.GetParentOrderStatus())
	}
}

func mapOrderMoneyDTO(money *orderv1.Money) orderMoneyDTO {
	if money == nil {
		return orderMoneyDTO{Currency: "INR"}
	}
	currency := strings.TrimSpace(money.GetCurrency())
	if currency == "" {
		currency = "INR"
	}
	return orderMoneyDTO{Amount: money.GetMinorUnits(), Currency: currency}
}

func timestampString(value *timestamppb.Timestamp) string {
	if value == nil {
		return ""
	}
	t := value.AsTime()
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func orderStatusString(status orderv1.OrderStatus) string {
	switch status {
	case orderv1.OrderStatus_ORDER_STATUS_CREATED:
		return "created"
	case orderv1.OrderStatus_ORDER_STATUS_PENDING_PAYMENT:
		return "pending_payment"
	case orderv1.OrderStatus_ORDER_STATUS_PAID:
		return "paid"
	case orderv1.OrderStatus_ORDER_STATUS_PACKED:
		return "packed"
	case orderv1.OrderStatus_ORDER_STATUS_SHIPPED:
		return "shipped"
	case orderv1.OrderStatus_ORDER_STATUS_DELIVERED:
		return "delivered"
	case orderv1.OrderStatus_ORDER_STATUS_CANCELLED:
		return "cancelled"
	case orderv1.OrderStatus_ORDER_STATUS_REFUNDED:
		return "refunded"
	case orderv1.OrderStatus_ORDER_STATUS_PAYMENT_FAILED:
		return "payment_failed"
	default:
		return "created"
	}
}

func sellerFulfillmentStatusString(status orderv1.SellerFulfillmentStatus) string {
	switch status {
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PENDING:
		return "pending"
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PARTIALLY_PACKED:
		return "partially_packed"
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PACKED:
		return "packed"
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PARTIALLY_SHIPPED:
		return "partially_shipped"
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_SHIPPED:
		return "shipped"
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PARTIALLY_DELIVERED:
		return "partially_delivered"
	case orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_DELIVERED:
		return "delivered"
	default:
		return ""
	}
}

func itemFulfillmentStatusString(status orderv1.ItemFulfillmentStatus) string {
	switch status {
	case orderv1.ItemFulfillmentStatus_ITEM_FULFILLMENT_STATUS_PENDING:
		return "pending"
	case orderv1.ItemFulfillmentStatus_ITEM_FULFILLMENT_STATUS_PACKED:
		return "packed"
	case orderv1.ItemFulfillmentStatus_ITEM_FULFILLMENT_STATUS_SHIPPED:
		return "shipped"
	case orderv1.ItemFulfillmentStatus_ITEM_FULFILLMENT_STATUS_DELIVERED:
		return "delivered"
	case orderv1.ItemFulfillmentStatus_ITEM_FULFILLMENT_STATUS_CANCELLED:
		return "cancelled"
	case orderv1.ItemFulfillmentStatus_ITEM_FULFILLMENT_STATUS_RETURNED:
		return "returned"
	default:
		return ""
	}
}

func shipmentStatusString(status orderv1.ShipmentStatus) string {
	switch status {
	case orderv1.ShipmentStatus_SHIPMENT_STATUS_PENDING:
		return "paid"
	case orderv1.ShipmentStatus_SHIPMENT_STATUS_PACKED:
		return "packed"
	case orderv1.ShipmentStatus_SHIPMENT_STATUS_SHIPPED:
		return "shipped"
	case orderv1.ShipmentStatus_SHIPMENT_STATUS_DELIVERED:
		return "delivered"
	case orderv1.ShipmentStatus_SHIPMENT_STATUS_FAILED:
		return "cancelled"
	default:
		return ""
	}
}
