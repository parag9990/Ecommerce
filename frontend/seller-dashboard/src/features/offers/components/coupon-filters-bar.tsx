import { RotateCcw, Search } from "lucide-react";

import type { CouponFilters, CouponStatusFilter } from "../types";
import { COUPON_STATUSES } from "../types";
import { formatStatusLabel } from "../utils/offer-formatters";

type CouponFiltersBarProps = {
  filters: CouponFilters;
  onChange: (filters: CouponFilters) => void;
};

const statuses: CouponStatusFilter[] = ["all", ...COUPON_STATUSES];
const pageSizes = [10, 20, 50];

export function CouponFiltersBar({ filters, onChange }: CouponFiltersBarProps) {
  const hasFilters = Boolean(
    filters.q ||
      filters.starts_from ||
      filters.ends_before ||
      (filters.status && filters.status !== "all"),
  );

  function updateFilters(patch: Partial<CouponFilters>) {
    onChange({
      ...filters,
      ...patch,
      page: 1,
    });
  }

  return (
    <div className="space-y-3 rounded-md border border-slate-200 bg-white p-3 shadow-sm">
      <div className="flex flex-wrap items-end gap-3">
        <label className="min-w-64 flex-1 space-y-1">
          <span className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Search
          </span>
          <div className="relative">
            <Search
              className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400"
              aria-hidden="true"
            />
            <input
              value={filters.q ?? ""}
              onChange={(event) =>
                updateFilters({ q: event.target.value.trimStart().toUpperCase() || undefined })
              }
              placeholder="Coupon code"
              className="h-9 w-full rounded-md border border-slate-300 bg-white pl-8 pr-3 text-sm uppercase text-slate-950 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
            />
          </div>
        </label>

        <label className="w-40 space-y-1">
          <span className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Starts from
          </span>
          <input
            type="date"
            value={filters.starts_from ?? ""}
            onChange={(event) => updateFilters({ starts_from: event.target.value || undefined })}
            className="h-9 w-full rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-950 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
          />
        </label>

        <label className="w-40 space-y-1">
          <span className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Ends before
          </span>
          <input
            type="date"
            value={filters.ends_before ?? ""}
            onChange={(event) => updateFilters({ ends_before: event.target.value || undefined })}
            className="h-9 w-full rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-950 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
          />
        </label>

        <label className="w-28 space-y-1">
          <span className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Page size
          </span>
          <select
            value={filters.page_size}
            onChange={(event) => updateFilters({ page_size: Number(event.target.value) })}
            className="h-9 w-full rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-950 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
          >
            {pageSizes.map((size) => (
              <option key={size} value={size}>
                {size}
              </option>
            ))}
          </select>
        </label>

        <button
          type="button"
          disabled={!hasFilters}
          onClick={() => onChange({ status: "all", page: 1, page_size: filters.page_size })}
          className="inline-flex h-9 items-center gap-2 rounded-md border border-slate-200 bg-white px-3 text-sm font-medium text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-50"
        >
          <RotateCcw className="h-4 w-4" aria-hidden="true" />
          Reset
        </button>
      </div>

      <div className="flex flex-wrap gap-2" role="group" aria-label="Coupon status">
        {statuses.map((status) => (
          <button
            key={status}
            type="button"
            onClick={() => updateFilters({ status })}
            className={
              (filters.status ?? "all") === status
                ? "h-8 rounded-md bg-blue-600 px-3 text-xs font-medium text-white transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
                : "h-8 rounded-md border border-slate-200 bg-white px-3 text-xs font-medium text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
            }
          >
            {status === "all" ? "All" : formatStatusLabel(status)}
          </button>
        ))}
      </div>
    </div>
  );
}
