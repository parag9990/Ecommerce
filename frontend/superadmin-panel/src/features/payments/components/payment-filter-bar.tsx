import { Search, X } from "lucide-react";

import type { PaymentFilters, PaymentProvider, PaymentStatus } from "../types";
import { PAYMENT_PROVIDERS, PAYMENT_STATUSES } from "../types";

type FilterPatch = Partial<Pick<PaymentFilters, "order_id" | "status" | "provider">>;

const inputClassName =
  "h-10 rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10";

export function PaymentFilterBar({
  filters,
  onChange,
  onReset
}: {
  filters: PaymentFilters;
  onChange: (patch: FilterPatch) => void;
  onReset: () => void;
}) {
  return (
    <div className="border-b border-slate-200 bg-white px-4 py-3">
      <div className="grid gap-3 md:grid-cols-[minmax(220px,1fr)_180px_160px_auto]">
        <label className="flex h-10 min-w-0 items-center gap-2 rounded-lg border border-slate-300 px-3 focus-within:border-slate-900 focus-within:ring-2 focus-within:ring-slate-900/10">
          <Search className="h-4 w-4 shrink-0 text-slate-500" aria-hidden="true" />
          <span className="sr-only">Payment order id</span>
          <input
            value={filters.order_id ?? ""}
            onChange={(event) => onChange({ order_id: event.target.value })}
            placeholder="Order ID"
            className="min-w-0 flex-1 bg-transparent text-sm outline-none"
          />
        </label>

        <label>
          <span className="sr-only">Payment status</span>
          <select
            value={filters.status ?? "all"}
            onChange={(event) => onChange({ status: event.target.value as PaymentStatus | "all" })}
            className={`${inputClassName} w-full`}
          >
            <option value="all">All payment states</option>
            {PAYMENT_STATUSES.map((status) => (
              <option key={status} value={status}>
                {status.replace(/_/g, " ")}
              </option>
            ))}
          </select>
        </label>

        <label>
          <span className="sr-only">Payment provider</span>
          <select
            value={filters.provider ?? "all"}
            onChange={(event) => onChange({ provider: event.target.value as PaymentProvider | "all" })}
            className={`${inputClassName} w-full`}
          >
            <option value="all">All providers</option>
            {PAYMENT_PROVIDERS.map((provider) => (
              <option key={provider} value={provider}>
                {provider.toUpperCase()}
              </option>
            ))}
          </select>
        </label>

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
