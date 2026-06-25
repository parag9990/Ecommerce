package http

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
)

type OrderPaymentUsecase interface {
	ListOrdersForAdmin(ctx context.Context, req domain.AdminOrderListRequest) (domain.AdminOrderListResponse, error)
	GetOrderForAdmin(ctx context.Context, orderID string) (domain.AdminOrderDetailResponse, error)
	ListOrderDisputes(ctx context.Context, orderID string) (domain.OrderDisputeListResponse, error)
	ListPaymentsForAdmin(ctx context.Context, req domain.AdminPaymentListRequest) (domain.AdminPaymentListResponse, error)
	GetPaymentForAdmin(ctx context.Context, paymentID string) (domain.AdminPaymentDetailResponse, error)
	ListRefundsForAdmin(ctx context.Context, req domain.RefundListRequest) (domain.RefundListResponse, error)
	ReviewRefund(ctx context.Context, refundID string, req domain.RefundReviewRequest) (domain.RefundSnapshot, error)
	MarkOrderManualReview(ctx context.Context, orderID string, req domain.ManualOrderReviewRequest) (*domain.ReviewTask, error)
	GetOrderDisputeView(ctx context.Context, orderID string) (domain.DisputeView, error)
	ListReconciliationAlerts(ctx context.Context, req domain.ReconciliationListRequest) (domain.ReconciliationListResponse, error)
	GetReconciliationAlert(ctx context.Context, reconciliationID string) (domain.ReconciliationDetailResponse, error)
}

type OrderPaymentHandler struct {
	controls OrderPaymentUsecase
	logger   logging.Logger
}

func NewOrderPaymentHandler(controls OrderPaymentUsecase, logger logging.Logger) *OrderPaymentHandler {
	if logger == nil {
		logger = logging.NewNop()
	}
	return &OrderPaymentHandler{controls: controls, logger: logger}
}

func (h *OrderPaymentHandler) Register(mux *http.ServeMux) {
	protected := ActorMiddleware
	mux.Handle("/api/v1/admin/orders", protected(http.HandlerFunc(h.listOrders)))
	mux.Handle("/api/v1/admin/orders/", protected(http.HandlerFunc(h.orderAction)))
	mux.Handle("/api/v1/admin/payments", protected(http.HandlerFunc(h.listPayments)))
	mux.Handle("/api/v1/admin/payments/", protected(http.HandlerFunc(h.paymentAction)))
	mux.Handle("/api/v1/admin/refunds", protected(http.HandlerFunc(h.listRefunds)))
	mux.Handle("/api/v1/admin/refunds/", protected(http.HandlerFunc(h.refundAction)))
	mux.Handle("/api/v1/admin/payment-reconciliations", protected(http.HandlerFunc(h.listReconciliationAlerts)))
	mux.Handle("/api/v1/admin/payment-reconciliations/", protected(http.HandlerFunc(h.reconciliationAction)))
}

func (h *OrderPaymentHandler) listOrders(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/admin/orders" {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/orders"))
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	req, err := orderListRequestFromQuery(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response, err := h.controls.ListOrdersForAdmin(r.Context(), req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, adminOrderListResponseDTO(response))
}

func (h *OrderPaymentHandler) listPayments(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/admin/payments" {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/payments"))
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	req, err := paymentListRequestFromQuery(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response, err := h.controls.ListPaymentsForAdmin(r.Context(), req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, adminPaymentListResponseDTO(response))
}

func (h *OrderPaymentHandler) orderAction(w http.ResponseWriter, r *http.Request) {
	orderID, action, hasAction, ok := resourceActionID(r.URL.Path, "/api/v1/admin/orders/")
	if !ok {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/orders/{order_id} or a supported order action"))
		return
	}
	if !hasAction {
		h.getOrder(w, r, orderID)
		return
	}

	switch action {
	case "manual-review", "review":
		h.markOrderManualReview(w, r, orderID)
	case "dispute-view":
		h.getOrderDisputeView(w, r, orderID)
	case "disputes":
		h.listOrderDisputes(w, r, orderID)
	default:
		writeError(w, r, domain.NewValidationError("unknown order admin action"))
	}
}

func (h *OrderPaymentHandler) getOrder(w http.ResponseWriter, r *http.Request, orderID string) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}
	response, err := h.controls.GetOrderForAdmin(r.Context(), orderID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, adminOrderDetailResponseDTO(response))
}

func (h *OrderPaymentHandler) markOrderManualReview(w http.ResponseWriter, r *http.Request, orderID string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	var req domain.ManualOrderReviewRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	response, err := h.controls.MarkOrderManualReview(r.Context(), orderID, req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "review_task": reviewTaskDTOFromDomain(*response)})
}

func (h *OrderPaymentHandler) getOrderDisputeView(w http.ResponseWriter, r *http.Request, orderID string) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	response, err := h.controls.GetOrderDisputeView(r.Context(), orderID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *OrderPaymentHandler) listOrderDisputes(w http.ResponseWriter, r *http.Request, orderID string) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}
	response, err := h.controls.ListOrderDisputes(r.Context(), orderID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *OrderPaymentHandler) paymentAction(w http.ResponseWriter, r *http.Request) {
	paymentID, _, hasAction, ok := resourceActionID(r.URL.Path, "/api/v1/admin/payments/")
	if !ok || hasAction {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/payments/{payment_id}"))
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}
	response, err := h.controls.GetPaymentForAdmin(r.Context(), paymentID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, adminPaymentDetailResponseDTO(response))
}

func (h *OrderPaymentHandler) listRefunds(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/admin/refunds" {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/refunds"))
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}
	req, err := refundListRequestFromQuery(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response, err := h.controls.ListRefundsForAdmin(r.Context(), req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, refundListResponseDTO(response))
}

func (h *OrderPaymentHandler) refundAction(w http.ResponseWriter, r *http.Request) {
	refundID, action, hasAction, ok := resourceActionID(r.URL.Path, "/api/v1/admin/refunds/")
	if !ok || !hasAction || action != "review" {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/refunds/{refund_id}/review"))
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	var req domain.RefundReviewRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	response, err := h.controls.ReviewRefund(r.Context(), refundID, req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, refundDTOFromDomain(response))
}

func (h *OrderPaymentHandler) listReconciliationAlerts(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/admin/payment-reconciliations" {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/payment-reconciliations"))
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}
	req, err := reconciliationListRequestFromQuery(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response, err := h.controls.ListReconciliationAlerts(r.Context(), req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, reconciliationListResponseDTO(response))
}

func (h *OrderPaymentHandler) reconciliationAction(w http.ResponseWriter, r *http.Request) {
	reconciliationID, action, hasAction, ok := resourceActionID(r.URL.Path, "/api/v1/admin/payment-reconciliations/")
	if !ok || hasAction || action != "" {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/payment-reconciliations/{reconciliation_id}"))
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}
	response, err := h.controls.GetReconciliationAlert(r.Context(), reconciliationID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, reconciliationDetailResponseDTO(response))
}

func orderListRequestFromQuery(r *http.Request) (domain.AdminOrderListRequest, error) {
	query := r.URL.Query()
	page, err := positiveIntQuery(query.Get("page"), "page")
	if err != nil {
		return domain.AdminOrderListRequest{}, err
	}
	pageSize, err := positiveIntQuery(query.Get("page_size"), "page_size")
	if err != nil {
		return domain.AdminOrderListRequest{}, err
	}

	req := domain.AdminOrderListRequest{
		Query:        query.Get("q"),
		Status:       domain.OrderStatus(strings.TrimSpace(query.Get("status"))),
		ReviewStatus: query.Get("review_status"),
		UserID:       query.Get("user_id"),
		SellerID:     query.Get("seller_id"),
		From:         query.Get("from"),
		To:           query.Get("to"),
		Pagination: domain.Pagination{
			Page:     page,
			PageSize: pageSize,
			Cursor:   query.Get("cursor"),
		},
	}
	return req.Normalize()
}

func refundListRequestFromQuery(r *http.Request) (domain.RefundListRequest, error) {
	query := r.URL.Query()
	page, err := positiveIntQuery(query.Get("page"), "page")
	if err != nil {
		return domain.RefundListRequest{}, err
	}
	pageSize, err := positiveIntQuery(query.Get("page_size"), "page_size")
	if err != nil {
		return domain.RefundListRequest{}, err
	}
	req := domain.RefundListRequest{
		Status: query.Get("status"),
		Pagination: domain.Pagination{
			Page:     page,
			PageSize: pageSize,
			Cursor:   query.Get("cursor"),
		},
	}
	return req.Normalize()
}

func reconciliationListRequestFromQuery(r *http.Request) (domain.ReconciliationListRequest, error) {
	query := r.URL.Query()
	page, err := positiveIntQuery(query.Get("page"), "page")
	if err != nil {
		return domain.ReconciliationListRequest{}, err
	}
	pageSize, err := positiveIntQuery(query.Get("page_size"), "page_size")
	if err != nil {
		return domain.ReconciliationListRequest{}, err
	}
	req := domain.ReconciliationListRequest{
		Status: query.Get("status"),
		Pagination: domain.Pagination{
			Page:     page,
			PageSize: pageSize,
			Cursor:   query.Get("cursor"),
		},
	}
	return req.Normalize()
}

func paymentListRequestFromQuery(r *http.Request) (domain.AdminPaymentListRequest, error) {
	query := r.URL.Query()
	page, err := positiveIntQuery(query.Get("page"), "page")
	if err != nil {
		return domain.AdminPaymentListRequest{}, err
	}
	pageSize, err := positiveIntQuery(query.Get("page_size"), "page_size")
	if err != nil {
		return domain.AdminPaymentListRequest{}, err
	}

	req := domain.AdminPaymentListRequest{
		Status:   domain.PaymentStatus(strings.TrimSpace(query.Get("status"))),
		Provider: query.Get("provider"),
		OrderID:  query.Get("order_id"),
		Pagination: domain.Pagination{
			Page:     page,
			PageSize: pageSize,
			Cursor:   query.Get("cursor"),
		},
	}
	return req.Normalize()
}

func resourceActionID(path string, prefix string) (string, string, bool, bool) {
	if !strings.HasPrefix(path, prefix) {
		return "", "", false, false
	}
	remaining := strings.TrimPrefix(path, prefix)
	id, action, ok := strings.Cut(remaining, "/")
	if strings.TrimSpace(id) == "" {
		return "", "", false, false
	}
	if ok && (strings.TrimSpace(action) == "" || strings.Contains(action, "/")) {
		return "", "", false, false
	}
	unescaped, err := url.PathUnescape(id)
	if err != nil || strings.TrimSpace(unescaped) == "" {
		return "", "", false, false
	}
	return unescaped, action, ok, true
}
