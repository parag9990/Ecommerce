package httptransport

import (
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/usecase"
)

type stateMachineResponse struct {
	Statuses         []paymentStatusDefinitionResponse `json:"statuses"`
	Transitions      []transitionRuleResponse          `json:"transitions"`
	InvalidExamples  []invalidTransitionResponse       `json:"invalid_examples"`
	ProviderMappings []providerEventMappingResponse    `json:"provider_mappings"`
	OrderMappings    []orderStateMappingResponse       `json:"order_mappings"`
}

type paymentStatusDefinitionResponse struct {
	Status          string `json:"status"`
	Meaning         string `json:"meaning"`
	UserFacing      string `json:"user_facing"`
	FinalForAttempt bool   `json:"final_for_attempt"`
	Helper          bool   `json:"helper"`
	Persisted       bool   `json:"persisted"`
}

type transitionRuleResponse struct {
	From          string `json:"from"`
	Event         string `json:"event"`
	To            string `json:"to"`
	TriggerSource string `json:"trigger_source"`
	Notes         string `json:"notes"`
}

type invalidTransitionResponse struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason"`
}

type providerEventMappingResponse struct {
	ProviderEvent string `json:"provider_event"`
	InternalEvent string `json:"internal_event"`
	NextStatus    string `json:"next_status"`
}

type orderStateMappingResponse struct {
	PaymentStatus string `json:"payment_status"`
	OrderStatus   string `json:"order_status"`
	Explanation   string `json:"explanation"`
}

type validateTransitionRequest struct {
	From          string `json:"from"`
	To            string `json:"to,omitempty"`
	Event         string `json:"event,omitempty"`
	ProviderEvent string `json:"provider_event,omitempty"`
}

type validateTransitionResponse struct {
	Allowed              bool   `json:"allowed"`
	Noop                 bool   `json:"noop"`
	From                 string `json:"from"`
	To                   string `json:"to"`
	Event                string `json:"event,omitempty"`
	ProviderEvent        string `json:"provider_event,omitempty"`
	SuggestedOrderStatus string `json:"suggested_order_status,omitempty"`
	Reason               string `json:"reason"`
}

type normalizeProviderEventRequest struct {
	ProviderEvent string `json:"provider_event"`
}

type normalizeProviderEventResponse struct {
	ProviderEvent string `json:"provider_event"`
	InternalEvent string `json:"internal_event"`
	NextStatus    string `json:"next_status"`
}

type paymentSchemaResponse struct {
	DatabaseName string                      `json:"database_name"`
	Tables       []paymentTableDefinitionDTO `json:"tables"`
}

type paymentTableDefinitionDTO struct {
	Name        string   `json:"name"`
	Columns     []string `json:"columns"`
	UniqueKeys  []string `json:"unique_keys"`
	Indexes     []string `json:"indexes"`
	ForeignKeys []string `json:"foreign_keys,omitempty"`
}

type paymentSchemaHealthResponse struct {
	DatabaseName       string   `json:"database_name"`
	Ready              bool     `json:"ready"`
	MissingTables      []string `json:"missing_tables,omitempty"`
	MissingIndexes     []string `json:"missing_indexes,omitempty"`
	MissingForeignKeys []string `json:"missing_foreign_keys,omitempty"`
}

type createPaymentIntentRequest struct {
	OrderID        string                       `json:"order_id"`
	UserID         string                       `json:"user_id"`
	Amount         int64                        `json:"amount"`
	Currency       string                       `json:"currency"`
	Customer       paymentIntentCustomerRequest `json:"customer"`
	IdempotencyKey string                       `json:"idempotency_key"`
	Provider       string                       `json:"provider,omitempty"`
	Metadata       map[string]string            `json:"metadata,omitempty"`
}

type paymentIntentCustomerRequest struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

type createPaymentIntentResponse struct {
	PaymentID        string                       `json:"payment_id"`
	OrderID          string                       `json:"order_id"`
	Provider         string                       `json:"provider"`
	ProviderIntentID string                       `json:"provider_intent_id"`
	Status           string                       `json:"status"`
	Amount           int64                        `json:"amount"`
	Currency         string                       `json:"currency"`
	ClientPayload    paymentClientPayloadResponse `json:"client_payload"`
}

type retryPaymentIntentRequest struct {
	PaymentID      string `json:"payment_id,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type retryPaymentIntentResponse struct {
	PaymentID        string                       `json:"payment_id"`
	OrderID          string                       `json:"order_id"`
	Provider         string                       `json:"provider"`
	ProviderIntentID string                       `json:"provider_intent_id"`
	Status           string                       `json:"status"`
	Amount           int64                        `json:"amount"`
	Currency         string                       `json:"currency"`
	ClientPayload    paymentClientPayloadResponse `json:"client_payload"`
	Retry            paymentRetryResponse         `json:"retry"`
}

type paymentRetryResponse struct {
	RetryOfPaymentID string `json:"retry_of_payment_id"`
	RootPaymentID    string `json:"root_payment_id"`
	AttemptNo        uint32 `json:"attempt_no"`
	Replayed         bool   `json:"replayed"`
}

type paymentClientPayloadResponse struct {
	PublicKey         string            `json:"public_key,omitempty"`
	ClientSecret      string            `json:"client_secret,omitempty"`
	ProviderOrderID   string            `json:"provider_order_id,omitempty"`
	CheckoutSessionID string            `json:"checkout_session_id,omitempty"`
	RedirectURL       string            `json:"redirect_url,omitempty"`
	Extra             map[string]string `json:"extra,omitempty"`
}

type handleWebhookResponse struct {
	Success          bool   `json:"success"`
	WebhookEventID   string `json:"webhook_event_id"`
	ProviderEventID  string `json:"provider_event_id"`
	ProcessingStatus string `json:"processing_status"`
	PaymentID        string `json:"payment_id,omitempty"`
	PaymentStatus    string `json:"payment_status,omitempty"`
	RefundID         string `json:"refund_id,omitempty"`
	RefundStatus     string `json:"refund_status,omitempty"`
}

type moneyRequest struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type refundPaymentRequest struct {
	PaymentID      string       `json:"payment_id,omitempty"`
	Amount         moneyRequest `json:"amount"`
	Reason         string       `json:"reason"`
	IdempotencyKey string       `json:"idempotency_key"`
}

type reviewRefundRequest struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type refundResponse struct {
	RefundID         string       `json:"refund_id"`
	PaymentID        string       `json:"payment_id"`
	ProviderRefundID string       `json:"provider_refund_id,omitempty"`
	Status           string       `json:"status"`
	Amount           moneyRequest `json:"amount"`
	Reason           string       `json:"reason"`
	RequestedBy      string       `json:"requested_by"`
	ReviewedBy       string       `json:"reviewed_by,omitempty"`
	ReviewReason     string       `json:"review_reason,omitempty"`
	ReviewedAt       *time.Time   `json:"reviewed_at,omitempty"`
	CreatedAt        *time.Time   `json:"created_at,omitempty"`
	Replayed         bool         `json:"replayed,omitempty"`
}

type healthResponse struct {
	Status string `json:"status"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func stateMachineResponseFromDomain(definition domain.StateMachineDefinition) stateMachineResponse {
	statuses := make([]paymentStatusDefinitionResponse, 0, len(definition.Statuses))
	for _, status := range definition.Statuses {
		statuses = append(statuses, paymentStatusDefinitionResponse{
			Status:          string(status.Status),
			Meaning:         status.Meaning,
			UserFacing:      status.UserFacing,
			FinalForAttempt: status.FinalForAttempt,
			Helper:          status.Helper,
			Persisted:       status.Persisted,
		})
	}

	transitions := make([]transitionRuleResponse, 0, len(definition.Transitions))
	for _, transition := range definition.Transitions {
		transitions = append(transitions, transitionRuleResponse{
			From:          string(transition.From),
			Event:         string(transition.Event),
			To:            string(transition.To),
			TriggerSource: transition.TriggerSource,
			Notes:         transition.Notes,
		})
	}

	invalidExamples := make([]invalidTransitionResponse, 0, len(definition.InvalidExamples))
	for _, invalid := range definition.InvalidExamples {
		invalidExamples = append(invalidExamples, invalidTransitionResponse{
			From:   string(invalid.From),
			To:     string(invalid.To),
			Reason: invalid.Reason,
		})
	}

	providerMappings := make([]providerEventMappingResponse, 0, len(definition.ProviderMappings))
	for _, mapping := range definition.ProviderMappings {
		providerMappings = append(providerMappings, providerEventMappingResponse{
			ProviderEvent: mapping.ProviderEvent,
			InternalEvent: string(mapping.InternalEvent),
			NextStatus:    string(mapping.NextStatus),
		})
	}

	orderMappings := make([]orderStateMappingResponse, 0, len(definition.OrderMappings))
	for _, mapping := range definition.OrderMappings {
		orderMappings = append(orderMappings, orderStateMappingResponse{
			PaymentStatus: string(mapping.PaymentStatus),
			OrderStatus:   string(mapping.OrderStatus),
			Explanation:   mapping.Explanation,
		})
	}

	return stateMachineResponse{
		Statuses:         statuses,
		Transitions:      transitions,
		InvalidExamples:  invalidExamples,
		ProviderMappings: providerMappings,
		OrderMappings:    orderMappings,
	}
}

func paymentSchemaResponseFromDomain(definition domain.PaymentSchemaDefinition) paymentSchemaResponse {
	tables := make([]paymentTableDefinitionDTO, 0, len(definition.Tables))
	for _, table := range definition.Tables {
		tables = append(tables, paymentTableDefinitionDTO{
			Name:        table.Name,
			Columns:     append([]string(nil), table.Columns...),
			UniqueKeys:  append([]string(nil), table.UniqueKeys...),
			Indexes:     append([]string(nil), table.Indexes...),
			ForeignKeys: append([]string(nil), table.ForeignKeys...),
		})
	}
	return paymentSchemaResponse{
		DatabaseName: definition.DatabaseName,
		Tables:       tables,
	}
}

func paymentSchemaHealthResponseFromUsecase(out usecase.PaymentSchemaHealthOutput) paymentSchemaHealthResponse {
	return paymentSchemaHealthResponse{
		DatabaseName:       out.Verification.DatabaseName,
		Ready:              out.Verification.Ready,
		MissingTables:      append([]string(nil), out.Verification.MissingTables...),
		MissingIndexes:     append([]string(nil), out.Verification.MissingIndexes...),
		MissingForeignKeys: append([]string(nil), out.Verification.MissingForeignKeys...),
	}
}

func createPaymentIntentInputFromRequest(req createPaymentIntentRequest, requestID string) usecase.CreatePaymentIntentInput {
	return usecase.CreatePaymentIntentInput{
		OrderID:        req.OrderID,
		UserID:         req.UserID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		IdempotencyKey: req.IdempotencyKey,
		Provider:       req.Provider,
		Metadata:       req.Metadata,
		RequestID:      requestID,
		Customer: usecase.PaymentIntentCustomer{
			Name:  req.Customer.Name,
			Email: req.Customer.Email,
			Phone: req.Customer.Phone,
		},
	}
}

func createPaymentIntentResponseFromUsecase(out usecase.CreatePaymentIntentOutput) createPaymentIntentResponse {
	return createPaymentIntentResponse{
		PaymentID:        out.PaymentID,
		OrderID:          out.OrderID,
		Provider:         out.Provider,
		ProviderIntentID: out.ProviderIntentID,
		Status:           string(out.Status),
		Amount:           out.Amount,
		Currency:         out.Currency,
		ClientPayload: paymentClientPayloadResponse{
			PublicKey:         out.ClientPayload.PublicKey,
			ClientSecret:      out.ClientPayload.ClientSecret,
			ProviderOrderID:   out.ClientPayload.ProviderOrderID,
			CheckoutSessionID: out.ClientPayload.CheckoutSessionID,
			RedirectURL:       out.ClientPayload.RedirectURL,
			Extra:             out.ClientPayload.Extra,
		},
	}
}

func retryPaymentIntentResponseFromUsecase(out usecase.RetryPaymentIntentOutput) retryPaymentIntentResponse {
	return retryPaymentIntentResponse{
		PaymentID:        out.PaymentID,
		OrderID:          out.OrderID,
		Provider:         out.Provider,
		ProviderIntentID: out.ProviderIntentID,
		Status:           string(out.Status),
		Amount:           out.Amount,
		Currency:         out.Currency,
		ClientPayload: paymentClientPayloadResponse{
			PublicKey:         out.ClientPayload.PublicKey,
			ClientSecret:      out.ClientPayload.ClientSecret,
			ProviderOrderID:   out.ClientPayload.ProviderOrderID,
			CheckoutSessionID: out.ClientPayload.CheckoutSessionID,
			RedirectURL:       out.ClientPayload.RedirectURL,
			Extra:             out.ClientPayload.Extra,
		},
		Retry: paymentRetryResponse{
			RetryOfPaymentID: out.RetryOfPaymentID,
			RootPaymentID:    out.RootPaymentID,
			AttemptNo:        out.AttemptNo,
			Replayed:         out.Replayed,
		},
	}
}

func handleWebhookResponseFromUsecase(out usecase.HandleWebhookOutput) handleWebhookResponse {
	return handleWebhookResponse{
		Success:          true,
		WebhookEventID:   out.WebhookEventID,
		ProviderEventID:  out.ProviderEventID,
		ProcessingStatus: string(out.ProcessingStatus),
		PaymentID:        out.PaymentID,
		PaymentStatus:    string(out.Status),
		RefundID:         out.RefundID,
		RefundStatus:     string(out.RefundStatus),
	}
}

func refundPaymentInputFromRequest(paymentID string, actorID string, requestID string, req refundPaymentRequest) usecase.RefundPaymentInput {
	if req.PaymentID != "" {
		paymentID = req.PaymentID
	}
	return usecase.RefundPaymentInput{
		PaymentID:      paymentID,
		AmountMinor:    req.Amount.Amount,
		Currency:       req.Amount.Currency,
		Reason:         req.Reason,
		RequestedBy:    actorID,
		IdempotencyKey: req.IdempotencyKey,
		RequestID:      requestID,
	}
}

func refundResponseFromDomain(refund domain.Refund, replayed bool) refundResponse {
	createdAt := refund.CreatedAt
	return refundResponse{
		RefundID:         refund.RefundID,
		PaymentID:        refund.PaymentID,
		ProviderRefundID: refund.ProviderRefundID,
		Status:           string(refund.Status),
		Amount:           moneyRequest{Amount: refund.Amount.Amount, Currency: refund.Amount.Currency},
		Reason:           refund.Reason,
		RequestedBy:      refund.RequestedBy,
		ReviewedBy:       refund.ReviewedBy,
		ReviewReason:     refund.ReviewReason,
		ReviewedAt:       refund.ReviewedAt,
		CreatedAt:        &createdAt,
		Replayed:         replayed,
	}
}
