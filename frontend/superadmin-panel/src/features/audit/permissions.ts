import { hasAnyRole, type AdminRole } from "../../lib/admin-rbac";
import { useAuthStore } from "../../stores/auth-store";

export const AUDIT_LOG_VIEW_ROLES: readonly AdminRole[] = [
  "superadmin",
  "readonly_admin"
] as const;

export const AUDIT_LOG_EXPORT_ROLES: readonly AdminRole[] = ["superadmin"] as const;

export function canViewAuditLogs(roles: readonly string[]): boolean {
  return hasAnyRole(roles, AUDIT_LOG_VIEW_ROLES);
}

export function canExportAuditLogs(roles: readonly string[]): boolean {
  return hasAnyRole(roles, AUDIT_LOG_EXPORT_ROLES);
}

export function useAuditPermissions() {
  const roles = useAuthStore((state) => state.adminRoles());

  return {
    roles,
    canViewAuditLogs: canViewAuditLogs(roles),
    canExportAuditLogs: canExportAuditLogs(roles)
  };
}
