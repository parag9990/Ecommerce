package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/repository"
)

type adminOrderRepository interface {
	GetOrder(context.Context, string) (domain.Order, error)
	ListOrdersForAdmin(context.Context, string, string, string, int, int) (repository.AdminOrderPage, error)
	GetOrderStatusHistoryForAdmin(context.Context, string) ([]domain.OrderStatusHistoryEntry, error)
}

type adminOrderHandler struct {
	token  string
	orders adminOrderRepository
}

func (h *adminOrderHandler) register(mux *http.ServeMux) {
	mux.HandleFunc("GET /internal/admin/orders", h.list)
	mux.HandleFunc("GET /internal/admin/orders/{order_id}", h.get)
	mux.HandleFunc("GET /internal/admin/orders/{order_id}/status-history", h.history)
}

func (h *adminOrderHandler) authorize(w http.ResponseWriter, r *http.Request) bool {
	if authorizedBearer(r.Header.Get("Authorization"), h.token) && strings.TrimSpace(r.Header.Get("X-Admin-Id")) != "" {
		return true
	}
	writeJSON(w, http.StatusUnauthorized, map[string]string{"code": "UNAUTHORIZED", "message": "Internal admin authorization is required"})
	return false
}

func (h *adminOrderHandler) list(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(w, r) {
		return
	}
	page, pageSize, err := orderAdminPagination(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}
	result, err := h.orders.ListOrdersForAdmin(r.Context(), r.URL.Query().Get("status"), r.URL.Query().Get("user_id"), r.URL.Query().Get("seller_id"), pageSize, (page-1)*pageSize)
	if err != nil {
		writeAdminOrderError(w, err)
		return
	}
	orders := make([]adminOrderResponse, 0, len(result.Orders))
	for _, order := range result.Orders {
		orders = append(orders, mapAdminOrder(order))
	}
	writeJSON(w, http.StatusOK, map[string]any{"orders": orders, "page": page, "page_size": pageSize, "total": result.Total})
}

func (h *adminOrderHandler) get(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(w, r) {
		return
	}
	order, err := h.orders.GetOrder(r.Context(), r.PathValue("order_id"))
	if err != nil {
		writeAdminOrderError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapAdminOrder(order))
}

func (h *adminOrderHandler) history(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(w, r) {
		return
	}
	history, err := h.orders.GetOrderStatusHistoryForAdmin(r.Context(), r.PathValue("order_id"))
	if err != nil {
		writeAdminOrderError(w, err)
		return
	}
	items := make([]adminOrderStatusEvent, 0, len(history))
	for _, entry := range history {
		created := entry.CreatedAt
		items = append(items, adminOrderStatusEvent{Status: string(entry.ToStatus), Reason: entry.Reason, ActorID: entry.ActorID, OccurredAt: &created, Metadata: entry.Metadata})
	}
	writeJSON(w, http.StatusOK, map[string]any{"history": items})
}

type adminOrderResponse struct {
	OrderID     string           `json:"order_id"`
	UserID      string           `json:"user_id"`
	SellerID    string           `json:"seller_id,omitempty"`
	Status      string           `json:"status"`
	TotalAmount adminMoney       `json:"total_amount"`
	Items       []adminOrderItem `json:"items"`
	CreatedAt   *time.Time       `json:"created_at,omitempty"`
	UpdatedAt   *time.Time       `json:"updated_at,omitempty"`
	Metadata    map[string]any   `json:"metadata,omitempty"`
}
type adminMoney struct {
	Currency string `json:"currency"`
	Amount   int64  `json:"amount"`
}
type adminOrderItem struct {
	OrderItemID string     `json:"order_item_id"`
	ProductID   string     `json:"product_id"`
	VariantID   string     `json:"variant_id"`
	SellerID    string     `json:"seller_id"`
	Quantity    int        `json:"quantity"`
	Status      string     `json:"status"`
	Amount      adminMoney `json:"amount"`
}
type adminOrderStatusEvent struct {
	Status     string         `json:"status"`
	Reason     string         `json:"reason,omitempty"`
	ActorID    string         `json:"actor_id,omitempty"`
	OccurredAt *time.Time     `json:"occurred_at,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

func mapAdminOrder(order domain.Order) adminOrderResponse {
	created, updated := order.CreatedAt, order.UpdatedAt
	response := adminOrderResponse{OrderID: order.OrderID, UserID: order.UserID, Status: string(order.Status), TotalAmount: adminMoney{Currency: order.Currency, Amount: order.TotalAmount}, Items: make([]adminOrderItem, 0, len(order.Items)), CreatedAt: &created, UpdatedAt: &updated}
	sellers := map[string]struct{}{}
	for _, item := range order.Items {
		sellers[item.SellerID] = struct{}{}
		response.Items = append(response.Items, adminOrderItem{OrderItemID: item.OrderItemID, ProductID: item.ProductID, VariantID: item.VariantID, SellerID: item.SellerID, Quantity: int(item.Quantity), Status: string(item.FulfillmentStatus), Amount: adminMoney{Currency: item.Currency, Amount: item.TotalAmount}})
	}
	if len(sellers) == 1 {
		for sellerID := range sellers {
			response.SellerID = sellerID
		}
	}
	if order.PaymentID != "" {
		response.Metadata = map[string]any{"payment_id": order.PaymentID}
	}
	return response
}

func orderAdminPagination(r *http.Request) (int, int, error) {
	page, pageSize := 1, 20
	var err error
	if raw := strings.TrimSpace(r.URL.Query().Get("page")); raw != "" {
		page, err = strconv.Atoi(raw)
		if err != nil || page < 1 {
			return 0, 0, errors.New("page must be a positive integer")
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("page_size")); raw != "" {
		pageSize, err = strconv.Atoi(raw)
		if err != nil || pageSize < 1 || pageSize > 100 {
			return 0, 0, errors.New("page_size must be between 1 and 100")
		}
	}
	return page, pageSize, nil
}
func writeAdminOrderError(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrOrderNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"code": "NOT_FOUND", "message": "Order not found"})
		return
	}
	if errors.Is(err, domain.ErrInvalidRequest) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"code": "INTERNAL_ERROR", "message": "Internal server error"})
}

var _ = json.Valid
