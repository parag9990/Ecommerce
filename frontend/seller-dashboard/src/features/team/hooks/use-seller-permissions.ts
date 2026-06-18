import { useMemo } from "react";

import { useSellerStore } from "../../../stores/seller-store";
import type { SellerPermission, SellerStaffRole } from "../types";
import { hasSellerPermission, isSellerStaffRole } from "../utils/seller-permissions";

type RoleAwareSeller = {
  role?: unknown;
  roles?: unknown;
  staff_role?: unknown;
};

function normalizeSellerRoles(activeSeller: RoleAwareSeller | null): SellerStaffRole[] {
  if (!activeSeller) {
    return [];
  }

  const rawRoles = Array.isArray(activeSeller.roles)
    ? activeSeller.roles
    : [activeSeller.staff_role, activeSeller.role].filter(Boolean);

  const roles = rawRoles.filter(isSellerStaffRole);

  return roles.length > 0 ? roles : ["seller"];
}

export function useSellerPermissions() {
  const activeSeller = useSellerStore((state) => state.activeSeller);

  const roles = useMemo(
    () => normalizeSellerRoles(activeSeller as RoleAwareSeller | null),
    [activeSeller],
  );

  return {
    roles,
    can(permission: SellerPermission) {
      return hasSellerPermission(roles, permission);
    },
  };
}
