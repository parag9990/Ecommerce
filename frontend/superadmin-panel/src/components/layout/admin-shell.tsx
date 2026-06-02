import { useState } from "react";
import { Outlet } from "react-router-dom";

import { AdminSidebar } from "./admin-sidebar";
import { AdminTopbar } from "./admin-topbar";

export function AdminShell() {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);

  return (
    <div className="min-h-screen bg-slate-50 text-slate-950">
      <div className="flex min-h-screen">
        <AdminSidebar isMobileOpen={isSidebarOpen} onClose={() => setIsSidebarOpen(false)} />

        <div className="flex min-w-0 flex-1 flex-col">
          <AdminTopbar onOpenSidebar={() => setIsSidebarOpen(true)} />

          <main className="flex-1 px-4 py-4 lg:px-6">
            <Outlet />
          </main>
        </div>
      </div>
    </div>
  );
}
