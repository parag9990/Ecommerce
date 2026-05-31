package usecase_test

import (
	"context"
	"reflect"
	"testing"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
	"ecommerce/superadmin-service/internal/usecase"
)

func TestReviewRefundApprovesRequestedRefund(t *testing.T) {
	paymentClient := &orderPaymentPaymentClient{
		refunds: map[string]domain.RefundSnapshot{
			"refund_1": {RefundID: "refund_1", PaymentID: "pay_1", Status: domain.RefundStatusRequested, Amount: domain.Money{Currency: "INR", Amount: 249900}},
		},
	}
	audit := &recordingAudit{}
	service := newOrderPaymentService(t, &orderPaymentOrderClient{}, paymentClient, &orderPaymentReviewTasks{}, audit)

	got, err := service.ReviewRefund(financeActorContext(), "refund_1", domain.RefundReviewRequest{
		Decision: string(domain.RefundDecisionApproved),
		Reason:   "duplicate charge verified by finance team",
	})
	if err != nil {
		t.Fatalf("ReviewRefund returned error: %v", err)
	}
	if got.Status != domain.RefundStatusApproved {
		t.Fatalf("status = %s, want approved", got.Status)
	}
	if paymentClient.reviewedRefundID != "refund_1" || paymentClient.reviewDecision != "approved" {
		t.Fatalf("review call = %s/%s", paymentClient.reviewedRefundID, paymentClient.reviewDecision)
	}
	if len(audit.records) != 1 || audit.records[0].Action != "refund.review.approved" {
		t.Fatalf("audit records = %+v", audit.records)
	}
}

func TestReviewRefundRejectsProcessingRefund(t *testing.T) {
	paymentClient := &orderPaymentPaymentClient{
		refunds: map[string]domain.RefundSnapshot{
			"refund_1": {RefundID: "refund_1", PaymentID: "pay_1", Status: domain.RefundStatusProcessing},
		},
	}
	service := newOrderPaymentService(t, &orderPaymentOrderClient{}, paymentClient, &orderPaymentReviewTasks{}, &recordingAudit{})

	_, err := service.ReviewRefund(financeActorContext(), "refund_1", domain.RefundReviewRequest{
		Decision: string(domain.RefundDecisionApproved),
		Reason:   "duplicate charge verified by finance team",
	})
	if appErr, ok := domain.AsAppError(err); !ok || appErr.Code != domain.CodeRefundNotReviewable {
		t.Fatalf("error = %v, want REFUND_NOT_REVIEWABLE", err)
	}
	if paymentClient.reviewedRefundID != "" {
		t.Fatalf("unexpected review call for %s", paymentClient.reviewedRefundID)
	}
}

func TestMarkOrderManualReviewCreatesReviewTask(t *testing.T) {
	orderClient := &orderPaymentOrderClient{
		orders: map[string]domain.OrderSnapshot{
			"ord_1": {OrderID: "ord_1", Status: domain.OrderStatusPendingPayment, TotalAmount: domain.Money{Currency: "INR", Amount: 249900}},
		},
	}
	paymentClient := &orderPaymentPaymentClient{
		paymentsByOrder: map[string][]domain.PaymentSnapshot{
			"ord_1": {{PaymentID: "pay_1", OrderID: "ord_1", Status: domain.PaymentStatusCaptured, Provider: "stripe"}},
		},
	}
	reviewTasks := &orderPaymentReviewTasks{}
	audit := &recordingAudit{}
	service := newOrderPaymentService(t, orderClient, paymentClient, reviewTasks, audit)

	task, err := service.MarkOrderManualReview(actorContext(), "ord_1", domain.ManualOrderReviewRequest{
		Category: string(domain.ManualReviewPaymentMismatch),
		Reason:   "payment captured but order is still pending",
	})
	if err != nil {
		t.Fatalf("MarkOrderManualReview returned error: %v", err)
	}
	if task.ResourceID != "ord_1" || task.Status != domain.ReviewTaskStatusOpen {
		t.Fatalf("task = %+v", task)
	}
	if len(reviewTasks.created) != 1 {
		t.Fatalf("created tasks = %+v", reviewTasks.created)
	}
	if reviewTasks.created[0].Metadata["category"] != string(domain.ManualReviewPaymentMismatch) {
		t.Fatalf("metadata = %+v", reviewTasks.created[0].Metadata)
	}
	if len(audit.records) != 1 || audit.records[0].Action != "order.manual_review.create" {
		t.Fatalf("audit records = %+v", audit.records)
	}
}

func TestGetOrderDisputeViewAggregatesRiskFlags(t *testing.T) {
	orderClient := &orderPaymentOrderClient{
		orders: map[string]domain.OrderSnapshot{
			"ord_1": {OrderID: "ord_1", Status: domain.OrderStatusPendingPayment, TotalAmount: domain.Money{Currency: "INR", Amount: 249900}},
		},
		historyByOrder: map[string][]domain.OrderStatusEvent{
			"ord_1": {{Status: domain.OrderStatusCreated}, {Status: domain.OrderStatusPendingPayment}},
		},
	}
	paymentClient := &orderPaymentPaymentClient{
		paymentsByOrder: map[string][]domain.PaymentSnapshot{
			"ord_1": {{PaymentID: "pay_1", OrderID: "ord_1", Status: domain.PaymentStatusCaptured, Amount: domain.Money{Currency: "INR", Amount: 249900}}},
		},
		refundsByPayment: map[string][]domain.RefundSnapshot{
			"pay_1": {{RefundID: "refund_1", PaymentID: "pay_1", Status: domain.RefundStatusRequested}},
		},
	}
	reviewTasks := &orderPaymentReviewTasks{
		byResource: map[string][]domain.ReviewTask{
			"order:ord_1": {{TaskID: "task_1", ResourceType: "order", ResourceID: "ord_1", Status: domain.ReviewTaskStatusOpen}},
		},
	}
	service := newOrderPaymentService(t, orderClient, paymentClient, reviewTasks, &recordingAudit{})

	got, err := service.GetOrderDisputeView(actorContext(), "ord_1")
	if err != nil {
		t.Fatalf("GetOrderDisputeView returned error: %v", err)
	}
	wantFlags := []string{"payment_mismatch", "refund_requested", "manual_review_open"}
	for _, flag := range wantFlags {
		if !containsString(got.RiskFlags, flag) {
			t.Fatalf("risk flags = %v, want %s", got.RiskFlags, flag)
		}
	}
	if len(got.StatusHistory) != 2 || len(got.Payments) != 1 || len(got.Refunds) != 1 || len(got.ReviewTasks) != 1 {
		t.Fatalf("view = %+v", got)
	}
}

func newOrderPaymentService(t *testing.T, orderClient usecase.OrderServiceClient, paymentClient usecase.PaymentServiceClient, reviewTasks usecase.OrderPaymentReviewTaskRepository, audit usecase.AuditRecorder) *usecase.OrderPaymentControlService {
	t.Helper()
	service, err := usecase.NewOrderPaymentControlService(&controlAuthorizer{}, orderClient, paymentClient, reviewTasks, audit, logging.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func financeActorContext() context.Context {
	actor := domain.AdminActor{
		AdminID:     "admin_fin_1",
		UserID:      "user_admin_fin_1",
		Roles:       []domain.AdminRole{domain.RoleFinance},
		SessionID:   "sess_fin_1",
		RequestID:   "req_fin_1",
		IPHash:      "hash_fin_1",
		MFAVerified: true,
	}
	return domain.ContextWithActor(context.Background(), actor)
}

type orderPaymentOrderClient struct {
	orders         map[string]domain.OrderSnapshot
	historyByOrder map[string][]domain.OrderStatusEvent
	listReq        domain.AdminOrderListRequest
}

func (c *orderPaymentOrderClient) ListOrdersForAdmin(ctx context.Context, req domain.AdminOrderListRequest, actor domain.AdminActor) (domain.AdminOrderListResponse, error) {
	c.listReq = req
	orders := make([]domain.OrderSnapshot, 0, len(c.orders))
	for _, order := range c.orders {
		orders = append(orders, order)
	}
	return domain.AdminOrderListResponse{Orders: orders}, nil
}

func (c *orderPaymentOrderClient) GetOrderForAdmin(ctx context.Context, orderID string, actor domain.AdminActor) (domain.OrderSnapshot, error) {
	order, ok := c.orders[orderID]
	if !ok {
		return domain.OrderSnapshot{}, domain.NewOrderNotFound(orderID)
	}
	return order, nil
}

func (c *orderPaymentOrderClient) GetOrderStatusHistory(ctx context.Context, orderID string, actor domain.AdminActor) ([]domain.OrderStatusEvent, error) {
	return c.historyByOrder[orderID], nil
}

type orderPaymentPaymentClient struct {
	paymentsByOrder  map[string][]domain.PaymentSnapshot
	refundsByPayment map[string][]domain.RefundSnapshot
	refunds          map[string]domain.RefundSnapshot
	listReq          domain.AdminPaymentListRequest
	reviewedRefundID string
	reviewDecision   string
}

func (c *orderPaymentPaymentClient) ListPaymentsForAdmin(ctx context.Context, req domain.AdminPaymentListRequest, actor domain.AdminActor) (domain.AdminPaymentListResponse, error) {
	c.listReq = req
	var payments []domain.PaymentSnapshot
	for _, orderPayments := range c.paymentsByOrder {
		payments = append(payments, orderPayments...)
	}
	return domain.AdminPaymentListResponse{Payments: payments}, nil
}

func (c *orderPaymentPaymentClient) ListPaymentsForOrder(ctx context.Context, orderID string, actor domain.AdminActor) ([]domain.PaymentSnapshot, error) {
	return c.paymentsByOrder[orderID], nil
}

func (c *orderPaymentPaymentClient) ListRefundsForPayment(ctx context.Context, paymentID string, actor domain.AdminActor) ([]domain.RefundSnapshot, error) {
	return c.refundsByPayment[paymentID], nil
}

func (c *orderPaymentPaymentClient) GetRefundForAdmin(ctx context.Context, refundID string, actor domain.AdminActor) (domain.RefundSnapshot, error) {
	refund, ok := c.refunds[refundID]
	if !ok {
		return domain.RefundSnapshot{}, domain.NewRefundNotFound(refundID)
	}
	return refund, nil
}

func (c *orderPaymentPaymentClient) ApplyRefundReview(ctx context.Context, refundID string, req domain.RefundReviewRequest, mutation domain.AdminMutationContext) (domain.RefundSnapshot, error) {
	refund, ok := c.refunds[refundID]
	if !ok {
		return domain.RefundSnapshot{}, domain.NewRefundNotFound(refundID)
	}
	c.reviewedRefundID = refundID
	c.reviewDecision = req.Decision
	refund.Status = domain.RefundStatus(req.Decision)
	c.refunds[refundID] = refund
	return refund, nil
}

type orderPaymentReviewTasks struct {
	created    []domain.ReviewTask
	closed     []domain.CloseReviewTaskRequest
	byResource map[string][]domain.ReviewTask
}

func (r *orderPaymentReviewTasks) Create(ctx context.Context, task domain.ReviewTask) (*domain.ReviewTask, error) {
	r.created = append(r.created, task)
	return &task, nil
}

func (r *orderPaymentReviewTasks) ListByResource(ctx context.Context, resourceType string, resourceID string) ([]domain.ReviewTask, error) {
	return r.byResource[resourceType+":"+resourceID], nil
}

func (r *orderPaymentReviewTasks) CloseForResource(ctx context.Context, req domain.CloseReviewTaskRequest) error {
	r.closed = append(r.closed, req)
	return nil
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if reflect.DeepEqual(value, want) {
			return true
		}
	}
	return false
}
