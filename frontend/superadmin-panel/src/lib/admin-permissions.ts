import { useAuthStore } from "../stores/auth-store";
import type { AdminRole } from "./admin-rbac";

export const USER_MANAGEMENT_ROLES: readonly AdminRole[] = [
  "superadmin",
  "operations_admin",
  "readonly_admin"
] as const;

export const USER_STATUS_MUTATION_ROLES: readonly AdminRole[] = [
  "superadmin",
  "operations_admin"
] as const;

export function canViewUserManagement(roles: readonly string[]): boolean {
  return roles.some((role) => USER_MANAGEMENT_ROLES.includes(role as AdminRole));
}

export function canViewUserProfile(roles: readonly string[]): boolean {
  return canViewUserManagement(roles);
}

export function canViewUserSessions(roles: readonly string[]): boolean {
  return canViewUserManagement(roles);
}

export function canMutateUserStatus(roles: readonly string[]): boolean {
  return roles.some((role) => USER_STATUS_MUTATION_ROLES.includes(role as AdminRole));
}

export function useCanMutateUserStatus(): boolean {
  const roles = useAuthStore((state) => state.user?.roles ?? []);

  return canMutateUserStatus(roles);
}
