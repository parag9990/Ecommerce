import type { ReactNode } from "react";
import { Navigate, Outlet, useLocation } from "react-router-dom";

import { PermissionDenied } from "../components/ui/permission-denied";
import { RouteLoader } from "../components/ui/route-loader";
import { hasAnyRole, isAdminUser, type AdminRole } from "../lib/admin-rbac";
import { useAuthStore } from "../stores/auth-store";

export function RequireAdmin() {
  const location = useLocation();
  const accessToken = useAuthStore((state) => state.accessToken);
  const user = useAuthStore((state) => state.user);
  const isHydrated = useAuthStore((state) => state.isHydrated);

  if (!isHydrated) {
    return <RouteLoader label="Checking admin session" />;
  }

  if (!accessToken || !user) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  if (!isAdminUser(user.roles)) {
    return <PermissionDenied />;
  }

  return <Outlet />;
}

export function RequireAdminRoles({
  roles,
  children
}: {
  roles: readonly AdminRole[];
  children: ReactNode;
}) {
  const userRoles = useAuthStore((state) => state.user?.roles ?? []);

  if (!hasAnyRole(userRoles, roles)) {
    return <PermissionDenied compact />;
  }

  return children;
}
