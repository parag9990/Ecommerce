package httptransport

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/usecase"
)

func TestStateMachineDefinitionEndpoint(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/internal/v1/payment-state-machine", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var got stateMachineResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal(stateMachineResponse) error = %v", err)
	}
	if len(got.Statuses) != 8 {
		t.Fatalf("statuses len = %d, want 8", len(got.Statuses))
	}
	if len(got.Transitions) == 0 {
		t.Fatal("expected transition rules")
	}
}

func TestValidateTransitionEndpointAllowsProviderEvent(t *testing.T) {
	router := newTestRouter(t)

	body := strings.NewReader(`{"from":"authorized","provider_event":"payment.captured"}`)
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/payment-state-machine/validate", body)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var got validateTransitionResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal(validateTransitionResponse) error = %v", err)
	}
	if !got.Allowed || got.To != "captured" || got.SuggestedOrderStatus != "paid" {
		t.Fatalf("response = %+v, want allowed captured paid", got)
	}
}

func TestValidateTransitionEndpointReturnsBlockedDecision(t *testing.T) {
	router := newTestRouter(t)

	body := strings.NewReader(`{"from":"failed","to":"captured"}`)
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/payment-state-machine/validate", body)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var got validateTransitionResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal(validateTransitionResponse) error = %v", err)
	}
	if got.Allowed {
		t.Fatalf("response = %+v, want blocked", got)
	}
	if !strings.Contains(got.Reason, "invalid payment transition") {
		t.Fatalf("reason = %q, want invalid transition", got.Reason)
	}
}

func TestValidateTransitionEndpointRejectsUnknownField(t *testing.T) {
	router := newTestRouter(t)

	body := strings.NewReader(`{"from":"initiated","to":"captured","raw_card":"4111111111111111"}`)
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/payment-state-machine/validate", body)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	assertAPIError(t, rr.Body.Bytes(), "INVALID_REQUEST")
}

func TestNormalizeProviderEventEndpoint(t *testing.T) {
	router := newTestRouter(t)

	body := strings.NewReader(`{"provider_event":"payment_intent.payment_failed"}`)
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/payment-state-machine/provider-event/normalize", body)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var got normalizeProviderEventResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal(normalizeProviderEventResponse) error = %v", err)
	}
	if got.InternalEvent != string(domain.PaymentEventProviderFailed) || got.NextStatus != string(domain.PaymentStatusFailed) {
		t.Fatalf("response = %+v, want provider_failed/failed", got)
	}
}

func TestPaymentSchemaDefinitionEndpoint(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/internal/v1/payment-schema", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var got paymentSchemaResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal(paymentSchemaResponse) error = %v", err)
	}
	if got.DatabaseName != domain.PaymentDatabaseName {
		t.Fatalf("database_name = %q, want %q", got.DatabaseName, domain.PaymentDatabaseName)
	}
	if len(got.Tables) != 5 {
		t.Fatalf("tables len = %d, want 5", len(got.Tables))
	}
}

func TestPaymentSchemaHealthEndpointReportsUnavailableWithoutVerifier(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/internal/v1/payment-schema/health", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusServiceUnavailable, rr.Body.String())
	}
	var got paymentSchemaHealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal(paymentSchemaHealthResponse) error = %v", err)
	}
	if got.Ready {
		t.Fatalf("schema health ready = true, want false")
	}
	if len(got.MissingTables) != 5 {
		t.Fatalf("missing tables len = %d, want 5", len(got.MissingTables))
	}
}

func TestCreatePaymentIntentEndpointReturnsClientPayload(t *testing.T) {
	intentUsecase := &fakePaymentIntentUsecase{
		output: usecase.CreatePaymentIntentOutput{
			PaymentID:        "pay_123",
			OrderID:          "ord_123",
			Provider:         "stripe_like",
			ProviderIntentID: "pi_123",
			Status:           domain.PaymentStatusRequiresAction,
			Amount:           1000,
			Currency:         "INR",
			ClientPayload: usecase.PaymentClientPayload{
				PublicKey:    "pk_test",
				ClientSecret: "client_secret_browser",
			},
		},
	}
	router := newTestRouterWithIntent(t, intentUsecase)
	body := strings.NewReader(`{"order_id":"ord_123","user_id":"usr_123","amount":1000,"currency":"INR","idempotency_key":"payment_intent:ord_123:attempt_1"}`)
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/payment-intents", body)
	req.Header.Set("X-Request-ID", "request_123")
	req.Header.Set("Authorization", "Bearer test-internal-payment-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	var got createPaymentIntentResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal(createPaymentIntentResponse) error = %v", err)
	}
	if got.PaymentID != "pay_123" || got.ClientPayload.ClientSecret != "client_secret_browser" {
		t.Fatalf("response = %+v, want payment client payload", got)
	}
	if intentUsecase.input.RequestID != "request_123" || intentUsecase.input.IdempotencyKey == "" {
		t.Fatalf("input = %+v, want mapped request and request id", intentUsecase.input)
	}
}

func TestCreatePaymentIntentEndpointMapsProcessingConflict(t *testing.T) {
	router := newTestRouterWithIntent(t, &fakePaymentIntentUsecase{err: usecase.ErrPaymentIntentProcessing})
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/payment-intents", strings.NewReader(`{"order_id":"ord_123"}`))
	req.Header.Set("Authorization", "Bearer test-internal-payment-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusConflict, rr.Body.String())
	}
	assertAPIError(t, rr.Body.Bytes(), "PAYMENT_INTENT_PROCESSING")
}

func TestCreatePaymentIntentEndpointRejectsUnauthenticatedCall(t *testing.T) {
	intentUsecase := &fakePaymentIntentUsecase{}
	router := newTestRouterWithIntent(t, intentUsecase)
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/payment-intents", strings.NewReader(`{"order_id":"ord_123"}`))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
	assertAPIError(t, rr.Body.Bytes(), "UNAUTHORIZED")
	if intentUsecase.input.OrderID != "" {
		t.Fatal("unauthenticated request reached payment intent usecase")
	}
}

func TestRetryPaymentEndpointMapsBuyerAndIdempotencyContract(t *testing.T) {
	retryUsecase := &fakeRetryPaymentUsecase{output: usecase.RetryPaymentIntentOutput{
		PaymentID: "pay_retry", OrderID: "ord_123", RetryOfPaymentID: "pay_failed", RootPaymentID: "pay_failed",
		AttemptNo: 2, Provider: provider.ProviderNameStripeLike, ProviderIntentID: "pi_retry",
		Status: domain.PaymentStatusRequiresAction, Amount: 1000, Currency: "INR",
		ClientPayload: usecase.PaymentClientPayload{ClientSecret: "browser_retry_secret"},
	}}
	router := newTestRouterWithRetry(t, retryUsecase)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/pay_failed/retry", strings.NewReader(`{"idempotency_key":"retry_client_123"}`))
	req.Header.Set("Authorization", "Bearer test-internal-payment-token")
	req.Header.Set("X-Actor-ID", "usr_123")
	req.Header.Set("X-Actor-Role", "buyer,admin")
	req.Header.Set("Idempotency-Key", "retry_client_123")
	req.Header.Set("X-Request-ID", "request_123")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	if retryUsecase.input.FailedPaymentID != "pay_failed" || retryUsecase.input.BuyerUserID != "usr_123" ||
		retryUsecase.input.RetryRequestKey != "retry_client_123" {
		t.Fatalf("input = %+v, want route buyer and key mapping", retryUsecase.input)
	}
	var got retryPaymentIntentResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal(retryPaymentIntentResponse) error = %v", err)
	}
	if got.Retry.AttemptNo != 2 || got.Retry.RetryOfPaymentID != "pay_failed" {
		t.Fatalf("response = %+v, want retry lineage", got)
	}
}

func TestRetryPaymentEndpointRejectsUntrustedBuyerAndKeyConflict(t *testing.T) {
	retryUsecase := &fakeRetryPaymentUsecase{}
	router := newTestRouterWithRetry(t, retryUsecase)
	unauthorized := httptest.NewRequest(http.MethodPost, "/api/v1/payments/pay_failed/retry", strings.NewReader(`{"idempotency_key":"retry_client_123"}`))
	unauthorized.Header.Set("Authorization", "Bearer test-internal-payment-token")
	unauthorized.Header.Set("X-Actor-ID", "usr_123")
	unauthorized.Header.Set("X-Actor-Role", "admin")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, unauthorized)
	if rr.Code != http.StatusUnauthorized || retryUsecase.input.FailedPaymentID != "" {
		t.Fatalf("status=%d input=%+v, want buyer authorization rejection", rr.Code, retryUsecase.input)
	}

	conflict := httptest.NewRequest(http.MethodPost, "/api/v1/payments/pay_failed/retry", strings.NewReader(`{"idempotency_key":"body_key"}`))
	conflict.Header.Set("Authorization", "Bearer test-internal-payment-token")
	conflict.Header.Set("X-Actor-ID", "usr_123")
	conflict.Header.Set("X-Actor-Role", "buyer")
	conflict.Header.Set("Idempotency-Key", "header_key")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, conflict)
	if rr.Code != http.StatusBadRequest || retryUsecase.input.FailedPaymentID != "" {
		t.Fatalf("status=%d input=%+v, want key conflict before usecase", rr.Code, retryUsecase.input)
	}
}

func TestRetryPaymentEndpointMapsPolicyRejection(t *testing.T) {
	router := newTestRouterWithRetry(t, &fakeRetryPaymentUsecase{err: domain.ErrPaymentRetryLimitReached})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/pay_failed/retry", strings.NewReader(`{"idempotency_key":"retry_client_123"}`))
	req.Header.Set("Authorization", "Bearer test-internal-payment-token")
	req.Header.Set("X-Actor-ID", "usr_123")
	req.Header.Set("X-Actor-Role", "buyer")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusTooManyRequests, rr.Body.String())
	}
	assertAPIError(t, rr.Body.Bytes(), "PAYMENT_RETRY_LIMIT_REACHED")
}

func TestWebhookEndpointPreservesRawBodyWithoutUserAuthorization(t *testing.T) {
	webhookUsecase := &fakeWebhookUsecase{output: usecase.HandleWebhookOutput{
		WebhookEventID:   "whe_123",
		ProviderEventID:  "evt_123",
		ProcessingStatus: domain.WebhookProcessingStatusProcessed,
		PaymentID:        "pay_123",
		Status:           domain.PaymentStatusCaptured,
	}}
	router := newTestRouterWithWebhook(t, webhookUsecase, maxRequestBodyBytes)
	raw := `{"id":"evt_123", "amount":1000}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/payments/stripe_like", strings.NewReader(raw))
	req.Header.Set("Stripe-Signature", "signed-raw-content")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if string(webhookUsecase.input.RawBody) != raw || webhookUsecase.input.Provider != "stripe_like" {
		t.Fatalf("input = %+v, want exact raw webhook body and route provider", webhookUsecase.input)
	}
	var got handleWebhookResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal(handleWebhookResponse) error = %v", err)
	}
	if !got.Success || got.ProcessingStatus != "processed" {
		t.Fatalf("response = %+v, want processed", got)
	}
}

func TestWebhookEndpointMapsInvalidSignature(t *testing.T) {
	router := newTestRouterWithWebhook(t, &fakeWebhookUsecase{err: provider.ErrInvalidWebhookSignature}, maxRequestBodyBytes)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/payments/stripe_like", strings.NewReader(`{"id":"evt_123"}`))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	assertAPIError(t, rr.Body.Bytes(), "INVALID_WEBHOOK_SIGNATURE")
}

func TestWebhookEndpointRejectsOversizedPayloadBeforeUsecase(t *testing.T) {
	webhookUsecase := &fakeWebhookUsecase{}
	router := newTestRouterWithWebhook(t, webhookUsecase, 4)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/payments/stripe_like", strings.NewReader(`{"oversized":true}`))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest || len(webhookUsecase.input.RawBody) != 0 {
		t.Fatalf("status=%d input=%+v, want body rejection before usecase", rr.Code, webhookUsecase.input)
	}
	assertAPIError(t, rr.Body.Bytes(), "INVALID_WEBHOOK_BODY")
}

func TestRefundEndpointRequiresFinanceAuthorizationAndMapsRequest(t *testing.T) {
	refundUsecase := &fakeRefundUsecase{output: usecase.RefundPaymentOutput{Refund: domain.Refund{
		RefundID: "rfnd_123", PaymentID: "pay_123", Status: domain.RefundStatusRequested,
		Amount: domain.Money{Amount: 250, Currency: "INR"}, Reason: "Cancelled item", RequestedBy: "finance_1",
	}}}
	router := newTestRouterWithRefund(t, refundUsecase)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/pay_123/refund", strings.NewReader(`{"amount":{"amount":250,"currency":"INR"},"reason":"Cancelled item","idempotency_key":"refund:pay_123:item_1"}`))
	req.Header.Set("Authorization", "Bearer test-internal-payment-token")
	req.Header.Set("X-Actor-ID", "finance_1")
	req.Header.Set("X-Actor-Role", "finance_admin")
	req.Header.Set("X-Request-ID", "request_123")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	if refundUsecase.input.PaymentID != "pay_123" || refundUsecase.input.RequestedBy != "finance_1" || refundUsecase.input.AmountMinor != 250 {
		t.Fatalf("refund input = %+v, want route/actor/amount mapped", refundUsecase.input)
	}
}

func TestRefundEndpointRejectsNonFinanceCaller(t *testing.T) {
	refundUsecase := &fakeRefundUsecase{}
	router := newTestRouterWithRefund(t, refundUsecase)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/pay_123/refund", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer test-internal-payment-token")
	req.Header.Set("X-Actor-ID", "buyer_1")
	req.Header.Set("X-Actor-Role", "buyer")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden || refundUsecase.input.PaymentID != "" {
		t.Fatalf("status=%d input=%+v, want authorization rejection", rr.Code, refundUsecase.input)
	}
}

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	uc, err := usecase.NewPaymentStateUsecase(domain.NewStateMachine(), slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	if err != nil {
		t.Fatalf("NewPaymentStateUsecase returned error: %v", err)
	}
	schemaUsecase, err := usecase.NewPaymentSchemaUsecase(nil, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	if err != nil {
		t.Fatalf("NewPaymentSchemaUsecase returned error: %v", err)
	}
	handler, err := NewHandler(uc, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), WithSchemaUsecase(schemaUsecase))
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}
	return NewRouter(handler)
}

func newTestRouterWithIntent(t *testing.T, intentUsecase PaymentIntentUsecase) http.Handler {
	t.Helper()

	uc, err := usecase.NewPaymentStateUsecase(domain.NewStateMachine(), slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	if err != nil {
		t.Fatalf("NewPaymentStateUsecase returned error: %v", err)
	}
	handler, err := NewHandler(
		uc,
		slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
		WithPaymentIntentUsecase(intentUsecase),
		WithPaymentIntentAuthorizationToken("test-internal-payment-token"),
	)
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}
	return NewRouter(handler)
}

func newTestRouterWithWebhook(t *testing.T, webhookUsecase WebhookUsecase, maxBodyBytes int64) http.Handler {
	t.Helper()

	uc, err := usecase.NewPaymentStateUsecase(domain.NewStateMachine(), slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	if err != nil {
		t.Fatalf("NewPaymentStateUsecase returned error: %v", err)
	}
	handler, err := NewHandler(
		uc,
		slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
		WithWebhookUsecase(webhookUsecase),
		WithWebhookMaxBodyBytes(maxBodyBytes),
	)
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}
	return NewRouter(handler)
}

func newTestRouterWithRetry(t *testing.T, retryUsecase RetryPaymentUsecase) http.Handler {
	t.Helper()
	uc, err := usecase.NewPaymentStateUsecase(domain.NewStateMachine(), slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	if err != nil {
		t.Fatalf("NewPaymentStateUsecase returned error: %v", err)
	}
	handler, err := NewHandler(
		uc,
		slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
		WithRetryPaymentUsecase(retryUsecase),
		WithPaymentIntentAuthorizationToken("test-internal-payment-token"),
	)
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}
	return NewRouter(handler)
}

func newTestRouterWithRefund(t *testing.T, refundUsecase RefundUsecase) http.Handler {
	t.Helper()
	uc, err := usecase.NewPaymentStateUsecase(domain.NewStateMachine(), slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	if err != nil {
		t.Fatalf("NewPaymentStateUsecase returned error: %v", err)
	}
	handler, err := NewHandler(
		uc,
		slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
		WithRefundUsecase(refundUsecase),
		WithPaymentIntentAuthorizationToken("test-internal-payment-token"),
	)
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}
	return NewRouter(handler)
}

type fakePaymentIntentUsecase struct {
	input  usecase.CreatePaymentIntentInput
	output usecase.CreatePaymentIntentOutput
	err    error
}

func (u *fakePaymentIntentUsecase) Execute(_ context.Context, input usecase.CreatePaymentIntentInput) (usecase.CreatePaymentIntentOutput, error) {
	u.input = input
	return u.output, u.err
}

type fakeWebhookUsecase struct {
	input  usecase.HandleWebhookInput
	output usecase.HandleWebhookOutput
	err    error
}

func (u *fakeWebhookUsecase) Execute(_ context.Context, input usecase.HandleWebhookInput) (usecase.HandleWebhookOutput, error) {
	u.input = input
	return u.output, u.err
}

type fakeRetryPaymentUsecase struct {
	input  usecase.RetryPaymentIntentInput
	output usecase.RetryPaymentIntentOutput
	err    error
}

func (u *fakeRetryPaymentUsecase) Execute(_ context.Context, input usecase.RetryPaymentIntentInput) (usecase.RetryPaymentIntentOutput, error) {
	u.input = input
	return u.output, u.err
}

type fakeRefundUsecase struct {
	input  usecase.RefundPaymentInput
	output usecase.RefundPaymentOutput
	err    error
}

func (u *fakeRefundUsecase) Execute(_ context.Context, input usecase.RefundPaymentInput) (usecase.RefundPaymentOutput, error) {
	u.input = input
	return u.output, u.err
}

func (u *fakeRefundUsecase) Get(_ context.Context, _ string) (domain.Refund, error) {
	return u.output.Refund, u.err
}

func (u *fakeRefundUsecase) Review(_ context.Context, _ usecase.ReviewRefundInput) (usecase.RefundPaymentOutput, error) {
	return u.output, u.err
}

func assertAPIError(t *testing.T, body []byte, wantCode string) {
	t.Helper()

	var got errorResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal(errorResponse) error = %v; body=%s", err, string(body))
	}
	if got.Error.Code != wantCode {
		t.Fatalf("api error code = %q, want %q; body=%s", got.Error.Code, wantCode, string(body))
	}
}
