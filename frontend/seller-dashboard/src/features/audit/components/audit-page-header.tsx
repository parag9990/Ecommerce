import { RefreshCw, ShieldCheck } from "lucide-react";

import { cn } from "../../../lib/cn";

type AuditPageHeaderProps = {
  sellerName?: string;
  isRefreshing: boolean;
  onRefresh: () => void;
};

export function AuditPageHeader({
  sellerName,
  isRefreshing,
  onRefresh,
}: AuditPageHeaderProps) {
  return (
    <div className="flex flex-wrap items-end justify-between gap-3 border-b border-slate-200 pb-4">
      <div className="min-w-0">
        <p className="inline-flex items-center gap-2 text-xs font-semibold uppercase text-slate-500">
          <ShieldCheck className="h-3.5 w-3.5" aria-hidden="true" />
          Audit Activity
        </p>
        <h1 className="mt-1 text-xl font-semibold text-slate-950">
          Recent seller actions
        </h1>
        <p className="mt-1 max-w-2xl text-sm text-slate-500">
          Product, order, offer, and team changes for {sellerName ?? "the active seller"}.
        </p>
      </div>

      <button
        type="button"
        onClick={onRefresh}
        disabled={isRefreshing}
        className="inline-flex h-9 items-center gap-2 rounded-md border border-slate-300 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-60"
      >
        <RefreshCw
          className={cn("h-4 w-4", isRefreshing && "animate-spin")}
          aria-hidden="true"
        />
        Refresh
      </button>
    </div>
  );
}
