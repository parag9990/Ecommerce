import {
  Activity,
  CreditCard,
  LayoutDashboard,
  PackageSearch,
  Search,
  Settings,
  ShieldCheck,
  Store,
  Users
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

import type { AdminMenuItem, AdminRole } from "../lib/admin-rbac";

export type AdminMenuConfigItem = AdminMenuItem & {
  icon: LucideIcon;
  description: string;
};

export const ALL_ADMIN_ROLES = [
  "superadmin",
  "operations_admin",
  "finance_admin",
  "catalog_admin",
  "readonly_admin"
] as const satisfies readonly AdminRole[];

export const adminMenu = [
  {
    id: "overview",
    label: "Overview",
    path: "/admin",
    icon: LayoutDashboard,
    description: "Protected shell status and platform control summary.",
    roles: [...ALL_ADMIN_ROLES]
  },
  {
    id: "users",
    label: "Users",
    path: "/admin/users",
    icon: Users,
    description: "Search users, inspect profiles, review sessions, and manage account status.",
    roles: ["superadmin", "operations_admin", "readonly_admin"]
  },
  {
    id: "sellers",
    label: "Sellers",
    path: "/admin/sellers",
    icon: Store,
    description: "Review seller KYC, lifecycle status, suspension, and catalog moderation.",
    roles: ["superadmin", "operations_admin", "catalog_admin", "readonly_admin"]
  },
  {
    id: "orders",
    label: "Orders",
    path: "/admin/orders",
    icon: PackageSearch,
    description: "Search orders, inspect disputes, and manage manual review queues.",
    roles: ["superadmin", "operations_admin", "readonly_admin"]
  },
  {
    id: "payments",
    label: "Payments",
    path: "/admin/payments",
    icon: CreditCard,
    description: "Payment lookup, refund review, and reconciliation mismatch operations.",
    roles: ["superadmin", "finance_admin", "readonly_admin"]
  },
  {
    id: "sessions",
    label: "Sessions",
    path: "/admin/sessions",
    icon: Activity,
    description: "Live sessions, suspicious activity, and journey oversight.",
    roles: ["superadmin", "operations_admin", "readonly_admin"]
  },
  {
    id: "search",
    label: "Search",
    path: "/admin/search",
    icon: Search,
    description: "Search operations and catalog discovery placeholder.",
    roles: ["superadmin", "catalog_admin", "readonly_admin"]
  },
  {
    id: "settings",
    label: "Platform Settings",
    path: "/admin/settings",
    icon: Settings,
    description: "Commission, search synonyms, feature flags, and maintenance controls.",
    roles: ["superadmin"]
  },
  {
    id: "audit-logs",
    label: "Audit Logs",
    path: "/admin/audit-logs",
    icon: ShieldCheck,
    description: "Filter, inspect, and export admin action audit logs.",
    roles: ["superadmin", "readonly_admin"]
  }
] as const satisfies readonly AdminMenuConfigItem[];

export function findAdminMenuItemByPath(pathname: string): AdminMenuConfigItem | undefined {
  const normalizedPath = pathname.replace(/\/+$/, "") || "/admin";

  return adminMenu.find((item) => item.path === normalizedPath);
}
