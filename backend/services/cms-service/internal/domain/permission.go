package domain

import "sort"

type Permission string

func (p Permission) String() string {
	return string(p)
}

const (
	PermissionSettingsRead         Permission = "cms:settings:read"
	PermissionSettingsUpdate       Permission = "cms:settings:update"
	PermissionAnalyticsRead        Permission = "cms:analytics:read"
	PermissionAuditLogsRead        Permission = "cms:audit_logs:read"
	PermissionStaffRead            Permission = "cms:staff:read"
	PermissionStaffInvite          Permission = "cms:staff:invite"
	PermissionStaffUpdateRole      Permission = "cms:staff:update_role"
	PermissionStaffRemove          Permission = "cms:staff:remove"
	PermissionProductsRead         Permission = "cms:products:read"
	PermissionProductsCreate       Permission = "cms:products:create"
	PermissionProductsUpdate       Permission = "cms:products:update"
	PermissionProductsSubmitReview Permission = "cms:products:submit_review"
	PermissionProductsUnpublish    Permission = "cms:products:unpublish"
	PermissionAdminCatalogModerate Permission = "admin:catalog:moderate"
	PermissionAdminForceUnpublish  Permission = "admin:catalog:force_unpublish"
	PermissionCouponsRead          Permission = "cms:coupons:read"
	PermissionCouponsCreate        Permission = "cms:coupons:create"
	PermissionCouponsUpdate        Permission = "cms:coupons:update"
	PermissionCouponsDisable       Permission = "cms:coupons:disable"
	PermissionCampaignsRead        Permission = "cms:campaigns:read"
	PermissionCampaignsCreate      Permission = "cms:campaigns:create"
	PermissionCampaignsUpdate      Permission = "cms:campaigns:update"
	PermissionCampaignsDisable     Permission = "cms:campaigns:disable"
	PermissionOrdersRead           Permission = "cms:orders:read"
	PermissionOrdersUpdateFulfill  Permission = "cms:orders:update_fulfillment"
	PermissionOrdersRespondReturn  Permission = "cms:orders:respond_return"
)

type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

type PermissionInfo struct {
	Permission  Permission
	Resource    string
	Action      string
	Description string
	Risk        RiskLevel
	Mutation    bool
}

type RoutePermission struct {
	Method     string
	Path       string
	Permission Permission
}

var permissionCatalog = map[Permission]PermissionInfo{
	PermissionSettingsRead:         {Permission: PermissionSettingsRead, Resource: "settings", Action: "read", Description: "Seller settings read", Risk: RiskLow},
	PermissionSettingsUpdate:       {Permission: PermissionSettingsUpdate, Resource: "settings", Action: "update", Description: "Seller settings update", Risk: RiskHigh, Mutation: true},
	PermissionAnalyticsRead:        {Permission: PermissionAnalyticsRead, Resource: "analytics", Action: "read", Description: "Seller metrics and dashboard read", Risk: RiskMedium},
	PermissionAuditLogsRead:        {Permission: PermissionAuditLogsRead, Resource: "audit_logs", Action: "read", Description: "Seller activity logs read", Risk: RiskMedium},
	PermissionStaffRead:            {Permission: PermissionStaffRead, Resource: "staff", Action: "read", Description: "Seller staff list read", Risk: RiskMedium},
	PermissionStaffInvite:          {Permission: PermissionStaffInvite, Resource: "staff", Action: "invite", Description: "Invite a seller staff member", Risk: RiskHigh, Mutation: true},
	PermissionStaffUpdateRole:      {Permission: PermissionStaffUpdateRole, Resource: "staff", Action: "update_role", Description: "Change a seller staff role", Risk: RiskHigh, Mutation: true},
	PermissionStaffRemove:          {Permission: PermissionStaffRemove, Resource: "staff", Action: "remove", Description: "Revoke seller staff access", Risk: RiskHigh, Mutation: true},
	PermissionProductsRead:         {Permission: PermissionProductsRead, Resource: "products", Action: "read", Description: "Own seller products read", Risk: RiskLow},
	PermissionProductsCreate:       {Permission: PermissionProductsCreate, Resource: "products", Action: "create", Description: "Create a product draft", Risk: RiskMedium, Mutation: true},
	PermissionProductsUpdate:       {Permission: PermissionProductsUpdate, Resource: "products", Action: "update", Description: "Update product draft or details", Risk: RiskHigh, Mutation: true},
	PermissionProductsSubmitReview: {Permission: PermissionProductsSubmitReview, Resource: "products", Action: "submit_review", Description: "Submit product for approval review", Risk: RiskMedium, Mutation: true},
	PermissionProductsUnpublish:    {Permission: PermissionProductsUnpublish, Resource: "products", Action: "unpublish", Description: "Unpublish seller product", Risk: RiskHigh, Mutation: true},
	PermissionAdminCatalogModerate: {Permission: PermissionAdminCatalogModerate, Resource: "catalog", Action: "moderate", Description: "Approve or reject submitted product reviews", Risk: RiskHigh, Mutation: true},
	PermissionAdminForceUnpublish:  {Permission: PermissionAdminForceUnpublish, Resource: "catalog", Action: "force_unpublish", Description: "Force unpublish a live product", Risk: RiskHigh, Mutation: true},
	PermissionCouponsRead:          {Permission: PermissionCouponsRead, Resource: "coupons", Action: "read", Description: "Coupons read", Risk: RiskLow},
	PermissionCouponsCreate:        {Permission: PermissionCouponsCreate, Resource: "coupons", Action: "create", Description: "Coupon draft or create", Risk: RiskHigh, Mutation: true},
	PermissionCouponsUpdate:        {Permission: PermissionCouponsUpdate, Resource: "coupons", Action: "update", Description: "Coupon rules update", Risk: RiskHigh, Mutation: true},
	PermissionCouponsDisable:       {Permission: PermissionCouponsDisable, Resource: "coupons", Action: "disable", Description: "Coupon disable", Risk: RiskMedium, Mutation: true},
	PermissionCampaignsRead:        {Permission: PermissionCampaignsRead, Resource: "campaigns", Action: "read", Description: "Campaigns read", Risk: RiskLow},
	PermissionCampaignsCreate:      {Permission: PermissionCampaignsCreate, Resource: "campaigns", Action: "create", Description: "Campaign create", Risk: RiskHigh, Mutation: true},
	PermissionCampaignsUpdate:      {Permission: PermissionCampaignsUpdate, Resource: "campaigns", Action: "update", Description: "Campaign update", Risk: RiskHigh, Mutation: true},
	PermissionCampaignsDisable:     {Permission: PermissionCampaignsDisable, Resource: "campaigns", Action: "disable", Description: "Campaign disable", Risk: RiskMedium, Mutation: true},
	PermissionOrdersRead:           {Permission: PermissionOrdersRead, Resource: "orders", Action: "read", Description: "Seller orders read", Risk: RiskMedium},
	PermissionOrdersUpdateFulfill:  {Permission: PermissionOrdersUpdateFulfill, Resource: "orders", Action: "update_fulfillment", Description: "Shipment and tracking status update", Risk: RiskHigh, Mutation: true},
	PermissionOrdersRespondReturn:  {Permission: PermissionOrdersRespondReturn, Resource: "orders", Action: "respond_return", Description: "Return or cancellation support response", Risk: RiskHigh, Mutation: true},
}

var rolePermissions = map[Role]map[Permission]struct{}{
	RoleSeller: {
		PermissionSettingsRead: {}, PermissionSettingsUpdate: {}, PermissionAnalyticsRead: {}, PermissionAuditLogsRead: {},
		PermissionStaffRead: {}, PermissionStaffInvite: {}, PermissionStaffUpdateRole: {}, PermissionStaffRemove: {},
		PermissionProductsRead: {}, PermissionProductsCreate: {}, PermissionProductsUpdate: {}, PermissionProductsSubmitReview: {}, PermissionProductsUnpublish: {},
		PermissionCouponsRead: {}, PermissionCouponsCreate: {}, PermissionCouponsUpdate: {}, PermissionCouponsDisable: {},
		PermissionCampaignsRead: {}, PermissionCampaignsCreate: {}, PermissionCampaignsUpdate: {}, PermissionCampaignsDisable: {},
		PermissionOrdersRead: {}, PermissionOrdersUpdateFulfill: {}, PermissionOrdersRespondReturn: {},
	},
	RoleSellerManager: {
		PermissionSettingsRead: {}, PermissionSettingsUpdate: {}, PermissionAnalyticsRead: {}, PermissionAuditLogsRead: {},
		PermissionStaffRead: {}, PermissionStaffInvite: {},
		PermissionProductsRead: {}, PermissionProductsCreate: {}, PermissionProductsUpdate: {}, PermissionProductsSubmitReview: {}, PermissionProductsUnpublish: {},
		PermissionCouponsRead: {}, PermissionCouponsCreate: {}, PermissionCouponsUpdate: {}, PermissionCouponsDisable: {},
		PermissionCampaignsRead: {}, PermissionCampaignsCreate: {}, PermissionCampaignsUpdate: {}, PermissionCampaignsDisable: {},
		PermissionOrdersRead: {}, PermissionOrdersUpdateFulfill: {}, PermissionOrdersRespondReturn: {},
	},
	RoleSellerCatalogEditor: {
		PermissionAnalyticsRead: {},
		PermissionProductsRead:  {}, PermissionProductsCreate: {}, PermissionProductsUpdate: {}, PermissionProductsSubmitReview: {}, PermissionProductsUnpublish: {},
		PermissionCouponsRead: {}, PermissionCouponsCreate: {}, PermissionCouponsUpdate: {}, PermissionCouponsDisable: {},
		PermissionCampaignsRead: {}, PermissionCampaignsCreate: {}, PermissionCampaignsUpdate: {}, PermissionCampaignsDisable: {},
	},
	RoleSellerOrderManager: {
		PermissionAnalyticsRead: {},
		PermissionOrdersRead:    {}, PermissionOrdersUpdateFulfill: {}, PermissionOrdersRespondReturn: {},
	},
	RoleCatalogAdmin: {
		PermissionAdminCatalogModerate: {},
	},
	RoleSuperadmin: {
		PermissionAdminCatalogModerate: {}, PermissionAdminForceUnpublish: {},
	},
}

var routePermissions = []RoutePermission{
	{Method: "GET", Path: "/api/v1/seller/dashboard/summary", Permission: PermissionAnalyticsRead},
	{Method: "GET", Path: "/api/v1/seller/audit-logs", Permission: PermissionAuditLogsRead},
	{Method: "GET", Path: "/api/v1/seller/settings", Permission: PermissionSettingsRead},
	{Method: "PATCH", Path: "/api/v1/seller/settings", Permission: PermissionSettingsUpdate},
	{Method: "GET", Path: "/api/v1/seller/staff", Permission: PermissionStaffRead},
	{Method: "POST", Path: "/api/v1/seller/staff/invites", Permission: PermissionStaffInvite},
	{Method: "PATCH", Path: "/api/v1/seller/staff/{staff_id}/role", Permission: PermissionStaffUpdateRole},
	{Method: "DELETE", Path: "/api/v1/seller/staff/{staff_id}", Permission: PermissionStaffRemove},
	{Method: "GET", Path: "/api/v1/seller/products", Permission: PermissionProductsRead},
	{Method: "POST", Path: "/api/v1/seller/products", Permission: PermissionProductsCreate},
	{Method: "PATCH", Path: "/api/v1/seller/products/{product_id}", Permission: PermissionProductsUpdate},
	{Method: "POST", Path: "/api/v1/seller/products/{product_id}/submit-review", Permission: PermissionProductsSubmitReview},
	{Method: "POST", Path: "/api/v1/seller/products/{product_id}/publish", Permission: PermissionProductsSubmitReview},
	{Method: "POST", Path: "/api/v1/seller/products/{product_id}/unpublish", Permission: PermissionProductsUnpublish},
	{Method: "GET", Path: "/api/v1/admin/catalog/reviews", Permission: PermissionAdminCatalogModerate},
	{Method: "POST", Path: "/api/v1/admin/catalog/reviews/{review_id}/decision", Permission: PermissionAdminCatalogModerate},
	{Method: "POST", Path: "/api/v1/admin/catalog/products/{product_id}/unpublish", Permission: PermissionAdminForceUnpublish},
	{Method: "GET", Path: "/api/v1/seller/coupons", Permission: PermissionCouponsRead},
	{Method: "POST", Path: "/api/v1/seller/coupons", Permission: PermissionCouponsCreate},
	{Method: "PATCH", Path: "/api/v1/seller/coupons/{coupon_id}", Permission: PermissionCouponsUpdate},
	{Method: "POST", Path: "/api/v1/seller/coupons/{coupon_id}/disable", Permission: PermissionCouponsDisable},
	{Method: "GET", Path: "/api/v1/seller/campaigns", Permission: PermissionCampaignsRead},
	{Method: "POST", Path: "/api/v1/seller/campaigns", Permission: PermissionCampaignsCreate},
	{Method: "PATCH", Path: "/api/v1/seller/campaigns/{campaign_id}", Permission: PermissionCampaignsUpdate},
	{Method: "POST", Path: "/api/v1/seller/campaigns/{campaign_id}/disable", Permission: PermissionCampaignsDisable},
	{Method: "GET", Path: "/api/v1/seller/orders", Permission: PermissionOrdersRead},
	{Method: "PATCH", Path: "/api/v1/seller/orders/{order_id}/fulfillment", Permission: PermissionOrdersUpdateFulfill},
	{Method: "POST", Path: "/api/v1/seller/orders/{order_id}/returns/respond", Permission: PermissionOrdersRespondReturn},
}

func IsKnownPermission(permission Permission) bool {
	_, ok := permissionCatalog[permission]
	return ok
}

func PermissionMetadata(permission Permission) (PermissionInfo, bool) {
	info, ok := permissionCatalog[permission]
	return info, ok
}

func IsMutationPermission(permission Permission) bool {
	info, ok := permissionCatalog[permission]
	return ok && info.Mutation
}

func RoleHasPermission(role Role, permission Permission) bool {
	permissions, ok := rolePermissions[role]
	if !ok {
		return false
	}
	_, ok = permissions[permission]
	return ok
}

func RolesHavePermission(roles []Role, permission Permission) bool {
	for _, role := range roles {
		if RoleHasPermission(role, permission) {
			return true
		}
	}
	return false
}

func RolePermissions(role Role) []Permission {
	permissions := make([]Permission, 0, len(rolePermissions[role]))
	for permission := range rolePermissions[role] {
		permissions = append(permissions, permission)
	}
	sort.Slice(permissions, func(i, j int) bool {
		return permissions[i] < permissions[j]
	})
	return permissions
}

func PermissionsByRole() map[Role][]Permission {
	out := make(map[Role][]Permission, len(rolePermissions))
	for _, role := range orderedRoles {
		out[role] = RolePermissions(role)
	}
	return out
}

func PermissionCatalog() []PermissionInfo {
	permissions := make([]PermissionInfo, 0, len(permissionCatalog))
	for _, info := range permissionCatalog {
		permissions = append(permissions, info)
	}
	sort.Slice(permissions, func(i, j int) bool {
		return permissions[i].Permission < permissions[j].Permission
	})
	return permissions
}

func RoutePermissions() []RoutePermission {
	return append([]RoutePermission(nil), routePermissions...)
}
