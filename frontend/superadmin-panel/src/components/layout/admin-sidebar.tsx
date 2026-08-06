import { X } from "lucide-react";
import { NavLink } from "react-router-dom";

import { adminMenu } from "../../config/admin-menu";
import { cn } from "../../lib/classnames";
import { filterAdminMenu } from "../../lib/admin-rbac";
import { useAdminRoles } from "../../stores/auth-store";

function SidebarContent({ onNavigate }: { onNavigate?: () => void }) {
  const roles = useAdminRoles();
  const visibleMenu = filterAdminMenu(roles, adminMenu);

  return (
    <>
      <div className="border-b border-slate-200 px-5 py-4">
        <p className="text-sm font-semibold text-slate-950">Superadmin</p>
        <p className="text-xs text-slate-500">Platform Control</p>
      </div>

      <nav className="space-y-1 px-3 py-4" aria-label="Admin navigation">
        {visibleMenu.map((item) => {
          const Icon = item.icon;

          return (
            <NavLink
              key={item.id}
              to={item.path}
              end={item.path === "/admin"}
              onClick={onNavigate}
              className={({ isActive }) =>
                cn(
                  "flex h-10 items-center gap-3 rounded-lg px-3 text-sm font-medium transition",
                  isActive
                    ? "bg-slate-950 text-white"
                    : "text-slate-600 hover:bg-slate-100 hover:text-slate-950"
                )
              }
            >
              <Icon size={18} aria-hidden="true" />
              <span className="truncate">{item.label}</span>
            </NavLink>
          );
        })}
      </nav>
    </>
  );
}

export function AdminSidebar({
  isMobileOpen,
  onClose
}: {
  isMobileOpen: boolean;
  onClose: () => void;
}) {
  return (
    <>
      <aside className="hidden w-64 shrink-0 border-r border-slate-200 bg-white lg:block">
        <SidebarContent />
      </aside>

      {isMobileOpen ? (
        <div className="fixed inset-0 z-40 lg:hidden" role="dialog" aria-modal="true">
          <button
            type="button"
            className="absolute inset-0 bg-slate-950/40"
            aria-label="Close navigation overlay"
            onClick={onClose}
          />
          <aside className="relative flex h-full w-72 max-w-[85vw] flex-col border-r border-slate-200 bg-white shadow-xl">
            <div className="absolute right-3 top-3">
              <button
                type="button"
                onClick={onClose}
                title="Close navigation"
                className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-100"
              >
                <X size={17} aria-hidden="true" />
                <span className="sr-only">Close navigation</span>
              </button>
            </div>
            <SidebarContent onNavigate={onClose} />
          </aside>
        </div>
      ) : null}
    </>
  );
}
