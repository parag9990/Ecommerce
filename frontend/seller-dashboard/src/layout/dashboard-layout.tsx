import { Outlet } from "react-router-dom";

import { cn } from "../lib/cn";
import { useSellerStore } from "../stores/seller-store";
import { Sidebar } from "./sidebar";
import { Topbar } from "./topbar";

export function DashboardLayout() {
  const sidebarCollapsed = useSellerStore((state) => state.sidebarCollapsed);
  const mobileSidebarOpen = useSellerStore((state) => state.mobileSidebarOpen);
  const closeMobileSidebar = useSellerStore((state) => state.closeMobileSidebar);

  return (
    <div className="min-h-screen bg-slate-100 text-slate-950">
      <Sidebar />

      {mobileSidebarOpen ? (
        <button
          type="button"
          aria-label="Close sidebar overlay"
          className="fixed inset-0 z-30 bg-slate-950/30 md:hidden"
          onClick={closeMobileSidebar}
        />
      ) : null}

      <div
        className={cn(
          "min-h-screen transition-[padding] duration-200",
          sidebarCollapsed ? "md:pl-16" : "md:pl-64",
        )}
      >
        <Topbar />

        <main className="px-3 py-3 sm:px-4 lg:px-6 lg:py-4">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
