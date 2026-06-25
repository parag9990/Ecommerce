package httptransport

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/repository"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/usecase"
)

const maxRequestBodyBytes = 1 << 20

type PaymentStateUsecase interface {
	Definition(ctx context.Context) (domain.StateMachineDefinition, error)
	EvaluateTransition(ctx context.Context, input usecase.EvaluateTransitionInput) (usecase.EvaluateTransitionOutput, error)
	NormalizeProviderEvent(ctx context.Context, providerEvent string) (usecase.ProviderEventOutput, error)
}

type PaymentSchemaUsecase interface {
	Definition(ctx context.Context) (domain.PaymentSchemaDefinition, error)
	Health(ctx context.Context) (usecase.PaymentSchemaHealthOutput, error)
}

type PaymentIntentUsecase interface {
	Execute(ctx context.Context, input usecase.CreatePaymentIntentInput) (usecase.CreatePaymentIntentOutput, error)
}

type RetryPaymentUsecase interface {
	Execute(ctx context.Context, input usecase.RetryPaymentIntentInput) (usecase.RetryPaymentIntentOutput, error)
}

type WebhookUsecase interface {
	Execute(ctx context.Context, input usecase.HandleWebhookInput) (usecase.HandleWebhookOutput, error)
}

type RefundUsecase interface {
	Execute(ctx context.Context, input usecase.RefundPaymentInput) (usecase.RefundPaymentOutput, error)
	Get(ctx context.Context, refundID string) (domain.Refund, error)
	Review(ctx context.Context, input usecase.ReviewRefundInput) (usecase.RefundPaymentOutput, error)
}

type AdminPaymentQuery interface {
	ListPaymentsForAdmin(context.Context, string, string, string, int, int) (repository.AdminPaymentPage, error)
	GetPaymentForAdmin(context.Context, string) (domain.Payment, error)
	ListRefundsForAdmin(context.Context, string, int, int) (repository.AdminRefundPage, error)
	ListRefundsForPaymentAdmin(context.Context, string) ([]domain.Refund, error)
	ListReconciliationsForAdmin(context.Context, string, int, int) (repository.AdminReconciliationPage, error)
	GetReconciliationByID(context.Context, string) (domain.PaymentReconciliation, error)
}

func (h *Handler) handleAdminPayments(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if !h.authorizeRefundRequest(r) {
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Finance admin authorization is required")
		return
	}
	if h.adminQuery == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "PAYMENT_QUERY_UNAVAILABLE", "Payment query is not configured")
		return
	}
	page, pageSize, err := paymentAdminPagination(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	result, err := h.adminQuery.ListPaymentsForAdmin(r.Context(), r.URL.Query().Get("status"), r.URL.Query().Get("provider"), r.URL.Query().Get("order_id"), pageSize, (page-1)*pageSize)
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}
	payments := make([]adminPaymentResponse, 0, len(result.Payments))
	for _, payment := range result.Payments {
		payments = append(payments, adminPaymentResponseFromDomain(payment))
	}
	writeJSON(w, http.StatusOK, map[string]any{"payments": payments, "page": page, "page_size": pageSize, "total": result.Total})
}

func (h *Handler) handleAdminPaymentDetail(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if !h.authorizeRefundRequest(r) {
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Finance admin authorization is required")
		return
	}
	if h.adminQuery == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "PAYMENT_QUERY_UNAVAILABLE", "Payment query is not configured")
		return
	}
	payment, err := h.adminQuery.GetPaymentForAdmin(r.Context(), r.PathValue("payment_id"))
	if err != nil {
		h.writeAdminQueryError(w, r, err)
		return
	}
	refunds, err := h.adminQuery.ListRefundsForPaymentAdmin(r.Context(), payment.PaymentID)
	if err != nil {
		h.writeAdminQueryError(w, r, err)
		return
	}
	refundResponses := make([]refundResponse, 0, len(refunds))
	for _, refund := range refunds {
		refundResponses = append(refundResponses, refundResponseFromDomain(refund, false))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"payment":  adminPaymentResponseFromDomain(payment),
		"attempts": []any{},
		"refunds":  refundResponses,
	})
}

func (h *Handler) handleAdminRefunds(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if !h.authorizeRefundRequest(r) {
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Finance admin authorization is required")
		return
	}
	if h.adminQuery == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "PAYMENT_QUERY_UNAVAILABLE", "Payment query is not configured")
		return
	}
	page, pageSize, err := paymentAdminPagination(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	result, err := h.adminQuery.ListRefundsForAdmin(r.Context(), r.URL.Query().Get("status"), pageSize, (page-1)*pageSize)
	if err != nil {
		h.writeAdminQueryError(w, r, err)
		return
	}
	refunds := make([]refundResponse, 0, len(result.Refunds))
	for _, refund := range result.Refunds {
		refunds = append(refunds, refundResponseFromDomain(refund, false))
	}
	writeJSON(w, http.StatusOK, map[string]any{"refunds": refunds, "page": page, "page_size": pageSize, "total": result.Total})
}

func (h *Handler) handleAdminPaymentRefunds(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if !h.authorizeRefundRequest(r) {
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Finance admin authorization is required")
		return
	}
	if h.adminQuery == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "PAYMENT_QUERY_UNAVAILABLE", "Payment query is not configured")
		return
	}
	refunds, err := h.adminQuery.ListRefundsForPaymentAdmin(r.Context(), r.PathValue("payment_id"))
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}
	responses := make([]refundResponse, 0, len(refunds))
	for _, refund := range refunds {
		responses = append(responses, refundResponseFromDomain(refund, false))
	}
	writeJSON(w, http.StatusOK, map[string]any{"refunds": responses})
}

func (h *Handler) handleAdminReconciliations(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if !h.authorizeRefundRequest(r) {
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Finance admin authorization is required")
		return
	}
	if h.adminQuery == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "PAYMENT_QUERY_UNAVAILABLE", "Payment query is not configured")
		return
	}
	page, pageSize, err := paymentAdminPagination(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	result, err := h.adminQuery.ListReconciliationsForAdmin(r.Context(), r.URL.Query().Get("status"), pageSize, (page-1)*pageSize)
	if err != nil {
		h.writeAdminQueryError(w, r, err)
		return
	}
	alerts := make([]map[string]any, 0, len(result.Reconciliations))
	for _, reconciliation := range result.Reconciliations {
		alerts = append(alerts, repository.ReconciliationAlertFromDomain(reconciliation))
	}
	writeJSON(w, http.StatusOK, map[string]any{"alerts": alerts, "page": page, "page_size": pageSize, "total": result.Total})
}

func (h *Handler) handleAdminReconciliationDetail(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if !h.authorizeRefundRequest(r) {
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Finance admin authorization is required")
		return
	}
	if h.adminQuery == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "PAYMENT_QUERY_UNAVAILABLE", "Payment query is not configured")
		return
	}
	reconciliation, err := h.adminQuery.GetReconciliationByID(r.Context(), r.PathValue("reconciliation_id"))
	if err != nil {
		h.writeAdminQueryError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"alert": repository.ReconciliationAlertFromDomain(reconciliation)})
}

type adminPaymentResponse struct {
	PaymentID         string       `json:"payment_id"`
	OrderID           string       `json:"order_id"`
	Provider          string       `json:"provider"`
	ProviderPaymentID string       `json:"provider_payment_id,omitempty"`
	Status            string       `json:"status"`
	Amount            moneyRequest `json:"amount"`
	CapturedAt        *time.Time   `json:"captured_at,omitempty"`
	CreatedAt         *time.Time   `json:"created_at,omitempty"`
	UpdatedAt         *time.Time   `json:"updated_at,omitempty"`
}

func adminPaymentResponseFromDomain(payment domain.Payment) adminPaymentResponse {
	created, updated := payment.CreatedAt, payment.UpdatedAt
	var capturedAt *time.Time
	if payment.CapturedAmount > 0 {
		capturedAt = &updated
	}
	return adminPaymentResponse{PaymentID: payment.PaymentID, OrderID: payment.OrderID, Provider: payment.Provider, ProviderPaymentID: payment.ProviderPaymentID, Status: string(payment.Status), Amount: moneyRequest{Amount: payment.Amount.Amount, Currency: payment.Amount.Currency}, CapturedAt: capturedAt, CreatedAt: &created, UpdatedAt: &updated}
}
func paymentAdminPagination(r *http.Request) (int, int, error) {
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

type Handler struct {
	stateUsecase         PaymentStateUsecase
	schemaUsecase        PaymentSchemaUsecase
	paymentIntentUsecase PaymentIntentUsecase
	retryPaymentUsecase  RetryPaymentUsecase
	webhookUsecase       WebhookUsecase
	refundUsecase        RefundUsecase
	adminQuery           AdminPaymentQuery
	paymentIntentToken   string
	maxWebhookBodyBytes  int64
	logger               *slog.Logger
}

type HandlerOption func(*Handler)

func WithSchemaUsecase(schemaUsecase PaymentSchemaUsecase) HandlerOption {
	return func(h *Handler) {
		h.schemaUsecase = schemaUsecase
	}
}

func WithPaymentIntentUsecase(paymentIntentUsecase PaymentIntentUsecase) HandlerOption {
	return func(h *Handler) {
		h.paymentIntentUsecase = paymentIntentUsecase
	}
}

func WithRetryPaymentUsecase(retryPaymentUsecase RetryPaymentUsecase) HandlerOption {
	return func(h *Handler) {
		h.retryPaymentUsecase = retryPaymentUsecase
	}
}

func WithPaymentIntentAuthorizationToken(token string) HandlerOption {
	return func(h *Handler) {
		h.paymentIntentToken = strings.TrimSpace(token)
	}
}

func WithWebhookUsecase(webhookUsecase WebhookUsecase) HandlerOption {
	return func(h *Handler) {
		h.webhookUsecase = webhookUsecase
	}
}

func WithRefundUsecase(refundUsecase RefundUsecase) HandlerOption {
	return func(h *Handler) {
		h.refundUsecase = refundUsecase
	}
}

func WithAdminPaymentQuery(query AdminPaymentQuery) HandlerOption {
	return func(h *Handler) { h.adminQuery = query }
}

func WithWebhookMaxBodyBytes(maxBodyBytes int64) HandlerOption {
	return func(h *Handler) {
		if maxBodyBytes > 0 {
			h.maxWebhookBodyBytes = maxBodyBytes
		}
	}
}

func NewHandler(stateUsecase PaymentStateUsecase, logger *slog.Logger, opts ...HandlerOption) (*Handler, error) {
	if stateUsecase == nil {
		return nil, errors.New("payment state usecase is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	handler := &Handler{
		stateUsecase:        stateUsecase,
		maxWebhookBodyBytes: maxRequestBodyBytes,
		logger:              logger,
	}
	return handler.withOptions(opts...), nil
}

func (h *Handler) withOptions(opts ...HandlerOption) *Handler {
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/internal/v1/payment-state-machine", h.handleStateMachine)
	mux.HandleFunc("/internal/v1/payment-state-machine/validate", h.handleValidateTransition)
	mux.HandleFunc("/internal/v1/payment-state-machine/provider-event/normalize", h.handleNormalizeProviderEvent)
	mux.HandleFunc("/internal/v1/payment-schema", h.handlePaymentSchema)
	mux.HandleFunc("/internal/v1/payment-schema/health", h.handlePaymentSchemaHealth)
	mux.HandleFunc("/internal/v1/payment-intents", h.handleCreatePaymentIntent)
	mux.HandleFunc("/api/v1/webhooks/payments/{provider}", h.handleWebhook)
	mux.HandleFunc("/api/v1/payments/{payment_id}/retry", h.handleRetryPayment)
	mux.HandleFunc("/api/v1/payments/{payment_id}/refund", h.handleRefundPayment)
	mux.HandleFunc("/api/v1/refunds/{refund_id}", h.handleGetRefund)
	mux.HandleFunc("/internal/v1/refunds/{refund_id}/review", h.handleReviewRefund)
	mux.HandleFunc("/internal/admin/payments", h.handleAdminPayments)
	mux.HandleFunc("/internal/admin/payments/{payment_id}", h.handleAdminPaymentDetail)
	mux.HandleFunc("/internal/admin/payments/{payment_id}/refunds", h.handleAdminPaymentRefunds)
	mux.HandleFunc("/internal/admin/refunds", h.handleAdminRefunds)
	mux.HandleFunc("/internal/admin/refunds/{refund_id}", h.handleGetRefund)
	mux.HandleFunc("/internal/admin/refunds/{refund_id}/review", h.handleReviewRefund)
	mux.HandleFunc("/internal/admin/payment-reconciliations", h.handleAdminReconciliations)
	mux.HandleFunc("/internal/admin/payment-reconciliations/{reconciliation_id}", h.handleAdminReconciliationDetail)
}

func (h *Handler) handleStateMachine(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	definition, err := h.stateUsecase.Definition(r.Context())
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, stateMachineResponseFromDomain(definition))
}

func (h *Handler) handleValidateTransition(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req validateTransitionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	out, err := h.stateUsecase.EvaluateTransition(r.Context(), usecase.EvaluateTransitionInput{
		From:          req.From,
		To:            req.To,
		Event:         req.Event,
		ProviderEvent: req.ProviderEvent,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPaymentTransition) {
			writeJSON(w, http.StatusOK, validateTransitionResponseFromUsecase(out))
			return
		}
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, validateTransitionResponseFromUsecase(out))
}

func (h *Handler) handleNormalizeProviderEvent(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req normalizeProviderEventRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	out, err := h.stateUsecase.NormalizeProviderEvent(r.Context(), req.ProviderEvent)
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, normalizeProviderEventResponse{
		ProviderEvent: out.ProviderEvent,
		InternalEvent: string(out.InternalEvent),
		NextStatus:    string(out.NextStatus),
	})
}

func (h *Handler) handlePaymentSchema(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if h.schemaUsecase == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SCHEMA_UNAVAILABLE", "Payment schema usecase is not configured")
		return
	}

	definition, err := h.schemaUsecase.Definition(r.Context())
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, paymentSchemaResponseFromDomain(definition))
}

func (h *Handler) handlePaymentSchemaHealth(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if h.schemaUsecase == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SCHEMA_UNAVAILABLE", "Payment schema usecase is not configured")
		return
	}

	out, err := h.schemaUsecase.Health(r.Context())
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}
	status := http.StatusOK
	if !out.Verification.Ready {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, paymentSchemaHealthResponseFromUsecase(out))
}

func (h *Handler) handleCreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	if !h.authorizePaymentIntentRequest(r) {
		h.logger.WarnContext(r.Context(), "payment.intent.unauthorized_request")
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Internal payment authorization is required")
		return
	}
	if h.paymentIntentUsecase == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "PAYMENT_INTENT_UNAVAILABLE", "Payment intent usecase is not configured")
		return
	}

	var req createPaymentIntentRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	out, err := h.paymentIntentUsecase.Execute(
		r.Context(),
		createPaymentIntentInputFromRequest(req, r.Header.Get("X-Request-ID")),
	)
	if err != nil {
		h.writeUsecaseError(w, r, err)
		return
	}
	status := http.StatusCreated
	if out.Replayed {
		status = http.StatusOK
	}
	writeJSON(w, status, createPaymentIntentResponseFromUsecase(out))
}

func (h *Handler) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	if h.webhookUsecase == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "WEBHOOK_UNAVAILABLE", "Payment webhook processing is not configured")
		return
	}
	rawBody, err := io.ReadAll(http.MaxBytesReader(w, r.Body, h.maxWebhookBodyBytes))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_WEBHOOK_BODY", "Webhook request body could not be read")
		return
	}
	out, err := h.webhookUsecase.Execute(r.Context(), usecase.HandleWebhookInput{
		Provider: r.PathValue("provider"),
		Headers:  flattenHeaders(r.Header),
		RawBody:  rawBody,
	})
	if err != nil {
		h.writeWebhookError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, handleWebhookResponseFromUsecase(out))
}

func (h *Handler) handleRetryPayment(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	if !h.authorizeBuyerRequest(r) {
		h.logger.WarnContext(r.Context(), "payment.retry.unauthorized_request")
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Buyer authorization is required")
		return
	}
	if h.retryPaymentUsecase == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "PAYMENT_RETRY_UNAVAILABLE", "Payment retry processing is not configured")
		return
	}
	var req retryPaymentIntentRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	pathPaymentID := strings.TrimSpace(r.PathValue("payment_id"))
	if strings.TrimSpace(req.PaymentID) != "" && strings.TrimSpace(req.PaymentID) != pathPaymentID {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Request payment_id does not match route payment_id")
		return
	}
	headerKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	bodyKey := strings.TrimSpace(req.IdempotencyKey)
	if headerKey != "" && bodyKey != "" && headerKey != bodyKey {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Idempotency-Key header does not match request idempotency_key")
		return
	}
	if bodyKey == "" {
		bodyKey = headerKey
	}
	out, err := h.retryPaymentUsecase.Execute(r.Context(), usecase.RetryPaymentIntentInput{
		FailedPaymentID: pathPaymentID,
		BuyerUserID:     r.Header.Get("X-Actor-ID"),
		RetryRequestKey: bodyKey,
		RequestID:       r.Header.Get("X-Request-ID"),
	})
	if err != nil {
		h.writeRetryError(w, r, err)
		return
	}
	status := http.StatusCreated
	if out.Replayed {
		status = http.StatusOK
	}
	writeJSON(w, status, retryPaymentIntentResponseFromUsecase(out))
}

func (h *Handler) handleRefundPayment(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	if !h.authorizeRefundRequest(r) {
		h.logger.WarnContext(r.Context(), "payment.refund.unauthorized_request")
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Finance admin authorization is required")
		return
	}
	if h.refundUsecase == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "REFUND_UNAVAILABLE", "Refund processing is not configured")
		return
	}
	var req refundPaymentRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	pathPaymentID := r.PathValue("payment_id")
	if strings.TrimSpace(req.PaymentID) != "" && strings.TrimSpace(req.PaymentID) != pathPaymentID {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Request payment_id does not match route payment_id")
		return
	}
	out, err := h.refundUsecase.Execute(r.Context(), refundPaymentInputFromRequest(
		pathPaymentID,
		r.Header.Get("X-Actor-ID"),
		r.Header.Get("X-Request-ID"),
		req,
	))
	if err != nil {
		h.writeRefundError(w, r, err)
		return
	}
	status := http.StatusCreated
	if out.Replayed {
		status = http.StatusOK
	}
	writeJSON(w, status, refundResponseFromDomain(out.Refund, out.Replayed))
}

func (h *Handler) handleGetRefund(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if !h.authorizeRefundRequest(r) {
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Finance admin authorization is required")
		return
	}
	if h.refundUsecase == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "REFUND_UNAVAILABLE", "Refund processing is not configured")
		return
	}
	refund, err := h.refundUsecase.Get(r.Context(), r.PathValue("refund_id"))
	if err != nil {
		h.writeRefundError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, refundResponseFromDomain(refund, false))
}

func (h *Handler) handleReviewRefund(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	if !h.authorizeRefundRequest(r) {
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Finance admin authorization is required")
		return
	}
	if h.refundUsecase == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "REFUND_UNAVAILABLE", "Refund processing is not configured")
		return
	}
	var req reviewRefundRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	out, err := h.refundUsecase.Review(r.Context(), usecase.ReviewRefundInput{
		RefundID:   r.PathValue("refund_id"),
		Decision:   req.Decision,
		Reason:     req.Reason,
		ReviewedBy: r.Header.Get("X-Actor-ID"),
	})
	if err != nil {
		h.writeRefundError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, refundResponseFromDomain(out.Refund, false))
}

func (h *Handler) authorizePaymentIntentRequest(r *http.Request) bool {
	if h.paymentIntentToken == "" {
		return false
	}
	const bearerPrefix = "Bearer "
	authorization := r.Header.Get("Authorization")
	if !strings.HasPrefix(authorization, bearerPrefix) {
		return false
	}
	provided := strings.TrimSpace(strings.TrimPrefix(authorization, bearerPrefix))
	expectedDigest := sha256.Sum256([]byte(h.paymentIntentToken))
	providedDigest := sha256.Sum256([]byte(provided))
	return subtle.ConstantTimeCompare(expectedDigest[:], providedDigest[:]) == 1
}

func (h *Handler) authorizeRefundRequest(r *http.Request) bool {
	actorID := firstNonEmptyHeader(r.Header, "X-Actor-ID", "X-Admin-ID")
	if !h.authorizePaymentIntentRequest(r) || actorID == "" {
		return false
	}
	roles := []string{r.Header.Get("X-Actor-Role"), r.Header.Get("X-Admin-Roles")}
	for _, raw := range roles {
		for _, role := range strings.Split(raw, ",") {
			switch strings.ToLower(strings.TrimSpace(role)) {
			case "admin", "finance_admin", "superadmin":
				return true
			}
		}
	}
	return false
}

func firstNonEmptyHeader(header http.Header, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(header.Get(name)); value != "" {
			return value
		}
	}
	return ""
}

func (h *Handler) authorizeBuyerRequest(r *http.Request) bool {
	if !h.authorizePaymentIntentRequest(r) || strings.TrimSpace(r.Header.Get("X-Actor-ID")) == "" {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Actor-Role")), "buyer")
}

func (h *Handler) writeUsecaseError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, provider.ErrInvalidProviderRequest):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, provider.ErrProviderNotFound):
		writeAPIError(w, http.StatusBadRequest, "UNSUPPORTED_PROVIDER", "Requested payment provider is not available")
	case errors.Is(err, usecase.ErrPaymentProviderUnavailable):
		writeAPIError(w, http.StatusServiceUnavailable, "PAYMENT_PROVIDER_UNAVAILABLE", "No payment provider is configured")
	case errors.Is(err, usecase.ErrPaymentIntentProcessing):
		writeAPIError(w, http.StatusConflict, "PAYMENT_INTENT_PROCESSING", "Payment intent is already being created")
	case errors.Is(err, usecase.ErrPaymentIntentConflict):
		writeAPIError(w, http.StatusConflict, "IDEMPOTENCY_CONFLICT", "Idempotency key was already used with different payment data")
	case errors.Is(err, usecase.ErrPaymentIntentPreviouslyFailed):
		writeAPIError(w, http.StatusConflict, "PAYMENT_INTENT_FAILED", "Payment intent creation previously failed for this idempotency key")
	case errors.Is(err, usecase.ErrPaymentClientPayloadUnavailable):
		writeAPIError(w, http.StatusServiceUnavailable, "PAYMENT_CLIENT_PAYLOAD_UNAVAILABLE", "Payment client payload cannot be recovered")
	case errors.Is(err, usecase.ErrTransitionTargetRequired), errors.Is(err, usecase.ErrTransitionTargetConflict):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, domain.ErrInvalidPaymentStatus), errors.Is(err, domain.ErrInvalidPaymentEvent), errors.Is(err, domain.ErrUnknownProviderEvent):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, domain.ErrPaymentSchemaUnavailable):
		writeAPIError(w, http.StatusServiceUnavailable, "SCHEMA_UNAVAILABLE", "Payment schema is unavailable")
	default:
		var providerErr *provider.Error
		if errors.As(err, &providerErr) {
			status := providerErr.StatusCode
			if status < http.StatusBadRequest || status > http.StatusNetworkAuthenticationRequired {
				status = http.StatusBadGateway
			}
			writeAPIError(w, status, string(providerErr.Code), "Payment provider could not create an intent")
			return
		}
		h.logger.ErrorContext(r.Context(), "payment.http_error",
			slog.String("error", err.Error()),
		)
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

func (h *Handler) writeWebhookError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, provider.ErrProviderNotFound):
		writeAPIError(w, http.StatusNotFound, "UNSUPPORTED_PROVIDER", "Requested payment provider is not available")
	case errors.Is(err, provider.ErrInvalidWebhookSignature):
		h.logger.WarnContext(r.Context(), "payment.webhook.signature_invalid",
			slog.String("provider", r.PathValue("provider")),
		)
		writeAPIError(w, http.StatusBadRequest, "INVALID_WEBHOOK_SIGNATURE", "Webhook signature is invalid")
	case errors.Is(err, provider.ErrInvalidProviderRequest), errors.Is(err, domain.ErrInvalidWebhookEvent), errors.Is(err, domain.ErrInvalidPayment):
		writeAPIError(w, http.StatusBadRequest, "INVALID_WEBHOOK", "Webhook payload is invalid")
	default:
		h.logger.ErrorContext(r.Context(), "payment.webhook.http_error",
			slog.String("provider", r.PathValue("provider")),
			slog.String("error", err.Error()),
		)
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

func (h *Handler) writeRetryError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, provider.ErrInvalidProviderRequest), errors.Is(err, domain.ErrInvalidPayment):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, domain.ErrPaymentRecordNotFound):
		writeAPIError(w, http.StatusNotFound, "PAYMENT_NOT_FOUND", "Payment record was not found")
	case errors.Is(err, domain.ErrPaymentNotRetryable):
		writeAPIError(w, http.StatusConflict, "PAYMENT_RETRY_NOT_ALLOWED", "Payment cannot be retried in its current state")
	case errors.Is(err, domain.ErrPaymentRetryInProgress):
		writeAPIError(w, http.StatusConflict, "PAYMENT_RETRY_IN_PROGRESS", "An active payment attempt already exists for this order")
	case errors.Is(err, domain.ErrPaymentAlreadyCaptured):
		writeAPIError(w, http.StatusConflict, "PAYMENT_ALREADY_CAPTURED", "Payment has already been captured for this order")
	case errors.Is(err, domain.ErrPaymentRetryLimitReached):
		writeAPIError(w, http.StatusTooManyRequests, "PAYMENT_RETRY_LIMIT_REACHED", "Maximum payment attempts have been reached")
	case errors.Is(err, domain.ErrPaymentRetryCooldown):
		writeAPIError(w, http.StatusTooManyRequests, "PAYMENT_RETRY_COOLDOWN", "Payment retry cooldown is active")
	case errors.Is(err, usecase.ErrPaymentIntentProcessing):
		writeAPIError(w, http.StatusConflict, "PAYMENT_INTENT_PROCESSING", "Payment retry intent is already being created")
	case errors.Is(err, usecase.ErrPaymentIntentPreviouslyFailed):
		writeAPIError(w, http.StatusConflict, "PAYMENT_INTENT_FAILED", "Payment retry intent creation previously failed")
	case errors.Is(err, usecase.ErrPaymentClientPayloadUnavailable):
		writeAPIError(w, http.StatusServiceUnavailable, "PAYMENT_CLIENT_PAYLOAD_UNAVAILABLE", "Payment client payload cannot be recovered")
	default:
		var providerErr *provider.Error
		if errors.As(err, &providerErr) {
			status := providerErr.StatusCode
			if status < http.StatusBadRequest || status > http.StatusNetworkAuthenticationRequired {
				status = http.StatusBadGateway
			}
			writeAPIError(w, status, string(providerErr.Code), "Payment provider could not create a retry intent")
			return
		}
		h.logger.ErrorContext(r.Context(), "payment.retry.http_error", slog.String("error", err.Error()))
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

func (h *Handler) writeRefundError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, provider.ErrInvalidProviderRequest), errors.Is(err, domain.ErrInvalidRefund), errors.Is(err, domain.ErrInvalidPayment):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, domain.ErrPaymentRecordNotFound), errors.Is(err, domain.ErrRefundRecordNotFound):
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Payment or refund record was not found")
	case errors.Is(err, domain.ErrPaymentNotRefundable):
		writeAPIError(w, http.StatusConflict, "PAYMENT_NOT_REFUNDABLE", "Payment is not eligible for a refund")
	case errors.Is(err, domain.ErrRefundCurrencyMismatch), errors.Is(err, domain.ErrRefundExceedsAvailable):
		writeAPIError(w, http.StatusConflict, "REFUND_AMOUNT_NOT_AVAILABLE", err.Error())
	case errors.Is(err, usecase.ErrRefundIdempotencyConflict):
		writeAPIError(w, http.StatusConflict, "IDEMPOTENCY_CONFLICT", "Idempotency key was already used with different refund data")
	case errors.Is(err, usecase.ErrRefundMakerChecker):
		writeAPIError(w, http.StatusForbidden, "MAKER_CHECKER_REQUIRED", "A different finance approver must review this refund")
	case errors.Is(err, usecase.ErrRefundReviewDecision), errors.Is(err, domain.ErrInvalidRefundTransition):
		writeAPIError(w, http.StatusConflict, "INVALID_REFUND_TRANSITION", err.Error())
	default:
		var providerErr *provider.Error
		if errors.As(err, &providerErr) {
			status := providerErr.StatusCode
			if status < http.StatusBadRequest || status > http.StatusNetworkAuthenticationRequired {
				status = http.StatusBadGateway
			}
			writeAPIError(w, status, string(providerErr.Code), "Payment provider could not process the refund")
			return
		}
		h.logger.ErrorContext(r.Context(), "payment.refund.http_error", slog.String("error", err.Error()))
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

func (h *Handler) writeAdminQueryError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrPaymentRecordNotFound), errors.Is(err, domain.ErrRefundRecordNotFound), errors.Is(err, domain.ErrReconciliationNotFound):
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Admin payment resource was not found")
	case errors.Is(err, domain.ErrInvalidPayment), errors.Is(err, domain.ErrInvalidRefund), errors.Is(err, domain.ErrInvalidReconciliation):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:
		h.logger.ErrorContext(r.Context(), "payment.admin_query.http_error", slog.String("error", err.Error()))
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

func validateTransitionResponseFromUsecase(out usecase.EvaluateTransitionOutput) validateTransitionResponse {
	return validateTransitionResponse{
		Allowed:              out.Allowed,
		Noop:                 out.Noop,
		From:                 string(out.From),
		To:                   string(out.To),
		Event:                string(out.Event),
		ProviderEvent:        out.ProviderEvent,
		SuggestedOrderStatus: string(out.SuggestedOrderStatus),
		Reason:               out.Reason,
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func flattenHeaders(headers http.Header) map[string]string {
	flattened := make(map[string]string, len(headers))
	for key, values := range headers {
		flattened[key] = strings.Join(values, ",")
	}
	return flattened
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	return false
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, errorResponse{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}
