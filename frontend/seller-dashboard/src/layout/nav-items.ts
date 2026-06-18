import {
  BarChart3,
  ClipboardList,
  LayoutDashboard,
  Megaphone,
  Package,
  ShieldCheck,
  Users,
  type LucideIcon,
} from "lucide-react";

export type SellerNavItem = {
  label: string;
  to: string;
  icon: LucideIcon;
  end?: boolean;
};

export const sellerNavItems: SellerNavItem[] = [
  { label: "Overview", to: "/seller", icon: LayoutDashboard, end: true },
  { label: "Products", to: "/seller/products", icon: Package },
  { label: "Orders", to: "/seller/orders", icon: ClipboardList },
  { label: "Offers", to: "/seller/offers", icon: Megaphone },
  { label: "Analytics", to: "/seller/analytics", icon: BarChart3 },
  { label: "Team", to: "/seller/team", icon: Users },
  { label: "Audit", to: "/seller/audit", icon: ShieldCheck },
];
