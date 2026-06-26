package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
	transporthttp "ecommerce/superadmin-service/internal/transport/http"
)

func TestListOrdersHandler(t *testing.T) {
	controls := &orderPaymentUsecaseStub{}
	mux := orderPaymentMux(controls)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/orders?status=paid&user_id=user_1&page=2&page_size=10", nil)
	addAdminHeaders(req, "superadmin_1", domain.RoleSuperadmin)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if controls.orderListReq.Status != domain.OrderStatusPaid || controls.orderListReq.UserID != "user_1" {
		t.Fatalf("request = %+v", controls.orderListReq)
	}
	if controls.orderListReq.Page != 2 || controls.orderListReq.PageSize != 10 {
		t.Fatalf("pagination = %+v", controls.orderListReq.Pagination)
	}
}

func TestRefundReviewHandler(t *testing.T) {
	controls := &orderPaymentUsecaseStub{}
	mux := orderPaymentMux(controls)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/refunds/refund_1/review", strings.NewReader(`{"decision":"approved","reason":"duplicate charge verified by finance"}`))
	addAdminHeaders(req, "finance_1", domain.RoleFinance)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"review_reason":"duplicate charge verified by finance"`) {
		t.Fatalf("response body missing review_reason: %s", rec.Body.String())
	}
	if controls.reviewedRefundID != "refund_1" || controls.refundReviewReq.Decision != "approved" {
		t.Fatalf("review request = %s/%+v", controls.reviewedRefundID, controls.refundReviewReq)
	}
}

func TestManualOrderReviewHandler(t *testing.T) {
	controls := &orderPaymentUsecaseStub{}
	mux := orderPaymentMux(controls)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/orders/ord_1/manual-review", strings.NewReader(`{"category":"payment_mismatch","reason":"payment captured but order is pending"}`))
	addAdminHeaders(req, "operations_1", domain.RoleOperations)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if controls.manualReviewOrderID != "ord_1" || controls.manualReviewReq.Category != "payment_mismatch" {
		t.Fatalf("manual review request = %s/%+v", controls.manualReviewOrderID, controls.manualReviewReq)
	}
}

func orderPaymentMux(controls *orderPaymentUsecaseStub) *http.ServeMux {
	rbacHandler := transporthttp.NewRBACHandler(&authorizationStub{}, nil, logging.NewNop())
	orderPaymentHandler := transporthttp.NewOrderPaymentHandler(controls, logging.NewNop())
	return transporthttp.NewServeMux(rbacHandler, orderPaymentHandler)
}

type orderPaymentUsecaseStub struct {
	orderListReq        domain.AdminOrderListRequest
	paymentListReq      domain.AdminPaymentListRequest
	reviewedRefundID    string
	refundReviewReq     domain.RefundReviewRequest
	manualReviewOrderID string
	manualReviewReq     domain.ManualOrderReviewRequest
	disputeViewOrderID  string
}

func (s *orderPaymentUsecaseStub) ListOrdersForAdmin(ctx context.Context, req domain.AdminOrderListRequest) (domain.AdminOrderListResponse, error) {
	s.orderListReq = req
	return domain.AdminOrderListResponse{
		Orders: []domain.OrderSnapshot{{OrderID: "ord_1", Status: domain.OrderStatusPaid}},
	}, nil
}

func (s *orderPaymentUsecaseStub) GetOrderForAdmin(ctx context.Context, orderID string) (domain.AdminOrderDetailResponse, error) {
	return domain.AdminOrderDetailResponse{
		Order: domain.OrderSnapshot{OrderID: orderID, Status: domain.OrderStatusPaid},
	}, nil
}

func (s *orderPaymentUsecaseStub) ListOrderDisputes(ctx context.Context, orderID string) (domain.OrderDisputeListResponse, error) {
	return domain.OrderDisputeListResponse{Disputes: []domain.OrderDispute{{DisputeID: "disp_1", OrderID: orderID, Type: "status_mismatch", Status: "open", OpenedBy: "system", Summary: "Payment captured but order is pending"}}}, nil
}

func (s *orderPaymentUsecaseStub) ListPaymentsForAdmin(ctx context.Context, req domain.AdminPaymentListRequest) (domain.AdminPaymentListResponse, error) {
	s.paymentListReq = req
	return domain.AdminPaymentListResponse{
		Payments: []domain.PaymentSnapshot{{PaymentID: "pay_1", Status: domain.PaymentStatusCaptured}},
	}, nil
}

func (s *orderPaymentUsecaseStub) GetPaymentForAdmin(ctx context.Context, paymentID string) (domain.AdminPaymentDetailResponse, error) {
	return domain.AdminPaymentDetailResponse{Payment: domain.PaymentSnapshot{PaymentID: paymentID, Status: domain.PaymentStatusCaptured}}, nil
}

func (s *orderPaymentUsecaseStub) ListRefundsForAdmin(ctx context.Context, req domain.RefundListRequest) (domain.RefundListResponse, error) {
	return domain.RefundListResponse{Refunds: []domain.RefundSnapshot{{RefundID: "refund_1", Status: domain.RefundStatusRequested}}}, nil
}

func (s *orderPaymentUsecaseStub) ReviewRefund(ctx context.Context, refundID string, req domain.RefundReviewRequest) (domain.RefundSnapshot, error) {
	s.reviewedRefundID = refundID
	s.refundReviewReq = req
	return domain.RefundSnapshot{RefundID: refundID, Status: domain.RefundStatusApproved, ReviewReason: req.Reason}, nil
}

func (s *orderPaymentUsecaseStub) MarkOrderManualReview(ctx context.Context, orderID string, req domain.ManualOrderReviewRequest) (*domain.ReviewTask, error) {
	s.manualReviewOrderID = orderID
	s.manualReviewReq = req
	return &domain.ReviewTask{TaskID: "task_1", ResourceType: "order", ResourceID: orderID, Status: domain.ReviewTaskStatusOpen}, nil
}

func (s *orderPaymentUsecaseStub) GetOrderDisputeView(ctx context.Context, orderID string) (domain.DisputeView, error) {
	s.disputeViewOrderID = orderID
	return domain.DisputeView{
		Order: domain.OrderSnapshot{OrderID: orderID, Status: domain.OrderStatusPaid},
	}, nil
}

func (s *orderPaymentUsecaseStub) ListReconciliationAlerts(ctx context.Context, req domain.ReconciliationListRequest) (domain.ReconciliationListResponse, error) {
	return domain.ReconciliationListResponse{Alerts: []domain.ReconciliationAlertSnapshot{{ReconciliationID: "rec_1", Provider: "stripe", Status: domain.ReconciliationStatusMismatch}}}, nil
}

func (s *orderPaymentUsecaseStub) GetReconciliationAlert(ctx context.Context, reconciliationID string) (domain.ReconciliationDetailResponse, error) {
	return domain.ReconciliationDetailResponse{Alert: domain.ReconciliationAlertSnapshot{ReconciliationID: reconciliationID, Provider: "stripe", Status: domain.ReconciliationStatusMismatch}}, nil
}
