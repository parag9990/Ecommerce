import type { RouteObject } from "react-router-dom";
import { Navigate } from "react-router-dom";

import { DashboardLayout } from "../layout/dashboard-layout";
import { DashboardHomePage } from "../pages/dashboard-home-page";
import { LoginRedirectPage } from "../pages/login-redirect-page";
import { PermissionDeniedPage } from "../pages/permission-denied-page";
import { OrderDetailPage } from "../features/orders/pages/order-detail-page";
import { OrderListPage } from "../features/orders/pages/order-list-page";
import { CampaignCreatePage } from "../features/offers/pages/campaign-create-page";
import { CouponEditorPage } from "../features/offers/pages/coupon-editor-page";
import { OffersPage } from "../features/offers/pages/offers-page";
import { RevenueAnalyticsPage } from "../features/analytics/pages/revenue-analytics-page";
import { ActivityLogPage } from "../features/audit/pages/activity-log-page";
import { ProductEditorPage } from "../features/products/pages/product-editor-page";
import { ProductListPage } from "../features/products/pages/product-list-page";
import { TeamMembersPage } from "../features/team/pages/team-members-page";
import { RequireSeller } from "./require-seller";

export const sellerRoutes: RouteObject[] = [
  {
    path: "/",
    element: <Navigate to="/seller" replace />,
  },
  {
    path: "/login",
    element: <LoginRedirectPage />,
  },
  {
    path: "/seller/permission-denied",
    element: <PermissionDeniedPage />,
  },
  {
    path: "/seller",
    element: (
      <RequireSeller>
        <DashboardLayout />
      </RequireSeller>
    ),
    children: [
      { index: true, element: <DashboardHomePage /> },
      { path: "products", element: <ProductListPage /> },
      { path: "products/new", element: <ProductEditorPage mode="create" /> },
      { path: "products/:productId/edit", element: <ProductEditorPage mode="edit" /> },
      { path: "orders", element: <OrderListPage /> },
      { path: "orders/:orderId", element: <OrderDetailPage /> },
      { path: "offers", element: <OffersPage /> },
      { path: "offers/coupons/new", element: <CouponEditorPage mode="create" /> },
      { path: "offers/coupons/:couponId/edit", element: <CouponEditorPage mode="edit" /> },
      { path: "offers/campaigns/new", element: <CampaignCreatePage /> },
      { path: "analytics", element: <RevenueAnalyticsPage /> },
      { path: "team", element: <TeamMembersPage /> },
      { path: "audit", element: <ActivityLogPage /> },
    ],
  },
  {
    path: "*",
    element: <Navigate to="/seller" replace />,
  },
];
