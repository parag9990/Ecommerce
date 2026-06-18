import type {
  AssignableSellerStaffRole,
  SellerPermission,
  SellerStaffRole,
} from "../types";

export const SELLER_STAFF_ROLES = [
  "seller",
  "seller_manager",
  "seller_catalog_editor",
  "seller_order_manager",
] as const satisfies readonly SellerStaffRole[];

export const ASSIGNABLE_SELLER_STAFF_ROLES = [
  "seller_manager",
  "seller_catalog_editor",
  "seller_order_manager",
] as const satisfies readonly AssignableSellerStaffRole[];

export const ROLE_LABELS: Record<SellerStaffRole, string> = {
  seller: "Owner",
  seller_manager: "Manager",
  seller_catalog_editor: "Catalog Editor",
  seller_order_manager: "Order Manager",
};

export const ROLE_DESCRIPTIONS: Record<SellerStaffRole, string> = {
  seller: "Full access across seller operations and team controls.",
  seller_manager: "Broad day-to-day access, including team and audit visibility.",
  seller_catalog_editor: "Catalog and offer access without orders or team controls.",
  seller_order_manager: "Order fulfillment and analytics access without catalog writes.",
};

export const ROLE_PERMISSIONS: Record<SellerStaffRole, SellerPermission[]> = {
  seller: [
    "dashboard:view",
    "products:view",
    "products:write",
    "orders:view",
    "orders:update_fulfillment",
    "offers:view",
    "offers:write",
    "analytics:view",
    "team:view",
    "team:invite",
    "team:update_role",
    "team:disable",
    "audit:view",
  ],
  seller_manager: [
    "dashboard:view",
    "products:view",
    "products:write",
    "orders:view",
    "orders:update_fulfillment",
    "offers:view",
    "offers:write",
    "analytics:view",
    "team:view",
    "team:invite",
    "team:update_role",
    "team:disable",
    "audit:view",
  ],
  seller_catalog_editor: [
    "dashboard:view",
    "products:view",
    "products:write",
    "offers:view",
    "offers:write",
  ],
  seller_order_manager: [
    "dashboard:view",
    "orders:view",
    "orders:update_fulfillment",
    "analytics:view",
  ],
};

export const PERMISSION_LABELS: Record<SellerPermission, string> = {
  "dashboard:view": "View dashboard",
  "products:view": "View products",
  "products:write": "Manage products",
  "orders:view": "View orders",
  "orders:update_fulfillment": "Update fulfillment",
  "offers:view": "View offers",
  "offers:write": "Manage offers",
  "analytics:view": "View analytics",
  "team:view": "View team",
  "team:invite": "Invite staff",
  "team:update_role": "Update roles",
  "team:disable": "Disable staff",
  "audit:view": "View audit",
};

export const ROLE_MATRIX_PERMISSIONS: SellerPermission[] = [
  "products:write",
  "orders:update_fulfillment",
  "offers:write",
  "analytics:view",
  "team:invite",
  "team:update_role",
  "team:disable",
  "audit:view",
];

export function isSellerStaffRole(value: unknown): value is SellerStaffRole {
  return typeof value === "string" && SELLER_STAFF_ROLES.includes(value as SellerStaffRole);
}

export function isAssignableSellerStaffRole(
  value: unknown,
): value is AssignableSellerStaffRole {
  return (
    typeof value === "string" &&
    ASSIGNABLE_SELLER_STAFF_ROLES.includes(value as AssignableSellerStaffRole)
  );
}

export function hasSellerPermission(
  roles: SellerStaffRole[],
  permission: SellerPermission,
) {
  return roles.some((role) => ROLE_PERMISSIONS[role].includes(permission));
}
