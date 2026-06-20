import { ShieldAlert } from "lucide-react";

import { cn } from "../../lib/classnames";

export function PermissionDenied({ compact = false }: { compact?: boolean }) {
  return (
    <main
      className={cn(
        "flex items-center justify-center bg-slate-50 px-4",
        compact ? "min-h-[320px]" : "min-h-screen"
      )}
    >
      <section className="w-full max-w-md rounded-lg border border-red-200 bg-white p-5 shadow-sm">
        <div className="flex items-center gap-2 text-sm font-semibold text-red-700">
          <ShieldAlert size={18} aria-hidden="true" />
          <span>Permission denied</span>
        </div>
        <h1 className="mt-2 text-lg font-semibold text-slate-950">Admin access required</h1>
        <p className="mt-1 text-sm text-slate-500">
          Your account does not have a valid admin role for this section.
        </p>
      </section>
    </main>
  );
}
