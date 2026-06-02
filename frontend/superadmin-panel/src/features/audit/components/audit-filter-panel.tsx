import { Search, X } from "lucide-react";

import { AUDIT_PAGE_SIZE_OPTIONS, AUDIT_RESOURCE_TYPES } from "../constants";
import type { AuditLogFilters } from "../types";

export type AuditFilterPatch = Partial<
  Pick<
    AuditLogFilters,
    | "actor_id"
    | "action"
    | "resource_type"
    | "resource_id"
    | "request_id"
    | "from"
    | "to"
    | "page_size"
  >
>;

const inputClassName =
  "h-10 rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10";

export function AuditFilterPanel({
  filters,
  onChange,
  onReset
}: {
  filters: AuditLogFilters;
  onChange: (patch: AuditFilterPatch) => void;
  onReset: () => void;
}) {
  return (
    <div className="border-b border-slate-200 bg-white px-4 py-3">
      <div className="grid gap-3 xl:grid-cols-[minmax(180px,1fr)_180px_180px_180px_180px_170px_170px_120px_auto]">
        <label className="flex h-10 min-w-0 items-center gap-2 rounded-lg border border-slate-300 px-3 focus-within:border-slate-900 focus-within:ring-2 focus-within:ring-slate-900/10">
          <Search className="h-4 w-4 shrink-0 text-slate-500" aria-hidden="true" />
          <span className="sr-only">Audit actor id</span>
          <input
            value={filters.actor_id ?? ""}
            onChange={(event) => onChange({ actor_id: event.target.value })}
            placeholder="Actor admin id"
            className="min-w-0 flex-1 bg-transparent text-sm outline-none"
          />
        </label>

        <label>
          <span className="sr-only">Audit action</span>
          <input
            value={filters.action ?? ""}
            onChange={(event) => onChange({ action: event.target.value })}
            placeholder="Action"
            className={`${inputClassName} w-full`}
          />
        </label>

        <label>
          <span className="sr-only">Audit resource type</span>
          <select
            value={filters.resource_type ?? "all"}
            onChange={(event) => onChange({ resource_type: event.target.value })}
            className={`${inputClassName} w-full`}
          >
            <option value="all">All resources</option>
            {AUDIT_RESOURCE_TYPES.map((resourceType) => (
              <option key={resourceType} value={resourceType}>
                {resourceType.replace(/_/g, " ")}
              </option>
            ))}
          </select>
        </label>

        <label>
          <span className="sr-only">Audit resource id</span>
          <input
            value={filters.resource_id ?? ""}
            onChange={(event) => onChange({ resource_id: event.target.value })}
            placeholder="Resource id"
            className={`${inputClassName} w-full`}
          />
        </label>

        <label>
          <span className="sr-only">Audit request id</span>
          <input
            value={filters.request_id ?? ""}
            onChange={(event) => onChange({ request_id: event.target.value })}
            placeholder="Request id"
            className={`${inputClassName} w-full`}
          />
        </label>

        <label>
          <span className="sr-only">Audit from date</span>
          <input
            type="datetime-local"
            value={filters.from ?? ""}
            onChange={(event) => onChange({ from: event.target.value })}
            className={`${inputClassName} w-full`}
          />
        </label>

        <label>
          <span className="sr-only">Audit to date</span>
          <input
            type="datetime-local"
            value={filters.to ?? ""}
            onChange={(event) => onChange({ to: event.target.value })}
            className={`${inputClassName} w-full`}
          />
        </label>

        <label>
          <span className="sr-only">Audit page size</span>
          <select
            value={filters.page_size}
            onChange={(event) => onChange({ page_size: Number(event.target.value) })}
            className={`${inputClassName} w-full`}
          >
            {AUDIT_PAGE_SIZE_OPTIONS.map((pageSize) => (
              <option key={pageSize} value={pageSize}>
                {pageSize} rows
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
