import { Search, X } from "lucide-react";

import type { AdminOrderFilters, OrderStatus, ReviewStatus } from "../types";
import { ORDER_STATUSES, REVIEW_STATUSES } from "../types";

type FilterPatch = Partial<Pick<AdminOrderFilters, "q" | "status" | "review_status" | "user_id" | "seller_id" | "from" | "to">>;

const inputClassName =
  "h-10 rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10";

export function OrderFilterBar({
  filters,
  onChange,
  onReset
}: {
  filters: AdminOrderFilters;
  onChange: (patch: FilterPatch) => void;
  onReset: () => void;
}) {
  return (
    <div className="border-b border-slate-200 bg-white px-4 py-3">
      <div className="grid gap-3 lg:grid-cols-[minmax(220px,1.3fr)_150px_170px_170px_150px_150px_150px_auto]">
        <label className="flex h-10 min-w-0 items-center gap-2 rounded-lg border border-slate-300 px-3 focus-within:border-slate-900 focus-within:ring-2 focus-within:ring-slate-900/10">
          <Search className="h-4 w-4 shrink-0 text-slate-500" aria-hidden="true" />
          <span className="sr-only">Search orders</span>
          <input
            value={filters.q ?? ""}
            onChange={(event) => onChange({ q: event.target.value })}
            placeholder="Order ID or support token"
            className="min-w-0 flex-1 bg-transparent text-sm outline-none"
          />
        </label>

        <label>
          <span className="sr-only">Order status</span>
          <select
            value={filters.status ?? "all"}
            onChange={(event) => onChange({ status: event.target.value as OrderStatus | "all" })}
            className={`${inputClassName} w-full`}
          >
            <option value="all">All statuses</option>
            {ORDER_STATUSES.map((status) => (
              <option key={status} value={status}>
                {status.replace(/_/g, " ")}
              </option>
            ))}
          </select>
        </label>

        <label>
          <span className="sr-only">Review status</span>
          <select
            value={filters.review_status ?? "all"}
            onChange={(event) => onChange({ review_status: event.target.value as ReviewStatus | "all" })}
            className={`${inputClassName} w-full`}
          >
            <option value="all">All review states</option>
            {REVIEW_STATUSES.map((status) => (
              <option key={status} value={status}>
                {status.replace(/_/g, " ")}
              </option>
            ))}
          </select>
        </label>

        <label>
          <span className="sr-only">Buyer user id</span>
          <input
            value={filters.user_id ?? ""}
            onChange={(event) => onChange({ user_id: event.target.value })}
            placeholder="Buyer user id"
            className={`${inputClassName} w-full`}
          />
        </label>

        <label>
          <span className="sr-only">Seller id</span>
          <input
            value={filters.seller_id ?? ""}
            onChange={(event) => onChange({ seller_id: event.target.value })}
            placeholder="Seller id"
            className={`${inputClassName} w-full`}
          />
        </label>

        <div className="grid grid-cols-2 gap-2 lg:grid-cols-1">
          <label>
            <span className="sr-only">Created from date</span>
            <input
              type="date"
              value={filters.from ?? ""}
              onChange={(event) => onChange({ from: event.target.value })}
              className={`${inputClassName} w-full`}
            />
          </label>
          <label className="lg:hidden">
            <span className="sr-only">Created to date</span>
            <input
              type="date"
              value={filters.to ?? ""}
              onChange={(event) => onChange({ to: event.target.value })}
              className={`${inputClassName} w-full`}
            />
          </label>
        </div>

        <div className="hidden lg:block">
          <label>
            <span className="sr-only">Created to date</span>
            <input
              type="date"
              value={filters.to ?? ""}
              onChange={(event) => onChange({ to: event.target.value })}
              className={`${inputClassName} w-full`}
            />
          </label>
        </div>

        <button
          type="button"
          onClick={onReset}
          className="inline-flex h-10 items-center justify-center gap-2 rounded-lg border border-slate-300 px-3 text-sm font-medium text-slate-700 hover:bg-slate-50"
        >
          <X className="h-4 w-4" aria-hidden="true" />
          Reset
        </button>
      </div>
    </div>
  );
}
