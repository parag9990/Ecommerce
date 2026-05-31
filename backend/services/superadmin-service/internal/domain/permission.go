package domain

import "sort"

type AdminRole string

const (
	RoleAdmin      AdminRole = "admin"
	RoleSuperadmin AdminRole = "superadmin"
	RoleOperations AdminRole = "operations_admin"
	RoleFinance    AdminRole = "finance_admin"
	RoleCatalog    AdminRole = "catalog_admin"
	RoleReadonly   AdminRole = "readonly_admin"
)

func (r AdminRole) Valid() bool {
	switch r {
	case RoleAdmin, RoleSuperadmin, RoleOperations, RoleFinance, RoleCatalog, RoleReadonly:
		return true
	default:
		return false
	}
}

func KnownRoles() []AdminRole {
	return []AdminRole{RoleAdmin, RoleSuperadmin, RoleOperations, RoleFinance, RoleCatalog, RoleReadonly}
}

type Permission string

const (
	PermissionAdminUsersRead      Permission = "admin:users:read"
	PermissionAdminUsersWrite     Permission = "admin:users:write"
	PermissionUsersRead           Permission = "users:read"
	PermissionUsersStatusUpdate   Permission = "users:status:update"
	PermissionSellersRead         Permission = "sellers:read"
	PermissionSellersStatusUpdate Permission = "sellers:status:update"
	PermissionOrdersRead          Permission = "orders:read"
	PermissionOrderManualReview   Permission = "orders:manual_review:write"
	PermissionPaymentsRead        Permission = "payments:read"
	PermissionRefundReview        Permission = "payments:refund:review"
	PermissionSessionsRead        Permission = "sessions:read"
	PermissionSessionRiskRead     Permission = "sessions:risk:read"
	PermissionCatalogRead         Permission = "catalog:moderation:read"
	PermissionCatalogWrite        Permission = "catalog:moderation:write"
	PermissionSearchSynonymsRead  Permission = "search:synonyms:read"
	PermissionSearchSynonymsWrite Permission = "search:synonyms:write"
	PermissionSettingsRead        Permission = "platform:settings:read"
	PermissionSettingsWrite       Permission = "platform:settings:write"
	PermissionAuditLogsRead       Permission = "audit:logs:read"
	PermissionAuditLogsExport     Permission = "audit:logs:export"
)

func (p Permission) Valid() bool {
	_, ok := PermissionByKey(p)
	return ok
}

type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

type PermissionDefinition struct {
	Key         Permission
	Description string
	Risk        RiskLevel
}

var permissionDefinitions = []PermissionDefinition{
	{Key: PermissionAdminUsersRead, Description: "View and list admin users", Risk: RiskMedium},
	{Key: PermissionAdminUsersWrite, Description: "Create, update, disable, or re-enable admin users", Risk: RiskCritical},
	{Key: PermissionUsersRead, Description: "Search and view platform users", Risk: RiskMedium},
	{Key: PermissionUsersStatusUpdate, Description: "Block or unblock platform users", Risk: RiskHigh},
	{Key: PermissionSellersRead, Description: "Search sellers and view seller/KYC metadata", Risk: RiskMedium},
	{Key: PermissionSellersStatusUpdate, Description: "Approve, reject, or suspend sellers", Risk: RiskHigh},
	{Key: PermissionOrdersRead, Description: "Search orders and view dispute context", Risk: RiskMedium},
	{Key: PermissionOrderManualReview, Description: "Create or update order manual review decisions", Risk: RiskHigh},
	{Key: PermissionPaymentsRead, Description: "Lookup payments and reconciliation metadata", Risk: RiskHigh},
	{Key: PermissionRefundReview, Description: "Approve or reject refund reviews", Risk: RiskCritical},
	{Key: PermissionSessionsRead, Description: "View admin session analytics", Risk: RiskMedium},
	{Key: PermissionSessionRiskRead, Description: "View suspicious session and risk data", Risk: RiskHigh},
	{Key: PermissionCatalogRead, Description: "View catalog moderation queues", Risk: RiskMedium},
	{Key: PermissionCatalogWrite, Description: "Apply catalog moderation decisions", Risk: RiskHigh},
	{Key: PermissionSearchSynonymsRead, Description: "View search synonym settings", Risk: RiskLow},
	{Key: PermissionSearchSynonymsWrite, Description: "Update search synonym settings", Risk: RiskMedium},
	{Key: PermissionSettingsRead, Description: "View platform settings", Risk: RiskMedium},
	{Key: PermissionSettingsWrite, Description: "Update platform settings", Risk: RiskCritical},
	{Key: PermissionAuditLogsRead, Description: "View admin audit logs", Risk: RiskHigh},
	{Key: PermissionAuditLogsExport, Description: "Export admin audit logs", Risk: RiskCritical},
}

func PermissionDefinitions() []PermissionDefinition {
	out := make([]PermissionDefinition, len(permissionDefinitions))
	copy(out, permissionDefinitions)
	return out
}

func KnownPermissions() []Permission {
	out := make([]Permission, 0, len(permissionDefinitions))
	for _, definition := range permissionDefinitions {
		out = append(out, definition.Key)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func PermissionByKey(permission Permission) (PermissionDefinition, bool) {
	for _, definition := range permissionDefinitions {
		if definition.Key == permission {
			return definition, true
		}
	}
	return PermissionDefinition{}, false
}
