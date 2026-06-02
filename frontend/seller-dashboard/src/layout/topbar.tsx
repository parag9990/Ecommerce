import { Menu, UserCircle } from "lucide-react";

import { IconButton } from "../components/ui/icon-button";
import { SellerSwitcher } from "./seller-switcher";
import { useSellerStore } from "../stores/seller-store";

export function Topbar() {
  const toggleSidebar = useSellerStore((state) => state.toggleSidebar);
  const openMobileSidebar = useSellerStore((state) => state.openMobileSidebar);

  return (
    <header className="sticky top-0 z-20 flex h-14 items-center justify-between gap-3 border-b border-slate-200 bg-white px-3 sm:px-4 lg:px-6">
      <div className="flex min-w-0 items-center gap-3">
        <IconButton
          label="Open sidebar"
          className="md:hidden"
          onClick={openMobileSidebar}
        >
          <Menu className="h-4 w-4" aria-hidden="true" />
        </IconButton>

        <IconButton
          label="Toggle sidebar"
          className="hidden md:inline-flex"
          onClick={toggleSidebar}
        >
          <Menu className="h-4 w-4" aria-hidden="true" />
        </IconButton>

        <div className="min-w-0">
          <h1 className="truncate text-sm font-semibold text-slate-950">
            Seller Dashboard
          </h1>
          <p className="hidden truncate text-xs text-slate-500 sm:block">
            Manage seller operations
          </p>
        </div>
      </div>

      <div className="flex min-w-0 items-center gap-2 sm:gap-3">
        <SellerSwitcher />
        <button
          type="button"
          aria-label="Open account menu"
          title="Open account menu"
          className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-slate-500 transition hover:bg-slate-100 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
        >
          <UserCircle className="h-5 w-5" aria-hidden="true" />
        </button>
      </div>
    </header>
  );
}
