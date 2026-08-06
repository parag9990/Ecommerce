import { hasAnyRole, type AdminRole } from "../../lib/admin-rbac";
import { useAdminRoles } from "../../stores/auth-store";

export const PLATFORM_SETTINGS_VIEW_ROLES: readonly AdminRole[] = [
  "superadmin",
  "readonly_admin"
] as const;

export const PLATFORM_SETTINGS_WRITE_ROLES: readonly AdminRole[] = ["superadmin"] as const;

export const SEARCH_SYNONYM_MANAGE_ROLES: readonly AdminRole[] = [
  "superadmin",
  "catalog_admin"
] as const;

export function canViewPlatformSettings(roles: readonly string[]): boolean {
  return hasAnyRole(roles, PLATFORM_SETTINGS_VIEW_ROLES);
}

export function canWritePlatformSettings(roles: readonly string[]): boolean {
  return hasAnyRole(roles, PLATFORM_SETTINGS_WRITE_ROLES);
}

export function canManageSearchSynonyms(roles: readonly string[]): boolean {
  return hasAnyRole(roles, SEARCH_SYNONYM_MANAGE_ROLES);
}

export function useSettingsPermissions() {
  const roles = useAdminRoles();

  return {
    roles,
    canViewPlatformSettings: canViewPlatformSettings(roles),
    canWritePlatformSettings: canWritePlatformSettings(roles),
    canManageSearchSynonyms: canManageSearchSynonyms(roles)
  };
}
