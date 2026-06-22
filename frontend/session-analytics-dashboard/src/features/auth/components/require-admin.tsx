import { Navigate, Outlet, useLocation } from "react-router-dom";

import { readAdminSession } from "../../../lib/auth-session";

export function RequireAdmin() {
  const location = useLocation();
  if (!readAdminSession()) {
    return <Navigate replace state={{ from: location.pathname }} to="/login" />;
  }
  return <Outlet />;
}
