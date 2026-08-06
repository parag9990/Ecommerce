import type { AdminRole } from "../../lib/admin-rbac";
import { useAdminRoles } from "../../stores/auth-store";

export const ORDER_VIEW_ROLES: readonly AdminRole[] = [
  "superadmin",
  "operations_admin",
  "readonly_admin"
] as const;

export const ORDER_REVIEW_ROLES: readonly AdminRole[] = ["superadmin", "operations_admin"] as const;

function hasRole(roles: readonly string[], allowedRoles: readonly AdminRole[]): boolean {
  return roles.some((role) => allowedRoles.includes(role as AdminRole));
}

export function canViewOrders(roles: readonly string[]): boolean {
  return hasRole(roles, ORDER_VIEW_ROLES);
}

export function canViewOrderDetail(roles: readonly string[]): boolean {
  return canViewOrders(roles);
}

export function canViewOrderDisputes(roles: readonly string[]): boolean {
  return canViewOrders(roles);
}

export function canReviewOrders(roles: readonly string[]): boolean {
  return hasRole(roles, ORDER_REVIEW_ROLES);
}

export function useOrderPermissions() {
  const roles = useAdminRoles();

  return {
    roles,
    canViewOrders: canViewOrders(roles),
    canViewOrderDetail: canViewOrderDetail(roles),
    canViewOrderDisputes: canViewOrderDisputes(roles),
    canReviewOrders: canReviewOrders(roles)
  };
}
