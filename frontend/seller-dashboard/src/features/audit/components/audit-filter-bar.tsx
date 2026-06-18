import { RotateCcw, Search } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import type { FormEvent } from "react";

import type { AuditFilters, AuditResourceType } from "../types";
import { DEFAULT_AUDIT_PAGE_SIZE } from "../types";

type AuditFilterBarProps = {
  filters: AuditFilters;
  onApply: (filters: AuditFilters) => void;
  onReset: () => void;
};

const resourceOptions: { label: string; value: "" | AuditResourceType }[] = [
  { label: "All resources", value: "" },
  { label: "Products", value: "product" },
  { label: "Variants", value: "variant" },
  { label: "Orders", value: "order" },
  { label: "Coupons", value: "coupon" },
  { label: "Campaigns", value: "campaign" },
  { label: "Team", value: "team" },
  { label: "Settings", value: "settings" },
];

const pageSizeOptions = [20, 50, 100];
const defaultDraftFilters: AuditFilters = {
  page_size: DEFAULT_AUDIT_PAGE_SIZE,
};

function toDateInput(value?: string) {
  return value ? value.slice(0, 10) : "";
}

function fromDateInput(value: string) {
  return value ? `${value}T00:00:00.000Z` : undefined;
}

function toDateInputEnd(value: string) {
  return value ? `${value}T23:59:59.999Z` : undefined;
}

function cleanDraft(draft: AuditFilters): AuditFilters {
  return {
    page_size: draft.page_size ?? DEFAULT_AUDIT_PAGE_SIZE,
    actor_id: draft.actor_id?.trim() || undefined,
    action: draft.action?.trim() || undefined,
    resource_type: draft.resource_type || undefined,
    resource_id: draft.resource_id?.trim() || undefined,
    from: draft.from,
    to: draft.to,
  };
}

export function AuditFilterBar({ filters, onApply, onReset }: AuditFilterBarProps) {
  const [draft, setDraft] = useState<AuditFilters>(filters);

  useEffect(() => {
    setDraft(filters);
  }, [filters]);

  const fromDate = toDateInput(draft.from);
  const toDate = toDateInput(draft.to);
  const dateRangeInvalid = Boolean(fromDate && toDate && fromDate > toDate);
  const hasDraftFilters = useMemo(
    () =>
      Boolean(
        draft.actor_id ||
          draft.action ||
          draft.resource_type ||
          draft.resource_id ||
          draft.from ||
          draft.to,
      ),
    [draft],
  );

  function submitFilters(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!dateRangeInvalid) {
      onApply(cleanDraft(draft));
    }
  }

  function resetFilters() {
    setDraft(defaultDraftFilters);
    onReset();
  }

  return (
    <form
      className="rounded-md border border-slate-200 bg-white p-3 shadow-sm"
      onSubmit={submitFilters}
    >
      <div className="grid gap-3 lg:grid-cols-6">
        <label className="space-y-1 text-sm text-slate-600">
          <span className="text-xs font-medium text-slate-500">Actor</span>
          <input
            value={draft.actor_id ?? ""}
            onChange={(event) =>
              setDraft((current) => ({ ...current, actor_id: event.target.value }))
            }
            placeholder="user_123"
            className="h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-900 outline-none transition placeholder:text-slate-400 focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
          />
        </label>

        <label className="space-y-1 text-sm text-slate-600">
          <span className="text-xs font-medium text-slate-500">Action</span>
          <input
            value={draft.action ?? ""}
            onChange={(event) =>
              setDraft((current) => ({ ...current, action: event.target.value }))
            }
            placeholder="product.updated"
            className="h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-900 outline-none transition placeholder:text-slate-400 focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
          />
        </label>

        <label className="space-y-1 text-sm text-slate-600">
          <span className="text-xs font-medium text-slate-500">Resource</span>
          <select
            value={draft.resource_type ?? ""}
            onChange={(event) =>
              setDraft((current) => ({
                ...current,
                resource_type: event.target.value || undefined,
              }))
            }
            className="h-9 w-full rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-900 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
          >
            {resourceOptions.map((option) => (
              <option key={option.label} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>

        <label className="space-y-1 text-sm text-slate-600">
          <span className="text-xs font-medium text-slate-500">Resource ID</span>
          <input
            value={draft.resource_id ?? ""}
            onChange={(event) =>
              setDraft((current) => ({ ...current, resource_id: event.target.value }))
            }
            placeholder="prod_123"
            className="h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-900 outline-none transition placeholder:text-slate-400 focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
          />
        </label>

        <label className="space-y-1 text-sm text-slate-600">
          <span className="text-xs font-medium text-slate-500">From</span>
          <input
            type="date"
            value={fromDate}
            onChange={(event) =>
              setDraft((current) => ({
                ...current,
                from: fromDateInput(event.target.value),
              }))
            }
            className="h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-900 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
          />
        </label>

        <label className="space-y-1 text-sm text-slate-600">
          <span className="text-xs font-medium text-slate-500">To</span>
          <input
            type="date"
            value={toDate}
            onChange={(event) =>
              setDraft((current) => ({
                ...current,
                to: toDateInputEnd(event.target.value),
              }))
            }
            className="h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-900 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
          />
        </label>
      </div>

      <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
        <label className="flex items-center gap-2 text-sm text-slate-600">
          Rows
          <select
            value={draft.page_size ?? DEFAULT_AUDIT_PAGE_SIZE}
            onChange={(event) =>
              setDraft((current) => ({
                ...current,
                page_size: Number(event.target.value),
              }))
            }
            className="h-9 rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-900 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
          >
            {pageSizeOptions.map((pageSize) => (
              <option key={pageSize} value={pageSize}>
                {pageSize}
              </option>
            ))}
          </select>
        </label>

        <div className="flex flex-wrap items-center gap-2">
          {dateRangeInvalid ? (
            <span className="text-sm text-red-600">From date must be before To date.</span>
          ) : null}

          <button
            type="button"
            onClick={resetFilters}
            disabled={
              !hasDraftFilters &&
              (draft.page_size ?? DEFAULT_AUDIT_PAGE_SIZE) === DEFAULT_AUDIT_PAGE_SIZE
            }
            className="inline-flex h-9 items-center gap-2 rounded-md border border-slate-300 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-60"
          >
            <RotateCcw className="h-4 w-4" aria-hidden="true" />
            Reset
          </button>

          <button
            type="submit"
            disabled={dateRangeInvalid}
            className="inline-flex h-9 items-center gap-2 rounded-md bg-slate-950 px-3 text-sm font-medium text-white transition hover:bg-slate-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-950 disabled:cursor-not-allowed disabled:opacity-60"
          >
            <Search className="h-4 w-4" aria-hidden="true" />
            Apply
          </button>
        </div>
      </div>
    </form>
  );
}
