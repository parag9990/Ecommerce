package rbac_test

import (
	"testing"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/rbac"
)

func TestRequireHighRiskContext(t *testing.T) {
	actor := domain.AdminActor{AdminID: "admin_1", RequestID: "req_1", MFAVerified: true}
	if err := rbac.RequireHighRiskContext(actor, "valid reason"); err != nil {
		t.Fatalf("RequireHighRiskContext returned error: %v", err)
	}
}

func TestHighRiskContextRequiresMFA(t *testing.T) {
	actor := domain.AdminActor{AdminID: "admin_1", RequestID: "req_1"}
	err := rbac.RequireHighRiskContext(actor, "valid reason")
	if appErr, ok := domain.AsAppError(err); !ok || appErr.Code != domain.CodeMFARequired {
		t.Fatalf("error = %v, want MFA_REQUIRED", err)
	}
}

func TestPermissionPolicyReasonOnly(t *testing.T) {
	actor := domain.AdminActor{AdminID: "admin_1", RequestID: "req_1"}
	err := rbac.ValidateHighRiskContext(rbac.PolicyForPermission(domain.PermissionUsersStatusUpdate), actor, "policy note")
	if err != nil {
		t.Fatalf("user status update should require reason/request but not MFA: %v", err)
	}
}

func TestPermissionPolicyRequiresReason(t *testing.T) {
	actor := domain.AdminActor{AdminID: "admin_1", RequestID: "req_1"}
	err := rbac.ValidateHighRiskContext(rbac.PolicyForPermission(domain.PermissionSellersStatusUpdate), actor, "")
	if appErr, ok := domain.AsAppError(err); !ok || appErr.Code != domain.CodeReasonRequired {
		t.Fatalf("error = %v, want REASON_REQUIRED", err)
	}
}
