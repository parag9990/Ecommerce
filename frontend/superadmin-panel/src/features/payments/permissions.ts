import type { AdminRole } from "../../lib/admin-rbac";
import { useAdminRoles } from "../../stores/auth-store";

export const PAYMENT_VIEW_ROLES: readonly AdminRole[] = [
  "superadmin",
  "finance_admin",
  "readonly_admin"
] as const;

export const PAYMENT_REFUND_REVIEW_ROLES: readonly AdminRole[] = [
  "superadmin",
  "finance_admin"
] as const;

export const PAYMENT_EXPORT_ROLES: readonly AdminRole[] = ["superadmin", "finance_admin"] as const;

function hasRole(roles: readonly string[], allowedRoles: readonly AdminRole[]): boolean {
  return roles.some((role) => allowedRoles.includes(role as AdminRole));
}

export function canViewPayments(roles: readonly string[]): boolean {
  return hasRole(roles, PAYMENT_VIEW_ROLES);
}

export function canViewRefunds(roles: readonly string[]): boolean {
  return canViewPayments(roles);
}

export function canReviewRefunds(roles: readonly string[]): boolean {
  return hasRole(roles, PAYMENT_REFUND_REVIEW_ROLES);
}

export function canViewReconciliation(roles: readonly string[]): boolean {
  return canViewPayments(roles);
}

export function canExportPayments(roles: readonly string[]): boolean {
  return hasRole(roles, PAYMENT_EXPORT_ROLES);
}

export function usePaymentPermissions() {
  const roles = useAdminRoles();

  return {
    roles,
    canViewPayments: canViewPayments(roles),
    canViewRefunds: canViewRefunds(roles),
    canReviewRefunds: canReviewRefunds(roles),
    canViewReconciliation: canViewReconciliation(roles),
    canExportPayments: canExportPayments(roles)
  };
}
