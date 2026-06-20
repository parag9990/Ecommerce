import type { ReactNode } from "react";
import { createBrowserRouter, Navigate } from "react-router-dom";

import { AdminShell } from "../components/layout/admin-shell";
import { adminMenu } from "../config/admin-menu";
import { AdminAuditLogPage } from "../features/audit/pages/admin-audit-log-page";
import { AdminLoginPage } from "../features/auth/pages/admin-login-page";
import { OrderDetailPage } from "../features/orders/pages/order-detail-page";
import { OrderOperationsPage } from "../features/orders/pages/order-operations-page";
import { PaymentOperationsPage } from "../features/payments/pages/payment-operations-page";
import { RefundReviewPage } from "../features/payments/pages/refund-review-page";
import { SellerListPage } from "../features/sellers/pages/seller-list-page";
import { SellerReviewPage } from "../features/sellers/pages/seller-review-page";
import { SessionOversightPage } from "../features/sessions/pages/session-oversight-page";
import { PlatformSettingsPage } from "../features/settings/pages/platform-settings-page";
import { AdminHomePage } from "../features/shell/pages/admin-home-page";
import { AdminModulePlaceholderPage } from "../features/shell/pages/admin-module-placeholder-page";
import { NotFoundPage } from "../features/shell/pages/not-found-page";
import { UserDetailPage } from "../features/users/pages/user-detail-page";
import { UserListPage } from "../features/users/pages/user-list-page";
import { RequireAdmin, RequireAdminRoles } from "../routes/require-admin";

function protectedRoute(menuId: (typeof adminMenu)[number]["id"], element: ReactNode) {
  const menuItem = adminMenu.find((item) => item.id === menuId);

  if (!menuItem) {
    throw new Error(`Missing admin menu config for ${menuId}`);
  }

  return <RequireAdminRoles roles={menuItem.roles}>{element}</RequireAdminRoles>;
}

function protectedModule(title: string, menuId: (typeof adminMenu)[number]["id"]) {
  const menuItem = adminMenu.find((item) => item.id === menuId);

  if (!menuItem) {
    throw new Error(`Missing admin menu config for ${menuId}`);
  }

  return protectedRoute(
    menuId,
    <AdminModulePlaceholderPage title={title} description={menuItem.description} />
  );
}

export const router = createBrowserRouter([
  {
    path: "/",
    element: <Navigate to="/admin" replace />
  },
  {
    path: "/login",
    element: <AdminLoginPage />
  },
  {
    element: <RequireAdmin />,
    children: [
      {
        path: "/admin",
        element: <AdminShell />,
        children: [
          { index: true, element: <AdminHomePage /> },
          { path: "users", element: protectedRoute("users", <UserListPage />) },
          { path: "users/:userId", element: protectedRoute("users", <UserDetailPage />) },
          { path: "sellers", element: protectedRoute("sellers", <SellerListPage />) },
          { path: "sellers/:sellerId", element: protectedRoute("sellers", <SellerReviewPage />) },
          { path: "orders", element: protectedRoute("orders", <OrderOperationsPage />) },
          { path: "orders/:orderId", element: protectedRoute("orders", <OrderDetailPage />) },
          { path: "payments", element: protectedRoute("payments", <PaymentOperationsPage />) },
          { path: "payments/refunds", element: protectedRoute("payments", <RefundReviewPage />) },
          { path: "sessions", element: protectedRoute("sessions", <SessionOversightPage />) },
          { path: "search", element: protectedModule("Search", "search") },
          { path: "settings", element: protectedRoute("settings", <PlatformSettingsPage />) },
          { path: "audit-logs", element: protectedRoute("audit-logs", <AdminAuditLogPage />) }
        ]
      }
    ]
  },
  {
    path: "*",
    element: <NotFoundPage />
  }
]);
