export type AdminRole =
  | "superadmin"
  | "operations_admin"
  | "finance_admin"
  | "catalog_admin"
  | "readonly_admin";

export type AdminMenuItem = {
  id: string;
  label: string;
  path: string;
  roles: AdminRole[];
};

export const ADMIN_ROLES = [
  "superadmin",
  "operations_admin",
  "finance_admin",
  "catalog_admin",
  "readonly_admin"
] as const satisfies readonly AdminRole[];

const ADMIN_ROLE_SET = new Set<string>(ADMIN_ROLES);

export function isAdminRole(role: string): role is AdminRole {
  return ADMIN_ROLE_SET.has(role);
}

export function normalizeAdminRoles(userRoles: readonly string[]): AdminRole[] {
  return userRoles.filter(isAdminRole);
}

export function hasAnyRole(
  userRoles: readonly string[],
  allowedRoles: readonly AdminRole[]
): boolean {
  const normalizedRoles = normalizeAdminRoles(userRoles);

  return allowedRoles.some((role) => normalizedRoles.includes(role));
}

export function isAdminUser(userRoles: readonly string[]): boolean {
  return normalizeAdminRoles(userRoles).length > 0;
}

export function canViewMenuItem(
  userRoles: readonly string[],
  item: Pick<AdminMenuItem, "roles">
): boolean {
  return hasAnyRole(userRoles, item.roles);
}

export function filterAdminMenu<T extends AdminMenuItem>(
  userRoles: readonly string[],
  menuItems: readonly T[]
): T[] {
  return menuItems.filter((item) => canViewMenuItem(userRoles, item));
}
