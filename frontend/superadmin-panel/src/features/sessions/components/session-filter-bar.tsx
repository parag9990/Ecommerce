import { Search, X } from "lucide-react";

import type { AdminSessionFilters, AdminSessionStatus, SessionRiskLevel } from "../types";

export type SessionFilterPatch = Partial<
  Pick<AdminSessionFilters, "user_id" | "from" | "to" | "status" | "risk_level">
>;

const inputClassName =
  "h-10 rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10";

export function SessionFilterBar({
  filters,
  onChange,
  onReset
}: {
  filters: AdminSessionFilters;
  onChange: (patch: SessionFilterPatch) => void;
  onReset: () => void;
}) {
  return (
    <div className="border-b border-slate-200 bg-white px-4 py-3">
      <div className="grid gap-3 xl:grid-cols-[minmax(220px,1fr)_150px_150px_180px_180px_auto]">
        <label className="flex h-10 min-w-0 items-center gap-2 rounded-lg border border-slate-300 px-3 focus-within:border-slate-900 focus-within:ring-2 focus-within:ring-slate-900/10">
          <Search className="h-4 w-4 shrink-0 text-slate-500" aria-hidden="true" />
          <span className="sr-only">Session user id</span>
          <input
            value={filters.user_id ?? ""}
            onChange={(event) => onChange({ user_id: event.target.value })}
            placeholder="User ID"
            className="min-w-0 flex-1 bg-transparent text-sm outline-none"
          />
        </label>

        <label>
          <span className="sr-only">Session status</span>
          <select
            value={filters.status}
            onChange={(event) =>
              onChange({ status: event.target.value as AdminSessionStatus | "all" })
            }
            className={`${inputClassName} w-full`}
          >
            <option value="all">All statuses</option>
            <option value="active">Active</option>
            <option value="revoked">Revoked</option>
            <option value="expired">Expired</option>
          </select>
        </label>

        <label>
          <span className="sr-only">Session risk level</span>
          <select
            value={filters.risk_level}
            onChange={(event) =>
              onChange({ risk_level: event.target.value as SessionRiskLevel | "all" })
            }
            className={`${inputClassName} w-full`}
          >
            <option value="all">All risk</option>
            <option value="high">High risk</option>
            <option value="medium">Medium risk</option>
            <option value="low">Low risk</option>
          </select>
        </label>

        <label>
          <span className="sr-only">Session from date</span>
          <input
            type="datetime-local"
            value={filters.from ?? ""}
            onChange={(event) => onChange({ from: event.target.value })}
            className={`${inputClassName} w-full`}
          />
        </label>

        <label>
          <span className="sr-only">Session to date</span>
          <input
            type="datetime-local"
            value={filters.to ?? ""}
            onChange={(event) => onChange({ to: event.target.value })}
            className={`${inputClassName} w-full`}
          />
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
