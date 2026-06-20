import { LogOut, Menu, Shield } from "lucide-react";
import { useNavigate } from "react-router-dom";

import { useAuthStore } from "../../stores/auth-store";

export function AdminTopbar({ onOpenSidebar }: { onOpenSidebar: () => void }) {
  const user = useAuthStore((state) => state.user);
  const clearSession = useAuthStore((state) => state.clearSession);
  const navigate = useNavigate();

  const handleLogout = () => {
    clearSession();
    navigate("/login", { replace: true });
  };

  return (
    <header className="flex h-14 shrink-0 items-center justify-between border-b border-slate-200 bg-white px-4">
      <div className="flex min-w-0 items-center gap-3">
        <button
          type="button"
          onClick={onOpenSidebar}
          title="Open navigation"
          className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-100 lg:hidden"
        >
          <Menu size={18} aria-hidden="true" />
          <span className="sr-only">Open navigation</span>
        </button>

        <div className="flex min-w-0 items-center gap-2 text-sm font-medium text-slate-700">
          <Shield size={18} aria-hidden="true" />
          <span className="truncate">Secure Admin Console</span>
        </div>
      </div>

      <div className="flex min-w-0 items-center gap-3">
        <div className="hidden min-w-0 text-right sm:block">
          <p className="truncate text-sm font-medium text-slate-950">{user?.name}</p>
          <p className="truncate text-xs text-slate-500">{user?.roles.join(", ")}</p>
        </div>

        <button
          type="button"
          onClick={handleLogout}
          title="Logout"
          className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-100"
          aria-label="Logout"
        >
          <LogOut size={17} aria-hidden="true" />
        </button>
      </div>
    </header>
  );
}
