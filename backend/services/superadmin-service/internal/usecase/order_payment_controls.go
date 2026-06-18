package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
)

const (
	orderManualReviewTaskType = "order_manual_review"
	refundReviewTaskType      = "refund_review"
	orderResourceType         = "order"
	refundResourceType        = "refund"
)

type OrderServiceClient interface {
	ListOrdersForAdmin(ctx context.Context, req domain.AdminOrderListRequest, actor domain.AdminActor) (domain.AdminOrderListResponse, error)
	GetOrderForAdmin(ctx context.Context, orderID string, actor domain.AdminActor) (domain.OrderSnapshot, error)
	GetOrderStatusHistory(ctx context.Context, orderID string, actor domain.AdminActor) ([]domain.OrderStatusEvent, error)
}

type PaymentServiceClient interface {
	ListPaymentsForAdmin(ctx context.Context, req domain.AdminPaymentListRequest, actor domain.AdminActor) (domain.AdminPaymentListResponse, error)
	ListPaymentsForOrder(ctx context.Context, orderID string, actor domain.AdminActor) ([]domain.PaymentSnapshot, error)
	ListRefundsForPayment(ctx context.Context, paymentID string, actor domain.AdminActor) ([]domain.RefundSnapshot, error)
	GetRefundForAdmin(ctx context.Context, refundID string, actor domain.AdminActor) (domain.RefundSnapshot, error)
	ApplyRefundReview(ctx context.Context, refundID string, req domain.RefundReviewRequest, mutation domain.AdminMutationContext) (domain.RefundSnapshot, error)
}

type OrderPaymentReviewTaskRepository interface {
	Create(ctx context.Context, task domain.ReviewTask) (*domain.ReviewTask, error)
	ListByResource(ctx context.Context, resourceType string, resourceID string) ([]domain.ReviewTask, error)
	CloseForResource(ctx context.Context, req domain.CloseReviewTaskRequest) error
}

type OrderPaymentControlService struct {
	authorizer  ControlAuthorizer
	orderClient OrderServiceClient
	payment     PaymentServiceClient
	reviewTasks OrderPaymentReviewTaskRepository
	audit       AuditRecorder
	logger      logging.Logger
}

func NewOrderPaymentControlService(
	authorizer ControlAuthorizer,
	orderClient OrderServiceClient,
	payment PaymentServiceClient,
	reviewTasks OrderPaymentReviewTaskRepository,
	audit AuditRecorder,
	logger logging.Logger,
) (*OrderPaymentControlService, error) {
	if authorizer == nil {
		return nil, errors.New("order payment control service requires authorizer")
	}
	if orderClient == nil {
		return nil, errors.New("order payment control service requires order service client")
	}
	if payment == nil {
		return nil, errors.New("order payment control service requires payment service client")
	}
	if logger == nil {
		logger = logging.NewNop()
	}
	if audit == nil {
		audit = NewLoggingAuditRecorder(logger)
	}
	return &OrderPaymentControlService{
		authorizer:  authorizer,
		orderClient: orderClient,
		payment:     payment,
		reviewTasks: reviewTasks,
		audit:       audit,
		logger:      logger,
	}, nil
}

func (s *OrderPaymentControlService) ListOrdersForAdmin(ctx context.Context, req domain.AdminOrderListRequest) (domain.AdminOrderListResponse, error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return domain.AdminOrderListResponse{}, err
	}
	if err := s.authorizer.RequirePermission(ctx, actor, domain.PermissionOrdersRead); err != nil {
		return domain.AdminOrderListResponse{}, err
	}

	cleanReq, err := req.Normalize()
	if err != nil {
		return domain.AdminOrderListResponse{}, err
	}
	return s.orderClient.ListOrdersForAdmin(ctx, cleanReq, actor)
}

func (s *OrderPaymentControlService) ListPaymentsForAdmin(ctx context.Context, req domain.AdminPaymentListRequest) (domain.AdminPaymentListResponse, error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return domain.AdminPaymentListResponse{}, err
	}
	if err := s.authorizer.RequirePermission(ctx, actor, domain.PermissionPaymentsRead); err != nil {
		return domain.AdminPaymentListResponse{}, err
	}

	cleanReq, err := req.Normalize()
	if err != nil {
		return domain.AdminPaymentListResponse{}, err
	}
	return s.payment.ListPaymentsForAdmin(ctx, cleanReq, actor)
}

func (s *OrderPaymentControlService) ReviewRefund(ctx context.Context, refundID string, req domain.RefundReviewRequest) (domain.RefundSnapshot, error) {
	refundID = strings.TrimSpace(refundID)
	normalized, err := domain.NormalizeRefundReviewRequest(refundID, req)
	if err != nil {
		return domain.RefundSnapshot{}, err
	}

	actor, err := actorFromContext(ctx)
	if err != nil {
		return domain.RefundSnapshot{}, err
	}
	if err := s.authorizer.RequireHighRiskPermission(ctx, actor, domain.PermissionRefundReview, normalized.Reason); err != nil {
		return domain.RefundSnapshot{}, err
	}

	before, err := s.payment.GetRefundForAdmin(ctx, refundID, actor)
	if err != nil {
		return domain.RefundSnapshot{}, err
	}
	if !before.Status.Valid() {
		return domain.RefundSnapshot{}, domain.NewInternal(fmt.Sprintf("payment service returned unknown refund status %q", before.Status), nil)
	}

	if before.Status != domain.RefundStatusRequested {
		if before.Status == normalized.Decision.RefundStatus() {
			if err := s.recordAudit(ctx, domain.AuditRecord{
				ActorAdminID: actor.AdminID,
				Action:       "refund.review.idempotent",
				ResourceType: refundResourceType,
				ResourceID:   refundID,
				RequestID:    actor.RequestID,
				SessionID:    actor.SessionID,
				IPHash:       actor.IPHash,
				Reason:       normalized.Reason,
				Before:       refundAuditSnapshot(before),
				After:        refundAuditSnapshot(before),
			}); err != nil {
				return domain.RefundSnapshot{}, err
			}
			return before, nil
		}
		return domain.RefundSnapshot{}, domain.NewRefundNotReviewable(string(before.Status))
	}

	after, err := s.payment.ApplyRefundReview(ctx, refundID, domain.RefundReviewRequest{
		RefundID: refundID,
		Decision: string(normalized.Decision),
		Reason:   normalized.Reason,
	}, domain.AdminMutationContext{Actor: actor, Reason: normalized.Reason})
	if err != nil {
		return domain.RefundSnapshot{}, err
	}

	s.closeRefundReviewTask(ctx, refundID, actor, normalized)
	if err := s.recordAudit(ctx, domain.AuditRecord{
		ActorAdminID: actor.AdminID,
		Action:       "refund.review." + string(normalized.Decision),
		ResourceType: refundResourceType,
		ResourceID:   refundID,
		RequestID:    actor.RequestID,
		SessionID:    actor.SessionID,
		IPHash:       actor.IPHash,
		Reason:       normalized.Reason,
		Before:       refundAuditSnapshot(before),
		After:        refundAuditSnapshot(after),
	}); err != nil {
		return domain.RefundSnapshot{}, err
	}

	return after, nil
}

func (s *OrderPaymentControlService) MarkOrderManualReview(ctx context.Context, orderID string, req domain.ManualOrderReviewRequest) (*domain.ReviewTask, error) {
	orderID = strings.TrimSpace(orderID)
	normalized, err := domain.NormalizeManualOrderReviewRequest(orderID, req)
	if err != nil {
		return nil, err
	}
	if s.reviewTasks == nil {
		return nil, domain.NewDownstreamUnavailable("admin review task storage is not configured", nil)
	}

	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.authorizer.RequireHighRiskPermission(ctx, actor, domain.PermissionOrderManualReview, normalized.Reason); err != nil {
		return nil, err
	}

	order, err := s.orderClient.GetOrderForAdmin(ctx, orderID, actor)
	if err != nil {
		return nil, err
	}
	if !order.Status.Valid() {
		return nil, domain.NewInternal(fmt.Sprintf("order service returned unknown order status %q", order.Status), nil)
	}

	taskID, err := newReviewTaskID()
	if err != nil {
		return nil, err
	}

	metadata := map[string]any{
		"category":             string(normalized.Category),
		"current_order_status": string(order.Status),
		"order_total":          order.TotalAmount,
	}
	if paymentSummary := s.paymentSummaryForOrder(ctx, orderID, actor); len(paymentSummary) > 0 {
		metadata["payment_statuses"] = paymentSummary
	}

	created, err := s.reviewTasks.Create(ctx, domain.ReviewTask{
		TaskID:       taskID,
		TaskType:     orderManualReviewTaskType,
		ResourceType: orderResourceType,
		ResourceID:   orderID,
		Status:       domain.ReviewTaskStatusOpen,
		CreatedBy:    actor.AdminID,
		Reason:       normalized.Reason,
		Metadata:     metadata,
	})
	if err != nil {
		return nil, err
	}

	if err := s.recordAudit(ctx, domain.AuditRecord{
		ActorAdminID: actor.AdminID,
		Action:       "order.manual_review.create",
		ResourceType: orderResourceType,
		ResourceID:   orderID,
		RequestID:    actor.RequestID,
		SessionID:    actor.SessionID,
		IPHash:       actor.IPHash,
		Reason:       normalized.Reason,
		Before:       map[string]string{"status": string(order.Status)},
		After: map[string]string{
			"task_id":       created.TaskID,
			"task_type":     created.TaskType,
			"resource_type": created.ResourceType,
			"resource_id":   created.ResourceID,
			"status":        string(created.Status),
		},
	}); err != nil {
		return nil, err
	}

	return created, nil
}

func (s *OrderPaymentControlService) GetOrderDisputeView(ctx context.Context, orderID string) (domain.DisputeView, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return domain.DisputeView{}, domain.NewValidationError("order_id is required")
	}

	actor, err := actorFromContext(ctx)
	if err != nil {
		return domain.DisputeView{}, err
	}
	if err := s.authorizer.RequirePermission(ctx, actor, domain.PermissionOrdersRead); err != nil {
		return domain.DisputeView{}, err
	}

	order, err := s.orderClient.GetOrderForAdmin(ctx, orderID, actor)
	if err != nil {
		return domain.DisputeView{}, err
	}
	if !order.Status.Valid() {
		return domain.DisputeView{}, domain.NewInternal(fmt.Sprintf("order service returned unknown order status %q", order.Status), nil)
	}

	history, err := s.orderClient.GetOrderStatusHistory(ctx, orderID, actor)
	if err != nil {
		return domain.DisputeView{}, err
	}
	payments, err := s.payment.ListPaymentsForOrder(ctx, orderID, actor)
	if err != nil {
		return domain.DisputeView{}, err
	}
	refunds, err := s.refundsForPayments(ctx, payments, actor)
	if err != nil {
		return domain.DisputeView{}, err
	}
	reviewTasks := s.reviewTasksForOrder(ctx, orderID, actor)

	return domain.BuildDisputeView(order, history, payments, refunds, reviewTasks), nil
}

func (s *OrderPaymentControlService) closeRefundReviewTask(ctx context.Context, refundID string, actor domain.AdminActor, review domain.NormalizedRefundReview) {
	if s.reviewTasks == nil {
		return
	}
	if err := s.reviewTasks.CloseForResource(ctx, domain.CloseReviewTaskRequest{
		TaskType:     refundReviewTaskType,
		ResourceType: refundResourceType,
		ResourceID:   refundID,
		Status:       review.Decision.ReviewTaskStatus(),
		ReviewedBy:   actor.AdminID,
		Reason:       review.Reason,
	}); err != nil {
		s.logger.Warn(ctx, "refund review task close failed",
			"refund_id", refundID,
			"request_id", actor.RequestID,
			"error", err,
		)
	}
}

func (s *OrderPaymentControlService) paymentSummaryForOrder(ctx context.Context, orderID string, actor domain.AdminActor) []map[string]string {
	payments, err := s.payment.ListPaymentsForOrder(ctx, orderID, actor)
	if err != nil {
		s.logger.Warn(ctx, "payment summary lookup for manual review failed",
			"order_id", orderID,
			"request_id", actor.RequestID,
			"error", err,
		)
		return nil
	}
	out := make([]map[string]string, 0, len(payments))
	for _, payment := range payments {
		out = append(out, map[string]string{
			"payment_id": payment.PaymentID,
			"status":     string(payment.Status),
			"provider":   payment.Provider,
		})
	}
	return out
}

func (s *OrderPaymentControlService) refundsForPayments(ctx context.Context, payments []domain.PaymentSnapshot, actor domain.AdminActor) ([]domain.RefundSnapshot, error) {
	refunds := make([]domain.RefundSnapshot, 0)
	for _, payment := range payments {
		paymentID := strings.TrimSpace(payment.PaymentID)
		if paymentID == "" {
			continue
		}
		paymentRefunds, err := s.payment.ListRefundsForPayment(ctx, paymentID, actor)
		if err != nil {
			return nil, err
		}
		refunds = append(refunds, paymentRefunds...)
	}
	return refunds, nil
}

func (s *OrderPaymentControlService) reviewTasksForOrder(ctx context.Context, orderID string, actor domain.AdminActor) []domain.ReviewTask {
	if s.reviewTasks == nil {
		return nil
	}
	tasks, err := s.reviewTasks.ListByResource(ctx, orderResourceType, orderID)
	if err != nil {
		s.logger.Warn(ctx, "order review task lookup failed",
			"order_id", orderID,
			"request_id", actor.RequestID,
			"error", err,
		)
		return nil
	}
	return tasks
}

func (s *OrderPaymentControlService) recordAudit(ctx context.Context, record domain.AuditRecord) error {
	if err := s.audit.RecordAdminMutation(ctx, record); err != nil {
		s.logger.Warn(ctx, "admin mutation audit record failed",
			"action", record.Action,
			"resource_type", record.ResourceType,
			"resource_id", record.ResourceID,
			"request_id", record.RequestID,
			"error", err,
		)
		return auditRecordFailure(err)
	}
	return nil
}

func newReviewTaskID() (string, error) {
	var randomBytes [16]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", domain.NewInternal("generate review task id failed", err)
	}
	return "task_" + hex.EncodeToString(randomBytes[:]), nil
}

func refundAuditSnapshot(refund domain.RefundSnapshot) map[string]string {
	return map[string]string{
		"refund_id":  refund.RefundID,
		"payment_id": refund.PaymentID,
		"order_id":   refund.OrderID,
		"status":     string(refund.Status),
		"currency":   refund.Amount.Currency,
		"amount":     fmt.Sprintf("%d", refund.Amount.Amount),
	}
}
