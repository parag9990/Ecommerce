package domain

import (
	"fmt"
	"strings"
	"time"
)

type Money struct {
	Currency string `json:"currency"`
	Amount   int64  `json:"amount"`
}

func (m Money) SameCurrency(other Money) bool {
	return strings.EqualFold(strings.TrimSpace(m.Currency), strings.TrimSpace(other.Currency))
}

type OrderStatus string

const (
	OrderStatusCreated        OrderStatus = "created"
	OrderStatusPendingPayment OrderStatus = "pending_payment"
	OrderStatusPaid           OrderStatus = "paid"
	OrderStatusPacked         OrderStatus = "packed"
	OrderStatusShipped        OrderStatus = "shipped"
	OrderStatusDelivered      OrderStatus = "delivered"
	OrderStatusCancelled      OrderStatus = "cancelled"
	OrderStatusRefunded       OrderStatus = "refunded"
	OrderStatusPaymentFailed  OrderStatus = "payment_failed"
)

func ParseOrderStatus(value string) (OrderStatus, error) {
	status := OrderStatus(strings.TrimSpace(value))
	if status == "" {
		return "", nil
	}
	if !status.Valid() {
		return "", NewInvalidStatus("order", value)
	}
	return status, nil
}

func (s OrderStatus) Valid() bool {
	switch s {
	case OrderStatusCreated,
		OrderStatusPendingPayment,
		OrderStatusPaid,
		OrderStatusPacked,
		OrderStatusShipped,
		OrderStatusDelivered,
		OrderStatusCancelled,
		OrderStatusRefunded,
		OrderStatusPaymentFailed:
		return true
	default:
		return false
	}
}

type PaymentStatus string

const (
	PaymentStatusInitiated         PaymentStatus = "initiated"
	PaymentStatusRequiresAction    PaymentStatus = "requires_action"
	PaymentStatusAuthorized        PaymentStatus = "authorized"
	PaymentStatusCaptured          PaymentStatus = "captured"
	PaymentStatusFailed            PaymentStatus = "failed"
	PaymentStatusRefunded          PaymentStatus = "refunded"
	PaymentStatusPartiallyRefunded PaymentStatus = "partially_refunded"
)

func ParsePaymentStatus(value string) (PaymentStatus, error) {
	status := PaymentStatus(strings.TrimSpace(value))
	if status == "" {
		return "", nil
	}
	if !status.Valid() {
		return "", NewInvalidStatus("payment", value)
	}
	return status, nil
}

func (s PaymentStatus) Valid() bool {
	switch s {
	case PaymentStatusInitiated,
		PaymentStatusRequiresAction,
		PaymentStatusAuthorized,
		PaymentStatusCaptured,
		PaymentStatusFailed,
		PaymentStatusRefunded,
		PaymentStatusPartiallyRefunded:
		return true
	default:
		return false
	}
}

type RefundStatus string

const (
	RefundStatusRequested  RefundStatus = "requested"
	RefundStatusApproved   RefundStatus = "approved"
	RefundStatusRejected   RefundStatus = "rejected"
	RefundStatusProcessing RefundStatus = "processing"
	RefundStatusSucceeded  RefundStatus = "succeeded"
	RefundStatusFailed     RefundStatus = "failed"
)

func (s RefundStatus) Valid() bool {
	switch s {
	case RefundStatusRequested,
		RefundStatusApproved,
		RefundStatusRejected,
		RefundStatusProcessing,
		RefundStatusSucceeded,
		RefundStatusFailed:
		return true
	default:
		return false
	}
}

type RefundDecision string

const (
	RefundDecisionApproved RefundDecision = "approved"
	RefundDecisionRejected RefundDecision = "rejected"
)

func ParseRefundDecision(value string) (RefundDecision, error) {
	decision := RefundDecision(strings.TrimSpace(value))
	if !decision.Valid() {
		return "", NewValidationError("decision must be approved or rejected")
	}
	return decision, nil
}

func (d RefundDecision) Valid() bool {
	switch d {
	case RefundDecisionApproved, RefundDecisionRejected:
		return true
	default:
		return false
	}
}

func (d RefundDecision) ReviewTaskStatus() ReviewTaskStatus {
	if d == RefundDecisionApproved {
		return ReviewTaskStatusApproved
	}
	return ReviewTaskStatusRejected
}

func (d RefundDecision) RefundStatus() RefundStatus {
	if d == RefundDecisionApproved {
		return RefundStatusApproved
	}
	return RefundStatusRejected
}

type ManualReviewCategory string

const (
	ManualReviewPaymentMismatch  ManualReviewCategory = "payment_mismatch"
	ManualReviewFulfillmentDelay ManualReviewCategory = "fulfillment_delay"
	ManualReviewRefundFollowup   ManualReviewCategory = "refund_followup"
	ManualReviewFraudRisk        ManualReviewCategory = "fraud_risk"
	ManualReviewCustomerDispute  ManualReviewCategory = "customer_dispute"
)

func ParseManualReviewCategory(value string) (ManualReviewCategory, error) {
	category := ManualReviewCategory(strings.TrimSpace(value))
	if !category.Valid() {
		return "", NewValidationError("category must be payment_mismatch, fulfillment_delay, refund_followup, fraud_risk, or customer_dispute")
	}
	return category, nil
}

func (c ManualReviewCategory) Valid() bool {
	switch c {
	case ManualReviewPaymentMismatch,
		ManualReviewFulfillmentDelay,
		ManualReviewRefundFollowup,
		ManualReviewFraudRisk,
		ManualReviewCustomerDispute:
		return true
	default:
		return false
	}
}

type AdminOrderListRequest struct {
	Status   OrderStatus `json:"status,omitempty"`
	UserID   string      `json:"user_id,omitempty"`
	SellerID string      `json:"seller_id,omitempty"`
	Pagination
}

func (r AdminOrderListRequest) Normalize() (AdminOrderListRequest, error) {
	status, err := ParseOrderStatus(string(r.Status))
	if err != nil {
		return AdminOrderListRequest{}, err
	}
	r.Status = status
	r.UserID = strings.TrimSpace(r.UserID)
	r.SellerID = strings.TrimSpace(r.SellerID)
	if len(r.UserID) > MaxSearchLength {
		return AdminOrderListRequest{}, NewValidationError(fmt.Sprintf("user_id must be at most %d characters", MaxSearchLength))
	}
	if len(r.SellerID) > MaxSearchLength {
		return AdminOrderListRequest{}, NewValidationError(fmt.Sprintf("seller_id must be at most %d characters", MaxSearchLength))
	}
	r.Pagination = NormalizePagination(r.Pagination.Page, r.Pagination.PageSize, r.Pagination.Cursor)
	return r, nil
}

type AdminPaymentListRequest struct {
	Status   PaymentStatus `json:"status,omitempty"`
	Provider string        `json:"provider,omitempty"`
	OrderID  string        `json:"order_id,omitempty"`
	Pagination
}

func (r AdminPaymentListRequest) Normalize() (AdminPaymentListRequest, error) {
	status, err := ParsePaymentStatus(string(r.Status))
	if err != nil {
		return AdminPaymentListRequest{}, err
	}
	r.Status = status
	r.Provider = strings.TrimSpace(r.Provider)
	r.OrderID = strings.TrimSpace(r.OrderID)
	if len(r.Provider) > MaxSearchLength {
		return AdminPaymentListRequest{}, NewValidationError(fmt.Sprintf("provider must be at most %d characters", MaxSearchLength))
	}
	if len(r.OrderID) > MaxSearchLength {
		return AdminPaymentListRequest{}, NewValidationError(fmt.Sprintf("order_id must be at most %d characters", MaxSearchLength))
	}
	r.Pagination = NormalizePagination(r.Pagination.Page, r.Pagination.PageSize, r.Pagination.Cursor)
	return r, nil
}

type OrderSnapshot struct {
	OrderID     string         `json:"order_id"`
	UserID      string         `json:"user_id,omitempty"`
	SellerID    string         `json:"seller_id,omitempty"`
	Status      OrderStatus    `json:"status"`
	TotalAmount Money          `json:"total_amount"`
	Items       []OrderItem    `json:"items,omitempty"`
	CreatedAt   *time.Time     `json:"created_at,omitempty"`
	UpdatedAt   *time.Time     `json:"updated_at,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type OrderItem struct {
	OrderItemID string `json:"order_item_id,omitempty"`
	ProductID   string `json:"product_id,omitempty"`
	VariantID   string `json:"variant_id,omitempty"`
	SellerID    string `json:"seller_id,omitempty"`
	Quantity    int    `json:"quantity,omitempty"`
	Status      string `json:"status,omitempty"`
	Amount      Money  `json:"amount,omitempty"`
}

type OrderStatusEvent struct {
	Status     OrderStatus    `json:"status"`
	Reason     string         `json:"reason,omitempty"`
	ActorID    string         `json:"actor_id,omitempty"`
	OccurredAt *time.Time     `json:"occurred_at,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type AdminOrderListResponse struct {
	Orders   []OrderSnapshot `json:"orders"`
	Page     int             `json:"page,omitempty"`
	PageSize int             `json:"page_size,omitempty"`
	Total    int64           `json:"total,omitempty"`
	Cursor   string          `json:"cursor,omitempty"`
}

type PaymentSnapshot struct {
	PaymentID string        `json:"payment_id"`
	OrderID   string        `json:"order_id,omitempty"`
	Provider  string        `json:"provider,omitempty"`
	Status    PaymentStatus `json:"status"`
	Amount    Money         `json:"amount"`
	CreatedAt *time.Time    `json:"created_at,omitempty"`
	UpdatedAt *time.Time    `json:"updated_at,omitempty"`
}

type AdminPaymentListResponse struct {
	Payments []PaymentSnapshot `json:"payments"`
	Page     int               `json:"page,omitempty"`
	PageSize int               `json:"page_size,omitempty"`
	Total    int64             `json:"total,omitempty"`
	Cursor   string            `json:"cursor,omitempty"`
}

type RefundSnapshot struct {
	RefundID  string       `json:"refund_id"`
	PaymentID string       `json:"payment_id,omitempty"`
	OrderID   string       `json:"order_id,omitempty"`
	Status    RefundStatus `json:"status"`
	Amount    Money        `json:"amount"`
	Reason    string       `json:"reason,omitempty"`
	CreatedAt *time.Time   `json:"created_at,omitempty"`
	UpdatedAt *time.Time   `json:"updated_at,omitempty"`
}

type RefundReviewRequest struct {
	RefundID string `json:"refund_id,omitempty"`
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type NormalizedRefundReview struct {
	Decision RefundDecision
	Reason   string
}

func NormalizeRefundReviewRequest(pathRefundID string, req RefundReviewRequest) (NormalizedRefundReview, error) {
	pathRefundID = strings.TrimSpace(pathRefundID)
	if pathRefundID == "" {
		return NormalizedRefundReview{}, NewValidationError("refund_id is required")
	}
	bodyRefundID := strings.TrimSpace(req.RefundID)
	if bodyRefundID != "" && bodyRefundID != pathRefundID {
		return NormalizedRefundReview{}, NewValidationError("refund_id in body must match path")
	}
	decision, err := ParseRefundDecision(req.Decision)
	if err != nil {
		return NormalizedRefundReview{}, err
	}
	reason, err := NormalizeMutationReason(req.Reason)
	if err != nil {
		return NormalizedRefundReview{}, err
	}
	return NormalizedRefundReview{Decision: decision, Reason: reason}, nil
}

type ManualOrderReviewRequest struct {
	Category string `json:"category"`
	Reason   string `json:"reason"`
}

type NormalizedManualOrderReview struct {
	Category ManualReviewCategory
	Reason   string
}

func NormalizeManualOrderReviewRequest(orderID string, req ManualOrderReviewRequest) (NormalizedManualOrderReview, error) {
	if strings.TrimSpace(orderID) == "" {
		return NormalizedManualOrderReview{}, NewValidationError("order_id is required")
	}
	category, err := ParseManualReviewCategory(req.Category)
	if err != nil {
		return NormalizedManualOrderReview{}, err
	}
	reason, err := NormalizeMutationReason(req.Reason)
	if err != nil {
		return NormalizedManualOrderReview{}, err
	}
	return NormalizedManualOrderReview{Category: category, Reason: reason}, nil
}

type ReviewTask struct {
	TaskID       string           `json:"task_id"`
	TaskType     string           `json:"task_type"`
	ResourceType string           `json:"resource_type"`
	ResourceID   string           `json:"resource_id"`
	Status       ReviewTaskStatus `json:"status"`
	AssignedTo   string           `json:"assigned_to,omitempty"`
	CreatedBy    string           `json:"created_by,omitempty"`
	ReviewedBy   string           `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time       `json:"reviewed_at,omitempty"`
	Reason       string           `json:"reason,omitempty"`
	Metadata     map[string]any   `json:"metadata,omitempty"`
	CreatedAt    *time.Time       `json:"created_at,omitempty"`
	UpdatedAt    *time.Time       `json:"updated_at,omitempty"`
}

type DisputeView struct {
	Order         OrderSnapshot      `json:"order"`
	StatusHistory []OrderStatusEvent `json:"status_history,omitempty"`
	Payments      []PaymentSnapshot  `json:"payments,omitempty"`
	Refunds       []RefundSnapshot   `json:"refunds,omitempty"`
	ReviewTasks   []ReviewTask       `json:"review_tasks,omitempty"`
	RiskFlags     []string           `json:"risk_flags,omitempty"`
}

func BuildDisputeView(order OrderSnapshot, history []OrderStatusEvent, payments []PaymentSnapshot, refunds []RefundSnapshot, reviewTasks []ReviewTask) DisputeView {
	return DisputeView{
		Order:         order,
		StatusHistory: history,
		Payments:      payments,
		Refunds:       refunds,
		ReviewTasks:   reviewTasks,
		RiskFlags:     disputeRiskFlags(order, payments, refunds, reviewTasks),
	}
}

func disputeRiskFlags(order OrderSnapshot, payments []PaymentSnapshot, refunds []RefundSnapshot, reviewTasks []ReviewTask) []string {
	flags := make([]string, 0, 4)
	seen := map[string]struct{}{}
	add := func(flag string) {
		if _, ok := seen[flag]; ok {
			return
		}
		seen[flag] = struct{}{}
		flags = append(flags, flag)
	}

	var capturedAmount int64
	var capturedCurrency string
	for _, payment := range payments {
		switch payment.Status {
		case PaymentStatusCaptured, PaymentStatusPartiallyRefunded, PaymentStatusRefunded:
			if capturedCurrency == "" {
				capturedCurrency = payment.Amount.Currency
			}
			if strings.EqualFold(capturedCurrency, payment.Amount.Currency) {
				capturedAmount += payment.Amount.Amount
			}
		case PaymentStatusFailed:
			add("payment_failed")
		}
	}

	if order.Status == OrderStatusPendingPayment && capturedAmount > 0 {
		add("payment_mismatch")
	}
	if order.TotalAmount.Amount > 0 && capturedAmount > 0 && order.TotalAmount.SameCurrency(Money{Currency: capturedCurrency}) && capturedAmount != order.TotalAmount.Amount {
		add("amount_mismatch")
	}

	for _, refund := range refunds {
		switch refund.Status {
		case RefundStatusRequested:
			add("refund_requested")
		case RefundStatusProcessing:
			add("refund_processing")
		case RefundStatusFailed:
			add("refund_failed")
		}
	}

	for _, task := range reviewTasks {
		if task.Status == ReviewTaskStatusOpen {
			add("manual_review_open")
			break
		}
	}

	return flags
}
