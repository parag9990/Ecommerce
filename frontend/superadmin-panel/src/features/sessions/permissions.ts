import { hasAnyRole, type AdminRole } from "../../lib/admin-rbac";
import { useAdminRoles } from "../../stores/auth-store";

export const SESSION_VIEW_ROLES: readonly AdminRole[] = [
  "superadmin",
  "operations_admin",
  "readonly_admin"
] as const;

export function canViewSessions(roles: readonly string[]): boolean {
  return hasAnyRole(roles, SESSION_VIEW_ROLES);
}

export function canViewLiveMetrics(roles: readonly string[]): boolean {
  return canViewSessions(roles);
}

export function canViewSessionJourney(roles: readonly string[]): boolean {
  return canViewSessions(roles);
}

export function canViewUnmaskedSessionPii(_roles: readonly string[]): boolean {
  return false;
}

export function useSessionPermissions() {
  const roles = useAdminRoles();

  return {
    roles,
    canViewSessions: canViewSessions(roles),
    canViewLiveMetrics: canViewLiveMetrics(roles),
    canViewSessionJourney: canViewSessionJourney(roles),
    canViewUnmaskedSessionPii: canViewUnmaskedSessionPii(roles)
  };
}
