import { NavLink } from "react-router-dom";

import { cn } from "../lib/cn";
import { useSellerStore } from "../stores/seller-store";
import { sellerNavItems } from "./nav-items";

export function Sidebar() {
  const collapsed = useSellerStore((state) => state.sidebarCollapsed);
  const mobileSidebarOpen = useSellerStore((state) => state.mobileSidebarOpen);
  const closeMobileSidebar = useSellerStore((state) => state.closeMobileSidebar);
  const showLabels = mobileSidebarOpen || !collapsed;

  return (
    <aside
      className={cn(
        "fixed inset-y-0 left-0 z-40 border-r border-slate-200 bg-white transition-[transform,width] duration-200",
        collapsed ? "md:w-16" : "md:w-64",
        mobileSidebarOpen ? "translate-x-0" : "-translate-x-full md:translate-x-0",
        "w-64",
      )}
    >
      <div className="flex h-14 items-center border-b border-slate-200 px-4">
        <span className="truncate text-sm font-semibold text-slate-950">
          {showLabels ? "Seller CMS" : "CMS"}
        </span>
      </div>

      <nav aria-label="Seller dashboard navigation" className="space-y-1 p-2">
        {sellerNavItems.map((item) => {
          const Icon = item.icon;

          return (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              onClick={closeMobileSidebar}
              className={({ isActive }) =>
                cn(
                  "flex h-9 items-center gap-3 rounded-md px-3 text-sm font-medium transition",
                  "text-slate-600 hover:bg-slate-100 hover:text-slate-950",
                  "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600",
                  isActive && "bg-blue-50 text-blue-700 hover:bg-blue-50 hover:text-blue-700",
                )
              }
              title={showLabels ? undefined : item.label}
            >
              <Icon className="h-4 w-4 shrink-0" aria-hidden="true" />
              {showLabels ? <span className="truncate">{item.label}</span> : null}
            </NavLink>
          );
        })}
      </nav>
    </aside>
  );
}
