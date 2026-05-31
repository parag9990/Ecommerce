package usecase_test

import (
	"context"
	"testing"
	"time"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
	"ecommerce/superadmin-service/internal/rbac"
	"ecommerce/superadmin-service/internal/usecase"
)

func TestSessionVisibilityAuthorization(t *testing.T) {
	service := newSessionVisibilityService(t)

	tests := []struct {
		name            string
		adminID         string
		role            domain.AdminRole
		route           domain.SessionAnalyticsRoute
		includeRisk     bool
		wantErrCode     domain.ErrorCode
		wantRiskAllowed bool
		wantMaskPII     bool
	}{
		{
			name:            "superadmin can read risk sessions",
			adminID:         "superadmin_1",
			role:            domain.RoleSuperadmin,
			route:           domain.SessionAnalyticsRouteSessions,
			includeRisk:     true,
			wantRiskAllowed: true,
			wantMaskPII:     false,
		},
		{
			name:            "operations can read risk sessions",
			adminID:         "operations_1",
			role:            domain.RoleOperations,
			route:           domain.SessionAnalyticsRouteSessions,
			includeRisk:     true,
			wantRiskAllowed: true,
			wantMaskPII:     false,
		},
		{
			name:            "readonly can read masked normal sessions",
			adminID:         "readonly_1",
			role:            domain.RoleReadonly,
			route:           domain.SessionAnalyticsRouteSessions,
			wantRiskAllowed: false,
			wantMaskPII:     true,
		},
		{
			name:        "readonly cannot read risk sessions",
			adminID:     "readonly_1",
			role:        domain.RoleReadonly,
			route:       domain.SessionAnalyticsRouteSessions,
			includeRisk: true,
			wantErrCode: domain.CodeForbidden,
		},
		{
			name:        "finance cannot read session analytics",
			adminID:     "finance_1",
			role:        domain.RoleFinance,
			route:       domain.SessionAnalyticsRouteSessions,
			wantErrCode: domain.CodeForbidden,
		},
		{
			name:        "catalog cannot read session analytics",
			adminID:     "catalog_1",
			role:        domain.RoleCatalog,
			route:       domain.SessionAnalyticsRouteSessions,
			wantErrCode: domain.CodeForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := service.AuthorizeSessionAnalytics(context.Background(), usecase.AuthorizeSessionAnalyticsRequest{
				Actor:       adminActor(tt.adminID, tt.role),
				Route:       tt.route,
				IncludeRisk: tt.includeRisk,
			})

			if tt.wantErrCode != "" {
				assertAppErrorCode(t, err, tt.wantErrCode)
				if decision == nil || decision.Allowed {
					t.Fatalf("decision = %+v, want denied decision", decision)
				}
				return
			}
			if err != nil {
				t.Fatalf("AuthorizeSessionAnalytics returned error: %v", err)
			}
			if !decision.Allowed {
				t.Fatalf("decision = %+v, want allowed", decision)
			}
			if decision.RiskAllowed != tt.wantRiskAllowed {
				t.Fatalf("RiskAllowed = %v, want %v", decision.RiskAllowed, tt.wantRiskAllowed)
			}
			if decision.MaskPII != tt.wantMaskPII {
				t.Fatalf("MaskPII = %v, want %v", decision.MaskPII, tt.wantMaskPII)
			}
			if decision.DownstreamHeaders["x-admin-permission"] != string(domain.PermissionSessionsRead) {
				t.Fatalf("downstream headers = %+v", decision.DownstreamHeaders)
			}
		})
	}
}

func TestSessionVisibilityValidatesRouteAndFilters(t *testing.T) {
	service := newSessionVisibilityService(t)
	actor := adminActor("superadmin_1", domain.RoleSuperadmin)

	_, err := service.AuthorizeSessionAnalytics(context.Background(), usecase.AuthorizeSessionAnalyticsRequest{
		Actor: actor,
		Route: "analytics.unknown",
	})
	assertAppErrorCode(t, err, domain.CodeValidationFailed)

	_, err = service.AuthorizeSessionAnalytics(context.Background(), usecase.AuthorizeSessionAnalyticsRequest{
		Actor: actor,
		Route: domain.SessionAnalyticsRouteJourney,
	})
	assertAppErrorCode(t, err, domain.CodeValidationFailed)

	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(48 * time.Hour)
	_, err = service.AuthorizeSessionAnalytics(context.Background(), usecase.AuthorizeSessionAnalyticsRequest{
		Actor: actor,
		Route: domain.SessionAnalyticsRouteSessions,
		Filters: domain.SessionAnalyticsFilters{
			From:     from,
			To:       to,
			PageSize: 101,
		},
	})
	assertAppErrorCode(t, err, domain.CodeValidationFailed)

	_, err = service.AuthorizeSessionAnalytics(context.Background(), usecase.AuthorizeSessionAnalyticsRequest{
		Actor: actor,
		Route: domain.SessionAnalyticsRouteHeatmaps,
	})
	assertAppErrorCode(t, err, domain.CodeValidationFailed)
}

func TestSessionVisibilityAuditsRiskAccess(t *testing.T) {
	audit := &recordingAudit{}
	service := newSessionVisibilityServiceWithAudit(t, audit)

	_, err := service.AuthorizeSessionAnalytics(context.Background(), usecase.AuthorizeSessionAnalyticsRequest{
		Actor:       adminActor("superadmin_1", domain.RoleSuperadmin),
		Route:       domain.SessionAnalyticsRouteSessions,
		IncludeRisk: true,
	})
	if err != nil {
		t.Fatalf("AuthorizeSessionAnalytics returned error: %v", err)
	}
	if len(audit.records) != 1 {
		t.Fatalf("audit records = %+v", audit.records)
	}
	got := audit.records[0]
	if got.Action != "session.analytics.access" || got.ResourceType != "session_analytics" {
		t.Fatalf("audit record = %+v", got)
	}
}

func TestSessionDashboardAccess(t *testing.T) {
	service := newSessionVisibilityService(t)

	readonlyAccess, err := service.DashboardAccess(context.Background(), adminActor("readonly_1", domain.RoleReadonly))
	if err != nil {
		t.Fatalf("DashboardAccess returned error: %v", err)
	}
	if !readonlyAccess.CanViewDashboard || readonlyAccess.RiskAllowed || !readonlyAccess.MaskPII {
		t.Fatalf("readonly access = %+v", readonlyAccess)
	}
	if moduleVisible(readonlyAccess.Modules, "suspicious_activity") {
		t.Fatalf("readonly suspicious activity module should be hidden: %+v", readonlyAccess.Modules)
	}

	operationsAccess, err := service.DashboardAccess(context.Background(), adminActor("operations_1", domain.RoleOperations))
	if err != nil {
		t.Fatalf("DashboardAccess returned error: %v", err)
	}
	if !operationsAccess.CanViewDashboard || !operationsAccess.RiskAllowed || operationsAccess.MaskPII {
		t.Fatalf("operations access = %+v", operationsAccess)
	}
	if !moduleVisible(operationsAccess.Modules, "suspicious_activity") {
		t.Fatalf("operations suspicious activity module should be visible: %+v", operationsAccess.Modules)
	}
}

func newSessionVisibilityService(t *testing.T) *usecase.SessionVisibilityService {
	t.Helper()
	return newSessionVisibilityServiceWithAudit(t, nil)
}

func newSessionVisibilityServiceWithAudit(t *testing.T, audit usecase.AuditRecorder) *usecase.SessionVisibilityService {
	t.Helper()
	repo := rbac.NewStaticPermissionRepository([]rbac.StaticAdmin{
		{AdminID: "superadmin_1", Role: domain.RoleSuperadmin, Active: true},
		{AdminID: "operations_1", Role: domain.RoleOperations, Active: true},
		{AdminID: "readonly_1", Role: domain.RoleReadonly, Active: true},
		{AdminID: "finance_1", Role: domain.RoleFinance, Active: true},
		{AdminID: "catalog_1", Role: domain.RoleCatalog, Active: true},
	})
	authz, err := usecase.NewAuthorizationService(repo, logging.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	service, err := usecase.NewSessionVisibilityServiceWithAudit(authz, usecase.SessionVisibilityConfig{
		MaxDateRange: 30 * 24 * time.Hour,
		MaxPageSize:  100,
	}, audit, logging.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func assertAppErrorCode(t *testing.T, err error, want domain.ErrorCode) {
	t.Helper()
	appErr, ok := domain.AsAppError(err)
	if !ok || appErr.Code != want {
		t.Fatalf("error = %v, want %s", err, want)
	}
}

func moduleVisible(modules []usecase.SessionDashboardModule, key string) bool {
	for _, module := range modules {
		if module.Key == key {
			return module.Visible
		}
	}
	return false
}
