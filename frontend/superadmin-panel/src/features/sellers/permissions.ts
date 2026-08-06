import type { AdminRole } from "../../lib/admin-rbac";
import { useAdminRoles } from "../../stores/auth-store";

export const SELLER_MANAGEMENT_ROLES: readonly AdminRole[] = [
  "superadmin",
  "operations_admin",
  "catalog_admin",
  "readonly_admin"
] as const;

export const SELLER_KYC_REVIEW_ROLES: readonly AdminRole[] = ["superadmin", "operations_admin"] as const;

export const SELLER_STATUS_MUTATION_ROLES: readonly AdminRole[] = [
  "superadmin",
  "operations_admin"
] as const;

export const SELLER_CATALOG_REVIEW_ROLES: readonly AdminRole[] = [
  "superadmin",
  "operations_admin",
  "catalog_admin"
] as const;

export const SELLER_CATALOG_VIEW_ROLES: readonly AdminRole[] = [
  ...SELLER_CATALOG_REVIEW_ROLES,
  "readonly_admin"
] as const;

function hasRole(roles: readonly string[], allowedRoles: readonly AdminRole[]): boolean {
  return roles.some((role) => allowedRoles.includes(role as AdminRole));
}

export function canViewSellers(roles: readonly string[]): boolean {
  return hasRole(roles, SELLER_MANAGEMENT_ROLES);
}

export function canReviewSellerKyc(roles: readonly string[]): boolean {
  return hasRole(roles, SELLER_KYC_REVIEW_ROLES);
}

export function canViewMaskedSellerKyc(roles: readonly string[]): boolean {
  return hasRole(roles, ["readonly_admin"]);
}

export function canSuspendSeller(roles: readonly string[]): boolean {
  return hasRole(roles, SELLER_STATUS_MUTATION_ROLES);
}

export function canReviewSellerCatalog(roles: readonly string[]): boolean {
  return hasRole(roles, SELLER_CATALOG_REVIEW_ROLES);
}

export function canViewSellerCatalog(roles: readonly string[]): boolean {
  return hasRole(roles, SELLER_CATALOG_VIEW_ROLES);
}

export function useSellerPermissions() {
  const roles = useAdminRoles();

  return {
    roles,
    canViewSellers: canViewSellers(roles),
    canReviewSellerKyc: canReviewSellerKyc(roles),
    canViewMaskedSellerKyc: canViewMaskedSellerKyc(roles),
    canSuspendSeller: canSuspendSeller(roles),
    canReviewSellerCatalog: canReviewSellerCatalog(roles),
    canViewSellerCatalog: canViewSellerCatalog(roles)
  };
}
