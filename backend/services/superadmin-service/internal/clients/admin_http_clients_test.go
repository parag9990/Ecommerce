package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
)

const testAdminToken = "test_superadmin_internal_token_32_chars"

func TestHTTPUserServiceClientUsesAuthenticatedAdminContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/admin/users" || r.URL.Query().Get("status") != "active" {
			t.Errorf("unexpected target %s", r.URL.String())
		}
		assertAdminRequest(t, r)
		_ = json.NewEncoder(w).Encode(map[string]any{"users": []map[string]any{{"user_id": "user_1", "status": "active"}}, "page": 1, "page_size": 20, "total": 1})
	}))
	defer server.Close()
	client, err := NewHTTPUserServiceClient(server.URL, testAdminToken, time.Second, logging.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.ListUsersForAdmin(context.Background(), domain.AdminUserListRequest{Status: domain.UserStatusActive, Pagination: domain.Pagination{Page: 1, PageSize: 20}}, testActor())
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Users) != 1 || response.Users[0].UserID != "user_1" {
		t.Fatalf("unexpected response %+v", response)
	}
}

func TestHTTPOrderServiceClientUsesAdminHistoryContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAdminRequest(t, r)
		if r.URL.Path != "/internal/admin/orders/order_1/status-history" {
			t.Errorf("path=%s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"history": []map[string]any{{"status": "paid"}}})
	}))
	defer server.Close()
	client, err := NewHTTPOrderServiceClient(server.URL, testAdminToken, time.Second, logging.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	history, err := client.GetOrderStatusHistory(context.Background(), "order_1", testActor())
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].Status != domain.OrderStatusPaid {
		t.Fatalf("history=%+v", history)
	}
}

func TestHTTPPaymentServiceClientReviewBodyMatchesStrictContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAdminRequest(t, r)
		if r.URL.Path != "/internal/admin/refunds/refund_1/review" {
			t.Errorf("path=%s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if _, exists := body["refund_id"]; exists {
			t.Errorf("refund_id must not be in body: %+v", body)
		}
		if len(body) != 2 {
			t.Errorf("unexpected body %+v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"refund_id": "refund_1", "status": "approved", "amount": map[string]any{"amount": 100, "currency": "INR"}})
	}))
	defer server.Close()
	client, err := NewHTTPPaymentServiceClient(server.URL, testAdminToken, time.Second, logging.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.ApplyRefundReview(context.Background(), "refund_1", domain.RefundReviewRequest{RefundID: "refund_1", Decision: "approved", Reason: "verified duplicate payment"}, domain.AdminMutationContext{Actor: testActor(), Reason: "verified duplicate payment"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != domain.RefundStatusApproved {
		t.Fatalf("result=%+v", result)
	}
}

func TestHTTPSessionServiceClientReplacesSpoofedHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAdminRequest(t, r)
		if r.Header.Get("X-Admin-Mask-PII") != "true" || r.Header.Get("X-Admin-Risk-Allowed") != "false" {
			t.Errorf("metadata=%v", r.Header)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	client, err := NewHTTPSessionServiceClient(server.URL, testAdminToken, time.Second, logging.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	incoming := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/live", nil)
	incoming.Header.Set("X-Admin-ID", "attacker")
	response, err := client.ForwardAnalytics(context.Background(), incoming, testActor(), map[string]string{"x-admin-mask-pii": "true", "x-admin-risk-allowed": "false"})
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
}

func assertAdminRequest(t *testing.T, r *http.Request) {
	t.Helper()
	if r.Header.Get("Authorization") != "Bearer "+testAdminToken {
		t.Errorf("authorization=%q", r.Header.Get("Authorization"))
	}
	if r.Header.Get("X-Admin-Id") != "admin_1" || !strings.Contains(r.Header.Get("X-Admin-Roles"), "superadmin") {
		t.Errorf("admin headers=%v", r.Header)
	}
}
func testActor() domain.AdminActor {
	return domain.AdminActor{AdminID: "admin_1", UserID: "user_1", Roles: []domain.AdminRole{domain.RoleSuperadmin}, SessionID: "sess_1", RequestID: "req_test_1", MFAVerified: true}
}
