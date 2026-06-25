package http

import (
	"fmt"
	"strings"
	"time"

	"ecommerce/superadmin-service/internal/domain"
)

type moneyDTO struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type orderItemDTO struct {
	ItemID            string    `json:"item_id"`
	ProductID         string    `json:"product_id,omitempty"`
	SellerID          string    `json:"seller_id,omitempty"`
	Title             string    `json:"title"`
	Quantity          int       `json:"quantity"`
	UnitPrice         *moneyDTO `json:"unit_price,omitempty"`
	FulfillmentStatus string    `json:"fulfillment_status,omitempty"`
}

type orderPaymentSummaryDTO struct {
	PaymentID    string    `json:"payment_id,omitempty"`
	Provider     string    `json:"provider,omitempty"`
	Status       string    `json:"status,omitempty"`
	RefundStatus string    `json:"refund_status,omitempty"`
	Amount       *moneyDTO `json:"amount,omitempty"`
}

type orderDTO struct {
	OrderID      string                  `json:"order_id"`
	UserID       string                  `json:"user_id"`
	Status       string                  `json:"status"`
	ReviewStatus string                  `json:"review_status,omitempty"`
	Total        moneyDTO                `json:"total"`
	Items        []orderItemDTO          `json:"items"`
	Payment      *orderPaymentSummaryDTO `json:"payment,omitempty"`
	CreatedAt    *time.Time              `json:"created_at,omitempty"`
	UpdatedAt    *time.Time              `json:"updated_at,omitempty"`
}

type orderStatusEventDTO struct {
	Status    string     `json:"status"`
	Note      string     `json:"note,omitempty"`
	ActorType string     `json:"actor_type"`
	ActorID   string     `json:"actor_id,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

type reviewTaskDTO struct {
	TaskID       string     `json:"task_id"`
	TaskType     string     `json:"task_type"`
	ResourceType string     `json:"resource_type"`
	ResourceID   string     `json:"resource_id"`
	Status       string     `json:"status"`
	Reason       string     `json:"reason,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type adminOrderListDTO struct {
	Orders   []orderDTO `json:"orders"`
	Total    int64      `json:"total,omitempty"`
	Page     int        `json:"page,omitempty"`
	Limit    int        `json:"limit,omitempty"`
	PageSize int        `json:"page_size,omitempty"`
	Cursor   string     `json:"cursor,omitempty"`
}

type adminOrderDetailDTO struct {
	Order         orderDTO                `json:"order"`
	StatusHistory []orderStatusEventDTO   `json:"status_history"`
	Shipments     []any                   `json:"shipments"`
	Payment       *orderPaymentSummaryDTO `json:"payment,omitempty"`
	ReviewTasks   []reviewTaskDTO         `json:"review_tasks,omitempty"`
	RiskFlags     []string                `json:"risk_flags,omitempty"`
}

type paymentDTO struct {
	PaymentID         string              `json:"payment_id"`
	OrderID           string              `json:"order_id"`
	Provider          string              `json:"provider"`
	ProviderPaymentID string              `json:"provider_payment_id,omitempty"`
	Status            string              `json:"status"`
	Amount            moneyDTO            `json:"amount"`
	CapturedAt        *time.Time          `json:"captured_at,omitempty"`
	CreatedAt         *time.Time          `json:"created_at,omitempty"`
	UpdatedAt         *time.Time          `json:"updated_at,omitempty"`
	Attempts          []paymentAttemptDTO `json:"attempts,omitempty"`
	Refunds           []refundDTO         `json:"refunds,omitempty"`
}

type paymentAttemptDTO struct {
	AttemptID         string     `json:"attempt_id"`
	PaymentID         string     `json:"payment_id"`
	Status            string     `json:"status"`
	ProviderReference string     `json:"provider_reference,omitempty"`
	ErrorCode         string     `json:"error_code,omitempty"`
	ErrorMessage      string     `json:"error_message,omitempty"`
	CreatedAt         *time.Time `json:"created_at,omitempty"`
}

type adminPaymentListDTO struct {
	Payments []paymentDTO `json:"payments"`
	Total    int64        `json:"total,omitempty"`
	Page     int          `json:"page,omitempty"`
	PageSize int          `json:"page_size,omitempty"`
	Cursor   string       `json:"cursor,omitempty"`
}

type adminPaymentDetailDTO struct {
	Payment  paymentDTO          `json:"payment"`
	Attempts []paymentAttemptDTO `json:"attempts"`
	Refunds  []refundDTO         `json:"refunds"`
}

type refundDTO struct {
	RefundID    string     `json:"refund_id"`
	PaymentID   string     `json:"payment_id"`
	OrderID     string     `json:"order_id,omitempty"`
	Status      string     `json:"status"`
	Amount      moneyDTO   `json:"amount"`
	Reason      string     `json:"reason"`
	RequestedBy string     `json:"requested_by,omitempty"`
	ReviewedBy  string     `json:"reviewed_by,omitempty"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

type refundListDTO struct {
	Refunds  []refundDTO `json:"refunds"`
	Total    int64       `json:"total,omitempty"`
	Page     int         `json:"page,omitempty"`
	PageSize int         `json:"page_size,omitempty"`
	Cursor   string      `json:"cursor,omitempty"`
}

type reconciliationDTO struct {
	ReconciliationID string     `json:"reconciliation_id"`
	PaymentID        string     `json:"payment_id,omitempty"`
	Provider         string     `json:"provider"`
	Status           string     `json:"status"`
	LocalAmount      *moneyDTO  `json:"local_amount,omitempty"`
	ProviderAmount   *moneyDTO  `json:"provider_amount,omitempty"`
	LocalStatus      string     `json:"local_status,omitempty"`
	ProviderStatus   string     `json:"provider_status,omitempty"`
	SettlementID     string     `json:"settlement_id,omitempty"`
	DetectedAt       *time.Time `json:"detected_at,omitempty"`
	Note             string     `json:"note,omitempty"`
}

type reconciliationListDTO struct {
	Alerts   []reconciliationDTO `json:"alerts"`
	Total    int64               `json:"total,omitempty"`
	Page     int                 `json:"page,omitempty"`
	PageSize int                 `json:"page_size,omitempty"`
	Cursor   string              `json:"cursor,omitempty"`
}

type reconciliationDetailDTO struct {
	Alert reconciliationDTO `json:"alert"`
}

func adminOrderListResponseDTO(response domain.AdminOrderListResponse) adminOrderListDTO {
	orders := make([]orderDTO, 0, len(response.Orders))
	for _, order := range response.Orders {
		orders = append(orders, orderDTOFromDomain(order, nil, nil))
	}
	return adminOrderListDTO{
		Orders:   orders,
		Total:    response.Total,
		Page:     response.Page,
		Limit:    response.PageSize,
		PageSize: response.PageSize,
		Cursor:   response.Cursor,
	}
}

func adminOrderDetailResponseDTO(response domain.AdminOrderDetailResponse) adminOrderDetailDTO {
	payment := orderPaymentSummaryFromPayments(response.Payments, response.Refunds)
	return adminOrderDetailDTO{
		Order:         orderDTOFromDomain(response.Order, payment, response.Refunds),
		StatusHistory: orderStatusEventDTOs(response.StatusHistory),
		Shipments:     []any{},
		Payment:       payment,
		ReviewTasks:   reviewTaskDTOs(response.ReviewTasks),
		RiskFlags:     append([]string(nil), response.RiskFlags...),
	}
}

func adminPaymentListResponseDTO(response domain.AdminPaymentListResponse) adminPaymentListDTO {
	payments := make([]paymentDTO, 0, len(response.Payments))
	for _, payment := range response.Payments {
		payments = append(payments, paymentDTOFromDomain(payment, nil, nil))
	}
	return adminPaymentListDTO{
		Payments: payments,
		Total:    response.Total,
		Page:     response.Page,
		PageSize: response.PageSize,
		Cursor:   response.Cursor,
	}
}

func adminPaymentDetailResponseDTO(response domain.AdminPaymentDetailResponse) adminPaymentDetailDTO {
	attempts := paymentAttemptDTOs(response.Attempts)
	refunds := refundDTOs(response.Refunds)
	return adminPaymentDetailDTO{
		Payment:  paymentDTOFromDomain(response.Payment, attempts, refunds),
		Attempts: attempts,
		Refunds:  refunds,
	}
}

func refundListResponseDTO(response domain.RefundListResponse) refundListDTO {
	return refundListDTO{
		Refunds:  refundDTOs(response.Refunds),
		Total:    response.Total,
		Page:     response.Page,
		PageSize: response.PageSize,
		Cursor:   response.Cursor,
	}
}

func reconciliationListResponseDTO(response domain.ReconciliationListResponse) reconciliationListDTO {
	alerts := make([]reconciliationDTO, 0, len(response.Alerts))
	for _, alert := range response.Alerts {
		alerts = append(alerts, reconciliationDTOFromDomain(alert))
	}
	return reconciliationListDTO{
		Alerts:   alerts,
		Total:    response.Total,
		Page:     response.Page,
		PageSize: response.PageSize,
		Cursor:   response.Cursor,
	}
}

func reconciliationDetailResponseDTO(response domain.ReconciliationDetailResponse) reconciliationDetailDTO {
	return reconciliationDetailDTO{Alert: reconciliationDTOFromDomain(response.Alert)}
}

func orderDTOFromDomain(order domain.OrderSnapshot, payment *orderPaymentSummaryDTO, refunds []domain.RefundSnapshot) orderDTO {
	items := make([]orderItemDTO, 0, len(order.Items))
	for i, item := range order.Items {
		itemID := firstNonEmptyString(item.OrderItemID, item.ProductID, fmt.Sprintf("item_%d", i+1))
		title := firstNonEmptyString(item.ProductID, item.VariantID, "Order item")
		unitPrice := moneyDTOFromDomainPtr(item.Amount)
		items = append(items, orderItemDTO{
			ItemID:            itemID,
			ProductID:         item.ProductID,
			SellerID:          item.SellerID,
			Title:             title,
			Quantity:          item.Quantity,
			UnitPrice:         unitPrice,
			FulfillmentStatus: item.Status,
		})
	}
	if payment == nil {
		payment = orderPaymentSummaryFromPayments(nil, refunds)
	}
	return orderDTO{
		OrderID:      order.OrderID,
		UserID:       order.UserID,
		Status:       string(order.Status),
		ReviewStatus: reviewStatusFromTasks(order.Metadata),
		Total:        moneyDTOFromDomain(order.TotalAmount),
		Items:        items,
		Payment:      payment,
		CreatedAt:    order.CreatedAt,
		UpdatedAt:    order.UpdatedAt,
	}
}

func orderPaymentSummaryFromPayments(payments []domain.PaymentSnapshot, refunds []domain.RefundSnapshot) *orderPaymentSummaryDTO {
	if len(payments) == 0 {
		return nil
	}
	payment := payments[0]
	return &orderPaymentSummaryDTO{
		PaymentID:    payment.PaymentID,
		Provider:     payment.Provider,
		Status:       string(payment.Status),
		RefundStatus: refundSummaryStatus(refunds),
		Amount:       moneyDTOFromDomainPtr(payment.Amount),
	}
}

func refundSummaryStatus(refunds []domain.RefundSnapshot) string {
	for _, refund := range refunds {
		if strings.TrimSpace(string(refund.Status)) != "" {
			return string(refund.Status)
		}
	}
	return ""
}

func orderStatusEventDTOs(events []domain.OrderStatusEvent) []orderStatusEventDTO {
	out := make([]orderStatusEventDTO, 0, len(events))
	for _, event := range events {
		out = append(out, orderStatusEventDTO{
			Status:    string(event.Status),
			Note:      event.Reason,
			ActorType: firstNonEmptyString(eventActorType(event.ActorID), "system"),
			ActorID:   event.ActorID,
			CreatedAt: event.OccurredAt,
		})
	}
	return out
}

func eventActorType(actorID string) string {
	switch {
	case strings.HasPrefix(actorID, "admin"):
		return "admin"
	case strings.HasPrefix(actorID, "seller"):
		return "seller"
	default:
		return "system"
	}
}

func paymentDTOFromDomain(payment domain.PaymentSnapshot, attempts []paymentAttemptDTO, refunds []refundDTO) paymentDTO {
	return paymentDTO{
		PaymentID: payment.PaymentID,
		OrderID:   payment.OrderID,
		Provider:  payment.Provider,
		Status:    string(payment.Status),
		Amount:    moneyDTOFromDomain(payment.Amount),
		CreatedAt: payment.CreatedAt,
		UpdatedAt: payment.UpdatedAt,
		Attempts:  attempts,
		Refunds:   refunds,
	}
}

func paymentAttemptDTOs(attempts []domain.PaymentAttemptSnapshot) []paymentAttemptDTO {
	out := make([]paymentAttemptDTO, 0, len(attempts))
	for _, attempt := range attempts {
		out = append(out, paymentAttemptDTO{
			AttemptID:         attempt.AttemptID,
			PaymentID:         attempt.PaymentID,
			Status:            string(attempt.Status),
			ProviderReference: attempt.ProviderReference,
			ErrorCode:         attempt.ErrorCode,
			ErrorMessage:      attempt.ErrorMessage,
			CreatedAt:         attempt.CreatedAt,
		})
	}
	return out
}

func refundDTOs(refunds []domain.RefundSnapshot) []refundDTO {
	out := make([]refundDTO, 0, len(refunds))
	for _, refund := range refunds {
		out = append(out, refundDTOFromDomain(refund))
	}
	return out
}

func refundDTOFromDomain(refund domain.RefundSnapshot) refundDTO {
	return refundDTO{
		RefundID:    refund.RefundID,
		PaymentID:   refund.PaymentID,
		OrderID:     refund.OrderID,
		Status:      string(refund.Status),
		Amount:      moneyDTOFromDomain(refund.Amount),
		Reason:      refund.Reason,
		RequestedBy: refund.RequestedBy,
		ReviewedBy:  refund.ReviewedBy,
		ReviewedAt:  refund.ReviewedAt,
		CreatedAt:   refund.CreatedAt,
	}
}

func reconciliationDTOFromDomain(alert domain.ReconciliationAlertSnapshot) reconciliationDTO {
	return reconciliationDTO{
		ReconciliationID: alert.ReconciliationID,
		PaymentID:        alert.PaymentID,
		Provider:         alert.Provider,
		Status:           string(alert.Status),
		LocalAmount:      moneyDTOFromDomainPointer(alert.LocalAmount),
		ProviderAmount:   moneyDTOFromDomainPointer(alert.ProviderAmount),
		LocalStatus:      alert.LocalStatus,
		ProviderStatus:   alert.ProviderStatus,
		SettlementID:     alert.SettlementID,
		DetectedAt:       alert.DetectedAt,
		Note:             alert.Note,
	}
}

func reviewTaskDTOs(tasks []domain.ReviewTask) []reviewTaskDTO {
	out := make([]reviewTaskDTO, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, reviewTaskDTOFromDomain(task))
	}
	return out
}

func reviewTaskDTOFromDomain(task domain.ReviewTask) reviewTaskDTO {
	return reviewTaskDTO{
		TaskID:       task.TaskID,
		TaskType:     task.TaskType,
		ResourceType: task.ResourceType,
		ResourceID:   task.ResourceID,
		Status:       string(task.Status),
		Reason:       task.Reason,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
	}
}

func moneyDTOFromDomain(money domain.Money) moneyDTO {
	return moneyDTO{Amount: money.Amount, Currency: firstNonEmptyString(money.Currency, "INR")}
}

func moneyDTOFromDomainPtr(money domain.Money) *moneyDTO {
	dto := moneyDTOFromDomain(money)
	return &dto
}

func moneyDTOFromDomainPointer(money *domain.Money) *moneyDTO {
	if money == nil {
		return nil
	}
	return moneyDTOFromDomainPtr(*money)
}

func reviewStatusFromTasks(metadata map[string]any) string {
	if raw, ok := metadata["review_status"].(string); ok {
		return raw
	}
	return "none"
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
