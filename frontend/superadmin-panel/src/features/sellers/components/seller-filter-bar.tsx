import { Search } from "lucide-react";

import type { SellerStatus } from "../types";

export function SellerFilterBar({
  query,
  status,
  onQueryChange,
  onStatusChange
}: {
  query: string;
  status: SellerStatus | "all";
  onQueryChange: (query: string) => void;
  onStatusChange: (status: SellerStatus | "all") => void;
}) {
  return (
    <div className="flex flex-col gap-3 border-b border-slate-200 bg-white px-4 py-3 md:flex-row md:items-center md:justify-between">
      <label className="flex min-h-10 w-full items-center gap-2 rounded-lg border border-slate-300 px-3 focus-within:border-slate-900 focus-within:ring-2 focus-within:ring-slate-900/10 md:max-w-md">
        <Search className="h-4 w-4 shrink-0 text-slate-500" aria-hidden="true" />
        <span className="sr-only">Search sellers</span>
        <input
          value={query}
          onChange={(event) => onQueryChange(event.target.value)}
          placeholder="Search by store, GST, email, or seller id"
          className="min-w-0 flex-1 bg-transparent text-sm outline-none"
        />
      </label>

      <label className="flex items-center gap-2 text-sm text-slate-600">
        <span className="shrink-0 font-medium">Status</span>
        <select
          value={status}
          onChange={(event) => onStatusChange(event.target.value as SellerStatus | "all")}
          className="h-10 rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10"
        >
          <option value="pending_review">Pending review</option>
          <option value="all">All statuses</option>
          <option value="draft">Draft</option>
          <option value="active">Active</option>
          <option value="suspended">Suspended</option>
          <option value="rejected">Rejected</option>
        </select>
      </label>
    </div>
  );
}
